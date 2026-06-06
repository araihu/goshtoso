package expense

// Sample returns a small, varied starter list shown on a visitor's first load
// (when no cookie is present yet) so the example never opens empty. Dates are
// fixed strings (no time.Now) to keep first paint deterministic.
func Sample() State {
	s := State{Page: 1}
	s.Add("Weekly groceries", "82.40", "Food", "2026-06-02")
	s.Add("Monthly transit pass", "55.00", "Transport", "2026-06-01")
	s.Add("Rent", "1200.00", "Housing", "2026-06-01")
	s.Add("Pharmacy", "18.75", "Health", "2026-06-03")
	s.Add("Cinema tickets", "32.00", "Entertainment", "2026-06-04")
	s.Add("Coffee with team", "14.20", "Food", "2026-06-04")
	return s
}
