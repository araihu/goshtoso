# Chat Example — Design

**Date:** 2026-05-31
**Status:** Approved (brainstorm), pending implementation plan
**Topic:** A realtime, full-duplex chat example app for Goshtoso (`/examples/chat`)

## Goal

Add a second example app (after Todo List) that demonstrates **full-duplex
realtime over htmx + websockets** while showcasing as many Goshtoso components in
real use as fits naturally. The transport is the novel bit; the components are
the point.

Mode (chosen): **realtime + ephemeral**. Messages are genuinely broadcast to all
connected clients, but the log lives only in server RAM (lost on restart). No
persistence, no auth, no accounts.

## Transport

- **Client:** htmx `ws` extension. The chat container carries
  `hx-ext="ws" ws-connect="/api/examples/chat/ws"`; the composer is a
  `<form ws-send>`. No bespoke client JS beyond the extension.
- **Up:** htmx sends a JSON frame `{message, HEADERS}` on submit.
- **Down:** server pushes **server-rendered templ HTML**. Each frame is an OOB
  swap (`hx-swap-oob="beforeend:#chat-log"`) so every client appends the same
  bubble. Goshtoso components remain the single source of rendering — bubbles are
  never client-built.
- **Server lib:** `github.com/coder/websocket` (new direct dep; only stdlib
  transitive deps). `Accept` → per-conn read pump + buffered write pump.
- **Vendored asset:** `assets/js/vendor/htmx-ext-ws.js` (no CDN, matching the
  Alpine/HTMX bundling rule). Loaded only on the chat page, not globally.

## State — RAM-only hub

Lives in `internal/examples/chat/` (pure, HTTP-free, unit-tested).

- **`Hub`** — owns the set of registered clients and three channels
  (`register`, `unregister`, `broadcast`). A single owner goroutine drains them,
  so the client set needs no mutex (idiomatic Go hub).
- **Ring buffer** — last **N = 50** messages. On connect, the server replays the
  backlog to the new socket *before* the client joins broadcast, so a fresh tab
  sees history. Cleared on restart — ephemeral by design.
- **`client`** — `conn` + `send chan []byte`; a read pump (up) and write pump
  (down). A slow or dead client (full send channel) is dropped and unregistered.
- **Limits** — server cap on total connections; per-message rune cap (mirrors
  todo's `MaxTitleLen`); read pump bounds inbound frame size.

## ELIZA bot

- Pure function `Reply(msg string) (reply string, matched bool)` in the domain
  package. Deterministic keyword → canned reply (ELIZA-style, ~40 lines, no
  deps, no AI).
- A matched bot reply is just another broadcast (rendered bubble → all clients →
  ring).
- **Per-connection toggle**: each client's own Toggle controls whether *their*
  tab requests bot replies. No shared/global bot state.

## Message flow (full-duplex)

1. Tab A `ws-send` → `{message, HEADERS}` JSON up.
2. Hub renders `MessageBubble` templ → broadcasts to **all** clients (including
   A) → appends to ring.
3. If the sender has ELIZA enabled, Hub renders the bot bubble → broadcasts it
   too.
4. Tabs A and B both see both bubbles live.

## Identity

- `gt_chat` cookie, base64url-JSON: `{nick, color}`. Reuses the todo cookie
  layering pattern.
- Auto-assigned `Guest-xxxx` + hashed color on first hit.
- Renamable via a Goshtoso **Text Input** that rewrites the cookie; the new nick
  labels subsequent messages. Persists across reload and fragment-nav.
- Read on ws connect to label outgoing messages.

## Goshtoso components used

| Component | Job in chat |
|-----------|-------------|
| **Card** | Chat panel shell (header + log + composer) |
| **Avatar** | Per-message sender avatar (initials + hashed color); bot has a distinct one |
| **Textarea** | Message composer (multi-line, Enter-to-send) inside `<form ws-send>` |
| **Button** | Send button; Rename submit |
| **Text Input** | Nickname rename field (writes `gt_chat` cookie) |
| **Badge** | Online count (`3 online`), `BOT` tag on ELIZA bubbles, `you` tag on own |
| **Toggle** | ELIZA bot on/off (per-connection) |
| **Banner** | Connection status — shown on ws disconnect ("Reconnecting…") |
| **Toast** | "X joined" / "X left" presence pings (OOB), auto-dismiss |
| **Spinner** | "ELIZA is typing…" indicator during the brief bot delay |
| **Tooltip** | Hover a bubble → full timestamp |

11 components, none forced.

## Files

```
internal/examples/chat/
  hub.go            # Hub, client, run loop, ring buffer (HTTP-free)
  hub_test.go       # register / broadcast / backlog replay / slow-client drop
  eliza.go          # Reply(msg) (string, bool) — keyword bot
  eliza_test.go     # deterministic input -> reply table
  identity.go       # gt_chat cookie encode/decode, Guest-xxxx + color
  identity_test.go
internal/pages/demo/examples/
  chat.templ        # ChatApp shell + MessageBubble + OnlineBadge + StatusBanner
                    # (exported; rendered by both the page shell AND ws push)
internal/server/
  chat_handler.go   # GET /examples/chat (page + fragment) ; WS /api/examples/chat/ws ; POST .../rename
assets/js/vendor/
  htmx-ext-ws.js    # vendored, loaded on the chat page only
tests/e2e/
  chat_test.go
```

Plus: Demos registry entry, sidebar "Examples" item, `/examples` gallery card.

## Edge cases (CLAUDE.md gotchas)

- **Fragment-nav + ws.** Arriving via sidebar fragment swap, htmx must `process`
  the new `ws-connect` node and the extension must be loaded. Verify connect
  fires on fragment-nav, not only direct load (todo OOB/Alpine gotcha, ws flavor).
- **OOB-on-first-paint.** Online badge and status banner must NOT carry
  `hx-swap-oob` on first paint (would throw `oobErrorNoTarget` on fragment-nav).
  Use the same `oob bool` gate as `CountBadge`/`ClearButton`.
- **Templ escaping.** No JSON inside Alpine attributes. ELIZA and nick text
  render as templ text nodes (auto-escaped); never injected into `x-data`.
- **Socket lifecycle.** Write pump uses ctx + bounded send chan; a full channel
  drops the client. Read pump bounds message size.
- **Asset loading.** `htmx-ext-ws.js` script tag is page-scoped (head
  Dependencies), not global.

## E2E (`chat_test.go`)

1. **Two-tab broadcast** — 2 Playwright pages; type in A, assert the bubble
   appears in **B**. The real full-duplex proof.
2. **ELIZA determinism** — toggle bot on, send a known input, assert the known
   reply bubble.
3. **Presence** — open a 2nd tab → online badge `2 online`; close → back to `1`.
4. **Rename** — set nick, send, assert the bubble is labeled with the new nick;
   persists across reload (cookie).
5. **Fragment-nav** — reach `/examples/chat` via sidebar click; assert ws
   connects and no console errors (the load-bearing gotcha test).
6. Reuse `clickUntil` where a swap rebinds a control.

## Out of scope (YAGNI)

- Persistence / database / message history beyond the 50-message ring.
- Auth, accounts, private rooms, DMs.
- Multiple chat rooms/channels.
- Drag-and-drop, file/image upload, reactions, edit/delete of sent messages.
- Real AI bot (ELIZA is deterministic and dependency-free on purpose).
