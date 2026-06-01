# Live Log Feed — example app design

**Date:** 2026-05-31
**Status:** Approved (brainstorm)
**Branch:** `worktree-example-live-log-feed`

## Goal

Add a second example app to Goshtoso (`/examples/*`): a **live log feed** that
streams synthetic log lines to the browser over **Server-Sent Events (SSE)**,
rendered with Goshtoso components. It is the first SSE / server-push example in
the repo and exercises a component combination the todo example does not.

It must satisfy the existing example-app contract (CLAUDE.md → "Example Apps"):
flow through the demo registry + `renderDemo`, get an Examples sidebar entry,
and stay stateless.

## Non-goals (YAGNI)

- No persistence (no cookie, no DB), no real log source, no shared event bus.
- No export / download / free-text search / per-row expand.
- No pause-buffering: paused = disconnected, so paused viewers miss events
  (realistic, and keeps the lifecycle honest).
- Does not touch skillgen (that is for `components/**`, not examples).

## Identity & routing

- Name: **"Live Log Feed"**, route `/examples/logs`, slug `logs`.
- Sidebar: second entry in the **Examples** section, after "Todo List"
  (`internal/pages/demo/layout.templ` → the Examples `sItem` block).
- Registry: add `"examples/logs": {"Live Log Feed", "logs", examples.LogsContent}`
  to `internal/pages/demo/components/registry.go`.
- Routing: `handleExample` (`internal/server/server.go`) gains
  `case "logs": s.renderLogsPage(w, r)`. First load picks full `Layout` vs HTMX
  `Fragment` on the `HX-Request` header, exactly like `renderTodoPage`.

## Stateless model

- **No cookie, no shared bus.** The only state is the live SSE connection: one
  goroutine per client, a synthetic generator on a `time.Ticker`, which exits
  on `r.Context().Done()`. This is ephemeral *connection* state, not server
  memory in the cookie sense — the stateless ethos holds. Each viewer gets an
  independent stream.

## Layering (mirrors the todo example)

1. **Domain — `internal/examples/logs/`** (HTTP-free, unit-tested)
   - `LogLine{ Time time.Time, Level Level, Source string, Message string }`.
   - `Level` enum: `Debug | Info | Warn | Error` with a severity ordering
     (`Debug < Info < Warn < Error`) used by the filter.
   - `Line(i int) LogLine` — deterministic generator cycling a fixed sample
     pool (same spirit as `todo.Sample()`), so E2E output is predictable. `Time`
     is supplied by the caller (handler passes the tick time) to keep the domain
     pure and testable.
   - Unit tests cover the sample-pool cycling and the severity ordering used by
     the filter.

