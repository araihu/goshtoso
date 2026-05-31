# Live Ticker (SSE) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `/examples/ticker` example app — a live stock-market watchlist board driven by Server-Sent Events, built from real Goshtoso components (Table, Badge, Card, Spinner, Toggle).

**Architecture:** Three layers mirroring the Todo example. A seeded, deterministic `Simulator` (random walk) is advanced by a single process-wide `Broker` goroutine that fans each snapshot out to all SSE subscribers — one shared stream, no per-user state. The SSE handler pushes rendered HTML per symbol (one named event per ticker). Each table price cell and the spotlight card price both subscribe to their symbol's event via `sse-swap`, so the shared stream updates everything; selection (which symbol is spotlighted) is purely client-side. Pause suppresses swaps via a cancelable `htmx:sseBeforeMessage` listener — the connection stays open.

**Tech Stack:** Go 1.26 (net/http ServeMux, `http.ResponseController.Flush`), templ v0.3, htmx 2.0.8 + the bundled `htmx-ext-sse` extension, Alpine.js 3, Tailwind v4, Playwright E2E.

---

## File Structure

**Create:**
- `internal/examples/ticker/types.go` — `Symbol`, `Snapshot`, `ChangePct`/`Direction`.
- `internal/examples/ticker/sim.go` — `Simulator` (seeded random walk), `InitialSymbols`.
- `internal/examples/ticker/broker.go` — pub/sub `Broker` (Run/Subscribe/Snapshot/Find).
- `internal/examples/ticker/sim_test.go` — determinism + change math.
- `internal/examples/ticker/broker_test.go` — fan-out, unsubscribe, slow-subscriber drop.
- `internal/pages/demo/examples/ticker.templ` — page shell + cell/spotlight/script templ components.
- `internal/pages/demo/examples/ticker_table.go` — Go helpers (table config, formatting, badge variant).
- `internal/server/ticker_handler.go` — SSE stream + spotlight handlers, route registration.
- `assets/js/vendor/sse.js` — vendored htmx SSE extension (offline, no CDN).
- `tests/e2e/ticker_test.go` — E2E (stream update, spotlight, pause/resume, fragment-nav no-errors).

**Modify:**
- `internal/server/server.go` — add `tickerBroker` field, start it in `New`, add `case "ticker"` in `handleExample`, call `registerTickerRoutes`.
- `internal/pages/demo/components/registry.go` — add `"examples/ticker"` entry.
- `internal/pages/demo/examples/index.templ` — add an `@exampleCard` for the ticker.
- `internal/pages/demo/layout.templ` — add a sidebar `sItem` under the Examples section.
- `assets/embed.go` — document the new vendor file in the package comment.

---

## Task 1: Domain types

**Files:**
- Create: `internal/examples/ticker/types.go`
- Test: `internal/examples/ticker/sim_test.go` (created here, expanded in Task 2)

- [ ] **Step 1: Write the failing test**

Create `internal/examples/ticker/sim_test.go`:

```go
package ticker_test

import (
	"testing"

	"github.com/araihu/goshtoso/internal/examples/ticker"
	"github.com/stretchr/testify/assert"
)

func TestChangePctAndDirection(t *testing.T) {
	up := ticker.Symbol{Ticker: "AAA", Price: 110, PrevPrice: 100}
	assert.InDelta(t, 10.0, up.ChangePct(), 0.0001)
	assert.Equal(t, "up", up.Direction())

	down := ticker.Symbol{Ticker: "BBB", Price: 90, PrevPrice: 100}
	assert.InDelta(t, -10.0, down.ChangePct(), 0.0001)
	assert.Equal(t, "down", down.Direction())

	flat := ticker.Symbol{Ticker: "CCC", Price: 100, PrevPrice: 100}
	assert.InDelta(t, 0.0, flat.ChangePct(), 0.0001)
	assert.Equal(t, "flat", flat.Direction())

	// Guard against divide-by-zero when PrevPrice is unset.
	zero := ticker.Symbol{Ticker: "DDD", Price: 5, PrevPrice: 0}
	assert.InDelta(t, 0.0, zero.ChangePct(), 0.0001)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/ticker/ -run TestChangePctAndDirection`
Expected: FAIL — `undefined: ticker.Symbol` (package does not compile yet).

- [ ] **Step 3: Write minimal implementation**

Create `internal/examples/ticker/types.go`:

