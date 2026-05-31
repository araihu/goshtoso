package ticker_test

import (
	"testing"

	"github.com/araihu/goshtoso/internal/examples/ticker"
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
