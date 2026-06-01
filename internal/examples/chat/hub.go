package chat

import (
	"sync"
	"sync/atomic"
)

// sendBuffer is the per-client outbound queue depth. A client whose queue is
// full is treated as too slow and the event is dropped for that client only.
// It must be >= the hub's ringMax (the handler constructs NewHub(50)) so that a
// full backlog replay performed atomically by Join never overflows and silently
// drops history.
const sendBuffer = 64

// EventKind distinguishes a chat message from a transient presence notice.
type EventKind uint8

const (
	EventMessage EventKind = iota
	EventPresence
)

// Event is one thing that happened in the room. The hub fans Events out to every
// client; each client's write pump renders the Event into HTML *for that client*,
// so a message is shown right-aligned ("mine") to its sender and left-aligned to
// everyone else with no client-side JavaScript deciding presentation.
type Event struct {
	Kind EventKind

	// Message fields.
	ConnID uint64 // the connection that sent it; 0 for the bot / system
	Nick   string
	Color  string // avatar variant token
	Text   string
	Time   string
	IsBot  bool

	// Presence fields.
	SystemText string // e.g. "Ada joined"
	Count      int    // current online count
}

// connSeq assigns a unique, stable id to each connection so "mine" detection
// (sender == viewer) survives renames and same-nick collisions — two tabs of the
// same person are different connections, so each sees only its own as mine.
var connSeq atomic.Uint64

// Client is one connected socket: a stable ConnID plus a buffered channel of
// Events the handler's write pump drains and renders for this client.
type Client struct {
	ConnID uint64
	Send   chan Event
}

// NewClient builds a Client with a unique connection id and a buffered channel.
func NewClient() *Client {
	return &Client{ConnID: connSeq.Add(1), Send: make(chan Event, sendBuffer)}
}

// Hub is a RAM-only broadcast fan-out plus a bounded ring buffer of recent
// Events (so a newly connected client can replay history). Mutex-guarded: no
// owner goroutine, so Count and Backlog are plain reads.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	ring    []Event
	ringMax int
}

// NewHub builds an empty Hub whose ring keeps at most ringMax recent events.
func NewHub(ringMax int) *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		ringMax: ringMax,
	}
}

// Register adds a client to the broadcast set.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// Join atomically registers a client AND replays the current ring into its send
// channel, all under a single lock. This is the correct way to admit a new
// client: it guarantees each event is delivered exactly once — either it is in
// the snapshot replayed here, or it is delivered live by Broadcast, never both.
// That holds because Broadcast appends-to-ring and sends-to-clients under the
// same mutex: a Broadcast cannot interleave between this client being added to
// the set and the ring being replayed.
//
// The replay uses a non-blocking send purely as a safety valve; with sendBuffer
// >= ringMax it cannot actually drop, but the non-blocking form guarantees we
// never block while holding the lock.
func (h *Hub) Join(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
	for _, ev := range h.ring {
		select {
		case c.Send <- ev:
		default: // impossibly full (buffer >= ringMax) — never block under the lock
		}
	}
}

// Unregister removes a client and closes its send channel so the write pump
// exits. Safe to call once per client.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.Send)
	}
	h.mu.Unlock()
}

// Broadcast appends the event to the ring and fans it out to every client. A
// client whose buffer is full is skipped (dropped), never blocked on.
func (h *Hub) Broadcast(ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ring = append(h.ring, ev)
	if len(h.ring) > h.ringMax {
		h.ring = h.ring[len(h.ring)-h.ringMax:]
	}
	for c := range h.clients {
		select {
		case c.Send <- ev:
		default: // slow client — drop this event for them
		}
	}
}

// BroadcastEphemeral fans an event out to every current client but does NOT
// append it to the ring, so it is never replayed to clients that join later.
// Use this for presence/transient events (join/leave notices, online-count
// updates) that would otherwise show up as stale lines in a new joiner's
// replayed backlog. Like Broadcast, a full client is skipped.
func (h *Hub) BroadcastEphemeral(ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.Send <- ev:
		default: // slow client — drop this event for them
		}
	}
}

// Backlog returns a snapshot copy of the recent-event ring, oldest first.
func (h *Hub) Backlog() []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Event, len(h.ring))
	copy(out, h.ring)
	return out
}

// Count returns the number of currently registered clients.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
