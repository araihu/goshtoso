package expense

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	in := State{
		Items: []Expense{
			{ID: 1, Desc: "Coffee", AmountCents: 450, Category: "Food", Date: "2026-06-01", Order: -1},
			{ID: 2, Desc: "Bus", AmountCents: 250, Category: "Transport", Order: -2},
		},
		Search:   "co",
		Category: "Food",
		Page:     2,
		Seq:      2,
	}
	out, err := Decode([]byte(Encode(in)))
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if out.Seq != in.Seq || out.Page != in.Page || out.Search != in.Search || out.Category != in.Category {
		t.Fatalf("scalar fields not round-tripped: got %+v want %+v", out, in)
	}
	if len(out.Items) != len(in.Items) {
		t.Fatalf("item count = %d, want %d", len(out.Items), len(in.Items))
	}
	if out.Items[0] != in.Items[0] {
		t.Errorf("item[0] = %+v, want %+v", out.Items[0], in.Items[0])
	}
}

func TestDecodeEmptyYieldsZeroState(t *testing.T) {
	s, err := Decode(nil)
	if err != nil {
		t.Fatalf("Decode(nil) error: %v", err)
	}
	if len(s.Items) != 0 || s.Seq != 0 {
		t.Errorf("Decode(nil) = %+v, want zero State", s)
	}
}

func TestDecodeCorruptIsError(t *testing.T) {
	if _, err := Decode([]byte("not-base64-$$$")); err == nil {
		t.Error("Decode of corrupt input should return an error")
	}
}
