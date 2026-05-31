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
