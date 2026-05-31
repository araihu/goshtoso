package chat

import (
	"fmt"
	"testing"
)

func drain(c *Client) []string {
	var out []string
	for {
		select {
		case m, ok := <-c.Send:
			if !ok {
				return out
			}
			out = append(out, string(m))
		default:
			return out
		}
	}
}

func TestHub_BroadcastReachesAllClients(t *testing.T) {
	h := NewHub(50)
	a := NewClient()
	b := NewClient()
	h.Register(a)
	h.Register(b)

	h.Broadcast([]byte("hi"))

	if got := drain(a); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("client a got %v", got)
	}
	if got := drain(b); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("client b got %v", got)
	}
}

func TestHub_Count(t *testing.T) {
	h := NewHub(50)
	if h.Count() != 0 {
		t.Fatalf("empty hub count = %d", h.Count())
	}
	a := NewClient()
	h.Register(a)
	if h.Count() != 1 {
		t.Fatalf("after one register count = %d", h.Count())
	}
	h.Unregister(a)
	if h.Count() != 0 {
		t.Fatalf("after unregister count = %d", h.Count())
	}
}

func TestHub_BacklogReplaysRecent(t *testing.T) {
	h := NewHub(2) // ring caps at 2
	h.Broadcast([]byte("m1"))
	h.Broadcast([]byte("m2"))
	h.Broadcast([]byte("m3")) // evicts m1

	got := h.Backlog()
	if len(got) != 2 || string(got[0]) != "m2" || string(got[1]) != "m3" {
		t.Fatalf("backlog = %q", got)
	}
}

func TestHub_UnregisterClosesSend(t *testing.T) {
	h := NewHub(50)
	a := NewClient()
	h.Register(a)
	h.Unregister(a)
	if _, ok := <-a.Send; ok {
		t.Fatalf("Send channel should be closed after Unregister")
	}
}

func TestHub_JoinReplaysBacklog(t *testing.T) {
	h := NewHub(50)
	// Push more frames than the old 16-deep send buffer so a too-small buffer
	// (or a Register-then-Backlog replay) would have dropped the oldest ones.
	const n = 30
	for i := range n {
		h.Broadcast(fmt.Appendf(nil, "m%d", i))
	}

	c := NewClient()
	h.Join(c)

	// Join must both register the client...
	if h.Count() != 1 {
		t.Fatalf("after Join count = %d, want 1", h.Count())
	}
	// ...and replay every ring frame, in order.
	got := drain(c)
	if len(got) != n {
		t.Fatalf("replayed %d frames, want %d", len(got), n)
	}
	for i := range n {
		want := fmt.Sprintf("m%d", i)
		if got[i] != want {
			t.Fatalf("frame %d = %q, want %q", i, got[i], want)
		}
	}
}

func TestHub_BroadcastEphemeralNotInRing(t *testing.T) {
	h := NewHub(50)
	c := NewClient()
	h.Register(c)

	h.BroadcastEphemeral([]byte("presence"))

	// Reaches the current client...
	if got := drain(c); len(got) != 1 || got[0] != "presence" {
		t.Fatalf("client got %v, want [presence]", got)
	}
	// ...but is never retained in the ring, so future joiners don't replay it.
	if got := h.Backlog(); len(got) != 0 {
		t.Fatalf("backlog = %q, want empty", got)
	}
}

func TestHub_SlowClientDoesNotBlock(t *testing.T) {
	h := NewHub(50)
	slow := NewClient()
	h.Register(slow)
	// Fill the slow client's buffer beyond capacity; Broadcast must not block.
	for i := 0; i < sendBuffer+5; i++ {
		h.Broadcast([]byte("flood"))
	}
	// If we reach here without deadlock, the drop-on-full path works.
	// The buffer must have been capped at sendBuffer, not grown unbounded.
	if len(slow.Send) != sendBuffer {
		t.Fatalf("slow client buffer = %d, want %d", len(slow.Send), sendBuffer)
	}
}
