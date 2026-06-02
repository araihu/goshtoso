package logs

import "testing"

func TestLineCyclesSamplePool(t *testing.T) {
	if PoolSize() == 0 {
		t.Fatal("expected a non-empty sample pool")
	}
	// Line(i) cycles deterministically: index i and i+PoolSize match.
	a := Line(0)
	b := Line(PoolSize())
	if a.Level != b.Level || a.Source != b.Source || a.Message != b.Message {
		t.Errorf("Line(0) and Line(PoolSize()) differ: %+v vs %+v", a, b)
	}
	// Negative / large indices never panic and stay in-pool.
	_ = Line(-3)
	_ = Line(1_000_000)
}

func TestSeverityOrdering(t *testing.T) {
	if Debug.Severity() >= Info.Severity() ||
		Info.Severity() >= Warn.Severity() ||
		Warn.Severity() >= Error.Severity() {
		t.Fatal("severity must order Debug < Info < Warn < Error")
	}
}

func TestLevelString(t *testing.T) {
	if Error.String() != "error" {
		t.Errorf("got %q, want %q", Error.String(), "error")
	}
}