```go
// Package ticker is the HTTP-free domain layer for the /examples/ticker
// example: a deterministic, seeded stock-price simulator plus a pub/sub broker
// that fans each tick out to all SSE subscribers. One shared stream, no
// per-user state.
package ticker

// Symbol is a single tracked instrument and its latest two prices.
type Symbol struct {
	Ticker    string
	Name      string
	Price     float64
	PrevPrice float64
}

// ChangePct is the percent change from PrevPrice to Price. Returns 0 when
// PrevPrice is 0 (guards the first tick / unset state).
func (s Symbol) ChangePct() float64 {
	if s.PrevPrice == 0 {
		return 0
	}
	return (s.Price - s.PrevPrice) / s.PrevPrice * 100
}

// Direction is "up", "down", or "flat" relative to PrevPrice.
func (s Symbol) Direction() string {
	switch {
	case s.Price > s.PrevPrice:
		return "up"
	case s.Price < s.PrevPrice:
		return "down"
	default:
		return "flat"
	}
}

// Snapshot is a point-in-time view of every symbol.
type Snapshot struct {
	Symbols []Symbol
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/ticker/ -run TestChangePctAndDirection`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/examples/ticker/types.go internal/examples/ticker/sim_test.go
git commit -m "feat(ticker): domain types (Symbol/Snapshot, change %, direction)"
```

---

## Task 2: Simulator (seeded random walk)

**Files:**
- Create: `internal/examples/ticker/sim.go`
- Test: `internal/examples/ticker/sim_test.go` (add to it)

- [ ] **Step 1: Write the failing test**

Append to `internal/examples/ticker/sim_test.go`:

```go
func TestSimulatorIsDeterministicForSeed(t *testing.T) {
	a := ticker.NewSimulator(42)
	b := ticker.NewSimulator(42)
	for i := 0; i < 5; i++ {
		sa := a.Tick()
		sb := b.Tick()
		assert.Equal(t, sb.Symbols, sa.Symbols, "same seed must produce identical ticks")
	}
}

func TestTickSetsPrevPriceAndStaysPositive(t *testing.T) {
	s := ticker.NewSimulator(1)
	before := s.Snapshot()
	snap := s.Tick()
	assert.Len(t, snap.Symbols, len(before.Symbols))
	for i, sym := range snap.Symbols {
		assert.Equal(t, before.Symbols[i].Price, sym.PrevPrice, "PrevPrice should be last tick's Price")
		assert.Greater(t, sym.Price, 0.0, "price must stay positive")
	}
}

