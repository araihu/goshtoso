# Examples — Todo List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an `/examples/*` route family to Goshtoso whose first app is a cookie-backed, HTMX-driven todo list that showcases the component library in real use.

**Architecture:** Stateless server — full todo state lives in a base64url-JSON cookie (`gt_todo`). A pure domain package (`internal/examples/todo`) owns the state model and all mutations; thin HTTP handlers read the cookie, mutate, re-write the cookie, and return HTML fragments. Pages render through the existing demo registry + `renderDemo`, reusing `Layout`/`Fragment`, theme, dark mode, and fragment-swap nav.

**Tech Stack:** Go 1.26, templ v0.3, HTMX v2, Alpine.js v3, Tailwind v4, Playwright E2E.

**Deviations from spec (discovered during planning, intentional):**
- Filter uses a **segmented button group** (3 HTMX POSTs), not the Tabs component — Tabs is Alpine-panel-driven and unsuited to a server-fragment list re-render.
- The per-row "done" toggle is a **styled `<input type=checkbox>` with `hx-post`**, not the Checkbox component — Checkbox/Button Configs expose no `Attrs` field for HTMX wiring. High-level components (Card, TextInput, Select, Button, Badge, Toast) are still used prominently in the page shell.

---

## File Structure

**Create:**
- `internal/examples/todo/state.go` — `Todo`, `State` types; constants (`MaxTodos=50`, `MaxTitleLen=200`); `Encode`/`Decode`.
- `internal/examples/todo/state_mutations.go` — `Add`, `Toggle`, `Delete`, `Edit`, `Reorder`, `ClearCompleted`, `SetFilter`, `Visible`, count helpers.
- `internal/examples/todo/cookie.go` — `CookieName`, `FromRequest(*http.Request) State`, `SetCookie(http.ResponseWriter, State)`.
- `internal/examples/todo/state_test.go` — encode/decode + cookie tests.
- `internal/examples/todo/state_mutations_test.go` — mutation + visible tests.
- `internal/pages/demo/examples/todo.templ` — `TodoContent`, `TodoApp`, `TodoList`, `TodoRow`, `CountBadge`, `priorityBadge`, the add form, the DnD `<script>`.
- `internal/pages/demo/examples/index.templ` — `IndexContent` (cards linking to each example).
- `internal/server/todo_handler.go` — the 7 `/api/examples/todo/*` handlers + `registerTodoRoutes`.
- `tests/e2e/todo_example_test.go` — E2E suite.

**Modify:**
- `internal/pages/demo/components/registry.go` — add `examples` + `examples/todo` entries (import the `examples` package).
- `internal/server/server.go` — register `/examples` + `/examples/` + `handleExample`; call `registerTodoRoutes`.
- `internal/pages/demo/layout.templ` — add an "Examples" sidebar section.

---

## Task 1: Domain types + cookie codec

**Files:**
- Create: `internal/examples/todo/state.go`
- Test: `internal/examples/todo/state_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/examples/todo/state_test.go
package todo

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	s := State{
		Todos:  []Todo{{ID: 1, Title: "Buy milk", Done: true, Priority: "high", Due: "2026-06-01", Order: 0}},
		Filter: "active",
		Seq:    2,
	}
	got, err := Decode([]byte(Encode(s)))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got.Todos) != 1 || got.Todos[0].Title != "Buy milk" || got.Seq != 2 || got.Filter != "active" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestDecodeMalformedReturnsEmpty(t *testing.T) {
	for _, in := range []string{"", "not-base64!!!", "YWJj" /* base64 "abc", invalid json */} {
		got, err := Decode([]byte(in))
		if err == nil && (len(got.Todos) != 0 || got.Seq != 0) {
			t.Fatalf("expected empty state for %q, got %+v", in, got)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/todo/ -run TestEncodeDecode -v`
Expected: FAIL — package/symbols undefined.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/examples/todo/state.go
// Package todo holds the pure, HTTP-free domain model for the /examples/todo
// app. State is serialized into a cookie, so the server keeps no per-user memory.
package todo

import (
	"encoding/base64"
	"encoding/json"
)

const (
	// MaxTodos caps the list so the encoded cookie stays well under the ~4KB limit.
	MaxTodos = 50
	// MaxTitleLen bounds a single title's stored length.
	MaxTitleLen = 200
)

// Todo is a single task. ID is assigned from State.Seq (deterministic, no rand).
type Todo struct {
	ID       int    `json:"i"`
	Title    string `json:"t"`
	Done     bool   `json:"d"`
	Priority string `json:"p"` // "low" | "med" | "high"
	Due      string `json:"u"` // "2006-01-02" or ""
	Order    int    `json:"o"` // explicit sort index
}

// State is the whole per-user todo list plus view filter and the next ID counter.
type State struct {
	Todos  []Todo `json:"todos"`
	Filter string `json:"filter"` // "all" | "active" | "done"
	Seq    int    `json:"seq"`
}

