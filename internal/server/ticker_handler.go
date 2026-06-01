package server

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/internal/examples/ticker"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// registerTickerRoutes wires the /api/examples/ticker/* endpoints.
func (s *Server) registerTickerRoutes() {
	s.mux.HandleFunc("/api/examples/ticker/stream", s.handleTickerStream)
	s.mux.HandleFunc("/api/examples/ticker/rows", s.handleTickerRows)
}

// handleTickerStream is the SSE endpoint. It subscribes to the shared broker and
// emits one named event per symbol (event name = ticker) on each tick, until the
// client disconnects.
func (s *Server) handleTickerStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Disable proxy buffering (nginx etc.) so each flushed event is forwarded
	// immediately instead of being held back, which would freeze the stream.
	w.Header().Set("X-Accel-Buffering", "no")

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

// handleTickerRows returns the filtered table rows (tbody innerHTML) for the
// search box. Rows carry live SSE price cells; htmx re-binds the swapped-in
// spans to the existing stream after the swap.
func (s *Server) handleTickerRows(w http.ResponseWriter, r *http.Request) {
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	snap := s.tickerBroker.Snapshot()
	matched := make([]ticker.Symbol, 0, len(snap.Symbols))
	for _, sym := range snap.Symbols {
		if search == "" ||
			strings.Contains(strings.ToLower(sym.Ticker), search) ||
			strings.Contains(strings.ToLower(sym.Name), search) {
			matched = append(matched, sym)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = table.TableRows(examples.TickerTableConfig(matched)).Render(r.Context(), w)
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
// multiple data: lines per the SSE spec.
func writeSSEEvent(w http.ResponseWriter, event, data string) {
	_, _ = fmt.Fprintf(w, "event: %s\n", event)
	for _, line := range strings.Split(data, "\n") {
		_, _ = fmt.Fprintf(w, "data: %s\n", line)
	}
	_, _ = fmt.Fprint(w, "\n")
}
