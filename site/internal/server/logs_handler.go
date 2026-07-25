package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/araihu/goshtoso/site/internal/examples/logs"
	"github.com/araihu/goshtoso/site/internal/pages/demo/examples"
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
	w.Header().Set("X-Accel-Buffering", "no")
	interval := durParam(r, "interval", 500*time.Millisecond)
	maxN := min(
		// 0 = unbounded; intParam is defined in todo_handler.go
		max(

			intParam(r, "max", 0), 0), 10000)
	streamLogs(r.Context(), w, flusher, interval, maxN)
}

// streamLogs is the testable core: emit synthetic LogRow SSE events on a ticker.
func streamLogs(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, interval time.Duration, maxN int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for i := 0; ; i++ {
		if maxN > 0 && i >= maxN {
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
			if err := writeSSEMessage(w, buf.String()); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// writeSSEMessage frames an HTML fragment as one SSE `message` event. templ
// output is multi-line, but a `data:` field is single-line, so each output
// line becomes its own `data:` field; the htmx SSE ext rejoins them with "\n".
func writeSSEMessage(w http.ResponseWriter, html string) error {
	if _, err := fmt.Fprint(w, "event: message\n"); err != nil {
		return err
	}
	for line := range strings.SplitSeq(strings.TrimRight(html, "\n"), "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}
