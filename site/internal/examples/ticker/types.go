// Package ticker is the HTTP-free domain layer for the /examples/ticker
// example: a deterministic, seeded stock-price simulator plus a pub/sub broker
// that fans each tick out to all SSE subscribers. One shared stream, no
// per-user state.
package ticker

// Symbol is a single tracked instrument and its latest two prices.
type Symbol struct {
	Ticker    string
	Name      string
	Price     float64
	PrevPrice float64
}

// ChangePct is the percent change from PrevPrice to Price. Returns 0 when
// PrevPrice is 0 (guards the first tick / unset state).
func (s Symbol) ChangePct() float64 {
	if s.PrevPrice == 0 {
		return 0
	}
	return (s.Price - s.PrevPrice) / s.PrevPrice * 100
}

// Direction is "up", "down", or "flat" relative to PrevPrice.
func (s Symbol) Direction() string {
	switch {
	case s.Price > s.PrevPrice:
		return "up"
	case s.Price < s.PrevPrice:
		return "down"
	default:
		return "flat"
	}
}

// Snapshot is a point-in-time view of every symbol.
type Snapshot struct {
	Symbols []Symbol
}
