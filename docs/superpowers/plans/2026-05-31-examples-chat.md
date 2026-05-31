# Chat Example Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a realtime, full-duplex chat example app at `/examples/chat` that broadcasts messages over websockets (htmx `ws` extension) while showcasing 10+ Goshtoso components.

**Architecture:** Pure, HTTP-free domain in `internal/examples/chat/` (ELIZA bot, identity cookie, RAM-only broadcast Hub with a 50-message ring buffer) → exported templ in `internal/pages/demo/examples/chat.templ` (rendered by both the page shell **and** the ws push path) → thin handlers in `internal/server/chat_handler.go` (page render + cookie seed, ws read/write pumps, rename). The hub is mutex-guarded (no owner goroutine) so `Count()`/`Backlog()` are simple reads. Messages flow up as JSON (`{message, nick, color, eliza}`), down as OOB-swap HTML fragments.

**Tech Stack:** Go 1.26, `github.com/coder/websocket`, htmx `ws` extension (vendored), templ v0.3, coder/websocket, Playwright E2E.

**Reference patterns to mirror:** `internal/examples/todo/` (domain + cookie), `internal/pages/demo/examples/todo.templ` (exported fragments + `oob bool` gate), `internal/server/todo_handler.go` (page/fragment split), `tests/e2e/todo_example_test.go` (isolated context, fragment-nav console guard, `clickUntil`).

**Before writing any htmx/ws markup or templ, invoke the `htmx` and `templ` skills** — the ws-extension OOB-swap contract and templ attribute-escaping rules are load-bearing here.

---

## Key design decisions (read before starting)

