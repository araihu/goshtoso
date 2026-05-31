package profile

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	in := State{Name: "Ada Lovelace", Bio: "First programmer."}
	out, err := Decode([]byte(Encode(in)))
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if out.Name != in.Name || out.Bio != in.Bio {
		t.Fatalf("round trip mismatch: got %+v want %+v", out, in)
	}
}

func TestDecodeEmptyYieldsZero(t *testing.T) {
	s, err := Decode(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "" || s.Bio != "" {
		t.Fatalf("expected zero State, got %+v", s)
	}
}

func TestDecodeCorruptReturnsError(t *testing.T) {
	if _, err := Decode([]byte("!!!not-base64!!!")); err == nil {
		t.Fatal("expected error for corrupt input")
	}
}
