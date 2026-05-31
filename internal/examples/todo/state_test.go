// internal/examples/todo/state_test.go
package todo

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	s := State{
		Todos:  []Todo{{ID: 1, Title: "Buy milk", Done: true, Priority: "high", Due: "2026-06-01", Order: 0}},
		Filter: "active",
		Seq:    2,
	}
	got, err := Decode([]byte(Encode(s)))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got.Todos) != 1 || got.Todos[0].Title != "Buy milk" || got.Seq != 2 || got.Filter != "active" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestDecodeMalformedReturnsEmpty(t *testing.T) {
	for _, in := range []string{"", "not-base64!!!", "YWJj" /* base64 "abc", invalid json */} {
		got, err := Decode([]byte(in))
		if err == nil && (len(got.Todos) != 0 || got.Seq != 0) {
			t.Fatalf("expected empty state for %q, got %+v", in, got)
		}
	}
}
