# Live Log Feed Example — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `/examples/logs` example app: synthetic log lines streamed to the browser over SSE, rendered with Goshtoso components.

**Architecture:** htmx-ext-sse opens an `EventSource` to a Go SSE endpoint that emits one server-rendered `LogRow` fragment per tick. htmx owns row *insertion* (append into `#log-feed`); Alpine (`x-data="logFeed"`) owns everything else — row cap, min-severity filter (CSS-class driven), auto-scroll, pause/resume (removes the connector element → closes the stream), and connection status. No cookie, no shared state; the only state is the per-connection goroutine, which exits on `ctx.Done()`.

**Tech Stack:** Go 1.26, templ, htmx 2.0.8 + htmx-ext-sse 2.2.x, Alpine.js 3, Tailwind v4, Playwright (E2E).

**Spec:** `docs/superpowers/specs/2026-05-31-live-log-feed-example-design.md`

---

## File Structure

- `assets/js/vendor/htmx-ext-sse.min.js` — **new**, vendored SSE extension.
- `internal/pages/demo/layout.templ` — **modify**, add `<script>` for the ext (after htmx).
- `assets/embed.go` — **modify**, add the new vendored file to the comment list.
- `internal/examples/logs/logs.go` — **new**, pure domain: `Level`, `LogLine`, `Line(i)`, severity ordering.
- `internal/examples/logs/logs_test.go` — **new**, domain unit tests.
- `internal/pages/demo/examples/logs.templ` — **new**, view: `LogRow`, `LogFeed`, control bar, `LogsApp`, `LogsContent`, the `logFeed` Alpine script, and the scoped filter `<style>`.
- `internal/server/logs_handler.go` — **new**, `registerLogsRoutes`, `handleLogsStream`, `streamLogs`, `durParam`.
- `internal/server/logs_handler_test.go` — **new**, `streamLogs` framing test.
- `internal/server/server.go` — **modify**, call `registerLogsRoutes()`, add `case "logs"` in `handleExample`.
- `internal/pages/demo/components/registry.go` — **modify**, add `examples/logs` entry.
- `internal/pages/demo/examples/index.templ` — **modify**, add a gallery card.
- `internal/pages/demo/layout.templ` — **modify**, add the Examples sidebar item.
- `tests/e2e/logs_test.go` — **new**, E2E.
- `CLAUDE.md` — **modify**, docs updates.

---

## Task 1: Vendor the htmx SSE extension + wire it

**Files:**
- Create: `assets/js/vendor/htmx-ext-sse.min.js`
- Modify: `internal/pages/demo/layout.templ` (after the htmx `<script>`, ~line 33)
- Modify: `assets/embed.go` (vendor comment list, ~line 14)

- [ ] **Step 1: Download the extension matching htmx 2.x**

```bash
cd /Users/guilhermecastro/repos/araihu/goshtoso/.claude/worktrees/example-live-log-feed
curl -fsSL https://unpkg.com/htmx-ext-sse@2.2.3/dist/sse.min.js -o assets/js/vendor/htmx-ext-sse.min.js
```
Expected: file exists, non-empty.
```bash
test -s assets/js/vendor/htmx-ext-sse.min.js && echo OK
```
Expected: `OK`. If `curl` is blocked by the sandbox, ask the user to run the command via `! curl …`.

- [ ] **Step 2: Record the exact SSE event names the ext dispatches**

Later tasks listen for SSE lifecycle events; the names must match the vendored file exactly.
```bash
grep -oE "htmx:sse[A-Za-z]+" assets/js/vendor/htmx-ext-sse.min.js | sort -u
```
Expected: a list including `htmx:sseOpen`, `htmx:sseError`, `htmx:sseMessage` (and possibly `htmx:sseClose`, `htmx:sseBeforeMessage`). **Note:** htmx dual-dispatches camelCase and kebab-case, so Alpine listens with kebab (`htmx:sse-open`, `htmx:sse-error`). If a name differs from what Task 5 uses, update Task 5's listeners to match.

- [ ] **Step 3: Add the script tag after htmx in the layout head**

In `internal/pages/demo/layout.templ`, immediately after the line:
```html
<script src="/assets/js/vendor/htmx.min.js"></script>
```
add:
```html
<script src="/assets/js/vendor/htmx-ext-sse.min.js"></script>
```
(htmx must load before its extension; keep it non-`defer` like htmx.)