func TestInitialSymbolsNonEmpty(t *testing.T) {
	assert.NotEmpty(t, ticker.InitialSymbols())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/ticker/ -run 'TestSimulator|TestTick|TestInitialSymbols'`
Expected: FAIL — `undefined: ticker.NewSimulator`.

- [ ] **Step 3: Write minimal implementation**

Create `internal/examples/ticker/sim.go`:

```go
package ticker

import "math/rand"

// Simulator advances a fixed set of symbols via a seeded random walk. It is
// deterministic for a given seed, which keeps the example offline and lets E2E
// rely on reproducible behaviour.
type Simulator struct {
	symbols []Symbol
	rng     *rand.Rand
}

// NewSimulator returns a Simulator seeded with the default symbol set.
func NewSimulator(seed int64) *Simulator {
	return &Simulator{
		symbols: InitialSymbols(),
		rng:     rand.New(rand.NewSource(seed)),
	}
}

// InitialSymbols is the starting watchlist, also used for the first server-side
// render before the SSE stream sends its first snapshot.
func InitialSymbols() []Symbol {
	return []Symbol{
		{Ticker: "AAPL", Name: "Apple Inc.", Price: 190.00, PrevPrice: 190.00},
		{Ticker: "GOOG", Name: "Alphabet Inc.", Price: 140.00, PrevPrice: 140.00},
		{Ticker: "MSFT", Name: "Microsoft Corp.", Price: 420.00, PrevPrice: 420.00},
		{Ticker: "AMZN", Name: "Amazon.com Inc.", Price: 178.00, PrevPrice: 178.00},
		{Ticker: "TSLA", Name: "Tesla Inc.", Price: 250.00, PrevPrice: 250.00},
		{Ticker: "NVDA", Name: "NVIDIA Corp.", Price: 120.00, PrevPrice: 120.00},
	}
}

// Tick advances every symbol one step (±1% random walk, floored at 1) and
// returns the new snapshot.
func (s *Simulator) Tick() Snapshot {
	for i := range s.symbols {
		prev := s.symbols[i].Price
		delta := (s.rng.Float64()*2 - 1) * 0.01 * prev
		next := prev + delta
		if next < 1 {
			next = 1
		}
		s.symbols[i].PrevPrice = prev
		s.symbols[i].Price = next
	}
	return s.snapshot()
}

// Snapshot returns the current state without advancing.
func (s *Simulator) Snapshot() Snapshot {
	return s.snapshot()
}

func (s *Simulator) snapshot() Snapshot {
	out := make([]Symbol, len(s.symbols))
	copy(out, s.symbols)
	return Snapshot{Symbols: out}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/ticker/`
Expected: PASS (all tests in the package).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/ticker/sim.go internal/examples/ticker/sim_test.go
git commit -m "feat(ticker): seeded deterministic price simulator"
```

---

## Task 3: Broker (pub/sub fan-out)

**Files:**
- Create: `internal/examples/ticker/broker.go`
- Test: `internal/examples/ticker/broker_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/examples/ticker/broker_test.go`:

```go
package ticker_test

import (
	"context"
	"testing"
	"time"

	"github.com/araihu/goshtoso/internal/examples/ticker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrokerBroadcastsToSubscribers(t *testing.T) {
	b := ticker.NewBroker(ticker.NewSimulator(1), 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)

	ch, unsub := b.Subscribe()
	defer unsub()

	select {
	case snap := <-ch:
		assert.NotEmpty(t, snap.Symbols)
	case <-time.After(time.Second):
		t.Fatal("expected a snapshot within 1s")
	}
}

func TestBrokerUnsubscribeClosesChannel(t *testing.T) {
	b := ticker.NewBroker(ticker.NewSimulator(1), 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)

	ch, unsub := b.Subscribe()
	unsub()
	// Drain any buffered value, then confirm the channel is closed.
	for range ch { //nolint:revive // intentional drain-until-closed
	}
}

func TestBrokerSlowSubscriberDoesNotBlockOthers(t *testing.T) {
	b := ticker.NewBroker(ticker.NewSimulator(1), 5*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)

	_, slowUnsub := b.Subscribe() // never drained
	defer slowUnsub()
	fast, fastUnsub := b.Subscribe()
	defer fastUnsub()

	got := 0
	deadline := time.After(time.Second)
	for got < 3 {
		select {
		case <-fast:
			got++
		case <-deadline:
			t.Fatalf("fast subscriber only got %d snapshots; slow one blocked the hub", got)
		}
	}
}

func TestBrokerFindReturnsCurrentSymbol(t *testing.T) {
	b := ticker.NewBroker(ticker.NewSimulator(1), time.Second)
	sym, ok := b.Find("AAPL")
	require.True(t, ok)
	assert.Equal(t, "AAPL", sym.Ticker)

	_, ok = b.Find("NOPE")
	assert.False(t, ok)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/ticker/ -run TestBroker`
Expected: FAIL — `undefined: ticker.NewBroker`.

- [ ] **Step 3: Write minimal implementation**

Create `internal/examples/ticker/broker.go`:

```go
package ticker

import (
	"context"
	"sync"
	"time"
)

// Broker runs a single Simulator on an interval and fans each snapshot out to
// all subscribers. One shared stream for every viewer → no per-user state.
type Broker struct {
	mu          sync.Mutex
	subscribers map[chan Snapshot]struct{}
	sim         *Simulator
	interval    time.Duration
}

// NewBroker creates a Broker over sim, ticking every interval. Call Run in a
// goroutine to start broadcasting.
func NewBroker(sim *Simulator, interval time.Duration) *Broker {
	return &Broker{
		subscribers: make(map[chan Snapshot]struct{}),
		sim:         sim,
		interval:    interval,
	}
}

// Run ticks until ctx is cancelled, broadcasting each snapshot to subscribers.
func (b *Broker) Run(ctx context.Context) {
	t := time.NewTicker(b.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			b.broadcast(b.tick())
		}
	}
}

// tick advances the simulator under the lock so Find/Snapshot stay consistent
// with what was just broadcast.
func (b *Broker) tick() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sim.Tick()
}

func (b *Broker) broadcast(snap Snapshot) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subscribers {
		select {
		case ch <- snap:
		default:
			// Slow subscriber: drop this tick rather than block the hub.
		}
	}
}

// Subscribe returns a buffered channel of snapshots and an unsubscribe func.
// The unsubscribe func is idempotent and closes the channel.
func (b *Broker) Subscribe() (<-chan Snapshot, func()) {
	ch := make(chan Snapshot, 4)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subscribers, ch)
			close(ch)
			b.mu.Unlock()
		})
	}
	return ch, unsub
}

// Snapshot returns the current state without advancing.
func (b *Broker) Snapshot() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sim.Snapshot()
}

// Find returns the current symbol with the given ticker.
func (b *Broker) Find(tkr string) (Symbol, bool) {
	for _, s := range b.Snapshot().Symbols {
		if s.Ticker == tkr {
			return s, true
		}
	}
	return Symbol{}, false
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/ticker/ -race`
Expected: PASS, no data races.

- [ ] **Step 5: Commit**

```bash
git add internal/examples/ticker/broker.go internal/examples/ticker/broker_test.go
git commit -m "feat(ticker): pub/sub broker with slow-subscriber drop"
```

---

## Task 4: Vendor the htmx SSE extension

**Files:**
- Create: `assets/js/vendor/sse.js`
- Modify: `assets/embed.go` (doc comment only), `internal/pages/demo/layout.templ`

- [ ] **Step 1: Download the extension matching htmx 2.x**

Run:
```bash
curl -sSL https://unpkg.com/htmx-ext-sse@2.2.3/dist/sse.js -o assets/js/vendor/sse.js
test -s assets/js/vendor/sse.js && echo OK
```
Expected: `OK` and a non-empty file. (If the network is unavailable, fetch `sse.js` from the `htmx-ext-sse@2.2.x` release by any means and place it at this path — it must be the un-minified or minified extension source, served locally so runtime needs no CDN.)

- [ ] **Step 2: Confirm the event names this plan relies on**

Run:
```bash
grep -o 'htmx:sse[A-Za-z]*' assets/js/vendor/sse.js | sort -u
```
Expected: includes `htmx:sseOpen` and `htmx:sseBeforeMessage`.
If the vendored version dispatches different names (e.g. kebab-case `htmx:sse-before-message`), note them — Task 6's Alpine script (`connect`) must listen for the exact names this grep prints. Adjust the two `addEventListener` calls accordingly before finishing Task 6.

- [ ] **Step 3: Add the script tag to the layout `<head>`**

In `internal/pages/demo/layout.templ`, after the htmx line (currently line 33, `<script src="/assets/js/vendor/htmx.min.js"></script>`), add:

```templ
			<script defer src="/assets/js/vendor/sse.js"></script>
```

(`defer` is fine — the SSE-connected element is processed by htmx after load; the ext only needs to be registered before htmx processes `hx-ext="sse"`, which the page-bottom rendering guarantees.)

- [ ] **Step 4: Document the file in embed.go**

In `assets/embed.go`, add this line to the package doc comment, right after the `htmx.min.js` bullet (line 12):

```go
//   - /assets/js/vendor/sse.js — HTMX Server-Sent Events extension
```

(The `//go:embed styles.css js fonts images` directive already embeds everything under `js/`, so no directive change is needed.)

- [ ] **Step 5: Regenerate templ + build**

Run:
```bash
templ generate
go build ./...
```
Expected: no errors; `internal/pages/demo/layout_templ.go` now emits the new script tag.

- [ ] **Step 6: Commit**

```bash
git add assets/js/vendor/sse.js assets/embed.go internal/pages/demo/layout.templ internal/pages/demo/layout_templ.go
git commit -m "feat(ticker): vendor htmx SSE extension locally"
```

---

## Task 5: Ticker view (templ components + table/format helpers)

**Files:**
- Create: `internal/pages/demo/examples/ticker_table.go`
- Create: `internal/pages/demo/examples/ticker.templ`

- [ ] **Step 1: Write the Go helpers (formatting, badge variant, table config)**

Create `internal/pages/demo/examples/ticker_table.go`:

```go
package examples

import (
	"fmt"

	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/internal/examples/ticker"
)

// tickerPrice formats a symbol's price as a fixed 2-decimal string.
func tickerPrice(s ticker.Symbol) string {
	return fmt.Sprintf("%.2f", s.Price)
}

// tickerChange formats the percent change with a sign, e.g. "+0.42%".
func tickerChange(s ticker.Symbol) string {
	return fmt.Sprintf("%+.2f%%", s.ChangePct())
}

// tickerBadgeVariant maps price direction to a badge color.
func tickerBadgeVariant(s ticker.Symbol) badge.Variant {
	switch s.Direction() {
	case "up":
		return badge.Success
	case "down":
		return badge.Danger
	default:
		return badge.Default
	}
}

// tickerTableConfig builds the live board: each price cell renders a span that
// subscribes to its symbol's SSE event; each row's click loads the spotlight.
func tickerTableConfig(symbols []ticker.Symbol) table.Config {
	rows := make([]table.Row, 0, len(symbols))
	for _, sym := range symbols {
		rows = append(rows, table.Row{
			ID:       sym.Ticker,
			HXGet:    "/api/examples/ticker/spotlight?symbol=" + sym.Ticker,
			HXTarget: "#ticker-spotlight",
			HXSwap:   "innerHTML",
			Cells: map[string]table.Cell{
				"symbol": {Text: sym.Ticker, Description: sym.Name},
				"price":  {Component: TickerCell(sym)},
			},
		})
	}
	return table.Config{
		ID: "ticker-table",
		Columns: []table.Column{
			{Key: "symbol", Label: "Symbol"},
			{Key: "price", Label: "Price", Align: "right"},
		},
		Rows: rows,
	}
}
```

- [ ] **Step 2: Write the templ components**

Create `internal/pages/demo/examples/ticker.templ`:

```templ
package examples

import (
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/card"
	"github.com/araihu/goshtoso/components/spinner"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/components/toggle"
	"github.com/araihu/goshtoso/internal/examples/ticker"
)

// TickerContent is the registry entry point for /examples/ticker.
templ TickerContent() {
	@tickerPane(ticker.InitialSymbols())
}

// tickerPane is the page shell: header (spinner + pause toggle), the SSE-connected
// board, and the spotlight slot. The Alpine component lives on the outer div so
// the spinner/toggle (outside the SSE root) share its connected/paused state.
templ tickerPane(symbols []ticker.Symbol) {
	<div id="ticker-fragment" x-data="tickerPane()" class="mx-auto max-w-4xl">
		<div class="mb-4 flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold text-on-surface dark:text-on-surface-dark">Live Ticker</h1>
				<p class="text-sm text-on-surface-muted dark:text-on-surface-dark-muted">
					Simulated prices streamed over Server-Sent Events.
				</p>
			</div>
			<div class="flex items-center gap-4">
				<span x-show="!connected" class="inline-flex items-center gap-2 text-sm text-on-surface-muted dark:text-on-surface-dark-muted">
					@spinner.Spinner(spinner.Config{Size: spinner.SizeSM})
					Connecting…
				</span>
				<div x-on:change="paused = $event.target.checked">
					@toggle.Toggle(toggle.Config{ID: "ticker-pause", Label: "Pause"})
				</div>
			</div>
		</div>
		<div
			x-ref="sse"
			x-init="$nextTick(() => connect($refs.sse))"
			hx-ext="sse"
			sse-connect="/api/examples/ticker/stream"
		>
			@table.Table(tickerTableConfig(symbols))
			<div id="ticker-spotlight" class="mt-6">
				@tickerSpotlightEmpty()
			</div>
		</div>
		@tickerPaneScript()
	</div>
}

// TickerCell is the live price cell placed in the table via Cell.Component. The
// span subscribes to the symbol's SSE event; ticks replace its inner content.
templ TickerCell(sym ticker.Symbol) {
	<span id={ "ticker-cell-" + sym.Ticker } sse-swap={ sym.Ticker } hx-swap="innerHTML">
		@TickerCellInner(sym)
	</span>
}

// TickerCellInner is the price + change-badge unit. It is rendered both at first
// paint and on every SSE tick (the handler pushes exactly this fragment).
templ TickerCellInner(sym ticker.Symbol) {
	<span class="inline-flex items-center justify-end gap-2">
		<span class="font-mono tabular-nums">{ tickerPrice(sym) }</span>
		@badge.Badge(badge.Config{
			Variant: tickerBadgeVariant(sym),
			Style:   badge.StyleSoft,
			Size:    badge.SizeSM,
			Text:    tickerChange(sym),
		})
	</span>
}

// TickerSpotlight is the card shown when a row is clicked. Its price span shares
// the same SSE event as the table cell, so it live-updates off the one stream.
templ TickerSpotlight(sym ticker.Symbol) {
	@card.Card(card.Config{
		Tag:         sym.Ticker,
		Title:       sym.Name,
		Description: "Live price",
		Footer:      tickerSpotlightPrice(sym),
	})
}

templ tickerSpotlightPrice(sym ticker.Symbol) {
	<div class="text-lg">
		<span id={ "ticker-card-" + sym.Ticker } sse-swap={ sym.Ticker } hx-swap="innerHTML">
			@TickerCellInner(sym)
		</span>
	</div>
}

// tickerSpotlightEmpty is the placeholder before any row is selected.
templ tickerSpotlightEmpty() {
	<p class="text-sm text-on-surface-muted dark:text-on-surface-dark-muted">
		Select a symbol to spotlight it.
	</p>
}

// tickerPaneScript registers the Alpine component. It uses templ.Raw + a
// <script> block (not an x-data string) to avoid templ attribute-escaping, and
// registers eagerly when Alpine is already running (fragment nav) — otherwise on
// alpine:init. `connect` attaches the SSE listeners that drive connected/paused.
templ tickerPaneScript() {
	@templ.Raw("<script>" + tickerPaneJS + "</script>")
}
```

- [ ] **Step 3: Add the JS constant**

Append to `internal/pages/demo/examples/ticker_table.go`:

```go
// tickerPaneJS registers the Alpine component for the ticker pane. `connected`
// hides the spinner once the stream is live; `paused` cancels swaps via the
// cancelable htmx:sseBeforeMessage event so the connection stays open while
// paused. Event names must match the vendored sse.js (see Task 4, Step 2).
const tickerPaneJS = `(() => {
	const register = () => {
		Alpine.data('tickerPane', () => ({
			connected: false,
			paused: false,
			connect(el) {
				if (!el) return;
				el.addEventListener('htmx:sseOpen', () => { this.connected = true; });
				el.addEventListener('htmx:sseBeforeMessage', (e) => {
					this.connected = true;
					if (this.paused) e.preventDefault();
				});
			},
		}));
	};
	if (window.Alpine && window.Alpine.version) {
		register();
	} else {
		document.addEventListener('alpine:init', register);
	}
})();`
```

- [ ] **Step 4: Generate templ + build**

Run:
```bash
rm -f internal/pages/demo/examples/ticker_templ.go
templ generate
go build ./...
```
Expected: no errors; `internal/pages/demo/examples/ticker_templ.go` exists.

- [ ] **Step 5: Commit**

```bash
git add internal/pages/demo/examples/ticker.templ internal/pages/demo/examples/ticker_templ.go internal/pages/demo/examples/ticker_table.go
git commit -m "feat(ticker): templ view (live cells, spotlight card, Alpine pane)"
```

---

## Task 6: Server handlers + routing

**Files:**
- Create: `internal/server/ticker_handler.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Add the broker to the Server struct and start it in New**

In `internal/server/server.go`, add the import (in the existing import block):

```go
	"github.com/araihu/goshtoso/internal/examples/ticker"
```

and (since `New` will start a goroutine) these to the existing imports if not present: `"context"` is already imported; add `"os"`, `"time"`, `"strconv"` — `strconv`, `time` are already imported, so add only:

```go
	"os"
```

Change the `Server` struct (currently lines 17–20) to:

```go
// Server handles HTTP requests for Goshtoso components
type Server struct {
	projectRoot  string
	mux          *http.ServeMux
	tickerBroker *ticker.Broker
}
```

Change `New` (currently lines 23–30) to:

```go
// New creates a new server instance
func New(projectRoot string) *Server {
	interval := time.Second
	if v := os.Getenv("GOSHTOSO_TICKER_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			interval = time.Duration(ms) * time.Millisecond
		}
	}
	broker := ticker.NewBroker(ticker.NewSimulator(1), interval)
	// Process-lifetime stream shared by all viewers; never cancelled.
	go broker.Run(context.Background())

	s := &Server{
		projectRoot:  projectRoot,
		mux:          http.NewServeMux(),
		tickerBroker: broker,
	}
	s.setupRoutes()
	return s
}
```

- [ ] **Step 2: Register the routes and the example page case**

In `internal/server/server.go`, in `setupRoutes`, just after `s.registerTodoRoutes()` (line ~48), add:

```go
	s.registerTickerRoutes()
```

In `handleExample`'s `switch sub` (currently the `case "todo":` block), add a case:

```go
	case "ticker":
		s.renderDemo(w, r, "examples/ticker")
```

- [ ] **Step 3: Write the handler file**

Create `internal/server/ticker_handler.go`:

```go
package server

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/araihu/goshtoso/internal/examples/ticker"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// registerTickerRoutes wires the /api/examples/ticker/* endpoints.
func (s *Server) registerTickerRoutes() {
	s.mux.HandleFunc("/api/examples/ticker/stream", s.handleTickerStream)
	s.mux.HandleFunc("/api/examples/ticker/spotlight", s.handleTickerSpotlight)
}

// handleTickerStream is the SSE endpoint. It subscribes to the shared broker and
// emits one named event per symbol (event name = ticker) on each tick, until the
// client disconnects.
func (s *Server) handleTickerStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)
	ch, unsubscribe := s.tickerBroker.Subscribe()
	defer unsubscribe()

	// Send the current state immediately so cells populate before the first tick.
	writeTickerSnapshot(r, w, rc, s.tickerBroker.Snapshot())

	for {
		select {
		case <-r.Context().Done():
			return
		case snap, ok := <-ch:
			if !ok {
				return
			}
			writeTickerSnapshot(r, w, rc, snap)
		}
	}
}

