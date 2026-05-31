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