2. **View — `internal/pages/demo/examples/logs.templ`** (exported templ)
   - `LogsContent` — registry entry point (control bar + feed panel; no page
     chrome, matches `TodoContent`).
   - `LogsApp(...)` — the example body rendered by both the page shell and used
     by the first-load handler.
   - `LogRow(LogLine)` — the per-event fragment streamed by SSE. Renders the
     level via `badge.Badge`; the row carries a `log-level-<level>` class for
     the CSS filter and the row text (time · source · message) is hand-rolled
     (no component exists for a log line — same call as todo's raw rows).
   - `LogFeed` — the `#log-feed` container.

3. **Handlers — `internal/server/logs_handler.go`** (thin)
   - `registerLogsRoutes()` wiring, called from `registerRoutes`.
   - `renderLogsPage` — first load (Layout vs Fragment).
   - `/api/examples/logs/stream` — the SSE endpoint (see below).

## Transport — htmx SSE extension

- **Vendor** `assets/js/vendor/htmx-ext-sse.min.js` — the `htmx-ext-sse` package
  at **2.2.x** (compatible with htmx 2.0.8). Record source + version the same
  way the other vendored files are noted.
- Wire a `<script>` tag **after** `htmx.min.js` in `internal/pages/demo/layout.templ`,
  and add it to the embedded-assets comment list in `assets/embed.go`.
- Markup:
  ```html
  <div hx-ext="sse" sse-connect="/api/examples/logs/stream">
    <div id="log-feed" sse-swap="message" hx-swap="beforeend"></div>
  </div>
  ```
  Each SSE `message` event carries one server-rendered `LogRow`, appended to
  `#log-feed`.

### SSE endpoint contract

- Headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`,
  `Connection: keep-alive`.
- Loop on a `time.Ticker`; each tick: render `LogRow(logs.Line(i))` to a buffer,
  emit it as an SSE `message` event, then `flusher.Flush()`. Return on
  `r.Context().Done()` (client disconnect / pause).
- **Framing gotcha:** templ output is multi-line, but a bare `data:` field is a
  single line. Render to a buffer, then write **each output line as its own
  `data: ` field**, followed by the blank-line event terminator. (htmx's SSE
  ext rejoins multiple `data:` lines with `\n` before swapping — harmless
  between HTML tags.)
- **Test affordances:** the endpoint reads `?interval=<dur>` (default e.g.
  500ms) and `?max=<n>` (default unbounded; when set, close after N events) so
  E2E runs fast and bounded.
- Guard against a missing `http.Flusher` (`w.(http.Flusher)`); if unavailable,
  return 500 — never silently buffer.

## Division of labor (the one novel integration)

htmx owns **insertion**; Alpine owns **everything else**. This is the only part
of the design not previously exercised in this repo, so the riskier pieces are
deliberately kept out of per-row Alpine.

- **Alpine registration:** via `<script>` + `Alpine.data('logFeed', …)` (the
  templ-escaping-safe pattern). Register **immediately if Alpine is already
  running**, not only inside `alpine:init`, or it is undefined on sidebar
  fragment-nav (the documented examples gotcha).
- **Cap (last 100 rows):** on `htmx:afterSwap` (scoped to `#log-feed`), Alpine
  trims oldest child nodes when the count exceeds 100. Bounds the DOM and keeps
  E2E row-count assertions stable under a fast ticker.
- **Filter:** **pure CSS**, driven by a `select.Select`. The feed wrapper
  carries a class reflecting the chosen minimum severity (`flt-all` / `flt-warn`
  / `flt-error`); rows carry `log-level-<level>`; CSS hides rows below the
  threshold (e.g. `.flt-error :is(.log-level-debug,.log-level-info,.log-level-warn){display:none}`).
  The Select lives **outside** `#log-feed` (so it survives every swap) and sets
  the wrapper class via `x-model` + a watcher. No per-row Alpine → robust against
  htmx insertions.
- **Auto-scroll:** a `toggle.Toggle` bound to Alpine `autoScroll`; when a row is
  appended and `autoScroll` is on, scroll `#log-feed` to the bottom.
- **Pause / Resume:** a `button.Button` toggles an Alpine flag that
  adds/removes the `sse-connect` element via `<template x-if>`. Removing it
  makes the ext **close** the EventSource → the server goroutine's context
  cancels and it exits (no wasted work). Resume re-adds → reconnect.
- **Clear:** a `button.Button` empties `#log-feed` (Alpine clears children).
- **Connection status:** a `badge.Badge` + `spinner.Spinner`, **outside** the
  feed, driven by ext lifecycle events: `htmx:sseOpen` → "Connected" (success
  Badge), `htmx:sseError` → "Reconnecting" (warning Badge + Spinner), paused →
  "Paused" (secondary Badge).

### OOB / fragment-nav guard

Any element that would carry `hx-swap-oob` on first paint must gate the
attribute to update-only (the `oob bool` pattern from todo's `CountBadge`),
to avoid `htmx:oobErrorNoTarget` when the page arrives via sidebar fragment
swap. (The current design has no first-paint OOB element, but keep this in mind
if a count/status badge is later wired via OOB.)

## Components showcased

| Element | Component |
|---|---|
| Feed panel + control-bar shell | **Card** |
| Per-row level (DEBUG/INFO/WARN/ERROR) | **Badge** |
| Connection state | **Badge** + **Spinner** |
| Min-level filter (`All / Warn+ / Error`) | **Select** |
| Auto-scroll | **Toggle** |
| Pause / Resume | **Button** |
| Clear | **Button** |
| Control-button labels | **Tooltip** |

7 distinct components (Card, Badge, Spinner, Select, Toggle, Button, Tooltip).
The filter uses **Select** specifically because todo already demonstrates a
segmented *radio* — this surfaces a different component. The only hand-rolled
markup is the log-row text layout (no component fits a log line).

## E2E — `tests/e2e/logs_test.go`

Use `?interval=50ms` (and `?max` where a bounded run helps) for speed; use
`WaitForFunction` for async status; use `clickUntil` for any click that triggers
an htmx swap which replaces the clicked control.

- **Direct load:** stream connects; rows appear within timeout; level Badges
  render.
- **Sidebar fragment-nav path:** navigate to the example via the sidebar (not a
  direct URL); assert the stream connects **and there are no console errors**
  (covers the Alpine-data-on-fragment-nav and OOB gotchas). This is mandatory
  per the example-app contract.
- **Pause/Resume:** Pause flips status to "Paused" and stops new rows; Resume
  reconnects and rows resume.
- **Filter:** selecting `Error` hides lower-severity rows; `All` shows them.
- **Clear:** empties the feed.
- **Cap:** under a fast ticker, the feed never exceeds 100 rows.

## Documentation updates

- `CLAUDE.md`:
  - "Example Apps" — note the SSE pattern + the htmx-ext-sse vendored dep + the
    htmx-owns-insert / Alpine-owns-the-rest division.
  - "Current Status → Example apps" — add "Live Log Feed (`/examples/logs`)".
  - Vendor list (`assets/js/vendor/`) — add `htmx-ext-sse.min.js`.
- `assets/embed.go` comment list — add the new vendored file.

## Build / verify checklist

- `templ generate` after editing `.templ` files.
- `tailwindcss -i css/main.css -o assets/styles.css` if any new utility class is
  introduced (CSS is embedded).
- `go build -o bin/server ./cmd/server`.
- `golangci-lint run` clean (keep new funcs under the cyclomatic ceiling of 20 —
  extract helpers; the SSE handler in particular should stay small).
- `go test ./internal/examples/logs/...` (domain units) and the new E2E.
- Codex review before finishing the branch (stop-time gate).