// handleTickerSpotlight returns the spotlight card for the requested symbol.
func (s *Server) handleTickerSpotlight(w http.ResponseWriter, r *http.Request) {
	sym, ok := s.tickerBroker.Find(r.URL.Query().Get("symbol"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.TickerSpotlight(sym).Render(r.Context(), w)
}

// writeTickerSnapshot renders each symbol's price fragment and emits it as a
// named SSE event, then flushes.
func writeTickerSnapshot(r *http.Request, w http.ResponseWriter, rc *http.ResponseController, snap ticker.Snapshot) {
	for _, sym := range snap.Symbols {
		var buf bytes.Buffer
		_ = examples.TickerCellInner(sym).Render(r.Context(), &buf)
		writeSSEEvent(w, sym.Ticker, buf.String())
	}
	_ = rc.Flush()
}

// writeSSEEvent writes a single named SSE event. Multi-line HTML is split into
// multiple `data:` lines per the SSE spec.
func writeSSEEvent(w http.ResponseWriter, event, data string) {
	_, _ = fmt.Fprintf(w, "event: %s\n", event)
	for _, line := range strings.Split(data, "\n") {
		_, _ = fmt.Fprintf(w, "data: %s\n", line)
	}
	_, _ = fmt.Fprint(w, "\n")
}
```

- [ ] **Step 4: Build**

Run: `go build ./...`
Expected: no errors. (If `os` is reported as already imported or unused, reconcile the import block — `time`, `strconv`, `context` are pre-existing; only `os` and the `ticker` package are new.)

- [ ] **Step 5: Commit**

```bash
git add internal/server/ticker_handler.go internal/server/server.go
git commit -m "feat(ticker): SSE stream + spotlight handlers, shared broker on Server"
```

---

## Task 7: Registry, index card, sidebar

**Files:**
- Modify: `internal/pages/demo/components/registry.go`, `internal/pages/demo/examples/index.templ`, `internal/pages/demo/layout.templ`

- [ ] **Step 1: Add the registry entry**

In `internal/pages/demo/components/registry.go`, in the `Demos` map, immediately after the `"examples/todo":` line, add:

```go
	"examples/ticker":        {"Live Ticker", "ticker", examples.TickerContent},
```

- [ ] **Step 2: Add the examples index card**

In `internal/pages/demo/examples/index.templ`, inside the `<div class="grid gap-4 sm:grid-cols-2">`, after the existing todo `@exampleCard(...)` line, add:

```templ
			@exampleCard("/examples/ticker", "Live Ticker", "A streaming stock-market watchlist powered by Server-Sent Events.")
```

- [ ] **Step 3: Add the sidebar item**

In `internal/pages/demo/layout.templ`, in `getSidebarSections`, inside the `Examples` section's `Items`, after the `sItem("todo", ...)` line, add:

```go
				sItem("ticker", "Live Ticker", "/examples/ticker", activeComponent),
```

- [ ] **Step 4: Generate + build + run unit tests**

Run:
```bash
templ generate
go build ./...
go test ./internal/... -count=1
```
Expected: build clean; all unit tests pass.

- [ ] **Step 5: Smoke-test the page manually (optional but recommended)**

Run (in one shell): `go run cmd/server/main.go`
Then: `curl -s http://localhost:8090/examples/ticker | grep -c 'sse-connect'`
Expected: `1` (the page renders with the SSE root). Stop the server afterward.

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/components/registry.go internal/pages/demo/examples/index.templ internal/pages/demo/examples/index_templ.go internal/pages/demo/layout.templ internal/pages/demo/layout_templ.go
git commit -m "feat(ticker): register ticker example (registry, index card, sidebar)"
```

---

## Task 8: E2E tests

**Files:**
- Create: `tests/e2e/ticker_test.go`

- [ ] **Step 1: Write the E2E tests**

Create `tests/e2e/ticker_test.go`:

```go
package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// waitForCellText returns the current text of a ticker price cell after it has
// rendered a price.
func tickerCellText(t *testing.T, page playwright.Page, symbol string) string {
	t.Helper()
	sel := "#ticker-cell-" + symbol
	_, err := page.WaitForFunction(
		"(s) => { const el = document.querySelector(s); return el && el.textContent.trim().length > 0; }",
		playwright.PageWaitForFunctionOptions{Arg: sel, Timeout: playwright.Float(6000)})
	require.NoError(t, err)
	txt, err := page.Locator(sel).TextContent()
	require.NoError(t, err)
	return txt
}

// TestTicker_StreamUpdatesCells loads the page and asserts a price cell changes
// as ticks stream in.
func TestTicker_StreamUpdatesCells(t *testing.T) {
	page := newPage(t, browser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	before := tickerCellText(t, page, "AAPL")
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() !== b; }",
		playwright.PageWaitForFunctionOptions{Arg: before, Timeout: playwright.Float(6000)})
	require.NoError(t, err, "price cell should change as SSE ticks arrive")
}

// TestTicker_SpotlightOnRowClick clicks a row and asserts the spotlight card
// rebinds to that symbol.
func TestTicker_SpotlightOnRowClick(t *testing.T) {
	page := newPage(t, browser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	tickerCellText(t, page, "MSFT") // ensure rows rendered

	require.NoError(t, page.Locator("#ticker-table tr#MSFT").First().Click())
	_, err = page.WaitForFunction(
		"() => { const el = document.querySelector('#ticker-spotlight'); return el && el.textContent.includes('Microsoft'); }",
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(4000)})
	require.NoError(t, err, "spotlight should show the clicked symbol")

	// The card's price span also subscribes to the symbol's SSE event.
	require.NoError(t, page.Locator("#ticker-card-MSFT").WaitFor())
}

// TestTicker_PauseStopsUpdates verifies pause freezes the cells and resume
// restarts them. Pause works by cancelling SSE swaps, not closing the stream.
func TestTicker_PauseStopsUpdates(t *testing.T) {
	page := newPage(t, browser)
	_, err := page.Goto(baseURL + "/examples/ticker")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	tickerCellText(t, page, "AAPL")

	// Pause via the toggle.
	require.NoError(t, page.Locator("label[for='ticker-pause']").Click())
	// Let the paused flag settle, then capture.
	_, err = page.WaitForFunction(
		"() => Alpine.$data(document.querySelector('#ticker-fragment')).paused === true", nil)
	require.NoError(t, err)
	paused := tickerCellText(t, page, "AAPL")

	// Across a couple of tick intervals the text must NOT change while paused.
	stable, err := page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() === b; }",
		playwright.PageWaitForFunctionOptions{Arg: paused, Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	require.NotNil(t, stable)

	// Resume and confirm updates restart.
	require.NoError(t, page.Locator("label[for='ticker-pause']").Click())
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() !== b; }",
		playwright.PageWaitForFunctionOptions{Arg: paused, Timeout: playwright.Float(6000)})
	require.NoError(t, err, "updates should resume after unpausing")
}

