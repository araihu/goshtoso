package ticker_test

import (
	"testing"

	"github.com/araihu/goshtoso/site/internal/examples/ticker"
	"github.com/stretchr/testify/assert"
)

func TestChangePctAndDirection(t *testing.T) {
	up := ticker.Symbol{Ticker: "AAA", Price: 110, PrevPrice: 100}
	assert.InDelta(t, 10.0, up.ChangePct(), 0.0001)
	assert.Equal(t, "up", up.Direction())

	down := ticker.Symbol{Ticker: "BBB", Price: 90, PrevPrice: 100}
	assert.InDelta(t, -10.0, down.ChangePct(), 0.0001)
	assert.Equal(t, "down", down.Direction())

	flat := ticker.Symbol{Ticker: "CCC", Price: 100, PrevPrice: 100}
	assert.InDelta(t, 0.0, flat.ChangePct(), 0.0001)
	assert.Equal(t, "flat", flat.Direction())

	zero := ticker.Symbol{Ticker: "DDD", Price: 5, PrevPrice: 0}
	assert.InDelta(t, 0.0, zero.ChangePct(), 0.0001)
}

func TestSimulatorIsDeterministicForSeed(t *testing.T) {
	a := ticker.NewSimulator(42)
	b := ticker.NewSimulator(42)
	for range 5 {
		sa := a.Tick()
		sb := b.Tick()
		assert.Equal(t, sb.Symbols, sa.Symbols, "same seed must produce identical ticks")
	}
}

func TestTickSetsPrevPriceAndStaysPositive(t *testing.T) {
	s := ticker.NewSimulator(1)
	before := s.Snapshot()
	snap := s.Tick()
	assert.Len(t, snap.Symbols, len(before.Symbols))
	for i, sym := range snap.Symbols {
		assert.Equal(t, before.Symbols[i].Price, sym.PrevPrice, "PrevPrice should be last tick's Price")
		assert.Greater(t, sym.Price, 0.0, "price must stay positive")
	}
}

func TestInitialSymbolsNonEmpty(t *testing.T) {
	assert.NotEmpty(t, ticker.InitialSymbols())
}