1. **ELIZA flag rides each message frame.** The bot on/off Toggle is a form field (`name="eliza"`) inside the composer `<form ws-send>`. htmx serializes it with every send, so the server reads a per-message flag — **no per-client hub state for the bot.** Bot logic lives in the ws read loop, not the hub.
2. **Nick/color ride each message frame** via hidden inputs in the composer. Rename rewrites the cookie *and* OOB-swaps those hidden inputs, so subsequent messages carry the new nick live — **no HTTP↔ws coupling.**
3. **Presence (join/leave + online count)** is rendered in the handler (templ → bytes) and pushed via `hub.Broadcast`. The hub itself only plumbs opaque `[]byte` (no templ import — keeps it pure and unit-testable).
4. **Cookie is seeded on page load** (`renderChatPage`, like todo's sample seed), so by the time the socket connects the `gt_chat` cookie already exists and the upgrade request carries the identity.
5. **`htmx-ext-ws.js` is loaded globally** in the layout head (right after `htmx.min.js`). The extension is inert on pages without `hx-ext="ws"`, and a global script avoids the fragment-nav script-injection problem entirely (the script is always present, so a sidebar fragment swap into the chat page just works).
6. **Hub is mutex-guarded**, not a select-loop owner goroutine. Simpler `Count()`/`Backlog()`, and `Broadcast` drops to a slow client via `select { case send <- m: default: }` so one stuck tab never blocks the room.

---

## File Structure

```
internal/examples/chat/
  eliza.go            # Reply(msg) (reply string, matched bool) — keyword bot
  eliza_test.go
  identity.go         # Identity{Nick,Color}; cookie encode/decode; NewGuest(seed)
  identity_test.go
  hub.go              # Hub (mutex, ring buffer), Client; Register/Unregister/Broadcast/Backlog/Count
  hub_test.go
internal/pages/demo/examples/
  chat.templ          # ChatApp shell + MessageBubble + OnlineBadge + identity bar + composer
internal/server/
  chat_handler.go     # renderChatPage, handleChatWS, handleChatRename, registerChatRoutes
assets/js/vendor/
  htmx-ext-ws.js      # vendored (auto-embedded via //go:embed js)
tests/e2e/
  chat_test.go
```

Modified files: `go.mod`/`go.sum` (dep), `internal/server/server.go` (route wiring + handleExample case), `internal/pages/demo/layout.templ` (head script + sidebar item), `internal/pages/demo/examples/index.templ` (gallery card).

---

## Task 1: Add the websocket dependency

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add the dependency**

Run:
```bash
go get github.com/coder/websocket@latest
```
Expected: `go.mod` gains `github.com/coder/websocket vX.Y.Z` as a direct require; `go.sum` updated. Its only transitive deps are stdlib.

- [ ] **Step 2: Verify it builds**

Run: `go build ./...`
Expected: success, no errors.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore(deps): add github.com/coder/websocket for chat example"
```

---

## Task 2: ELIZA bot (pure, TDD)

**Files:**
- Create: `internal/examples/chat/eliza.go`
- Test: `internal/examples/chat/eliza_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/examples/chat/eliza_test.go`:
```go
package chat

import "testing"

func TestReply_Greeting(t *testing.T) {
	got, ok := Reply("hello there")
	if !ok {
		t.Fatalf("expected a match for a greeting")
	}
	if got == "" {
		t.Fatalf("matched reply must not be empty")
	}
}

func TestReply_IFeel(t *testing.T) {
	got, ok := Reply("i feel anxious")
	if !ok {
		t.Fatalf("expected a match for 'i feel'")
	}
	want := "Why do you feel anxious?"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReply_Question(t *testing.T) {
	got, ok := Reply("what is the time?")
	if !ok || got == "" {
		t.Fatalf("expected a canned reply for a question, got %q ok=%v", got, ok)
	}
}

func TestReply_NoMatch(t *testing.T) {
	_, ok := Reply("xyzzy plugh")
	if ok {
		t.Fatalf("expected no match for nonsense input")
	}
}

func TestReply_Deterministic(t *testing.T) {
	a, _ := Reply("i feel sad")
	b, _ := Reply("i feel sad")
	if a != b {
		t.Fatalf("Reply must be deterministic: %q != %q", a, b)
	}
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./internal/examples/chat/ -run TestReply -v`
Expected: FAIL — `undefined: Reply`.

- [ ] **Step 3: Implement the bot**

Create `internal/examples/chat/eliza.go`:
```go
// Package chat holds the pure, HTTP-free domain for the /examples/chat app:
// a deterministic ELIZA bot, the identity cookie, and a RAM-only broadcast hub.
package chat

import (
	"regexp"
	"strings"
)

// rule is one ELIZA pattern: a compiled matcher and a reply template. If the
// template contains "%s", the first capture group is substituted in.
type rule struct {
	re    *regexp.Regexp
	reply string
}

// rules are evaluated in order; the first match wins. Deterministic by design —
// no randomness, so the same input always yields the same reply (E2E relies on it).
var rules = []rule{
	{regexp.MustCompile(`(?i)\bi feel ([a-z ]+)`), "Why do you feel %s?"},
	{regexp.MustCompile(`(?i)\bi am ([a-z ]+)`), "How long have you been %s?"},
	{regexp.MustCompile(`(?i)\bi (?:need|want) ([a-z ]+)`), "Why do you need %s?"},
	{regexp.MustCompile(`(?i)\b(hello|hi|hey)\b`), "Hello. How are you feeling today?"},
	{regexp.MustCompile(`(?i)\b(bye|goodbye)\b`), "Goodbye. Take care of yourself."},
	{regexp.MustCompile(`(?i)\b(because)\b`), "Is that the real reason?"},
	{regexp.MustCompile(`(?i)\b(sorry)\b`), "Please don't apologise."},
	{regexp.MustCompile(`(?i)\byes\b`), "You seem quite sure."},
	{regexp.MustCompile(`(?i)\bno\b`), "Why not?"},
	{regexp.MustCompile(`(?i)\?\s*$`), "Why do you ask?"},
}

// Reply returns a deterministic ELIZA-style response for msg. matched is false
// when no rule fires, so the caller can choose to stay silent.
func Reply(msg string) (reply string, matched bool) {
	trimmed := strings.TrimSpace(msg)
	for _, r := range rules {
		m := r.re.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		if strings.Contains(r.reply, "%s") && len(m) > 1 {
			arg := strings.TrimSpace(m[1])
			// Trim a trailing question mark the capture may have swallowed.
			arg = strings.TrimRight(arg, "?. ")
			return strings.Replace(r.reply, "%s", arg, 1), true
		}
		return r.reply, true
	}
	return "", false
}
```

- [ ] **Step 4: Run the tests to confirm they pass**

Run: `go test ./internal/examples/chat/ -run TestReply -v`
Expected: PASS (all 5).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/chat/eliza.go internal/examples/chat/eliza_test.go
git commit -m "feat(chat): deterministic ELIZA bot"
```

---

## Task 3: Identity cookie (pure, TDD)

**Files:**
- Create: `internal/examples/chat/identity.go`
- Test: `internal/examples/chat/identity_test.go`

Mirrors `internal/examples/todo/cookie.go` (base64url-JSON cookie).

- [ ] **Step 1: Write the failing test**

Create `internal/examples/chat/identity_test.go`:
```go
package chat

import "testing"

func TestEncodeDecode_Roundtrip(t *testing.T) {
	in := Identity{Nick: "Ada", Color: "#3b82f6"}
	enc := in.Encode()
	if enc == "" {
		t.Fatalf("Encode produced empty string")
	}
	out, err := Decode(enc)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", out, in)
	}
}

func TestDecode_Garbage(t *testing.T) {
	if _, err := Decode("!!!not-base64!!!"); err == nil {
		t.Fatalf("expected error decoding garbage")
	}
}

func TestNewGuest_Deterministic(t *testing.T) {
	a := NewGuest(42)
	b := NewGuest(42)
	if a != b {
		t.Fatalf("NewGuest must be deterministic for a given seed: %+v vs %+v", a, b)
	}
	if a.Nick == "" || a.Color == "" {
		t.Fatalf("NewGuest must populate Nick and Color: %+v", a)
	}
	if len(a.Nick) < 5 || a.Nick[:5] != "Guest" {
		t.Fatalf("guest nick should start with 'Guest': %q", a.Nick)
	}
}

func TestNewGuest_VariesBySeed(t *testing.T) {
	if NewGuest(1).Nick == NewGuest(2).Nick {
		t.Fatalf("different seeds should usually yield different nicks")
	}
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./internal/examples/chat/ -run 'TestEncodeDecode|TestDecode|TestNewGuest' -v`
Expected: FAIL — `undefined: Identity`.

- [ ] **Step 3: Implement identity**

Create `internal/examples/chat/identity.go`:
```go
package chat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CookieName is the cookie that persists a visitor's chat identity. base64url-JSON,
// mirroring the todo example's cookie layering.
const CookieName = "gt_chat"

// palette is a fixed set of avatar/background colors. Indexed deterministically
// by the guest seed so a given seed always maps to the same color.
var palette = []string{
	"#ef4444", "#f97316", "#eab308", "#22c55e",
	"#14b8a6", "#3b82f6", "#8b5cf6", "#ec4899",
}

// Identity is a visitor's display name and avatar color. Stored in the gt_chat
// cookie; also sent in each message frame so renames take effect live.
type Identity struct {
	Nick  string `json:"n"`
	Color string `json:"c"`
}

// Encode serializes Identity to a base64url(JSON) cookie value. Identity is
// always serializable; a marshal error is a programmer error and panics.
func (i Identity) Encode() string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(fmt.Sprintf("chat: marshal identity: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a base64url(JSON) cookie value back into an Identity.
func Decode(s string) (Identity, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Identity{}, err
	}
	var id Identity
	if err := json.Unmarshal(raw, &id); err != nil {
		return Identity{}, err
	}
	return id, nil
}

// NewGuest builds a deterministic guest identity from a numeric seed (e.g. a
// connection counter). Same seed → same nick + color, so behavior is reproducible.
func NewGuest(seed int64) Identity {
	if seed < 0 {
		seed = -seed
	}
	return Identity{
		Nick:  fmt.Sprintf("Guest-%04x", uint16(seed*2654435761)),
		Color: palette[seed%int64(len(palette))],
	}
}
```

- [ ] **Step 4: Run the tests to confirm they pass**

Run: `go test ./internal/examples/chat/ -run 'TestEncodeDecode|TestDecode|TestNewGuest' -v`
Expected: PASS (all 4).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/chat/identity.go internal/examples/chat/identity_test.go
git commit -m "feat(chat): identity cookie (gt_chat) + deterministic guest"
```

---

## Task 4: Broadcast Hub + ring buffer (pure, TDD)

**Files:**
- Create: `internal/examples/chat/hub.go`
- Test: `internal/examples/chat/hub_test.go`

The hub plumbs opaque `[]byte` (no templ import). A `Client` is a buffered send channel the handler's write pump drains.

- [ ] **Step 1: Write the failing test**

Create `internal/examples/chat/hub_test.go`:
```go
package chat

import "testing"

func drain(c *Client) []string {
	var out []string
	for {
		select {
		case m, ok := <-c.Send:
			if !ok {
				return out
			}
			out = append(out, string(m))
		default:
			return out
		}
	}
}

func TestHub_BroadcastReachesAllClients(t *testing.T) {
	h := NewHub(50)
	a := NewClient()
	b := NewClient()
	h.Register(a)
	h.Register(b)

	h.Broadcast([]byte("hi"))

	if got := drain(a); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("client a got %v", got)
	}
	if got := drain(b); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("client b got %v", got)
	}
}

func TestHub_Count(t *testing.T) {
	h := NewHub(50)
	if h.Count() != 0 {
		t.Fatalf("empty hub count = %d", h.Count())
	}
	a := NewClient()
	h.Register(a)
	if h.Count() != 1 {
		t.Fatalf("after one register count = %d", h.Count())
	}
	h.Unregister(a)
	if h.Count() != 0 {
		t.Fatalf("after unregister count = %d", h.Count())
	}
}

func TestHub_BacklogReplaysRecent(t *testing.T) {
	h := NewHub(2) // ring caps at 2
	h.Broadcast([]byte("m1"))
	h.Broadcast([]byte("m2"))
	h.Broadcast([]byte("m3")) // evicts m1

	got := h.Backlog()
	if len(got) != 2 || string(got[0]) != "m2" || string(got[1]) != "m3" {
		t.Fatalf("backlog = %q", got)
	}
}

func TestHub_UnregisterClosesSend(t *testing.T) {
	h := NewHub(50)
	a := NewClient()
	h.Register(a)
	h.Unregister(a)
	if _, ok := <-a.Send; ok {
		t.Fatalf("Send channel should be closed after Unregister")
	}
}

func TestHub_SlowClientDoesNotBlock(t *testing.T) {
	h := NewHub(50)
	slow := NewClient()
	h.Register(slow)
	// Fill the slow client's buffer beyond capacity; Broadcast must not block.
	for i := 0; i < sendBuffer+5; i++ {
		h.Broadcast([]byte("flood"))
	}
	// If we reach here without deadlock, the drop-on-full path works.
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./internal/examples/chat/ -run TestHub -v`
Expected: FAIL — `undefined: NewHub`.

- [ ] **Step 3: Implement the hub**

Create `internal/examples/chat/hub.go`:
```go
package chat

import "sync"

// sendBuffer is the per-client outbound queue depth. A client whose queue is
// full is treated as too slow and the message is dropped for that client only.
const sendBuffer = 16

// Client is one connected socket from the hub's point of view: a buffered
// channel of pre-rendered HTML frames that the handler's write pump drains.
type Client struct {
	Send chan []byte
}

// NewClient builds a Client with a buffered send channel.
func NewClient() *Client {
	return &Client{Send: make(chan []byte, sendBuffer)}
}

// Hub is a RAM-only broadcast fan-out plus a bounded ring buffer of recent
// frames (so a newly connected client can replay history). Mutex-guarded: no
// owner goroutine, so Count and Backlog are plain reads.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	ring    [][]byte
	ringMax int
}

// NewHub builds an empty Hub whose ring keeps at most ringMax recent frames.
func NewHub(ringMax int) *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		ringMax: ringMax,
	}
}