// TestTicker_FragmentNavNoErrors lands elsewhere, navigates to the ticker via the
// sidebar (htmx fragment swap), and asserts no console/page errors plus a live
// cell update — the SPA path a direct load can't catch.
func TestTicker_FragmentNavNoErrors(t *testing.T) {
	page := newPage(t, browser)

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

	require.NoError(t, page.Locator("a[href='/examples/ticker']").First().Click())
	before := tickerCellText(t, page, "AAPL")
	_, err = page.WaitForFunction(
		"(b) => { const el = document.querySelector('#ticker-cell-AAPL'); return el && el.textContent.trim() !== b; }",
		playwright.PageWaitForFunctionOptions{Arg: before, Timeout: playwright.Float(6000)})
	require.NoError(t, err, "SSE should update cells after fragment nav")

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav ticker page: %v", jsErrors)
}
```

- [ ] **Step 2: Verify selector assumptions against the rendered table**

The tests assume each row's `<tr>` has `id="<TICKER>"` (from `table.Row.ID`). Confirm the table component renders `Row.ID` as the `<tr>` id:

Run: `grep -n 'id=' components/table/table.templ | head`
Expected: a row line emitting `id={ row.ID }` (or equivalent). If the table instead prefixes the id (e.g. `row-<ID>`) or omits it on `<tr>`, update the `#ticker-table tr#MSFT` selector in `TestTicker_SpotlightOnRowClick` to match the real markup. If `Row.ID` is only used for the checkbox input, switch the click target to `#ticker-table tr:has(#ticker-cell-MSFT)`.

