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