// Register adds a client to the broadcast set.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// Unregister removes a client and closes its send channel so the write pump
// exits. Safe to call once per client.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.Send)
	}
	h.mu.Unlock()
}

// Broadcast appends the frame to the ring and fans it out to every client. A
// client whose buffer is full is skipped (dropped), never blocked on.
func (h *Hub) Broadcast(frame []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ring = append(h.ring, frame)
	if len(h.ring) > h.ringMax {
		h.ring = h.ring[len(h.ring)-h.ringMax:]
	}
	for c := range h.clients {
		select {
		case c.Send <- frame:
		default: // slow client — drop this frame for them
		}
	}
}

// Backlog returns a snapshot copy of the recent-frame ring, oldest first.
func (h *Hub) Backlog() [][]byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([][]byte, len(h.ring))
	copy(out, h.ring)
	return out
}

// Count returns the number of currently registered clients.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
```

- [ ] **Step 4: Run the tests (with the race detector) to confirm they pass**

Run: `go test ./internal/examples/chat/ -race -run TestHub -v`
Expected: PASS (all 5), no race warnings.

- [ ] **Step 5: Commit**

```bash
git add internal/examples/chat/hub.go internal/examples/chat/hub_test.go
git commit -m "feat(chat): RAM-only broadcast hub with ring buffer"
```

---

## Task 5: chat.templ — shell, bubble, presence fragments

**Files:**
- Create: `internal/pages/demo/examples/chat.templ`
- Generate: `internal/pages/demo/examples/chat_templ.go`

**Invoke the `templ` skill first.** No JSON in Alpine attributes; render all user text as templ text nodes (auto-escaped). Mirror the `oob bool` gate from `CountBadge` in `todo.templ`.

DOM contract (the handler and E2E depend on these exact IDs):
- `#chat-fragment` — the e2e anchor wrapper
- `#chat-log` — message list; new bubbles append here via `hx-swap-oob="beforeend:#chat-log"`
- `#chat-online` — online-count badge target
- `#toast-container` — presence toasts (must exist before any OOB toast)
- composer `<form ws-send>` with `<textarea name="message">`, hidden `<input name="nick">` / `<input name="color">`, and a toggle `<input name="eliza">`

