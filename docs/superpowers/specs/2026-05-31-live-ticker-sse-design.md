# Live Ticker (SSE) — Example App Design

**Date:** 2026-05-31
**Branch:** `example-live-ticker-sse`
**Status:** Approved design, pre-implementation

## Summary

A new Goshtoso example app at `/examples/ticker`: a live stock-market
**watchlist board** driven by **Server-Sent Events**. It is the second entry in
the `/examples/*` family (after Todo List) and the first to use a server-push
transport instead of request/response HTMX.

Prices come from a **server-side simulated ticker** (seeded random walk) — no
external market API. This keeps the example deterministic, offline, and
CDN-free, matching the repo's hard E2E constraints.

The board showcases ~5 components in real use: **Table** (the live board),
**Badge** (up/down change indicator), **Card** (spotlight for the selected
symbol), **Spinner** (connection state), **Toggle** (pause/resume).

## Goals & Non-Goals

**Goals**
- Demonstrate SSE + htmx in a Goshtoso example, end to end.
- Deterministic, offline, no-CDN — runnable and E2E-testable like every other example.
- Stay htmx-native: server pushes rendered HTML, not JSON.
- Reuse the established example layering (domain → templ → handler → registry).

**Non-Goals (explicit YAGNI cuts)**
- No real market API.
- No cookie-backed personal watchlist (that was the rejected "Trading desk" scope).
- No price-alert toasts, no sparklines, no charting.
- No per-user server state — one shared simulated stream for all viewers.

## Decisions (locked during brainstorming)

| Question | Decision |
|----------|----------|
| Data source | Server-side simulated ticker, seeded random walk (deterministic) |
| SSE transport | Bundle the htmx SSE extension (`sse.js`); server pushes HTML |
| Scope | Watchlist board (Table + spotlight Card + Spinner + pause Toggle) |

## Architecture

Three layers, mirroring the Todo example.

### 1. Domain — `internal/examples/ticker/` (HTTP-free, unit-tested)

- **`types.go`** — `Symbol{Ticker, Name, Price, PrevPrice}`, `Snapshot`
  (a point-in-time list of symbols), `ChangePct()` / direction helper.
- **`sim.go`** — `Simulator`: holds `[]Symbol`; `Tick()` advances each price by a
  seeded random walk (`math/rand` with a fixed seed → reproducible sequence).
  Pure and deterministic; unit-tested for monotonic seed behavior and
  change-percent math.
- **`broker.go`** — pub/sub hub goroutine. `Subscribe() (<-chan Snapshot, cancel)`
  / internal unsubscribe. Ticks on a **configurable interval**; on each tick it
  advances the simulator and broadcasts the snapshot to all subscribers.
  Per-subscriber buffered channel; slow/closed subscribers are dropped, never
  block the hub. **One shared stream for all clients → no per-user state**
  (still "stateless per user").

### 2. Templ — `internal/pages/demo/examples/ticker.templ`

- **`TickerContent`** — page shell: header, the live `Table`, the spotlight
  `Card`, a connection `Spinner`, and a pause/resume `Toggle`. Follows the
  example/demo registry render path so theme, dark mode, and fragment-swap nav
  come for free.
- **Fragment helpers** rendered server-side and reused by the SSE handler:
  - per-symbol price/change cell (the unit pushed on each tick),
  - spotlight card body for a given symbol.

### 3. Handler — `internal/server/ticker_handler.go`

Registered via a `registerTickerRoutes()` method on `Server`, same pattern as
`registerTodoRoutes()` (`s.mux.HandleFunc`).

- **`GET /api/examples/ticker/stream`** — SSE endpoint.
  - Sets `Content-Type: text/event-stream`, `Cache-Control: no-cache`,
    `Connection: keep-alive`; uses `http.Flusher` (Go 1.26, net/http ServeMux).
  - Subscribes to the broker. On each tick, emits **one named SSE event per
    symbol** — event name = ticker symbol:
    ```
    event: AAPL
    data: <td>…rendered cell html…</td>

    ```
  - Loops until `r.Context().Done()` (client disconnect), then unsubscribes.
