package chat

import "sync"

// sendBuffer is the per-client outbound queue depth. A client whose queue is
// full is treated as too slow and the message is dropped for that client only.
// It must be >= the hub's ringMax (the handler constructs NewHub(50)) so that a
// full backlog replay performed atomically by Join never overflows and silently
// drops history.
const sendBuffer = 64

// Client is one connected socket from the hub's point of view: a buffered
// channel of pre-rendered HTML frames that the handler's write pump drains.
type Client struct {
	Send chan []byte
}

// NewClient builds a Client with a buffered send channel.
func NewClient() *Client {
	return &Client{Send: make(chan []byte, sendBuffer)}
}

// Hub is a RAM-only broadcast fan-out plus a bounded ring buffer of recent
// frames (so a newly connected client can replay history). Mutex-guarded: no
// owner goroutine, so Count and Backlog are plain reads.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	ring    [][]byte
	ringMax int
}

// NewHub builds an empty Hub whose ring keeps at most ringMax recent frames.
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
// client: it guarantees each frame is delivered exactly once — either it is in
// the snapshot replayed here, or it is delivered live by Broadcast, never both.
// That holds because Broadcast appends-to-ring and sends-to-clients under the
// same mutex: a Broadcast cannot interleave between this client being added to
// the set and the ring being replayed, so no frame can both land in the ring
// snapshot AND be fanned out to this client live.
//
// The replay uses a non-blocking send purely as a safety valve; with sendBuffer
// >= ringMax it cannot actually drop, but the non-blocking form guarantees we
// never block while holding the lock.
func (h *Hub) Join(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
	for _, frame := range h.ring {
		select {
		case c.Send <- frame:
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

// Broadcast appends the frame to the ring and fans it out to every client. A
// client whose buffer is full is skipped (dropped), never blocked on.
func (h *Hub) Broadcast(frame []byte) {
	frame = append([]byte(nil), frame...) // own the bytes; callers may reuse their buffer
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ring = append(h.ring, frame)
	if len(h.ring) > h.ringMax {
		h.ring = h.ring[len(h.ring)-h.ringMax:]
	}
	for c := range h.clients {
		select {
		case c.Send <- frame:
		default: // slow client — drop this frame for them
		}
	}
}

// BroadcastEphemeral fans a frame out to every current client but does NOT
// append it to the ring, so it is never replayed to clients that join later.
// Use this for presence/transient frames (join/leave toasts, online-count
// updates) that would otherwise show up as ghost toasts and stale counts in a
// new joiner's replayed backlog. Like Broadcast, a full client is skipped.
func (h *Hub) BroadcastEphemeral(frame []byte) {
	frame = append([]byte(nil), frame...) // own the bytes; callers may reuse their buffer
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.Send <- frame:
		default: // slow client — drop this frame for them
		}
	}
}

// Backlog returns a snapshot copy of the recent-frame ring, oldest first.
func (h *Hub) Backlog() [][]byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([][]byte, len(h.ring))
	copy(out, h.ring)
	return out
}

// Count returns the number of currently registered clients.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