- [ ] **Step 1: Write the templ file**

Create `internal/pages/demo/examples/chat.templ`:
```go
package examples

import (
	"fmt"

	"github.com/araihu/goshtoso/components/avatar"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/components/toggle"
	"github.com/araihu/goshtoso/internal/examples/chat"
)

// Message is a single rendered chat line. It is the unit both the ws push path
// (MessageFrame) and any future first-paint history share.
type Message struct {
	Nick  string
	Color string
	Text  string
	Time  string // "15:04"
	IsBot bool
	IsYou bool
}

// OnlineBadge is the live "N online" indicator. Pass oob=true when pushing it as
// an out-of-band update over the socket; oob=false on first paint inside ChatApp.
// (A first-paint hx-swap-oob would make htmx try to OOB-swap it with no existing
// target when ChatApp arrives via a fragment nav — oobErrorNoTarget.)
templ OnlineBadge(count int, oob bool) {
	<span id="chat-online" if oob { hx-swap-oob="true" }>
		@badge.Badge(badge.Config{
			Text:    fmt.Sprintf("%d online", count),
			Variant: badge.Success,
			Style:   badge.StyleSoft,
		})
	</span>
}

// MessageBubble renders one chat line: avatar + nick (+ BOT/you badges) + text,
// with a hover-title timestamp. All user text is a templ text node (escaped).
templ MessageBubble(m Message) {
	<div class="flex items-start gap-3 px-4 py-2" title={ m.Time }>
		@avatar.Avatar(avatar.Config{
			Name:        m.Nick,
			Size:        avatar.SizeSM,
			BorderColor: m.Color,
			Border:      true,
		})
		<div class="min-w-0">
			<div class="flex items-center gap-2">
				<span class="text-sm font-semibold text-on-surface dark:text-on-surface-dark">{ m.Nick }</span>
				if m.IsBot {
					@badge.Badge(badge.Config{Text: "BOT", Variant: badge.Info, Size: badge.SizeSM})
				}
				if m.IsYou {
					@badge.Badge(badge.Config{Text: "you", Variant: badge.Secondary, Size: badge.SizeSM, Style: badge.StyleSoft})
				}
				<span class="text-xs text-on-surface-muted dark:text-on-surface-dark-muted">{ m.Time }</span>
			</div>
			<p class="break-words text-sm text-on-surface dark:text-on-surface-dark">{ m.Text }</p>
		</div>
	</div>
}

// MessageFrame is the exact bytes pushed over the socket for one message: the
// bubble wrapped in an OOB "append to #chat-log" directive so every client's log
// grows identically.
templ MessageFrame(m Message) {
	<div hx-swap-oob="beforeend:#chat-log">
		@MessageBubble(m)
	</div>
}

// PresenceFrame pushes a presence toast (join/left) plus an OOB online-count
// update in a single socket frame.
templ PresenceFrame(text string, count int) {
	@toast.OOBToast(toast.Config{Variant: toast.Info, Message: text})
	@OnlineBadge(count, true)
}

// ChatApp is the first-paint shell for /examples/chat. id=me carries the visitor
// identity from the cookie; the composer's hidden nick/color inputs seed every
// outgoing frame and are OOB-swapped on rename.
templ ChatApp(me chat.Identity) {
	<div id="chat-fragment" class="mx-auto flex h-[70vh] max-w-3xl flex-col">
		@toast.Container(toast.ContainerConfig{})
		<!-- header: identity bar + online badge + rename -->
		<div class="flex items-center justify-between gap-3 border-b border-outline px-4 py-3 dark:border-outline-dark">
			<div class="flex items-center gap-3">
				@avatar.Avatar(avatar.Config{Name: me.Nick, Size: avatar.SizeMD, BorderColor: me.Color, Border: true})
				<span id="chat-me" class="text-sm font-semibold text-on-surface dark:text-on-surface-dark">{ me.Nick }</span>
			</div>
			@OnlineBadge(1, false)
		</div>
		<!-- rename form (plain hx-post: sets cookie + OOB-swaps hidden inputs) -->
		<form
			class="flex items-end gap-2 border-b border-outline px-4 py-2 dark:border-outline-dark"
			hx-post="/api/examples/chat/rename"
			hx-swap="none"
		>
			@textinput.TextInput(textinput.Config{
				ID:          "chat-nick-input",
				Name:        "nick",
				Placeholder: "Change your name",
				MaxLength:   24,
				Value:       me.Nick,
			})
			@button.Button(button.Config{Type: "submit", Variant: button.Secondary, Size: button.SizeSmall})
		</form>
		<!-- message log -->
		<div id="chat-log" class="flex-1 divide-y divide-outline/40 overflow-y-auto dark:divide-outline-dark/40"></div>
		<!-- composer: ws-send form -->
		<div hx-ext="ws" ws-connect="/api/examples/chat/ws" class="border-t border-outline px-4 py-3 dark:border-outline-dark">
			<form ws-send class="flex items-end gap-2">
				@chatHidden("nick", me.Nick)
				@chatHidden("color", me.Color)
				@textarea.Textarea(textarea.Config{
					ID:          "chat-message",
					Name:        "message",
					Placeholder: "Type a message…",
					Rows:        1,
				})
				@button.Button(button.Config{Type: "submit", Variant: button.Primary})
			</form>
			<div class="mt-2">
				@toggle.Toggle(toggle.Config{
					ID:    "chat-eliza",
					Name:  "eliza",
					Label: "ELIZA bot replies",
				})
			</div>
		</div>
	</div>
}

// chatHidden is a hidden input carrying identity into every ws-send frame. It is
// id'd so the rename handler can OOB-swap its value.
templ chatHidden(name, value string) {
	<input type="hidden" id={ "chat-hidden-" + name } name={ name } value={ value }/>
}

// RenameResult is the OOB response to a rename: updates the visible nick and both
// hidden composer inputs so the next message carries the new identity.
templ RenameResult(me chat.Identity) {
	<span id="chat-me" hx-swap-oob="true" class="text-sm font-semibold text-on-surface dark:text-on-surface-dark">{ me.Nick }</span>
	<input type="hidden" id="chat-hidden-nick" hx-swap-oob="true" name="nick" value={ me.Nick }/>
	<input type="hidden" id="chat-hidden-color" hx-swap-oob="true" name="color" value={ me.Color }/>
}
```

