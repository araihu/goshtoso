package todo

// Sample returns a small, varied starter list shown on a visitor's first load
// (when no cookie is present yet) so the example never opens empty. Dates are
// fixed strings (no time.Now) to keep first paint deterministic.
func Sample() State {
	s := State{Filter: "all"}
	s.Add("Read the Goshtoso component docs", "low", "")
	s.Add("Wire the table HTMX endpoint", "high", "2026-06-03")
	s.Add("Review the open pull request", "med", "2026-06-02")
	s.Add("Ship the examples gallery", "med", "")
	s.Toggle(1) // first task already done, to show the completed state
	return s
}
