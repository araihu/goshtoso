package ticker

import "math/rand"

// Simulator advances a fixed set of symbols via a seeded random walk. It is
// deterministic for a given seed, which keeps the example offline and lets E2E
// rely on reproducible behaviour.
type Simulator struct {
	symbols []Symbol
	rng     *rand.Rand
}

// NewSimulator returns a Simulator seeded with the default symbol set.
func NewSimulator(seed int64) *Simulator {
	return &Simulator{
		symbols: InitialSymbols(),
		rng:     rand.New(rand.NewSource(seed)),
	}
}

// InitialSymbols is the starting watchlist, also used for the first server-side
// render before the SSE stream sends its first snapshot.
func InitialSymbols() []Symbol {
	return []Symbol{
		{Ticker: "AAPL", Name: "Apple Inc.", Price: 190.00, PrevPrice: 190.00},
		{Ticker: "GOOG", Name: "Alphabet Inc.", Price: 140.00, PrevPrice: 140.00},
		{Ticker: "MSFT", Name: "Microsoft Corp.", Price: 420.00, PrevPrice: 420.00},
		{Ticker: "AMZN", Name: "Amazon.com Inc.", Price: 178.00, PrevPrice: 178.00},
		{Ticker: "TSLA", Name: "Tesla Inc.", Price: 250.00, PrevPrice: 250.00},
		{Ticker: "NVDA", Name: "NVIDIA Corp.", Price: 120.00, PrevPrice: 120.00},
	}
}

// Tick advances every symbol one step (±1% random walk, floored at 1) and
// returns the new snapshot.
func (s *Simulator) Tick() Snapshot {
	for i := range s.symbols {
		prev := s.symbols[i].Price
		delta := (s.rng.Float64()*2 - 1) * 0.01 * prev
		next := prev + delta
		if next < 1 {
			next = 1
		}
		s.symbols[i].PrevPrice = prev
		s.symbols[i].Price = next
	}
	return s.snapshot()
}

// Snapshot returns the current state without advancing.
func (s *Simulator) Snapshot() Snapshot {
	return s.snapshot()
}

func (s *Simulator) snapshot() Snapshot {
	out := make([]Symbol, len(s.symbols))
	copy(out, s.symbols)
	return Snapshot{Symbols: out}
}