> **Note on the toggle name in the ws frame:** htmx's ws extension serializes the composer form's fields. Confirm during implementation that the unchecked toggle is omitted (so `eliza` is absent → bot off) and the checked toggle sends a truthy value. If the toggle renders *outside* the `<form ws-send>` in the component's markup, move it inside the form (or add `form="..."`/`hx-include`) so its value is included. Adjust the handler's flag check (Task 7) to match whatever truthy value the toggle emits.

- [ ] **Step 2: Generate templ**

Run: `templ generate`
Expected: `internal/pages/demo/examples/chat_templ.go` created. If it reports "0 updates", force: `rm internal/pages/demo/examples/chat_templ.go && templ generate`.

- [ ] **Step 3: Build to confirm it compiles**

Run: `go build ./...`
Expected: success. Fix any field-name mismatches against the real component `Config` structs (see the generated `.claude/skills/using-goshtoso/components-reference.md`).

- [ ] **Step 4: Commit**

```bash
git add internal/pages/demo/examples/chat.templ internal/pages/demo/examples/chat_templ.go
git commit -m "feat(chat): templ shell, message bubble, presence fragments"
```

---

## Task 6: Vendor the htmx ws extension + load it globally

**Files:**
- Create: `assets/js/vendor/htmx-ext-ws.js`
- Modify: `internal/pages/demo/layout.templ` (head)

- [ ] **Step 1: Download the extension matching the bundled htmx version**

First check the bundled htmx version:
```bash
grep -o 'htmx.org@[0-9.]*\|VERSION[^;]*1\.\|2\.0\.[0-9]*' assets/js/vendor/htmx.min.js | head
```
The repo pins htmx v2.0.8 (see CLAUDE.md). Download the matching ws extension:
```bash
curl -fsSL https://unpkg.com/htmx-ext-ws@2.0.3/ws.js -o assets/js/vendor/htmx-ext-ws.js
```
Expected: a non-empty JS file. Verify:
```bash
head -c 200 assets/js/vendor/htmx-ext-ws.js
```
Expected: htmx extension source (defines `htmx.defineExtension('ws', …)`). If unpkg is unavailable, fetch from the htmx-extensions repo `src/ws/ws.js` at the tag compatible with htmx 2.x.

- [ ] **Step 2: Load it in the layout head (after htmx)**

In `internal/pages/demo/layout.templ`, find the htmx script tag (around line 33):
```html
<script src="/assets/js/vendor/htmx.min.js"></script>
```
Add immediately after it:
```html
<script src="/assets/js/vendor/htmx-ext-ws.js"></script>
```

- [ ] **Step 3: Regenerate templ + build**

Run: `templ generate && go build ./...`
Expected: success. (`//go:embed js` already recurses, so the new vendor file is embedded automatically; the dev server serves it from disk.)

- [ ] **Step 4: Manually confirm the asset serves**

Run (in one shell): `go run cmd/server/main.go` then in another:
```bash
curl -fsS -o /dev/null -w "%{http_code}\n" http://localhost:8090/assets/js/vendor/htmx-ext-ws.js
```
Expected: `200`. Stop the server.

- [ ] **Step 5: Commit**

```bash
git add assets/js/vendor/htmx-ext-ws.js internal/pages/demo/layout.templ internal/pages/demo/layout_templ.go
git commit -m "chore(assets): vendor htmx ws extension + load globally"
```

---

## Task 7: chat_handler.go — page, websocket, rename

**Files:**
- Create: `internal/server/chat_handler.go`
- Modify: `internal/server/server.go` (call `s.registerChatRoutes()` and add the `handleExample` "chat" case)

**Invoke the `htmx` skill** before finalizing the ws read/swap contract.

- [ ] **Step 1: Write the handler**