- [ ] **Step 4: Add the file to the embed comment list**

In `assets/embed.go`, after the `htmx.min.js` comment line, add:
```go
//   - /assets/js/vendor/htmx-ext-sse.min.js — HTMX SSE extension
```

- [ ] **Step 5: Build to confirm nothing broke**

```bash
templ generate && go build ./...
```
Expected: exit 0.

- [ ] **Step 6: Commit**

```bash
git add assets/js/vendor/htmx-ext-sse.min.js internal/pages/demo/layout.templ assets/embed.go
git commit -m "chore(examples): vendor htmx-ext-sse for the live log feed

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Domain package `internal/examples/logs`

**Files:**
- Create: `internal/examples/logs/logs.go`
- Test: `internal/examples/logs/logs_test.go`

- [ ] **Step 1: Write the failing test**

`internal/examples/logs/logs_test.go`:
```go
package logs

import "testing"

func TestLineCyclesSamplePool(t *testing.T) {
	if PoolSize() == 0 {
		t.Fatal("expected a non-empty sample pool")
	}
	// Line(i) cycles deterministically: index i and i+PoolSize match.
	a := Line(0)
	b := Line(PoolSize())
	if a.Level != b.Level || a.Source != b.Source || a.Message != b.Message {
		t.Errorf("Line(0) and Line(PoolSize()) differ: %+v vs %+v", a, b)
	}
	// Negative / large indices never panic and stay in-pool.
	_ = Line(-3)
	_ = Line(1_000_000)
}

func TestSeverityOrdering(t *testing.T) {
	if !(Debug.Severity() < Info.Severity() &&
		Info.Severity() < Warn.Severity() &&
		Warn.Severity() < Error.Severity()) {
		t.Fatal("severity must order Debug < Info < Warn < Error")
	}
}