- **`GET /api/examples/ticker/spotlight?symbol=X`** — normal htmx swap. Returns a
  Card wrapper bound to symbol `X`'s SSE event.

### Registry & sidebar

- Add `"examples/ticker"` registry entry → `examples.TickerContent`.
- Add a sidebar item under the **Examples** section.

### Vendor / assets

- Add the htmx SSE extension (`htmx-ext-sse`, `sse.js`) to `assets/js/vendor/`.
- Wire it into `assets/embed.go`, the layout `<script>` includes, and the
  component head `Dependencies` so it loads with core htmx (no CDN at runtime).

## SSE wiring (htmx-native, no per-client server state)

The key trick: the stream is **shared** (same events for everyone) but
**selection is per-client**, achieved purely through which event each element
subscribes to.

- **Connection root** (tbody wrapper): `hx-ext="sse"
  sse-connect="/api/examples/ticker/stream"`.
- **Each table cell**: `sse-swap="AAPL"` (etc.) → that cell live-updates from its
  symbol's event, `hx-swap="innerHTML"`.
- **Spotlight Card wrapper**: `sse-swap="<selected>"`. Clicking a table row fires
  `hx-get="/api/examples/ticker/spotlight?symbol=X"` which swaps in a card whose
  `sse-swap` is now symbol `X` — it then live-updates off the **same shared
  stream**. No server-side notion of "who selected what."

### Pause / resume

A small Alpine wrapper toggles the SSE connection by adding/removing
`sse-connect` (closing/reopening the underlying `EventSource`). **This is the one
fiddly bit** — flagged for careful validation during implementation; if the
add/remove-attribute approach proves unreliable with the htmx ext, fall back to
toggling `hx-ext`/re-processing the element via the htmx JS API.

## Data flow

1. Server boot → one `Simulator` + `broker` goroutine starts ticking.
2. Browser loads `/examples/ticker` → htmx opens the SSE connection.
3. Each tick → broker broadcasts snapshot → handler writes one event per symbol
   → htmx swaps each subscribed cell. Badge color reflects direction.
4. User clicks a row → htmx GET spotlight → card rebinds to that symbol's event →
   card live-updates from the same stream.
5. User toggles pause → Alpine closes the EventSource; resume reopens it.

## Error handling

- **Null/empty guards:** snapshot always a non-nil slice; no JSON-in-attribute
  paths (HTML fragments only), avoiding the templ-escaping class of bugs.
- **Client disconnect:** handler watches `r.Context().Done()`, unsubscribes,
  returns — no goroutine leak.
- **Slow subscriber:** buffered per-subscriber channel; hub drops rather than
  blocks.
- **Fragment-nav OOB gotcha:** per CLAUDE.md, any element carrying an OOB-style
  attribute on first paint must gate it to update-only; register any
  `Alpine.data()` immediately (not only on `alpine:init`) so it is defined on
  fragment nav.

## Testing

Unit (domain, fast, deterministic):
- `sim_test.go` — seeded `Tick()` reproducibility; `ChangePct`/direction math.
- broker — subscribe/unsubscribe, broadcast fan-out, slow-subscriber drop.

E2E (`tests/e2e/ticker_test.go`) — seeded sim + short test tick interval:
- SSE connects and a known cell mutates after a tick.
- Row click rebinds the spotlight card to the selected symbol.
- Pause stops updates; resume restarts them.
- **Sidebar fragment-nav path** loads the example with **no console errors**
  (not just the direct URL load) — per CLAUDE.md examples rule.
- Assert direction / Badge **color classes**, not exact float values, for
  robustness under a streaming source.

## Implementation order (rough)

1. Domain pkg (`types`, `sim`, `broker`) + unit tests.
2. Bundle `sse.js` into vendor + embed + layout wiring.
3. `ticker.templ` (shell + fragment helpers) + `templ generate`.
4. `ticker_handler.go` (stream + spotlight) + `registerTickerRoutes`.
5. Registry entry + sidebar item.
6. `go run ./scripts/skillgen` if any component entry points changed (none expected).
7. E2E tests; `templ generate && go build`; full suite.

## Open risk

- **Pause/resume reliability** with the htmx SSE ext (noted above) — the only
  part not directly modeled on an existing example.