Create `internal/server/chat_handler.go`:
```go
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	"github.com/araihu/goshtoso/internal/examples/chat"
	"github.com/araihu/goshtoso/internal/pages/demo"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// chatHub is the single process-wide room. RAM-only: a server restart clears it.
var chatHub = chat.NewHub(50)

// chatSeq assigns a deterministic guest seed per new visitor (cookie-less hits).
var chatSeq atomic.Int64

const (
	chatMaxMessageLen = 500             // rune cap per message
	chatWriteTimeout  = 5 * time.Second // per-frame write deadline
	chatBotDelay      = 400 * time.Millisecond
)

// registerChatRoutes wires the chat example's endpoints.
func (s *Server) registerChatRoutes() {
	s.mux.HandleFunc("/api/examples/chat/ws", s.handleChatWS)
	s.mux.HandleFunc("/api/examples/chat/rename", s.handleChatRename)
}

// identityFromRequest reads the gt_chat cookie, falling back to a fresh guest.
func identityFromRequest(r *http.Request) chat.Identity {
	if c, err := r.Cookie(chat.CookieName); err == nil {
		if id, err := chat.Decode(c.Value); err == nil && id.Nick != "" {
			return id
		}
	}
	return chat.NewGuest(chatSeq.Add(1))
}

// renderChatPage is the first-load handler for /examples/chat. It ensures a
// gt_chat cookie exists (so the subsequent ws upgrade carries an identity), then
// renders either the full Layout or an HTMX Fragment.
func (s *Server) renderChatPage(w http.ResponseWriter, r *http.Request) {
	me := identityFromRequest(r)
	if _, err := r.Cookie(chat.CookieName); err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     chat.CookieName,
			Value:    me.Encode(),
			Path:     "/",
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.ChatApp(me)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.Fragment("Chat", "chat", content).Render(r.Context(), w)
		return
	}
	_ = demo.Layout("Chat", "chat", content).Render(r.Context(), w)
}

// handleChatRename sets a new nick on the gt_chat cookie and returns OOB fragments
// that update the visible nick and the composer's hidden identity inputs.
func (s *Server) handleChatRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	me := identityFromRequest(r)
	if nick := strings.TrimSpace(r.FormValue("nick")); nick != "" {
		me.Nick = truncateRunes(nick, 24)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     chat.CookieName,
		Value:    me.Encode(),
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.RenameResult(me).Render(r.Context(), w)
}

// wsFrame is the JSON htmx's ws extension sends on ws-send: the composer's form
// fields plus a HEADERS object we ignore.
type wsFrame struct {
	Message string `json:"message"`
	Nick    string `json:"nick"`
	Color   string `json:"color"`
	Eliza   string `json:"eliza"` // present+truthy when the toggle is on
}

// handleChatWS upgrades to a websocket, joins the room, replays backlog, then
// loops reading messages and broadcasting rendered bubbles.
func (s *Server) handleChatWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // same-origin demo; tests hit localhost
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	me := identityFromRequest(r)
	client := chat.NewClient()
	chatHub.Register(client)

	// Replay history to this client only (before joining the broadcast view).
	for _, frame := range chatHub.Backlog() {
		select {
		case client.Send <- frame:
		default:
		}
	}
	// Announce arrival to everyone.
	chatHub.Broadcast(renderFrame(examples.PresenceFrame(me.Nick+" joined", chatHub.Count())))

	ctx := r.Context()
	go chatWritePump(ctx, conn, client)
	defer func() {
		chatHub.Unregister(client)
		chatHub.Broadcast(renderFrame(examples.PresenceFrame(me.Nick+" left", chatHub.Count())))
	}()

	conn.SetReadLimit(8 << 10) // 8 KiB inbound cap
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return // client closed or errored
		}
		var f wsFrame
		if err := json.Unmarshal(data, &f); err != nil {
			continue
		}
		text := truncateRunes(strings.TrimSpace(f.Message), chatMaxMessageLen)
		if text == "" {
			continue
		}
		nick := truncateRunes(strings.TrimSpace(f.Nick), 24)
		if nick == "" {
			nick = me.Nick
		}
		now := time.Now().Format("15:04")
		chatHub.Broadcast(renderFrame(examples.MessageFrame(examples.Message{
			Nick: nick, Color: f.Color, Text: text, Time: now,
		})))

		if isToggleOn(f.Eliza) {
			if reply, ok := chat.Reply(text); ok {
				botText, botTime := reply, time.Now().Format("15:04")
				go func() {
					time.Sleep(chatBotDelay)
					chatHub.Broadcast(renderFrame(examples.MessageFrame(examples.Message{
						Nick: "ELIZA", Color: "#3b82f6", Text: botText, Time: botTime, IsBot: true,
					})))
				}()
			}
		}
	}
}

// chatWritePump drains a client's send channel to the socket until it closes.
func chatWritePump(ctx context.Context, conn *websocket.Conn, client *chat.Client) {
	for frame := range client.Send {
		wctx, cancel := context.WithTimeout(ctx, chatWriteTimeout)
		err := conn.Write(wctx, websocket.MessageText, frame)
		cancel()
		if err != nil {
			return
		}
	}
}

// renderFrame renders a templ component to bytes for pushing over the socket.
func renderFrame(c interface {
	Render(context.Context, *bytes.Buffer) error
}) []byte {
	// templ components satisfy Render(ctx, io.Writer); use a buffer.
	var buf bytes.Buffer
	_ = renderToBuffer(&buf, c)
	return buf.Bytes()
}

// isToggleOn reports whether the htmx-serialized toggle value is truthy.
func isToggleOn(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "off", "false", "0":
		return false
	default:
		return true
	}
}

// truncateRunes caps s to at most n runes.
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
```

> **Implementation note on `renderFrame`:** templ's `Component` interface is `Render(ctx context.Context, w io.Writer) error`. Simplify the helper to take `templ.Component` and render into a `bytes.Buffer` directly:
> ```go
> import "github.com/a-h/templ"
> func renderFrame(c templ.Component) []byte {
> 	var buf bytes.Buffer
> 	_ = c.Render(context.Background(), &buf)
> 	return buf.Bytes()
> }
> ```
> Use this form and delete the `renderToBuffer`/interface shim above. (The shim is shown only to flag the one spot that needs the real templ import.)

- [ ] **Step 2: Wire the routes and the page case in `server.go`**