func TestLevelString(t *testing.T) {
	if Error.String() != "error" {
		t.Errorf("got %q, want %q", Error.String(), "error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/examples/logs/...
```
Expected: FAIL — package/symbols undefined.

- [ ] **Step 3: Write the implementation**

`internal/examples/logs/logs.go`:
```go
// Package logs is the pure (HTTP-free) domain for the Live Log Feed example:
// a deterministic synthetic log-line generator. It holds no state and is fully
// unit-testable; the SSE handler supplies wall-clock time per tick.
package logs

import "time"

// Level is a log severity.
type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
)

// String returns the lowercase level name (also the CSS-class suffix).
func (l Level) String() string { return string(l) }

// Severity is the ordering used by the min-level filter (higher = more severe).
func (l Level) Severity() int {
	switch l {
	case Error:
		return 3
	case Warn:
		return 2
	case Info:
		return 1
	default: // Debug
		return 0
	}
}

// LogLine is one rendered feed row. Time is supplied by the caller.
type LogLine struct {
	Time    time.Time
	Level   Level
	Source  string
	Message string
}

// pool is the fixed, deterministic sample of log lines (Time is filled per tick).
var pool = []LogLine{
	{Level: Info, Source: "http", Message: "GET /examples/logs 200 in 4ms"},
	{Level: Debug, Source: "cache", Message: "hit key=session:1f2c ttl=300s"},
	{Level: Info, Source: "auth", Message: "user 4821 signed in"},
	{Level: Warn, Source: "ratelimit", Message: "client 10.0.0.5 nearing quota (92%)"},
	{Level: Info, Source: "worker", Message: "job email.send completed in 120ms"},
	{Level: Error, Source: "db", Message: "connection reset; retrying (attempt 2/3)"},
	{Level: Debug, Source: "router", Message: "matched route /api/examples/logs/stream"},
	{Level: Info, Source: "http", Message: "POST /api/examples/todo/add 200 in 7ms"},
	{Level: Warn, Source: "tls", Message: "certificate for cdn expires in 9 days"},
	{Level: Error, Source: "payments", Message: "charge declined: insufficient_funds"},
	{Level: Info, Source: "worker", Message: "scheduled cleanup removed 12 rows"},
	{Level: Debug, Source: "cache", Message: "evicted 3 cold entries"},
}

// PoolSize is the number of distinct sample lines.
func PoolSize() int { return len(pool) }

// Line returns the i-th synthetic line, cycling the pool. Safe for any int.
func Line(i int) LogLine {
	n := len(pool)
	idx := ((i % n) + n) % n // floored modulo, handles negatives
	return pool[idx]
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/examples/logs/...
```
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/logs/
git commit -m "feat(examples): synthetic log-line domain for the live log feed

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: View — `LogRow` fragment (needed by the handler)

**Files:**
- Create: `internal/pages/demo/examples/logs.templ` (partial — `LogRow` + the level→badge helper only; rest added in Task 5)

> `LogRow` is built first because Task 4's SSE handler renders it. The control bar / `LogsApp` / Alpine script come in Task 5.

- [ ] **Step 1: Write the `LogRow` fragment + helper**

`internal/pages/demo/examples/logs.templ`:
```go
package examples

import (
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/internal/examples/logs"
)

// levelBadge maps a log level to a badge variant.
func levelBadge(l logs.Level) badge.Variant {
	switch l {
	case logs.Error:
		return badge.Danger
	case logs.Warn:
		return badge.Warning
	case logs.Info:
		return badge.Info
	default: // Debug
		return badge.Secondary
	}
}

// LogRow is one streamed feed row. The `log-level-<level>` class drives the
// pure-CSS min-severity filter; the row text (time · source · message) is
// hand-rolled because no component models a log line.
templ LogRow(line logs.LogLine) {
	<div class={ "log-row log-level-" + line.Level.String(),
		"flex items-center gap-3 border-b border-outline/50 px-3 py-1.5 font-mono text-xs dark:border-outline-dark/50" }>
		<time class="shrink-0 tabular-nums text-on-surface-muted dark:text-on-surface-dark-muted">
			{ line.Time.Format("15:04:05") }
		</time>
		<span class="w-16 shrink-0">
			@badge.Badge(badge.Config{
				Variant: levelBadge(line.Level),
				Style:   badge.StyleSoft,
				Size:    badge.SizeSM,
				Text:    line.Level.String(),
			})
		</span>
		<span class="w-20 shrink-0 truncate text-on-surface-muted dark:text-on-surface-dark-muted">{ line.Source }</span>
		<span class="min-w-0 flex-1 truncate text-on-surface dark:text-on-surface-dark">{ line.Message }</span>
	</div>
}
```

- [ ] **Step 2: Generate + build**

```bash
templ generate && go build ./...
```
Expected: exit 0. (`LogRow` is exported and compiles even though it is not yet referenced by a page.)

- [ ] **Step 3: Commit**

```bash
git add internal/pages/demo/examples/logs.templ internal/pages/demo/examples/logs_templ.go
git commit -m "feat(examples): LogRow fragment for the live log feed

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: SSE endpoint

**Files:**
- Create: `internal/server/logs_handler.go`
- Test: `internal/server/logs_handler_test.go`
- Modify: `internal/server/server.go` (`setupRoutes` — register the route)

- [ ] **Step 1: Write the failing test**

`internal/server/logs_handler_test.go`:
```go
package server

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamLogsFraming(t *testing.T) {
	rec := httptest.NewRecorder()
	// max=3 → returns after 3 events without needing cancellation.
	streamLogs(context.Background(), rec, rec, time.Millisecond, 3)

	body := rec.Body.String()
	if got := strings.Count(body, "event: message"); got != 3 {
		t.Fatalf("expected 3 events, got %d\n%s", got, body)
	}
	// Every payload line is a `data:` field (SSE framing); no bare HTML lines.
	for _, line := range strings.Split(strings.TrimSpace(body), "\n") {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "event:") && !strings.HasPrefix(line, "data:") {
			t.Fatalf("unframed SSE line: %q", line)
		}
	}
	// A known sample message and the level badge text both appear.
	if !strings.Contains(body, "GET /examples/logs") {
		t.Errorf("expected a sample log message in output")
	}
}

func TestStreamLogsStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { streamLogs(ctx, rec, rec, time.Millisecond, 0); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("streamLogs did not return on cancelled context")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/server/ -run TestStreamLogs
```
Expected: FAIL — `streamLogs` undefined.

- [ ] **Step 3: Write the implementation**

`internal/server/logs_handler.go`:
```go
package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/araihu/goshtoso/internal/examples/logs"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// registerLogsRoutes wires the live log feed's SSE endpoint.
func (s *Server) registerLogsRoutes() {
	s.mux.HandleFunc("/api/examples/logs/stream", s.handleLogsStream)
}

// durParam reads a named query param as a time.Duration, clamped to a sane
// minimum so a hostile ?interval=0 cannot spin a tight loop.
func durParam(r *http.Request, name string, def time.Duration) time.Duration {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d < 10*time.Millisecond {
		return def
	}
	return d
}

// handleLogsStream is the SSE endpoint. It streams one rendered LogRow per tick
// until the client disconnects (ctx cancel) or ?max is reached. ?interval and
// ?max keep e2e fast and bounded.
func (s *Server) handleLogsStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	interval := durParam(r, "interval", 500*time.Millisecond)
	max := intParam(r, "max", 0) // 0 = unbounded; intParam is defined in todo_handler.go
	streamLogs(r.Context(), w, flusher, interval, max)
}

// streamLogs is the testable core: emit synthetic LogRow SSE events on a ticker.
func streamLogs(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, interval time.Duration, max int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for i := 0; ; i++ {
		if max > 0 && i >= max {
			return
		}
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			line := logs.Line(i)
			line.Time = t
			var buf bytes.Buffer
			if err := examples.LogRow(line).Render(ctx, &buf); err != nil {
				return
			}
			writeSSEMessage(w, buf.String())
			flusher.Flush()
		}
	}
}

// writeSSEMessage frames an HTML fragment as one SSE `message` event. templ
// output is multi-line, but a `data:` field is single-line, so each output
// line becomes its own `data:` field; the htmx SSE ext rejoins them with "\n".
func writeSSEMessage(w http.ResponseWriter, html string) {
	fmt.Fprint(w, "event: message\n")
	for _, line := range strings.Split(strings.TrimRight(html, "\n"), "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}
```

- [ ] **Step 4: Register the route**

In `internal/server/server.go`, inside `setupRoutes`, right after `s.registerTodoRoutes()`:
```go
	s.registerLogsRoutes()
```

- [ ] **Step 5: Run tests + build**

```bash
go test ./internal/server/ -run TestStreamLogs && go build ./...
```
Expected: PASS (2 tests), build exit 0.

- [ ] **Step 6: Commit**

```bash
git add internal/server/logs_handler.go internal/server/logs_handler_test.go internal/server/server.go
git commit -m "feat(examples): SSE endpoint streaming synthetic log rows

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: View — control bar, `LogsApp`, Alpine `logFeed`, filter CSS

**Files:**
- Modify: `internal/pages/demo/examples/logs.templ` (append to the file from Task 3)

The `logFeed` Alpine component is registered via `<script>` + `Alpine.data()` (templ-escaping-safe) and registers **immediately if Alpine is already running** (so sidebar fragment-nav works). The filter is pure CSS via a scoped `<style>`. Pause removes the SSE *connector* element (closing the EventSource) while `#log-feed` and its rows persist, because the connector targets `#log-feed` via `hx-target`.

- [ ] **Step 1: Append the script, styles, control bar, and page entry points**

Append to `internal/pages/demo/examples/logs.templ`. Add the new imports to the existing `import (...)` block at the top of the file:
```go
	"github.com/araihu/goshtoso/components/button"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/spinner"
	"github.com/araihu/goshtoso/components/toggle"
	"github.com/araihu/goshtoso/components/tooltip"
```

Then append the following templ definitions:
```go
// logFeedScript registers the Alpine component that owns everything htmx does
// not: row cap, min-severity filter, auto-scroll, pause/resume, and status.
// Registered immediately when Alpine is already running (fragment-nav), else on
// alpine:init (full page load).
templ logFeedScript() {
	@templ.Raw(`<script>
	(function () {
		function register() {
			if (!window.Alpine || Alpine.__logFeedRegistered) return;
			Alpine.__logFeedRegistered = true;
			Alpine.data('logFeed', () => ({
				cap: 100,
				paused: false,
				autoScroll: true,
				connected: false,
				minLevel: 'all',
				init() {
					this.$watch('minLevel', (v) => this.applyFilter(v));
					this.applyFilter(this.minLevel);
				},
				applyFilter(v) {
					const wrap = this.$refs.feedWrap;
					if (!wrap) return;
					wrap.classList.remove('flt-all', 'flt-warn', 'flt-error');
					wrap.classList.add('flt-' + v);
				},
				onSwap(e) {
					if (!e.detail || !e.detail.target || e.detail.target.id !== 'log-feed') return;
					const feed = document.getElementById('log-feed');
					if (!feed) return;
					while (feed.children.length > this.cap) feed.removeChild(feed.firstElementChild);
					if (this.autoScroll) feed.scrollTop = feed.scrollHeight;
				},
				togglePause() { this.paused = !this.paused; if (this.paused) this.connected = false; },
				clearFeed() { const f = document.getElementById('log-feed'); if (f) f.replaceChildren(); },
				get statusText() { return this.paused ? 'Paused' : (this.connected ? 'Connected' : 'Connecting'); },
			}));
		}
		if (window.Alpine) register();
		else document.addEventListener('alpine:init', register);
	})();
	</script>`)
}

// logFilterStyle scopes the pure-CSS min-severity filter. The feed wrapper
// carries flt-all / flt-warn / flt-error; rows carry log-level-<level>.
templ logFilterStyle() {
	@templ.Raw(`<style>
	.flt-warn .log-level-debug, .flt-warn .log-level-info { display: none; }
	.flt-error .log-level-debug, .flt-error .log-level-info, .flt-error .log-level-warn { display: none; }
	</style>`)
}

// LogsApp is the example body, rendered both on first load and as the fragment.
templ LogsApp() {
	@logFilterStyle()
	@logFeedScript()
	<div id="logs-fragment" x-data="logFeed" class="mx-auto max-w-3xl">
		<header class="mb-6">
			<h1 class="text-2xl font-bold text-on-surface dark:text-on-surface-dark">Live Log Feed</h1>
			<p class="mt-2 text-on-surface-muted dark:text-on-surface-dark-muted">
				A streaming log viewer built from Goshtoso components. The server pushes synthetic
				log lines over Server-Sent Events; htmx appends each rendered row while Alpine owns
				filtering, auto-scroll, pause, and connection status. No state is stored — the only
				state is the live connection, which closes when you pause or leave.
			</p>
		</header>
		<div class="rounded-radius border border-outline bg-surface-alt dark:border-outline-dark dark:bg-surface-dark-alt">
			<div class="flex flex-col gap-4 p-4">
				<!-- Control bar (OUTSIDE #log-feed so it survives every swap) -->
				<div class="flex flex-wrap items-end gap-3">
					<div class="w-40">
						@selectfield.Select(selectfield.Config{
							ID:          "log-filter",
							Label:       "Minimum level",
							AlpineModel: "minLevel",
							Options: []selectfield.Option{
								{Value: "all", Label: "All", Selected: true},
								{Value: "warn", Label: "Warn+"},
								{Value: "error", Label: "Error"},
							},
						})
					</div>
					<div class="self-center pt-5" x-on:change="autoScroll = $event.target.checked">
						@toggle.Toggle(toggle.Config{ID: "log-autoscroll", Label: "Auto-scroll", Checked: true})
					</div>
					<div class="ml-auto flex items-center gap-3 self-center pt-5">
						<!-- Connection status: Spinner while connecting/reconnecting, Badge for state -->
						<span class="flex items-center gap-2 text-sm text-on-surface-muted dark:text-on-surface-dark-muted">
							<span x-show="!connected && !paused">
								@spinner.Spinner(spinner.Config{Size: spinner.SizeSM})
							</span>
							<span x-text="statusText"></span>
						</span>
						@tooltip.Tooltip(tooltip.Config{
							ID:   "log-pause-tip",
							Text: "Pause closes the stream; resume reconnects",
							TriggerContent: pauseButton(),
						})
						@button.Button(button.Config{
							Variant: button.Secondary,
							Size:    button.SizeSmall,
							Type:    "button",
							Alpine:  &button.AlpineConfig{OnClick: "clearFeed()"},
						}) {
							Clear
						}
					</div>
				</div>
				<!-- Persistent feed + filter wrapper -->
				<div
					x-ref="feedWrap"
					x-on:htmx:after-swap="onSwap($event)"
					x-on:htmx:sse-open="connected = true"
					x-on:htmx:sse-error="connected = false"
					class="flt-all rounded-radius border border-outline dark:border-outline-dark"
				>
					<div id="log-feed" class="h-80 overflow-y-auto"></div>
				</div>
				<!-- SSE connector: removed on pause (closes EventSource), targets the persistent feed -->
				<template x-if="!paused">
					<div hx-ext="sse" sse-connect="/api/examples/logs/stream">
						<div sse-swap="message" hx-target="#log-feed" hx-swap="beforeend" hidden></div>
					</div>
				</template>
			</div>
		</div>
	</div>
}

// pauseButton is the Pause/Resume button used as the tooltip trigger; its label
// flips with the Alpine `paused` flag.
templ pauseButton() {
	@button.Button(button.Config{
		Variant: button.Secondary,
		Size:    button.SizeSmall,
		Type:    "button",
		Alpine:  &button.AlpineConfig{OnClick: "togglePause()"},
	}) {
		<span x-text="paused ? 'Resume' : 'Pause'">Pause</span>
	}
}

// LogsContent is the registry entry point.
templ LogsContent() {
	@LogsApp()
}
```

> **Note on `x-on:htmx:sse-open` / `sse-error`:** kebab-case per Task 1 Step 2. If the grep there showed different event names, use those.

- [ ] **Step 2: Generate + build**

```bash
templ generate && go build ./...
```
Expected: exit 0. If `templ generate` reports "0 updates" but rendering is stale, force it:
```bash
rm internal/pages/demo/examples/logs_templ.go && templ generate && go build ./...
```

- [ ] **Step 3: Smoke-test in the dev server**

```bash
go run cmd/server/main.go &
sleep 2
curl -s http://localhost:8090/api/examples/logs/stream?interval=10ms\&max=2 | head -20
kill %1
```
Expected: SSE output with `event: message` and `data:` lines containing a `log-level-` div. (Port 8090 is the manual-dev port; fine for a one-off smoke test.)

- [ ] **Step 4: Commit**

```bash
git add internal/pages/demo/examples/logs.templ internal/pages/demo/examples/logs_templ.go
git commit -m "feat(examples): live log feed view, controls, and Alpine logFeed

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Wire routing, registry, sidebar, gallery

**Files:**
- Modify: `internal/pages/demo/components/registry.go` (add entry)
- Modify: `internal/server/server.go` (`handleExample` — add case)
- Modify: `internal/pages/demo/layout.templ` (Examples sidebar block, ~line 469)
- Modify: `internal/pages/demo/examples/index.templ` (gallery grid)

- [ ] **Step 1: Add the registry entry**

In `internal/pages/demo/components/registry.go`, in the `Demos` map after the `"examples/todo"` line:
```go
	"examples/logs":              {"Live Log Feed", "logs", examples.LogsContent},
```

- [ ] **Step 2: Route `/examples/logs`**

In `internal/server/server.go`, in `handleExample`'s `switch sub`, add a case before `default`:
```go
	case "logs":
		s.renderDemo(w, r, "examples/logs")
```
(No custom page handler: the feed has no per-request state, so `renderDemo` handles Layout-vs-Fragment like the examples index.)

- [ ] **Step 3: Add the sidebar item**

In `internal/pages/demo/layout.templ`, in the Examples section `Items`, after the Todo List `sItem`:
```go
				sItem("logs", "Live Log Feed", "/examples/logs", activeComponent),
```

- [ ] **Step 4: Add the gallery card**

In `internal/pages/demo/examples/index.templ`, inside the grid `<div>`, after the todo `@exampleCard(...)`:
```go
			@exampleCard("/examples/logs", "Live Log Feed", "Streaming log viewer over Server-Sent Events: htmx appends rows, Alpine filters, pauses, and tracks connection status.")
```

- [ ] **Step 5: Generate + build + run lint**

```bash
templ generate && go build -o bin/server ./cmd/server && golangci-lint run
```
Expected: build exit 0; lint clean. (Watch `handleLogsStream`/`streamLogs` cyclomatic complexity — both are small; if lint flags one, extract a helper rather than suppress.)

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/components/registry.go internal/server/server.go internal/pages/demo/layout.templ internal/pages/demo/examples/index.templ internal/pages/demo/examples/index_templ.go
git commit -m "feat(examples): register live log feed (route, registry, sidebar, gallery)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: E2E tests

**Files:**
- Create: `tests/e2e/logs_test.go`

Use `?interval=50ms` for speed. Wait for Alpine and for the first row. The pause button is wrapped in a tooltip; target it by its text. Use `WaitForFunction` for async (row count, status). Console-error capture mirrors the existing examples tests.

- [ ] **Step 1: Write the E2E tests**

`tests/e2e/logs_test.go`:
```go
package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// logsURL builds the example URL with a fast stream interval for tests.
func logsURL(base string) string {
	return base + "/examples/logs?interval=50ms"
}

func TestLogFeed_StreamsRows(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(logsURL(serverURL)); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	// At least one row with a level badge arrives.
	if _, err := page.WaitForSelector("#log-feed .log-row"); err != nil {
		t.Fatalf("no log rows streamed: %v", err)
	}
}

func TestLogFeed_FragmentNavNoConsoleErrors(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})
	// Land on the examples index, then navigate to the feed via the sidebar link
	// (fragment swap), exercising the Alpine-register-on-fragment-nav path.
	if _, err := page.Goto(serverURL + "/examples"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	link := page.Locator("a[href='/examples/logs']").First()
	if err := link.Click(); err != nil {
		t.Fatalf("sidebar click: %v", err)
	}
	if _, err := page.WaitForSelector("#log-feed .log-row"); err != nil {
		t.Fatalf("no rows after fragment nav: %v", err)
	}
	if len(consoleErrors) > 0 {
		t.Fatalf("console errors after fragment nav: %v", consoleErrors)
	}
}

func TestLogFeed_PauseStopsRows(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(logsURL(serverURL)); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#log-feed .log-row")
	// Pause.
	if err := page.Locator("button", playwright.PageLocatorOptions{HasText: "Pause"}).First().Click(); err != nil {
		t.Fatalf("pause click: %v", err)
	}
	// Status reads "Paused".
	if _, err := page.WaitForFunction(
		"() => document.querySelector('#logs-fragment [x-text=\"statusText\"]').textContent.trim() === 'Paused'", nil,
	); err != nil {
		t.Fatalf("status did not reach Paused: %v", err)
	}
	// Row count is stable across a window in which un-paused it would have grown.
	count := func() int {
		v, _ := page.Evaluate("() => document.querySelectorAll('#log-feed .log-row').length")
		n, _ := v.(int)
		return n
	}
	before := count()
	page.WaitForTimeout(400) // > several 50ms ticks
	if after := count(); after != before {
		t.Fatalf("rows grew while paused: %d -> %d", before, after)
	}
	// Resume → rows grow again.
	page.Locator("button", playwright.PageLocatorOptions{HasText: "Resume"}).First().Click()
	if _, err := page.WaitForFunction(
		fmt.Sprintf("() => document.querySelectorAll('#log-feed .log-row').length > %d", before), nil,
	); err != nil {
		t.Fatalf("rows did not resume: %v", err)
	}
}

func TestLogFeed_FilterHidesLowerLevels(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(logsURL(serverURL)); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#log-feed .log-row")
	// Drive the filter directly via Alpine state (the Select's x-model target),
	// avoiding custom-dropdown click choreography.
	page.Evaluate("() => { const r = document.getElementById('logs-fragment'); Alpine.$data(r).minLevel = 'error'; }")
	// Wait until the wrapper reflects the filter, then assert no debug/info rows are visible.
	if _, err := page.WaitForFunction(
		"() => document.querySelector('[x-ref=feedWrap]').classList.contains('flt-error')", nil,
	); err != nil {
		t.Fatalf("filter class not applied: %v", err)
	}
	visibleLower, _ := page.Evaluate(`() => {
		return [...document.querySelectorAll('#log-feed .log-level-debug, #log-feed .log-level-info')]
			.filter(el => el.offsetParent !== null).length;
	}`)
	if n, _ := visibleLower.(int); n != 0 {
		t.Fatalf("expected lower-severity rows hidden, %d visible", n)
	}
}

func TestLogFeed_ClearEmptiesFeed(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(logsURL(serverURL)); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#log-feed .log-row")
	page.Locator("button", playwright.PageLocatorOptions{HasText: "Clear"}).First().Click()
	// Immediately after clear, the feed is empty (it may refill on the next tick).
	v, _ := page.Evaluate("() => document.querySelectorAll('#log-feed .log-row').length")
	if n, _ := v.(int); n > 1 {
		t.Fatalf("expected feed cleared, found %d rows", n)
	}
}
```

> **Note:** confirm the shared-server URL variable name used by the existing e2e tests (`serverURL` here is a guess). Open `tests/e2e/e2e_test.go` and `tests/e2e/sidebar_test.go` and match the exact helper/variable names (`setupPlaywright`, `newPage`, base-URL accessor) before running. Adjust the `Evaluate` int-cast pattern to match how other tests read numeric results (Playwright-go returns `interface{}`).

- [ ] **Step 2: Run the new tests**

```bash
go test ./tests/e2e/... -count=1 -timeout 5m -run TestLogFeed
```
Expected: PASS (5 tests). If a just-rendered control is clicked during an htmx swap and the click is lost, switch that interaction to the `clickUntil` helper (per CLAUDE.md's HTMX rebind-race note).

- [ ] **Step 3: Commit**

```bash
git add tests/e2e/logs_test.go
git commit -m "test(e2e): live log feed — stream, fragment-nav, pause, filter, clear

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Documentation

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update the Example Apps section**

In `CLAUDE.md` → "Example Apps (`/examples/*`)", add a bullet describing the SSE example: htmx-ext-sse transport, the htmx-owns-insertion / Alpine-owns-the-rest split, pause = disconnect-via-connector-removal, and the scoped-CSS min-severity filter.

- [ ] **Step 2: Update Current Status**

In "Current Status → Example apps", add:
```
Live Log Feed (`/examples/logs`) — SSE-streamed synthetic logs; htmx-ext-sse +
Alpine; the first server-push example.
```

- [ ] **Step 3: Update the vendor list**

In the Tech Stack / repo-structure notes where `assets/js/vendor/` is described, add `htmx-ext-sse.min.js` alongside alpine/htmx.

- [ ] **Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: document the live log feed SSE example

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: Full verification + review gate

- [ ] **Step 1: Full build + lint + unit tests**

```bash
templ generate && go build ./... && golangci-lint run && go test ./internal/... -count=1
```
Expected: all green.

- [ ] **Step 2: Full E2E suite (catches cross-test browser pressure / rebind races)**

```bash
go test ./tests/e2e/... -count=1 -timeout 15m
```
Expected: all pass, no new flakes. If a log-feed test flakes only under the full suite, apply `clickUntil` / `WaitForFunction` hardening.

- [ ] **Step 3: Codex review (stop-time gate)**

Hand the branch diff to Codex (`codex:rescue` skill or the review gate) for an independent pass. Confirm findings against the code before acting.

- [ ] **Step 4: Finish the branch**

Use the `superpowers:finishing-a-development-branch` skill to choose merge / PR / cleanup.

---

## Self-Review

**Spec coverage:**
- Identity & routing → Task 6. Stateless model → Task 4 (`ctx.Done()`, no cookie). Domain layering → Task 2. View layering → Tasks 3 & 5. Handler layering → Task 4. Transport / htmx-ext-sse → Task 1 + Task 5 markup. SSE framing → Task 4 (`writeSSEMessage` + test). `?interval`/`?max` → Task 4. Division of labor (cap/filter/autoscroll/pause/status) → Task 5 `logFeed`. Components showcased → Task 5 (Badge, Button, Select, Spinner, Toggle, Tooltip). E2E incl. fragment-nav + no-console-errors → Task 7. Docs → Task 8. Build/verify checklist → Task 9.
- **Deviation from spec:** spec listed **Card** for the panel shell; Card is an ecommerce product-card (Image/Title/Price), so the panel uses a plain bordered div (as todo does) and Card is dropped. Showcase is 6 components, not 7. Documented here and in Task 5.

**Placeholder scan:** No TBD/TODO. Two explicit "confirm against existing code" notes (Task 1 Step 2 event names; Task 7 e2e helper names) are verification steps with concrete fallbacks, not unfilled blanks.

**Type consistency:** `Line(i int) LogLine`, `Level.String()`, `Level.Severity()` defined in Task 2 and used in Tasks 3/4. `streamLogs(ctx, w, flusher, interval, max)` signature matches between Task 4 impl and test. `LogRow(logs.LogLine)` defined Task 3, called Task 4. `logFeed` Alpine methods (`onSwap`, `togglePause`, `clearFeed`, `applyFilter`, `statusText`, `minLevel`, `autoScroll`, `connected`, `paused`) defined and referenced consistently in Task 5 and asserted in Task 7. `intParam` reused from `todo_handler.go` (same package). `examples.LogsContent` matches the registry field type `func() templ.Component`.