// Encode serializes State to a base64url(JSON) string for cookie storage.
func Encode(s State) string {
	b, _ := json.Marshal(s)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a cookie value back into State. Any error yields the zero State
// so a corrupt/absent cookie degrades gracefully to "empty list".
func Decode(raw []byte) (State, error) {
	var s State
	if len(raw) == 0 {
		return s, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(string(raw))
	if err != nil {
		return State{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	return s, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/todo/ -run TestEncodeDecode -v && go test ./internal/examples/todo/ -run TestDecodeMalformed -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/examples/todo/state.go internal/examples/todo/state_test.go
git commit -m "feat(examples): todo domain types + cookie codec"
```

---

## Task 2: Add (cap + truncate + Seq)

**Files:**
- Create: `internal/examples/todo/state_mutations.go`
- Test: `internal/examples/todo/state_mutations_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/examples/todo/state_mutations_test.go
package todo

import (
	"strings"
	"testing"
)

func TestAddAssignsIncrementingIDsAndOrder(t *testing.T) {
	var s State
	s.Add("first", "low", "")
	s.Add("second", "high", "2026-06-01")
	if len(s.Todos) != 2 {
		t.Fatalf("want 2 todos, got %d", len(s.Todos))
	}
	if s.Todos[0].ID != 1 || s.Todos[1].ID != 2 {
		t.Fatalf("ids not monotonic: %+v", s.Todos)
	}
	if s.Todos[0].Order != 0 || s.Todos[1].Order != 1 {
		t.Fatalf("order not sequential: %+v", s.Todos)
	}
	if s.Seq != 2 {
		t.Fatalf("seq want 2, got %d", s.Seq)
	}
}

func TestAddTruncatesTitleAndDefaultsPriority(t *testing.T) {
	var s State
	s.Add(strings.Repeat("x", MaxTitleLen+50), "bogus", "")
	if len(s.Todos[0].Title) != MaxTitleLen {
		t.Fatalf("title not truncated: %d", len(s.Todos[0].Title))
	}
	if s.Todos[0].Priority != "med" {
		t.Fatalf("unknown priority should default to med, got %q", s.Todos[0].Priority)
	}
}

func TestAddRejectsBlankAndRespectsCap(t *testing.T) {
	var s State
	s.Add("   ", "low", "") // blank after trim → ignored
	if len(s.Todos) != 0 {
		t.Fatalf("blank title should be ignored")
	}
	for i := 0; i < MaxTodos+10; i++ {
		s.Add("t", "low", "")
	}
	if len(s.Todos) != MaxTodos {
		t.Fatalf("cap not enforced: %d", len(s.Todos))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/todo/ -run TestAdd -v`
Expected: FAIL — `Add` undefined.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/examples/todo/state_mutations.go
package todo

import (
	"slices"
	"strings"
)

func normalizePriority(p string) string {
	switch p {
	case "low", "med", "high":
		return p
	default:
		return "med"
	}
}

// Add appends a todo. Blank titles are ignored; the title is trimmed and capped
// to MaxTitleLen; the list is capped to MaxTodos. ID comes from the Seq counter.
func (s *State) Add(title, priority, due string) {
	title = strings.TrimSpace(title)
	if title == "" || len(s.Todos) >= MaxTodos {
		return
	}
	if len(title) > MaxTitleLen {
		title = title[:MaxTitleLen]
	}
	s.Seq++
	s.Todos = append(s.Todos, Todo{
		ID:       s.Seq,
		Title:    title,
		Priority: normalizePriority(priority),
		Due:      due,
		Order:    len(s.Todos),
	})
}

// indexByID returns the slice index of the todo with id, or -1.
func (s *State) indexByID(id int) int {
	return slices.IndexFunc(s.Todos, func(t Todo) bool { return t.ID == id })
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/todo/ -run TestAdd -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/examples/todo/state_mutations.go internal/examples/todo/state_mutations_test.go
git commit -m "feat(examples): todo Add with cap, truncation, Seq IDs"
```

---

## Task 3: Toggle, Delete, Edit

**Files:**
- Modify: `internal/examples/todo/state_mutations.go`
- Test: `internal/examples/todo/state_mutations_test.go`

- [ ] **Step 1: Write the failing test**

```go
// append to internal/examples/todo/state_mutations_test.go

func seeded() State {
	var s State
	s.Add("a", "low", "")  // ID 1
	s.Add("b", "high", "") // ID 2
	return s
}

func TestToggleFlipsDone(t *testing.T) {
	s := seeded()
	s.Toggle(1)
	if !s.Todos[0].Done {
		t.Fatalf("toggle should set done")
	}
	s.Toggle(1)
	if s.Todos[0].Done {
		t.Fatalf("toggle should clear done")
	}
	s.Toggle(999) // unknown id: no-op, no panic
}

func TestDeleteRemovesByID(t *testing.T) {
	s := seeded()
	s.Delete(1)
	if len(s.Todos) != 1 || s.Todos[0].ID != 2 {
		t.Fatalf("delete failed: %+v", s.Todos)
	}
	s.Delete(999) // unknown id: no-op
}

func TestEditUpdatesFieldsAndDefaults(t *testing.T) {
	s := seeded()
	s.Edit(2, "  renamed  ", "bogus", "2026-07-01")
	got := s.Todos[1]
	if got.Title != "renamed" || got.Priority != "med" || got.Due != "2026-07-01" {
		t.Fatalf("edit mismatch: %+v", got)
	}
	s.Edit(2, "   ", "low", "") // blank title ignored, other fields still applied
	if s.Todos[1].Title != "renamed" {
		t.Fatalf("blank title should not overwrite")
	}
	if s.Todos[1].Priority != "low" {
		t.Fatalf("priority should update even when title blank")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/todo/ -run 'TestToggle|TestDelete|TestEdit' -v`
Expected: FAIL — methods undefined.

- [ ] **Step 3: Write minimal implementation**

```go
// append to internal/examples/todo/state_mutations.go

// Toggle flips Done for the todo with id. Unknown id is a no-op.
func (s *State) Toggle(id int) {
	if i := s.indexByID(id); i >= 0 {
		s.Todos[i].Done = !s.Todos[i].Done
	}
}

// Delete removes the todo with id. Unknown id is a no-op.
func (s *State) Delete(id int) {
	if i := s.indexByID(id); i >= 0 {
		s.Todos = slices.Delete(s.Todos, i, i+1)
	}
}

// Edit updates priority and due always; the title only when the trimmed input
// is non-empty (a blank title never overwrites an existing one).
func (s *State) Edit(id int, title, priority, due string) {
	i := s.indexByID(id)
	if i < 0 {
		return
	}
	if t := strings.TrimSpace(title); t != "" {
		if len(t) > MaxTitleLen {
			t = t[:MaxTitleLen]
		}
		s.Todos[i].Title = t
	}
	s.Todos[i].Priority = normalizePriority(priority)
	s.Todos[i].Due = due
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/todo/ -run 'TestToggle|TestDelete|TestEdit' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/examples/todo/state_mutations.go internal/examples/todo/state_mutations_test.go
git commit -m "feat(examples): todo Toggle/Delete/Edit"
```

---

## Task 4: Reorder, ClearCompleted, SetFilter, Visible, counts

**Files:**
- Modify: `internal/examples/todo/state_mutations.go`
- Test: `internal/examples/todo/state_mutations_test.go`

- [ ] **Step 1: Write the failing test**

```go
// append to internal/examples/todo/state_mutations_test.go

func TestReorderReassignsOrderAndTrailsOmitted(t *testing.T) {
	var s State
	s.Add("a", "low", "") // ID 1
	s.Add("b", "low", "") // ID 2
	s.Add("c", "low", "") // ID 3
	s.Reorder([]int{3, 1}) // 2 omitted, unknown 99 ignored if present
	byID := map[int]int{}
	for _, td := range s.Todos {
		byID[td.ID] = td.Order
	}
	if byID[3] != 0 || byID[1] != 1 {
		t.Fatalf("explicit order wrong: %+v", byID)
	}
	if byID[2] < 2 {
		t.Fatalf("omitted todo should trail: %+v", byID)
	}
}

func TestVisibleAppliesFilterSortedByOrder(t *testing.T) {
	var s State
	s.Add("a", "low", "") // ID1 order0
	s.Add("b", "low", "") // ID2 order1
	s.Toggle(1)           // a done
	s.Reorder([]int{2, 1})

	s.SetFilter("active")
	vis := s.Visible()
	if len(vis) != 1 || vis[0].ID != 2 {
		t.Fatalf("active filter wrong: %+v", vis)
	}
	s.SetFilter("done")
	if v := s.Visible(); len(v) != 1 || v[0].ID != 1 {
		t.Fatalf("done filter wrong: %+v", v)
	}
	s.SetFilter("all")
	if v := s.Visible(); len(v) != 2 || v[0].ID != 2 || v[1].ID != 1 {
		t.Fatalf("all filter / order wrong: %+v", v)
	}
}

func TestClearCompletedAndCounts(t *testing.T) {
	var s State
	s.Add("a", "low", "")
	s.Add("b", "low", "")
	s.Toggle(1)
	if a, d := s.ActiveCount(), s.DoneCount(); a != 1 || d != 1 {
		t.Fatalf("counts wrong: active=%d done=%d", a, d)
	}
	s.ClearCompleted()
	if len(s.Todos) != 1 || s.Todos[0].ID != 2 {
		t.Fatalf("clear-completed failed: %+v", s.Todos)
	}
}

func TestSetFilterRejectsUnknown(t *testing.T) {
	var s State
	s.SetFilter("garbage")
	if s.Filter != "all" {
		t.Fatalf("unknown filter should fall back to all, got %q", s.Filter)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/todo/ -run 'TestReorder|TestVisible|TestClear|TestSetFilter' -v`
Expected: FAIL — methods undefined.

- [ ] **Step 3: Write minimal implementation**

```go
// append to internal/examples/todo/state_mutations.go

// Reorder reassigns Order so the todos named in ids come first in that sequence;
// any todo not listed keeps its relative order and trails after. Unknown ids are
// ignored.
func (s *State) Reorder(ids []int) {
	rank := make(map[int]int, len(ids))
	for pos, id := range ids {
		rank[id] = pos
	}
	next := len(ids)
	// Stable trailing order: walk current Order, assign trailing ranks in order.
	trailing := make([]Todo, 0, len(s.Todos))
	for _, td := range s.Todos {
		if _, ok := rank[td.ID]; !ok {
			trailing = append(trailing, td)
		}
	}
	slices.SortStableFunc(trailing, func(a, b Todo) int { return a.Order - b.Order })
	for _, td := range trailing {
		rank[td.ID] = next
		next++
	}
	for i := range s.Todos {
		s.Todos[i].Order = rank[s.Todos[i].ID]
	}
}

// ClearCompleted removes all done todos.
func (s *State) ClearCompleted() {
	s.Todos = slices.DeleteFunc(s.Todos, func(t Todo) bool { return t.Done })
}

// SetFilter sets the view filter; unknown values fall back to "all".
func (s *State) SetFilter(f string) {
	switch f {
	case "all", "active", "done":
		s.Filter = f
	default:
		s.Filter = "all"
	}
}

// Visible returns the todos for the current filter, sorted by Order.
func (s State) Visible() []Todo {
	out := make([]Todo, 0, len(s.Todos))
	for _, t := range s.Todos {
		switch s.Filter {
		case "active":
			if t.Done {
				continue
			}
		case "done":
			if !t.Done {
				continue
			}
		}
		out = append(out, t)
	}
	slices.SortStableFunc(out, func(a, b Todo) int { return a.Order - b.Order })
	return out
}

// ActiveCount returns the number of not-done todos.
func (s State) ActiveCount() int {
	n := 0
	for _, t := range s.Todos {
		if !t.Done {
			n++
		}
	}
	return n
}

// DoneCount returns the number of done todos.
func (s State) DoneCount() int { return len(s.Todos) - s.ActiveCount() }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/todo/ -v`
Expected: PASS (all domain tests).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/todo/state_mutations.go internal/examples/todo/state_mutations_test.go
git commit -m "feat(examples): todo Reorder/ClearCompleted/SetFilter/Visible/counts"
```

---

## Task 5: Cookie request/response helpers

**Files:**
- Create: `internal/examples/todo/cookie.go`
- Test: `internal/examples/todo/state_test.go` (append)

- [ ] **Step 1: Write the failing test**

```go
// append to internal/examples/todo/state_test.go
import_block_note := 0
_ = import_block_note
```

(Replace the file's import block with the following, then append the test.)

```go
// internal/examples/todo/state_test.go — imports
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)
```

```go
// append to internal/examples/todo/state_test.go

func TestCookieRoundTripThroughHTTP(t *testing.T) {
	var s State
	s.Add("ship it", "high", "")

	rec := httptest.NewRecorder()
	SetCookie(rec, s)

	cookie := rec.Result().Cookies()[0]
	if cookie.Name != CookieName || cookie.Path != "/" || !cookie.HttpOnly {
		t.Fatalf("cookie attrs wrong: %+v", cookie)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	got := FromRequest(req)
	if len(got.Todos) != 1 || got.Todos[0].Title != "ship it" {
		t.Fatalf("did not round-trip via http: %+v", got)
	}
}

func TestFromRequestNoCookieIsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := FromRequest(req); len(got.Todos) != 0 {
		t.Fatalf("missing cookie should be empty state")
	}
}

func TestFromRequestCorruptCookieIsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: strings.Repeat("!", 20)})
	if got := FromRequest(req); len(got.Todos) != 0 {
		t.Fatalf("corrupt cookie should be empty state")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/todo/ -run TestCookie -v`
Expected: FAIL — `SetCookie`/`FromRequest`/`CookieName` undefined.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/examples/todo/cookie.go
package todo

import "net/http"

// CookieName is the cookie that carries the encoded todo State.
const CookieName = "gt_todo"

// cookieMaxAge is ~30 days in seconds.
const cookieMaxAge = 30 * 24 * 60 * 60

// FromRequest reads and decodes State from the request cookie. A missing or
// corrupt cookie yields the zero State.
func FromRequest(r *http.Request) State {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return State{}
	}
	s, err := Decode([]byte(c.Value))
	if err != nil {
		return State{}
	}
	return s
}

// SetCookie writes the encoded State as a cookie. Path is "/" so it is sent to
// both /examples/* and /api/examples/todo/* (a narrower path would not reach
// the API endpoints).
func SetCookie(w http.ResponseWriter, s State) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    Encode(s),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   cookieMaxAge,
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/todo/ -v`
Expected: PASS (all domain + cookie tests).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/todo/cookie.go internal/examples/todo/state_test.go
git commit -m "feat(examples): todo cookie request/response helpers"
```

---

## Task 6: Examples package — row, list, and count partials

**Files:**
- Create: `internal/pages/demo/examples/todo.templ`

> These partials are rendered by BOTH the page shell and the HTTP handlers, so they live in the `examples` package and are exported. `templ generate` produces `todo_templ.go`.

- [ ] **Step 1: Write the partials**

```go
// internal/pages/demo/examples/todo.templ
package examples

import (
	"fmt"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/internal/examples/todo"
)

// rowID is the DOM id for a single todo's <li>, used as the htmx swap target.
func rowID(id int) string { return fmt.Sprintf("todo-row-%d", id) }

// priorityVariant maps a priority string to a badge variant.
func priorityVariant(p string) badge.Variant {
	switch p {
	case "high":
		return badge.Danger
	case "low":
		return badge.Secondary
	default:
		return badge.Warning
	}
}

// CountBadge is the live "N active" indicator. It carries hx-swap-oob so handlers
// can replace it out of band on every mutation.
templ CountBadge(active int) {
	<span id="todo-count" hx-swap-oob="true" class="text-sm text-on-surface-variant dark:text-on-surface-dark-variant">
		{ fmt.Sprintf("%d active", active) }
	</span>
}

// TodoRow renders one todo as a draggable <li> with toggle, title, priority
// badge, optional due date, up/down reorder, and delete. Interactive controls
// are explicit HTMX markup (the Checkbox/Button components expose no Attrs field).
templ TodoRow(t todo.Todo) {
	<li
		id={ rowID(t.ID) }
		data-todo-id={ fmt.Sprintf("%d", t.ID) }
		draggable="true"
		x-on:dragstart="onDragStart($event, $el)"
		x-on:dragover.prevent="onDragOver($event, $el)"
		x-on:drop.prevent="onDrop($event, $el)"
		x-on:dragend="onDragEnd()"
		class="flex items-center gap-3 rounded-radius border border-outline bg-surface px-3 py-2 dark:border-outline-dark dark:bg-surface-dark"
	>
		<span class="cursor-grab select-none text-on-surface-variant dark:text-on-surface-dark-variant" aria-hidden="true">⠿</span>
		<input
			type="checkbox"
			aria-label="Toggle done"
			class="size-4 shrink-0 rounded-radius accent-primary dark:accent-primary-dark"
			if t.Done {
				checked
			}
			hx-post={ fmt.Sprintf("/api/examples/todo/toggle?id=%d", t.ID) }
			hx-target={ "#" + rowID(t.ID) }
			hx-swap="outerHTML"
		/>
		<span
			class={ "flex-1 text-sm text-on-surface dark:text-on-surface-dark",
				templ.KV("line-through opacity-60", t.Done) }
		>
			{ t.Title }
		</span>
		if t.Due != "" {
			<span class="text-xs text-on-surface-variant dark:text-on-surface-dark-variant">{ t.Due }</span>
		}
		@badge.Badge(badge.Config{Variant: priorityVariant(t.Priority), Text: t.Priority})
		<button
			type="button"
			aria-label="Move up"
			class="rounded-radius px-1 text-on-surface-variant hover:bg-surface-alt dark:text-on-surface-dark-variant dark:hover:bg-surface-dark-alt"
			hx-post={ fmt.Sprintf("/api/examples/todo/move?id=%d&dir=up", t.ID) }
			hx-target="#todo-list"
			hx-swap="outerHTML"
		>↑</button>
		<button
			type="button"
			aria-label="Move down"
			class="rounded-radius px-1 text-on-surface-variant hover:bg-surface-alt dark:text-on-surface-dark-variant dark:hover:bg-surface-dark-alt"
			hx-post={ fmt.Sprintf("/api/examples/todo/move?id=%d&dir=down", t.ID) }
			hx-target="#todo-list"
			hx-swap="outerHTML"
		>↓</button>
		<button
			type="button"
			aria-label="Delete"
			class="rounded-radius px-1 text-danger hover:bg-danger/10 dark:text-danger-dark"
			hx-post={ fmt.Sprintf("/api/examples/todo/delete?id=%d", t.ID) }
			hx-target={ "#" + rowID(t.ID) }
			hx-swap="outerHTML"
		>✕</button>
	</li>
}

// TodoList renders the <ul> of visible todos (or an empty state). It is the
// reorder/filter swap target (#todo-list).
templ TodoList(s todo.State) {
	<ul id="todo-list" class="flex flex-col gap-2" x-data="todoApp()">
		if len(s.Visible()) == 0 {
			<li class="rounded-radius border border-dashed border-outline px-3 py-8 text-center text-sm text-on-surface-variant dark:border-outline-dark dark:text-on-surface-dark-variant">
				Nothing here yet — add your first task above.
			</li>
		} else {
			for _, t := range s.Visible() {
				@TodoRow(t)
			}
		}
	</ul>
}
```

> Note: `badge.Config` uses field `Text` (verify the exact field name with `grep -n "Text\|Label" components/badge/types.go` before running generate; if the field is `Label`, use that). `rounded-radius`, `surface-alt`, `danger` etc. are existing theme utilities used elsewhere in the repo.

- [ ] **Step 2: Generate templ**

Run: `templ generate`
Expected: creates `internal/pages/demo/examples/todo_templ.go`, no errors.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/pages/demo/examples/`
Expected: builds clean. (If `badge.Config` field name is wrong, fix and re-generate.)

- [ ] **Step 4: Commit**

```bash
git add internal/pages/demo/examples/todo.templ internal/pages/demo/examples/todo_templ.go
git commit -m "feat(examples): todo row/list/count templ partials"
```

---

## Task 7: Examples page shell (TodoApp + TodoContent) and DnD script

**Files:**
- Modify: `internal/pages/demo/examples/todo.templ`

- [ ] **Step 1: Add the shell, add-form, DnD script**

```go
// append to internal/pages/demo/examples/todo.templ — add these imports to the existing import block:
//   "github.com/araihu/goshtoso/components/button"
//   "github.com/araihu/goshtoso/components/card"
//   "github.com/araihu/goshtoso/components/select"
//   "github.com/araihu/goshtoso/components/textinput"
//   "github.com/araihu/goshtoso/components/toast"

// filterButton renders one segment of the All/Active/Done filter group.
templ filterButton(label, value, current string) {
	<button
		type="button"
		class={ "rounded-radius px-3 py-1 text-sm",
			templ.KV("bg-primary text-on-primary dark:bg-primary-dark dark:text-on-primary-dark", value == current),
			templ.KV("text-on-surface-variant hover:bg-surface-alt dark:text-on-surface-dark-variant dark:hover:bg-surface-dark-alt", value != current) }
		hx-post={ "/api/examples/todo/filter?f=" + value }
		hx-target="#todo-list"
		hx-swap="outerHTML"
	>{ label }</button>
}

// TodoApp is the interactive card (add form, filter group, list, count). It is
// the content rendered both on first load and as the registry fragment.
templ TodoApp(s todo.State) {
	<div id="todo-fragment" class="mx-auto max-w-2xl">
		@card.Card(card.Config{}) {
			<div class="flex flex-col gap-4 p-2">
				<div class="flex items-center justify-between">
					<h1 class="text-xl font-semibold text-on-surface dark:text-on-surface-dark">Todo List</h1>
					@CountBadge(s.ActiveCount())
				</div>
				<form
					class="flex flex-col gap-3 sm:flex-row sm:items-end"
					hx-post="/api/examples/todo/add"
					hx-target="#todo-list"
					hx-swap="beforeend"
					hx-on::after-request="if(event.detail.successful) this.reset()"
				>
					<div class="flex-1">
						@textinput.TextInput(textinput.Config{Name: "title", Label: "Task", Placeholder: "What needs doing?"})
					</div>
					@select.Select(select.Config{
						Name:  "priority",
						Label: "Priority",
						Options: []select.Option{
							{Value: "low", Label: "Low"},
							{Value: "med", Label: "Medium", Selected: true},
							{Value: "high", Label: "High"},
						},
					})
					<div>
						<label class="mb-1 block text-sm text-on-surface dark:text-on-surface-dark">Due</label>
						<input type="date" name="due" class="rounded-radius border border-outline bg-surface px-2 py-2 text-sm text-on-surface dark:border-outline-dark dark:bg-surface-dark dark:text-on-surface-dark"/>
					</div>
					@button.Button(button.Config{Type: "submit", Variant: button.Primary, Label: "Add"})
				</form>
				<div class="flex items-center justify-between border-t border-outline pt-3 dark:border-outline-dark">
					<div class="flex gap-1">
						@filterButton("All", "all", s.Filter)
						@filterButton("Active", "active", s.Filter)
						@filterButton("Done", "done", s.Filter)
					</div>
					<button
						type="button"
						class="text-sm text-on-surface-variant hover:underline dark:text-on-surface-dark-variant"
						hx-post="/api/examples/todo/clear-completed"
						hx-target="#todo-list"
						hx-swap="outerHTML"
					>Clear completed</button>
				</div>
				@TodoList(s)
			</div>
		}
		@toast.Container(toast.ContainerConfig{})
		@todoScript()
	</div>
}

// TodoContent is the registry entry point: it reads no request here (the handler
// seeds first load), so it renders an empty-state app. Server first-load uses
// TodoApp(state); the registry uses this zero-state version for direct nav when
// no cookie exists. See server handleExample which prefers TodoApp.
templ TodoContent() {
	@TodoApp(todo.State{Filter: "all"})
}

// todoScript registers the Alpine component powering native drag-and-drop. Wrapped
// in a <script> + templ.Raw so templ does not escape the JS (per CLAUDE.md).
templ todoScript() {
	@templ.Raw("<script>" + todoScriptJS + "</script>")
}
```

```go
// also append to internal/pages/demo/examples/todo.templ (Go const, outside any templ block):

const todoScriptJS = `
document.addEventListener('alpine:init', () => {
  Alpine.data('todoApp', () => ({
    dragId: null,
    onDragStart(e, el) { this.dragId = el.dataset.todoId; e.dataTransfer.effectAllowed = 'move'; },
    onDragOver(e, el) { e.dataTransfer.dropEffect = 'move'; },
    onDrop(e, el) {
      const list = this.$el;
      const dragged = list.querySelector('[data-todo-id="' + this.dragId + '"]');
      if (!dragged || dragged === el) return;
      const rect = el.getBoundingClientRect();
      const after = e.clientY > rect.top + rect.height / 2;
      list.insertBefore(dragged, after ? el.nextSibling : el);
    },
    onDragEnd() {
      const ids = Array.from(this.$el.querySelectorAll('[data-todo-id]')).map(n => n.dataset.todoId);
      this.dragId = null;
      htmx.ajax('POST', '/api/examples/todo/reorder', { values: { ids: ids.join(',') }, swap: 'none' });
    },
  }));
});
`
```

> Note: `select.Option` field names (`Value`, `Label`, `Selected`) — verify with `grep -n "type Option" -A8 components/select/types.go`; adjust if different. `card.Config{}` with children: confirm `Card` accepts `{ children... }` (it renders a card shell — `grep -n "children" components/card/card.templ`). If `Card` does not take children, wrap the body in a plain `<div class="rounded-radius border border-outline bg-surface p-4 dark:border-outline-dark dark:bg-surface-dark">` instead.

- [ ] **Step 2: Generate + build**

Run: `templ generate && go build ./internal/pages/demo/examples/`
Expected: clean build. Fix any field-name mismatches flagged in the notes.

- [ ] **Step 3: Commit**

```bash
git add internal/pages/demo/examples/todo.templ internal/pages/demo/examples/todo_templ.go
git commit -m "feat(examples): todo app shell, add form, filter group, DnD script"
```

---

## Task 8: Examples index page

**Files:**
- Create: `internal/pages/demo/examples/index.templ`

- [ ] **Step 1: Write the index content**

```go
// internal/pages/demo/examples/index.templ
package examples

// exampleCard is one entry in the examples gallery.
templ exampleCard(href, title, desc string) {
	<a
		href={ templ.SafeURL(href) }
		hx-get={ href }
		hx-target="#main-content"
		hx-push-url="true"
		class="block rounded-radius border border-outline bg-surface p-5 transition hover:border-primary dark:border-outline-dark dark:bg-surface-dark dark:hover:border-primary-dark"
	>
		<h3 class="mb-1 text-lg font-semibold text-on-surface dark:text-on-surface-dark">{ title }</h3>
		<p class="text-sm text-on-surface-variant dark:text-on-surface-dark-variant">{ desc }</p>
	</a>
}

// IndexContent is the /examples landing page: a gallery of example apps.
templ IndexContent() {
	<div id="examples-fragment" class="mx-auto max-w-4xl">
		<h1 class="mb-2 text-2xl font-bold text-on-surface dark:text-on-surface-dark">Examples</h1>
		<p class="mb-6 text-on-surface-variant dark:text-on-surface-dark-variant">
			Small, real apps built from Goshtoso components.
		</p>
		<div class="grid gap-4 sm:grid-cols-2">
			@exampleCard("/examples/todo", "Todo List", "Cookie-backed, HTMX-driven task list with priorities, filters, and drag reorder.")
		</div>
	</div>
}
```

> Note: confirm the fragment-nav attribute pattern by checking how `sItem`/`navHxAttrs` build links in `layout.templ`; the `hx-get`/`hx-target="#main-content"`/`hx-push-url` triplet matches sidebar nav. If the repo uses a different target id, match it.

- [ ] **Step 2: Generate + build**

Run: `templ generate && go build ./internal/pages/demo/examples/`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add internal/pages/demo/examples/index.templ internal/pages/demo/examples/index_templ.go
git commit -m "feat(examples): examples index gallery page"
```

---

## Task 9: Register routes, registry entries, sidebar section

**Files:**
- Modify: `internal/pages/demo/components/registry.go`
- Modify: `internal/server/server.go`
- Modify: `internal/pages/demo/layout.templ`

- [ ] **Step 1: Add registry entries**

In `internal/pages/demo/components/registry.go`, add the import and two map entries:

```go
// add to imports
import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// add inside the Demos map literal:
	"examples":      {"Examples", "examples", examples.IndexContent},
	"examples/todo": {"Todo List", "todo", examples.TodoContent},
```

- [ ] **Step 2: Add the /examples handler in server.go**

In `setupRoutes()` (after the `/components/` registration) add:

```go
	// Example apps
	s.mux.HandleFunc("/examples", s.handleExample)
	s.mux.HandleFunc("/examples/", s.handleExample)
	s.registerTodoRoutes()
```

Add the handler (mirrors `handleComponent`, but seeds the todo page with the cookie state on first load):

```go
// handleExample resolves /examples and /examples/<name> to a registry page.
// For the todo app it renders the cookie-backed state directly so a returning
// user sees their list on first load (not the empty registry version).
func (s *Server) handleExample(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/examples")
	path = strings.TrimPrefix(path, "/")
	name := ""
	if path != "" {
		name = strings.Split(path, "/")[0]
	}
	if name == "todo" {
		s.renderTodoPage(w, r)
		return
	}
	key := "examples"
	if name != "" {
		key = "examples/" + name
	}
	s.renderDemo(w, r, key)
}
```

> `renderTodoPage` is added in Task 10 (it needs the `examples.TodoApp(state)` + Layout/Fragment split). For this step, temporarily render the zero-state via `s.renderDemo(w, r, "examples/todo")` for `name == "todo"` so the build is green, then Task 10 replaces it.

For now, use:

```go
	if name == "todo" {
		s.renderDemo(w, r, "examples/todo")
		return
	}
```

- [ ] **Step 3: Add `registerTodoRoutes` stub**

Create `internal/server/todo_handler.go` with just the route registration returning 200 empty, so the build compiles (filled in Task 10):

```go
// internal/server/todo_handler.go
package server

import "net/http"

// registerTodoRoutes wires the /api/examples/todo/* endpoints.
func (s *Server) registerTodoRoutes() {
	// filled in Task 10
	_ = http.MethodPost
}
```

- [ ] **Step 4: Add the sidebar section**

In `internal/pages/demo/layout.templ`, inside `getSidebarSections`'s returned slice, add a new section (place it first so Examples sits at the top):

```go
		{
			Title:       "Examples",
			Collapsible: true,
			Items: []sidebar.Item{
				sItem("examples", "Overview", "/examples", activeComponent),
				sItem("todo", "Todo List", "/examples/todo", activeComponent),
			},
		},
```

- [ ] **Step 5: Generate, build, run, eyeball**

```bash
templ generate && go build -o bin/server ./cmd/server
```
Expected: clean build. Then `go run cmd/server/main.go`, open `http://localhost:8090/examples/todo`, confirm the app renders, add a task in the browser, and confirm the sidebar "Examples" section is present.

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/components/registry.go internal/server/server.go internal/server/todo_handler.go internal/pages/demo/layout.templ internal/pages/demo/layout_templ.go
git commit -m "feat(examples): register /examples routes, registry entries, sidebar section"
```

---

## Task 10: Todo API handlers + cookie-backed first-load render

**Files:**
- Modify: `internal/server/todo_handler.go`
- Modify: `internal/server/server.go` (swap the temporary todo render for `renderTodoPage`)

- [ ] **Step 1: Implement the handlers**

Replace `internal/server/todo_handler.go` with:

```go
// internal/server/todo_handler.go
package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/components/toast"
	todod "github.com/araihu/goshtoso/internal/examples/todo"
	"github.com/araihu/goshtoso/internal/pages/demo"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// registerTodoRoutes wires the /api/examples/todo/* endpoints.
func (s *Server) registerTodoRoutes() {
	s.mux.HandleFunc("/api/examples/todo/add", s.todoAdd)
	s.mux.HandleFunc("/api/examples/todo/toggle", s.todoToggle)
	s.mux.HandleFunc("/api/examples/todo/delete", s.todoDelete)
	s.mux.HandleFunc("/api/examples/todo/edit", s.todoEdit)
	s.mux.HandleFunc("/api/examples/todo/filter", s.todoFilter)
	s.mux.HandleFunc("/api/examples/todo/move", s.todoMove)
	s.mux.HandleFunc("/api/examples/todo/clear-completed", s.todoClearCompleted)
	s.mux.HandleFunc("/api/examples/todo/reorder", s.todoReorder)
}

func (s *Server) renderTodoPage(w http.ResponseWriter, r *http.Request) {
	state := todod.FromRequest(r)
	if state.Filter == "" {
		state.Filter = "all"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.TodoApp(state)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.Fragment("Todo List", "todo", content).Render(r.Context(), w)
		return
	}
	_ = demo.Layout("Todo List", "todo", content).Render(r.Context(), w)
}

// idParam reads the ?id= query as an int (0 if absent/invalid).
func idParam(r *http.Request) int {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	return id
}

// onlyPost guards a handler to POST.
func onlyPost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func (s *Server) todoAdd(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	st.Add(r.FormValue("title"), r.FormValue("priority"), r.FormValue("due"))
	todod.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	added := st.Todos[len(st.Todos)-1]
	_ = examples.TodoRow(added).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
	_ = toast.OOBToast(toast.Config{Variant: toast.Success, Title: "Added", Message: added.Title}).Render(r.Context(), w)
}

func (s *Server) todoToggle(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	id := idParam(r)
	st.Toggle(id)
	todod.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if i := indexOf(st, id); i >= 0 {
		_ = examples.TodoRow(st.Todos[i]).Render(r.Context(), w)
	}
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
}

func (s *Server) todoDelete(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	st.Delete(idParam(r))
	todod.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Empty body removes the targeted row (hx-swap outerHTML); update count + toast OOB.
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
	_ = toast.OOBToast(toast.Config{Variant: toast.Info, Title: "Deleted", Message: "Task removed."}).Render(r.Context(), w)
}

func (s *Server) todoEdit(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	id := idParam(r)
	st.Edit(id, r.FormValue("title"), r.FormValue("priority"), r.FormValue("due"))
	todod.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if i := indexOf(st, id); i >= 0 {
		_ = examples.TodoRow(st.Todos[i]).Render(r.Context(), w)
	}
}

func (s *Server) todoFilter(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	st.SetFilter(r.URL.Query().Get("f"))
	todod.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.TodoList(st).Render(r.Context(), w)
}

func (s *Server) todoMove(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	moveByButton(&st, idParam(r), r.URL.Query().Get("dir"))
	todod.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.TodoList(st).Render(r.Context(), w)
}

func (s *Server) todoClearCompleted(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	st.ClearCompleted()
	todod.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.TodoList(st).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
	_ = toast.OOBToast(toast.Config{Variant: toast.Info, Title: "Cleared", Message: "Completed tasks removed."}).Render(r.Context(), w)
}

func (s *Server) todoReorder(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	st := todod.FromRequest(r)
	ids := parseIDs(r.FormValue("ids"))
	st.Reorder(ids)
	todod.SetCookie(w, st)
	// UI already moved optimistically; just confirm + refresh count OOB.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
}

// indexOf returns the slice index of id in st.Todos, or -1.
func indexOf(st todod.State, id int) int {
	for i, t := range st.Todos {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// parseIDs parses "3,1,2" into []int, skipping blanks/invalid tokens.
func parseIDs(csv string) []int {
	var out []int
	for _, tok := range strings.Split(csv, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if n, err := strconv.Atoi(tok); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// moveByButton reorders st by moving id one slot up or down within the current
// visible (Order-sorted) sequence, then persists the new order via Reorder.
func moveByButton(st *todod.State, id int, dir string) {
	vis := st.Visible()
	pos := -1
	for i, t := range vis {
		if t.ID == id {
			pos = i
			break
		}
	}
	if pos < 0 {
		return
	}
	swap := pos - 1
	if dir == "down" {
		swap = pos + 1
	}
	if swap < 0 || swap >= len(vis) {
		return
	}
	vis[pos], vis[swap] = vis[swap], vis[pos]
	ids := make([]int, len(vis))
	for i, t := range vis {
		ids[i] = t.ID
	}
	st.Reorder(ids)
}
```

- [ ] **Step 2: Swap the temporary todo render**

In `internal/server/server.go` `handleExample`, replace the temporary line from Task 9:

```go
	if name == "todo" {
		s.renderDemo(w, r, "examples/todo")   // <-- remove this
		return
	}
```
with:
```go
	if name == "todo" {
		s.renderTodoPage(w, r)
		return
	}
```

- [ ] **Step 3: Confirm `demo.Fragment` signature**

Run: `grep -n "func Fragment\|templ Fragment" internal/pages/demo/fragment.templ`
Expected: `Fragment(title, active string, content templ.Component)` (3 args). If the signature differs, match the existing `renderDemo` call in `server.go`.

- [ ] **Step 4: Build + manual smoke**

```bash
templ generate && go build -o bin/server ./cmd/server
```
Then `go run cmd/server/main.go`; at `http://localhost:8090/examples/todo`: add task (toast appears, count updates, row appended), toggle (strike-through), ↑/↓ reorder, drag reorder, delete, filter All/Active/Done, clear completed. Reload page — todos persist (cookie).

- [ ] **Step 5: Commit**

```bash
git add internal/server/todo_handler.go internal/server/server.go
git commit -m "feat(examples): todo API handlers + cookie-backed first load"
```

---

## Task 11: E2E tests

**Files:**
- Create: `tests/e2e/todo_example_test.go`

- [ ] **Step 1: Write the E2E test**

```go
// tests/e2e/todo_example_test.go
package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// addTodo fills the add form and submits, waiting for the row to appear.
func addTodo(t *testing.T, page playwright.Page, title string) {
	t.Helper()
	require.NoError(t, page.Locator("input[name='title']").Fill(title))
	require.NoError(t, page.Locator("#todo-fragment form button[type='submit']").Click())
	_, err := page.WaitForFunction(
		fmt.Sprintf("() => Array.from(document.querySelectorAll('#todo-list li span')).some(s => s.textContent.trim() === %q)", title),
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

func gotoTodo(t *testing.T, page playwright.Page) {
	_, err := page.Goto(baseURL + "/examples/todo")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
}

func TestTodoExample_AddAndCount(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoTodo(t, page)

	addTodo(t, page, "write tests")
	count, err := page.Locator("#todo-list > li").Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)

	badge, err := page.Locator("#todo-count").TextContent()
	require.NoError(t, err)
	require.Contains(t, badge, "1 active")
}

func TestTodoExample_ToggleDelete(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoTodo(t, page)

	addTodo(t, page, "toggle me")
	row := page.Locator("#todo-list > li").First()
	checkbox := row.Locator("input[type='checkbox']")

	// Toggle via clickUntil (HTMX outerHTML swap rebind race).
	clickUntil(t, page, checkbox,
		"() => { const s = document.querySelector('#todo-list li span'); return s && s.className.includes('line-through'); }")

	// Delete the (re-queried) row.
	clickUntil(t, page, page.Locator("#todo-list > li button[aria-label='Delete']").First(),
		"() => document.querySelectorAll('#todo-list > li').length === 0 || document.querySelector('#todo-list li')?.textContent.includes('Nothing here yet')")
}

func TestTodoExample_Filters(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoTodo(t, page)

	addTodo(t, page, "active one")
	addTodo(t, page, "done one")

	// Mark the second one done.
	clickUntil(t, page, page.Locator("#todo-list > li").Nth(1).Locator("input[type='checkbox']"),
		"() => document.querySelectorAll('#todo-list li span.line-through').length === 1")

	// Active filter → only the not-done todo.
	clickUntil(t, page, page.Locator("#todo-fragment button:has-text('Active')"),
		"() => document.querySelectorAll('#todo-list > li').length === 1")

	// Done filter → only the done todo.
	clickUntil(t, page, page.Locator("#todo-fragment button:has-text('Done')"),
		"() => { const li = document.querySelectorAll('#todo-list > li'); return li.length === 1 && li[0].querySelector('span.line-through'); }")
}

func TestTodoExample_ReorderButtons(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoTodo(t, page)

	addTodo(t, page, "first")
	addTodo(t, page, "second")

	// Move "second" (row index 1) up; deterministic HTMX path.
	clickUntil(t, page, page.Locator("#todo-list > li").Nth(1).Locator("button[aria-label='Move up']"),
		"() => document.querySelector('#todo-list > li span')?.textContent.trim() === 'second'")
}

func TestTodoExample_ClearCompleted(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoTodo(t, page)

	addTodo(t, page, "keep")
	addTodo(t, page, "remove")
	clickUntil(t, page, page.Locator("#todo-list > li").Nth(1).Locator("input[type='checkbox']"),
		"() => document.querySelectorAll('#todo-list li span.line-through').length === 1")

	clickUntil(t, page, page.Locator("#todo-fragment button:has-text('Clear completed')"),
		"() => document.querySelectorAll('#todo-list > li').length === 1")
}

func TestTodoExample_SidebarPresent(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL + "/examples/todo")
	require.NoError(t, err)
	require.NoError(t, page.Locator("text=Examples").First().WaitFor())
	require.NoError(t, page.Locator("a[href='/examples/todo']").First().WaitFor())
}
```

> Note: each test must start with a fresh page (fresh cookie). `newPage` creates a new browser context tab; if cookies leak between todo tests in the shared browser, switch to `browser.NewContext()` per test (see how other stateful tests isolate — `grep -n "NewContext" tests/e2e/*.go`). If isolation is needed, replace `newPage(t, browser)` with a context + page and clear cookies via `context.ClearCookies()`.

- [ ] **Step 2: Run the todo E2E subset**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestTodoExample`
Expected: all PASS. Debug with `takeScreenshot` if a selector misses; fix selectors/JS conditions to match the actual rendered HTML (verify class names like `line-through` exist on the rendered span).

- [ ] **Step 3: Commit**

```bash
git add tests/e2e/todo_example_test.go
git commit -m "test(examples): e2e coverage for todo app"
```

---

## Task 12: Tailwind rebuild, full suite, lint

**Files:**
- Modify: `assets/styles.css` (generated)

- [ ] **Step 1: Rebuild Tailwind**

Any new utility classes used (`accent-primary`, etc.) must be in the embedded CSS.

Run: `tailwindcss -i css/main.css -o assets/styles.css`
Expected: regenerated CSS.

- [ ] **Step 2: Full build + lint**

```bash
templ generate && go build -o bin/server ./cmd/server && golangci-lint run
```
Expected: clean build, lint passes (keep handlers under cyclomatic complexity 20 — they are small; `moveByButton` is the densest, verify it's under 20 or extract).

- [ ] **Step 3: Run the full E2E suite**

Run: `go test ./tests/e2e/... -count=1 -timeout 15m`
Expected: all pass, including the existing 381 tests + new todo tests, no skips. If a new flake appears under full-suite pressure, it is almost certainly a post-swap click missing `clickUntil` — fix per CLAUDE.md.

- [ ] **Step 4: Run domain unit tests once more**

Run: `go test ./internal/examples/todo/... -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add assets/styles.css
git commit -m "chore(examples): rebuild Tailwind for todo utilities"
```

---

## Task 13: Update CLAUDE.md status

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Document the examples family**

Add a short "Example Apps" subsection under Repository Structure / Current Status noting: `/examples` route family, cookie-backed stateless pattern (`internal/examples/todo`), how to add a new example (domain pkg + `examples` templ + registry entry + sidebar item), and that the todo app is the reference example.

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: document /examples app family and todo reference"
```

---

## Self-Review notes (addressed)

- **Spec coverage:** routes/index/sidebar (T9), cookie state model + guards (T1–T5), all 7 endpoints + OOB toast/count (T10), reorder native-DnD + buttons (T7/T10/T11), unit + E2E testing (T1–T5, T11–T12), components showcased (T6–T7). All spec sections map to a task.
- **Deviations** (Tabs→button group, Checkbox→styled input) documented at the top with rationale.
- **Verification gates** flagged inline for field names that must be confirmed against the actual component Configs (`badge.Text`, `select.Option`, `card.Card` children, `demo.Fragment` arity) — these are read-and-confirm steps, not placeholders.
- **Type consistency:** `State`/`Todo`, method names (`Add/Toggle/Delete/Edit/Reorder/ClearCompleted/SetFilter/Visible/ActiveCount/DoneCount`), partial names (`TodoRow/TodoList/CountBadge/TodoApp/TodoContent/IndexContent`), and endpoint paths are consistent across tasks. `/move` endpoint (button reorder) added to handlers + row markup + tests consistently.
