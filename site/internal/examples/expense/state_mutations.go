// internal/examples/expense/state_mutations.go
package expense

import (
	"slices"
	"strings"
)

// truncateDesc caps s to MaxDescLen runes without splitting a multibyte rune.
func truncateDesc(s string) string {
	if len(s) <= MaxDescLen {
		return s
	}
	r := []rune(s)
	if len(r) <= MaxDescLen {
		return s // already <= MaxDescLen runes; long only in bytes
	}
	return string(r[:MaxDescLen])
}

// normalizeCategory returns c when it is a known category, else "Other".
func normalizeCategory(c string) string {
	if slices.Contains(Categories, c) {
		return c
	}
	return "Other"
}

// ParseAmountCents parses a user-entered amount like "12", "12.5", "12.34", or
// the currency-masked "$1,234.56" into an integer count of cents. A leading "$",
// thousands "," separators, and surrounding spaces are tolerated (the client
// applies an Alpine $money mask). ok is false for blank, negative, or
// unparseable input, or for more than two fractional digits. This keeps all
// money handling in integers.
func ParseAmountCents(s string) (int, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	whole, frac, hasDot := strings.Cut(s, ".")
	cents := 0
	for _, r := range whole {
		if r < '0' || r > '9' {
			return 0, false
		}
		cents = cents*10 + int(r-'0')
	}
	fracCents := 0
	if hasDot {
		if len(frac) == 0 || len(frac) > 2 {
			return 0, false
		}
		for _, r := range frac {
			if r < '0' || r > '9' {
				return 0, false
			}
			fracCents = fracCents*10 + int(r-'0')
		}
		if len(frac) == 1 {
			fracCents *= 10
		}
	}
	return cents*100 + fracCents, true
}

// Add prepends an expense so the newest item sorts first. Blank descriptions or
// unparseable amounts are ignored. The list is capped to MaxItems (coarse
// fast-path) and also refused when the resulting encoded state would exceed
// maxCookieBytes. ID comes from the Seq counter. Callers can detect whether the
// add succeeded by comparing s.Seq before and after.
func (s *State) Add(desc, amount, category, date string) {
	desc = truncateDesc(strings.TrimSpace(desc))
	cents, ok := ParseAmountCents(amount)
	if desc == "" || !ok || len(s.Items) >= MaxItems {
		return
	}
	// New items get Order one below the current minimum so they sort first.
	minOrder := 0
	for _, it := range s.Items {
		if it.Order < minOrder {
			minOrder = it.Order
		}
	}
	candidate := append(slices.Clone(s.Items), Expense{
		ID:          s.Seq + 1,
		Desc:        desc,
		AmountCents: cents,
		Category:    normalizeCategory(category),
		Date:        date,
		Order:       minOrder - 1,
	})
	probe := State{Items: candidate, Search: s.Search, Category: s.Category, Page: s.Page, Seq: s.Seq + 1}
	if len(Encode(probe)) > maxCookieBytes {
		return
	}
	s.Seq++
	s.Items = candidate
}

// indexByID returns the slice index of the expense with id, or -1.
func (s *State) indexByID(id int) int {
	return slices.IndexFunc(s.Items, func(e Expense) bool { return e.ID == id })
}

// Find returns a copy of the expense with id and whether it was found.
func (s State) Find(id int) (Expense, bool) {
	if i := slices.IndexFunc(s.Items, func(e Expense) bool { return e.ID == id }); i >= 0 {
		return s.Items[i], true
	}
	return Expense{}, false
}

// Delete removes the expense with id. Unknown id is a no-op.
func (s *State) Delete(id int) {
	if i := s.indexByID(id); i >= 0 {
		s.Items = slices.Delete(s.Items, i, i+1)
	}
}

// Edit updates amount, category, and date always; the description only when the
// trimmed input is non-empty and an unparseable amount leaves the row untouched.
func (s *State) Edit(id int, desc, amount, category, date string) {
	i := s.indexByID(id)
	if i < 0 {
		return
	}
	cents, ok := ParseAmountCents(amount)
	if !ok {
		return
	}
	if d := truncateDesc(strings.TrimSpace(desc)); d != "" {
		s.Items[i].Desc = d
	}
	s.Items[i].AmountCents = cents
	s.Items[i].Category = normalizeCategory(category)
	s.Items[i].Date = date
}

// Restore re-inserts an expense, preserving its original fields. It respects
// MaxItems and the cookie-size budget (like Add); if the list is full or the
// encoded state would exceed maxCookieBytes, Restore is a no-op. The list is
// re-sorted by Order after insertion so the item lands back in place.
func (s *State) Restore(e Expense) {
	if len(s.Items) >= MaxItems {
		return
	}
	candidate := append(slices.Clone(s.Items), e)
	slices.SortStableFunc(candidate, func(a, b Expense) int { return a.Order - b.Order })
	probe := State{Items: candidate, Search: s.Search, Category: s.Category, Page: s.Page, Seq: s.Seq}
	if len(Encode(probe)) > maxCookieBytes {
		return
	}
	s.Items = candidate
}

// Clear removes every expense but keeps the filters and Seq counter.
func (s *State) Clear() {
	s.Items = nil
	s.Page = 1
}

// SetFilter sets the search text and category filter, then resets to page 1 so
// the visitor never lands on an out-of-range page after the result set shrinks.
// An unknown category falls back to "" (all categories).
func (s *State) SetFilter(search, category string) {
	s.Search = strings.TrimSpace(search)
	if category != "" && slices.Contains(Categories, category) {
		s.Category = category
	} else {
		s.Category = ""
	}
	s.Page = 1
}

// matches reports whether e satisfies the active search and category filters.
func (s State) matches(e Expense) bool {
	if s.Category != "" && e.Category != s.Category {
		return false
	}
	if s.Search != "" && !strings.Contains(strings.ToLower(e.Desc), strings.ToLower(s.Search)) {
		return false
	}
	return true
}

// Filtered returns every expense matching the active filters, sorted by Order
// (newest first), independent of pagination.
func (s State) Filtered() []Expense {
	out := make([]Expense, 0, len(s.Items))
	for _, e := range s.Items {
		if s.matches(e) {
			out = append(out, e)
		}
	}
	slices.SortStableFunc(out, func(a, b Expense) int { return a.Order - b.Order })
	return out
}

// PageCount returns the number of pages of filtered results (at least 1).
func (s State) PageCount() int {
	n := len(s.Filtered())
	if n == 0 {
		return 1
	}
	return (n + PerPage - 1) / PerPage
}

// CurrentPage clamps the stored Page into [1, PageCount].
func (s State) CurrentPage() int {
	p := max(s.Page, 1)
	return min(p, s.PageCount())
}

// SetPage stores a requested page; CurrentPage clamps it at read time.
func (s *State) SetPage(p int) {
	s.Page = max(p, 1)
}

// Visible returns the filtered expenses for the current (clamped) page.
func (s State) Visible() []Expense {
	all := s.Filtered()
	start := (s.CurrentPage() - 1) * PerPage
	if start >= len(all) {
		return nil
	}
	end := min(start+PerPage, len(all))
	return all[start:end]
}

// TotalCents sums the amounts of all filtered expenses (every page), so the
// running total reflects the current filter, not just the visible page.
func (s State) TotalCents() int {
	sum := 0
	for _, e := range s.Filtered() {
		sum += e.AmountCents
	}
	return sum
}
