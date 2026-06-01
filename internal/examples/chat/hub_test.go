package chat

import (
	"fmt"
	"testing"
)

// drain returns the Text of every buffered Event on the client, non-blocking.
func drain(c *Client) []string {
	var out []string
	for {
		select {
		case ev, ok := <-c.Send:
			if !ok {
				return out
			}
			out = append(out, ev.Text)
		default:
			return out
		}
	}
}

func msg(text string) Event { return Event{Kind: EventMessage, Text: text} }

func TestHub_BroadcastReachesAllClients(t *testing.T) {
	h := NewHub(50)
	a := NewClient()
	b := NewClient()
	h.Register(a)
	h.Register(b)

	h.Broadcast(msg("hi"))

	if got := drain(a); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("client a got %v", got)
	}
	if got := drain(b); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("client b got %v", got)
	}
}

func TestHub_ClientsHaveUniqueConnIDs(t *testing.T) {
	a := NewClient()
	b := NewClient()
	if a.ConnID == b.ConnID {
		t.Fatalf("two clients share ConnID %d", a.ConnID)
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
	h.Broadcast(msg("m1"))
	h.Broadcast(msg("m2"))
	h.Broadcast(msg("m3")) // evicts m1

	got := h.Backlog()
	if len(got) != 2 || got[0].Text != "m2" || got[1].Text != "m3" {
		t.Fatalf("backlog = %v", got)
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
	// Push more events than the old 16-deep send buffer so a too-small buffer
	// (or a Register-then-Backlog replay) would have dropped the oldest ones.
	const n = 30
	for i := range n {
		h.Broadcast(msg(fmt.Sprintf("m%d", i)))
	}

	c := NewClient()
	h.Join(c)

	// Join must both register the client...
	if h.Count() != 1 {
		t.Fatalf("after Join count = %d, want 1", h.Count())
	}
	// ...and replay every ring event, in order.
	got := drain(c)
	if len(got) != n {
		t.Fatalf("replayed %d events, want %d", len(got), n)
	}
	for i := range n {
		want := fmt.Sprintf("m%d", i)
		if got[i] != want {
			t.Fatalf("event %d = %q, want %q", i, got[i], want)
		}
	}
}

func TestHub_BroadcastEphemeralNotInRing(t *testing.T) {
	h := NewHub(50)
	c := NewClient()
	h.Register(c)

	h.BroadcastEphemeral(Event{Kind: EventPresence, SystemText: "presence"})

	// Reaches the current client...
	got := <-c.Send
	if got.SystemText != "presence" {
		t.Fatalf("client got %q, want presence", got.SystemText)
	}
	// ...but is never retained in the ring, so future joiners don't replay it.
	if rb := h.Backlog(); len(rb) != 0 {
		t.Fatalf("backlog = %v, want empty", rb)
	}
}

func TestHub_SlowClientDoesNotBlock(t *testing.T) {
	h := NewHub(50)
	slow := NewClient()
	h.Register(slow)
	// Fill the slow client's buffer beyond capacity; Broadcast must not block.
	for range sendBuffer + 5 {
		h.Broadcast(msg("flood"))
	}
	// If we reach here without deadlock, the drop-on-full path works.
	// The buffer must have been capped at sendBuffer, not grown unbounded.
	if len(slow.Send) != sendBuffer {
		t.Fatalf("slow client buffer = %d, want %d", len(slow.Send), sendBuffer)
	}
}