- [ ] **Step 3: Run the ticker E2E tests**

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestTicker`
Expected: all four `TestTicker_*` PASS. If `TestTicker_StreamUpdatesCells` is flaky under the 1s default interval, set a faster interval for the suite by exporting `GOSHTOSO_TICKER_MS=300` before the run (note: only effective if TestMain forwards env to the server subprocess — see Step 4).

- [ ] **Step 4: (If needed) speed up ticks for E2E**

If the stream-update test needs a faster tick than the 1s default, check how `tests/e2e/e2e_test.go` starts the server and ensure the server subprocess inherits `GOSHTOSO_TICKER_MS`. `os/exec` inherits the parent env by default, so `GOSHTOSO_TICKER_MS=300 go test ...` should already reach the server. Document the variable in the e2e run command if you rely on it.

- [ ] **Step 5: Commit**

```bash
git add tests/e2e/ticker_test.go
git commit -m "test(ticker): e2e for stream update, spotlight, pause/resume, fragment nav"
```

---

## Task 9: Full verification, lint, docs sync

**Files:** none new — verification + any doc regen.

- [ ] **Step 1: Regenerate everything and build the binary**

Run:
```bash
templ generate
tailwindcss -i css/main.css -o assets/styles.css
go build -o bin/server ./cmd/server
```
Expected: no errors. (Tailwind rebuild picks up any new utility classes used in `ticker.templ`, e.g. `tabular-nums`.)

- [ ] **Step 2: Lint**

Run: `golangci-lint run`
Expected: clean. Watch the cyclomatic ceiling (20) — the new functions are all small; if `New` trips it, extract the interval-parsing into a helper `tickerInterval() time.Duration`.

- [ ] **Step 3: Sync the usage skill (only if a component entry point changed)**

This example consumes components but does not change any `components/**/types.go` or entry points, so the generated reference should be unaffected. Run the generator to be safe:

Run: `go run ./scripts/skillgen`
Expected: no diff to `.claude/skills/using-goshtoso/components-reference.md`. If there is a diff, something unexpected changed — investigate before committing.

- [ ] **Step 4: Run the full E2E suite**

Run: `go test ./tests/e2e/... -count=1 -timeout 15m`
Expected: all tests pass, no skips (existing 381 + the new ticker tests).

- [ ] **Step 5: Commit any regenerated artifacts**

```bash
git add -A
git commit -m "chore(ticker): regenerate styles.css + templ after ticker example" || echo "nothing to commit"
```

- [ ] **Step 6: Codex review (repo gate)**

Per CLAUDE.md, hand the branch diff to Codex for an independent review before finishing (`codex:rescue` skill or the stop-time gate). Address findings, re-run the relevant tests, then proceed to branch finishing (`superpowers:finishing-a-development-branch`).

---

## Notes / Known Risks

- **Pause approach:** chosen mechanism is cancelling swaps via `htmx:sseBeforeMessage` `preventDefault()` — the EventSource stays open. This is far more reliable than adding/removing `sse-connect`. The only dependency is the exact event names dispatched by the vendored `sse.js` (verified in Task 4 Step 2; adjust the two listeners if they differ).
- **Broker lifetime:** the broker runs for the process lifetime (`context.Background()`), appropriate for a demo server. Per-client SSE handlers correctly unsubscribe on disconnect (`r.Context().Done()` + `defer unsubscribe()`), so there is no per-connection leak.
- **Determinism:** the simulator is seeded (seed `1`); the *sequence* is reproducible, but E2E asserts that a cell *changes* and that pause *freezes* it — never exact float values — so tests stay robust.
- **Tick rate:** default 1s; override with `GOSHTOSO_TICKER_MS` for faster E2E.
```
