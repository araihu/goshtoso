package expense

import (
	"strings"
	"testing"
)

func TestParseAmountCents(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"12", 1200, true},
		{"12.5", 1250, true},
		{"12.34", 1234, true},
		{"0.09", 9, true},
		{"$1,234.56", 123456, true}, // currency-masked input
		{"1,200", 120000, true},     // thousands separators stripped
		{"$15.00", 1500, true},
		{"  7.7 ", 770, true},
		{"", 0, false},
		{"abc", 0, false},
		{"1.234", 0, false}, // too many fractional digits
		{"1.", 0, false},    // dangling dot
		{"-5", 0, false},    // negative not accepted
	}
	for _, c := range cases {
		got, ok := ParseAmountCents(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("ParseAmountCents(%q) = (%d,%v), want (%d,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestFormatCents(t *testing.T) {
	cases := map[int]string{
		0:         "$0.00",
		5:         "$0.05",
		1234:      "$12.34",
		120000:    "$1,200.00",
		123456789: "$1,234,567.89",
		-450:      "-$4.50",
	}
	for in, want := range cases {
		if got := FormatCents(in); got != want {
			t.Errorf("FormatCents(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestAddPrependsAndAssignsID(t *testing.T) {
	var s State
	s.Add("First", "10.00", "Food", "")
	s.Add("Second", "20.00", "Transport", "")
	if s.Seq != 2 || len(s.Items) != 2 {
		t.Fatalf("after 2 adds: Seq=%d len=%d, want 2/2", s.Seq, len(s.Items))
	}
	vis := s.Filtered()
	if vis[0].Desc != "Second" {
		t.Errorf("newest item should sort first, got %q", vis[0].Desc)
	}
}

func TestAddRejectsBadInput(t *testing.T) {
	var s State
	s.Add("", "10.00", "Food", "")     // blank desc
	s.Add("Valid", "nope", "Food", "") // bad amount
	if s.Seq != 0 || len(s.Items) != 0 {
		t.Errorf("bad adds should be no-ops, got Seq=%d len=%d", s.Seq, len(s.Items))
	}
}

func TestAddUnknownCategoryFallsBackToOther(t *testing.T) {
	var s State
	s.Add("Mystery", "1.00", "Bananas", "")
	if got := s.Items[0].Category; got != "Other" {
		t.Errorf("unknown category = %q, want Other", got)
	}
}

func TestEditAndDelete(t *testing.T) {
	var s State
	s.Add("Lunch", "10.00", "Food", "")
	id := s.Items[0].ID
	s.Edit(id, "Dinner", "25.50", "Entertainment", "2026-06-05")
	e, _ := s.Find(id)
	if e.Desc != "Dinner" || e.AmountCents != 2550 || e.Category != "Entertainment" {
		t.Errorf("edit not applied: %+v", e)
	}
	s.Delete(id)
	if len(s.Items) != 0 {
		t.Errorf("delete failed, len=%d", len(s.Items))
	}
}

func TestEditBlankDescKeepsOld(t *testing.T) {
	var s State
	s.Add("Keep me", "10.00", "Food", "")
	id := s.Items[0].ID
	s.Edit(id, "   ", "11.00", "Food", "")
	if e, _ := s.Find(id); e.Desc != "Keep me" {
		t.Errorf("blank desc overwrote existing: %q", e.Desc)
	}
}

func TestFilterByCategoryAndSearch(t *testing.T) {
	var s State
	s.Add("Groceries", "10.00", "Food", "")
	s.Add("Train ticket", "5.00", "Transport", "")
	s.Add("Snacks", "3.00", "Food", "")

	s.SetFilter("", "Food")
	if got := len(s.Filtered()); got != 2 {
		t.Errorf("category filter Food = %d items, want 2", got)
	}
	s.SetFilter("snack", "")
	f := s.Filtered()
	if len(f) != 1 || f[0].Desc != "Snacks" {
		t.Errorf("search 'snack' = %+v, want [Snacks]", f)
	}
	s.SetFilter("", "NotACategory")
	if s.Category != "" {
		t.Errorf("unknown category should reset to all, got %q", s.Category)
	}
}

func TestTotalCentsTracksFilter(t *testing.T) {
	var s State
	s.Add("A", "10.00", "Food", "")
	s.Add("B", "5.00", "Transport", "")
	if s.TotalCents() != 1500 {
		t.Errorf("unfiltered total = %d, want 1500", s.TotalCents())
	}
	s.SetFilter("", "Food")
	if s.TotalCents() != 1000 {
		t.Errorf("filtered total = %d, want 1000", s.TotalCents())
	}
}

func TestPagination(t *testing.T) {
	var s State
	for range PerPage + 3 {
		s.Add("item", "1.00", "Food", "")
	}
	if s.PageCount() != 2 {
		t.Fatalf("PageCount = %d, want 2", s.PageCount())
	}
	s.SetPage(1)
	if len(s.Visible()) != PerPage {
		t.Errorf("page 1 visible = %d, want %d", len(s.Visible()), PerPage)
	}
	s.SetPage(2)
	if len(s.Visible()) != 3 {
		t.Errorf("page 2 visible = %d, want 3", len(s.Visible()))
	}
	// Out-of-range page clamps instead of returning an empty list.
	s.SetPage(99)
	if s.CurrentPage() != 2 {
		t.Errorf("CurrentPage clamp = %d, want 2", s.CurrentPage())
	}
}

func TestSetFilterResetsPage(t *testing.T) {
	var s State
	for range PerPage + 3 {
		s.Add("item", "1.00", "Food", "")
	}
	s.SetPage(2)
	s.SetFilter("item", "")
	if s.Page != 1 {
		t.Errorf("SetFilter should reset to page 1, got %d", s.Page)
	}
}

func TestRestoreReinserts(t *testing.T) {
	var s State
	s.Add("Gone", "9.99", "Food", "")
	e := s.Items[0]
	s.Delete(e.ID)
	s.Restore(e)
	if len(s.Items) != 1 || s.Items[0].ID != e.ID {
		t.Errorf("restore failed: %+v", s.Items)
	}
}

func TestClear(t *testing.T) {
	var s State
	s.Add("A", "1.00", "Food", "")
	s.Add("B", "2.00", "Food", "")
	seqBefore := s.Seq
	s.Clear()
	if len(s.Items) != 0 || s.Seq != seqBefore {
		t.Errorf("Clear should empty items but keep Seq: len=%d seq=%d", len(s.Items), s.Seq)
	}
}

func TestAddRespectsCookieBudget(t *testing.T) {
	var s State
	longDesc := strings.Repeat("x", MaxDescLen)
	added := 0
	for range MaxItems {
		before := s.Seq
		s.Add(longDesc, "999.99", "Entertainment", "2026-06-05")
		if s.Seq == before {
			break // refused by the cookie-size budget
		}
		added++
	}
	if added == 0 {
		t.Fatal("expected at least some adds to succeed")
	}
	if len(Encode(s)) > maxCookieBytes {
		t.Errorf("encoded state %d bytes exceeds budget %d", len(Encode(s)), maxCookieBytes)
	}
}