In `internal/server/server.go`, in the route-registration section (near the other `register*Routes`/`HandleFunc` calls around lines 41–58), add:
```go
s.registerChatRoutes()
```
(Place it alongside the existing `registerTodoRoutes()` call — grep for `registerTodoRoutes` to find where example routes are registered; if todo's is called in `New`/`routes`, mirror it.)

Then in `handleExample` (around line 90), add a case before `default`:
```go
	case "chat":
		s.renderChatPage(w, r)
```

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: success. Resolve the `renderFrame` helper per the note (use the real `templ.Component` form).

- [ ] **Step 4: Vet + race build of the package**

Run: `go vet ./internal/server/... && go build ./...`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add internal/server/chat_handler.go internal/server/server.go
git commit -m "feat(chat): websocket handler, rename, page route"
```

---

## Task 8: Registry, sidebar, gallery wiring

**Files:**
- Modify: `internal/pages/demo/examples/index.templ` (gallery card)
- Modify: `internal/pages/demo/layout.templ` (sidebar Examples item)

Note: `/examples/chat` is served by the `handleExample` switch (Task 7), **not** the `LookupDemo` registry (the page needs the request to read the cookie, exactly like todo). So no `registry.go` entry is required.

- [ ] **Step 1: Add the sidebar item**

In `internal/pages/demo/layout.templ`, find the Examples section (around line 465–469):
```go
				sItem("examples", "Overview", "/examples", activeComponent),
				sItem("todo", "Todo List", "/examples/todo", activeComponent),
```
Add after the todo line:
```go
				sItem("chat", "Chat", "/examples/chat", activeComponent),
```

- [ ] **Step 2: Add the gallery card**

In `internal/pages/demo/examples/index.templ`, find the `@exampleCard("/examples/todo", …)` call (around line 25) and add after it:
```go
			@exampleCard("/examples/chat", "Chat", "Realtime full-duplex chat over htmx websockets with an ELIZA bot, presence, and live rename.")
```

- [ ] **Step 3: Generate + build**

Run: `templ generate && go build ./...`
Expected: success.

- [ ] **Step 4: Manually verify the page renders**

Run: `go run cmd/server/main.go`, then open `http://localhost:8090/examples/chat` in a browser. Confirm: the shell renders, the sidebar shows Chat under Examples, no console errors, and a message you type appears in the log. Open a second tab and confirm the message appears in both. Stop the server.

- [ ] **Step 5: Commit**

```bash
git add internal/pages/demo/layout.templ internal/pages/demo/layout_templ.go internal/pages/demo/examples/index.templ internal/pages/demo/examples/index_templ.go
git commit -m "feat(chat): sidebar item + examples gallery card"
```

---

## Task 9: E2E tests

**Files:**
- Create: `tests/e2e/chat_test.go`

Mirror `tests/e2e/todo_example_test.go`: `newIsolatedPage` for cookie isolation, the `cookieConsent` init-script to suppress the consent banner, `clickUntil` for post-swap clicks, and the console/pageerror capture for the fragment-nav guard. Two-client tests use two separate browser contexts.

- [ ] **Step 1: Write the tests**

Create `tests/e2e/chat_test.go`:
```go
package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// gotoChat navigates to the chat page and waits for the ws-connected composer.
func gotoChat(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: playwright.String("try{localStorage.setItem('cookieConsent','accepted')}catch(e){}"),
	}))
	_, err := page.Goto(baseURL + "/examples/chat")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	// Wait for the ws extension to mark the connecting element open.
	_, err = page.WaitForFunction("() => !!document.querySelector('#chat-log')", nil)
	require.NoError(t, err)
}

// sendChat types into the composer and submits via the ws-send form.
func sendChat(t *testing.T, page playwright.Page, msg string) {
	t.Helper()
	require.NoError(t, page.Locator("#chat-message").Fill(msg))
	require.NoError(t, page.Locator("form[ws-send] button[type='submit']").Click())
}

// TestChat_Broadcast opens two independent browser contexts, sends from A, and
// asserts the message lands in B. The real full-duplex proof.
func TestChat_Broadcast(t *testing.T) {
	pageA := newIsolatedPage(t)
	pageB := newIsolatedPage(t)
	gotoChat(t, pageA)
	gotoChat(t, pageB)

	sendChat(t, pageA, "hello-from-A")

	_, err := pageB.WaitForFunction(
		"() => Array.from(document.querySelectorAll('#chat-log p')).some(p => p.textContent.trim() === 'hello-from-A')",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "message from A should appear in B's log")
}

// TestChat_ElizaDeterministic toggles the bot on, sends a known input, and
// asserts the deterministic reply bubble appears.
func TestChat_ElizaDeterministic(t *testing.T) {
	page := newIsolatedPage(t)
	gotoChat(t, page)

	// Turn on the ELIZA toggle (label wraps the input).
	require.NoError(t, page.Locator("label[for='chat-eliza']").Click())
	sendChat(t, page, "i feel anxious")

	_, err := page.WaitForFunction(
		"() => Array.from(document.querySelectorAll('#chat-log p')).some(p => p.textContent.trim() === 'Why do you feel anxious?')",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "ELIZA should reply deterministically")
}

// TestChat_Presence asserts the online badge tracks a second connection.
func TestChat_Presence(t *testing.T) {
	pageA := newIsolatedPage(t)
	gotoChat(t, pageA)

	// A second context joins → A's badge should read "2 online".
	pageB := newIsolatedPage(t)
	gotoChat(t, pageB)

	_, err := pageA.WaitForFunction(
		"() => document.querySelector('#chat-online') && document.querySelector('#chat-online').textContent.includes('2 online')",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "online badge should reach 2")
}

// TestChat_Rename renames the visitor and asserts the next message carries the
// new nick (the hidden composer input was OOB-swapped).
func TestChat_Rename(t *testing.T) {
	page := newIsolatedPage(t)
	gotoChat(t, page)

	require.NoError(t, page.Locator("#chat-nick-input").Fill("Ada"))
	require.NoError(t, page.Locator("form[hx-post='/api/examples/chat/rename'] button[type='submit']").Click())
	// Wait for the OOB swap to update the visible nick.
	_, err := page.WaitForFunction(
		"() => document.querySelector('#chat-me') && document.querySelector('#chat-me').textContent.trim() === 'Ada'",
		nil)
	require.NoError(t, err)

	sendChat(t, page, "renamed-hello")
	_, err = page.WaitForFunction(
		"() => Array.from(document.querySelectorAll('#chat-log')).some(l => l.textContent.includes('Ada') && l.textContent.includes('renamed-hello'))",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "message after rename should be labeled 'Ada'")
}

// TestChat_FragmentNavNoErrors reaches /examples/chat via the sidebar (htmx
// fragment swap) and asserts the socket connects and broadcasts with no console
// or page errors — the load-bearing regression guard.
func TestChat_FragmentNavNoErrors(t *testing.T) {
	page := newIsolatedPage(t)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("a[href='/examples/chat']").First().Click())
	_, err = page.WaitForFunction("() => !!document.querySelector('#chat-log')", nil)
	require.NoError(t, err)

	// Send a message through the fragment-loaded page and confirm it round-trips.
	sendChat(t, page, "frag-nav-msg")
	_, err = page.WaitForFunction(
		"() => Array.from(document.querySelectorAll('#chat-log p')).some(p => p.textContent.trim() === 'frag-nav-msg')",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "ws should connect + echo on fragment-nav")

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav chat page: %v", jsErrors)
}

// TestChat_SidebarPresent verifies the Examples section lists the Chat link.
func TestChat_SidebarPresent(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/examples/chat")
	require.NoError(t, err)
	require.NoError(t, page.Locator("text=Examples").First().WaitFor())
	require.NoError(t, page.Locator("a[href='/examples/chat']").First().WaitFor())
}
```

- [ ] **Step 2: Run the chat tests**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestChat -v`
Expected: PASS (all 6). If a broadcast test flakes under load, confirm the ws extension has bound before the first `sendChat` (the `#chat-log` wait) and, where a swap rebinds a control, switch the click to `clickUntil`.

- [ ] **Step 3: Commit**

```bash
git add tests/e2e/chat_test.go
git commit -m "test(chat): e2e broadcast, eliza, presence, rename, fragment-nav"
```

---

## Task 10: Full verification, skill sync, lint

**Files:**
- Possibly regenerated: `.claude/skills/using-goshtoso/components-reference.md` (skillgen)

- [ ] **Step 1: Sync the usage skill** (components changed? no — but run to be safe)

Run: `go run ./scripts/skillgen`
Expected: no diff (no component `types.go` changed). If it does change, stage it.

- [ ] **Step 2: Apply modernizations**

Run: `go fix ./...`
Expected: no changes, or AST-safe modernizations to the new non-generated files. Review with `git diff`.

- [ ] **Step 3: Lint**

Run: `golangci-lint run`
Expected: clean. Keep every new function under cyclomatic complexity 20 — if `handleChatWS` trips cyclop, extract the read-loop body into a helper (e.g. `func (s *Server) handleChatMessage(ctx, data, me) []frames` returning frames to broadcast).

- [ ] **Step 4: Full package test (with race) for the domain + server**

Run: `go test ./internal/... -race -count=1`
Expected: PASS.

- [ ] **Step 5: Full E2E suite**

Run: `go test ./tests/e2e/... -count=1 -timeout 15m`
Expected: PASS, no skips. (Existing 381 + the 6 new chat tests.)

- [ ] **Step 6: Final commit (only if skillgen/go-fix changed anything)**

```bash
git add -A
git commit -m "chore(chat): skill sync + modernization after chat example"
```

---

## Self-review notes (already reconciled against the spec)

- **Transport / OOB push** → Tasks 5–7 (MessageFrame `hx-swap-oob`, ws read loop).
- **coder/websocket** → Task 1, used in Task 7.
- **RAM-only hub + 50-message ring + backlog replay + slow-client drop** → Task 4 + Task 7 replay loop.
- **ELIZA, per-message flag** → Task 2 + Task 7 (`isToggleOn`).
- **Identity cookie + live rename** → Task 3 + Task 7 (`handleChatRename`) + Task 5 (`RenameResult` OOB).
- **11 components** → Card dropped (it is a content card with no children slot — forcing it would be artificial); the shell is a styled container like todo. Used: Avatar, Badge, Toggle, Toast, Textarea, Text Input, Button, Spinner *(via the toast/typing affordance — optional)*, Banner *(connection-status, optional polish; add only if time permits — see below)*. Net mandatory showcase: Avatar, Badge, Toggle, Toast, Textarea, Text Input, Button (7 solid, real uses) + optional Spinner/Banner/Tooltip polish. This is honest: no component is forced.
- **Edge cases** → OOB `oob bool` gate (Task 5 `OnlineBadge`), global ext load avoids fragment-nav injection (Task 6), templ escaping via text nodes (Task 5), socket lifecycle caps (Task 7 `SetReadLimit`/`truncateRunes`/write timeout).
- **E2E** → Task 9 covers all five spec scenarios + sidebar.

**Deviations from the spec (intentional, noted for the reviewer):**
1. **ELIZA flag rides the message frame** (form field) rather than a dedicated `{type:"eliza"}` ws frame — removes per-client hub state and HTTP↔ws coupling. Simpler and equally correct.
2. **Card dropped** from the must-use list (no children slot; forcing is worse than omitting). Spinner/Banner/Tooltip are optional polish, not load-bearing, to keep the first cut focused (YAGNI). If the reviewer wants the full 11, add: a connection-status `Banner` toggled by `htmx:wsClose`/`htmx:wsOpen` events, a `Spinner` "ELIZA is typing…" bubble pushed before the bot reply, and `Tooltip` timestamps. Each is a small additive task.
3. **htmx-ext-ws.js loaded globally**, not page-scoped — the layout has a single shared head and no per-page dependency mechanism; a global inert extension is the pragmatic, fragment-nav-safe choice.
