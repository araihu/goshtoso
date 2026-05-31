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
// multiple data: lines per the SSE spec.
func writeSSEEvent(w http.ResponseWriter, event, data string) {
	_, _ = fmt.Fprintf(w, "event: %s\n", event)
	for _, line := range strings.Split(data, "\n") {
		_, _ = fmt.Fprintf(w, "data: %s\n", line)
	}
	_, _ = fmt.Fprint(w, "\n")
}
