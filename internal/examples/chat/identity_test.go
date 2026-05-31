package chat

import "testing"

func TestEncodeDecode_Roundtrip(t *testing.T) {
	in := Identity{Nick: "Ada", Color: "#3b82f6"}
	enc := in.Encode()
	if enc == "" {
		t.Fatalf("Encode produced empty string")
	}
	out, err := Decode(enc)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", out, in)
	}
}

func TestDecode_Garbage(t *testing.T) {
	if _, err := Decode("!!!not-base64!!!"); err == nil {
		t.Fatalf("expected error decoding garbage")
	}
}

func TestNewGuest_Deterministic(t *testing.T) {
	a := NewGuest(42)
	b := NewGuest(42)
	if a != b {
		t.Fatalf("NewGuest must be deterministic for a given seed: %+v vs %+v", a, b)
	}
	if a.Nick == "" || a.Color == "" {
		t.Fatalf("NewGuest must populate Nick and Color: %+v", a)
	}
	if len(a.Nick) < 5 || a.Nick[:5] != "Guest" {
		t.Fatalf("guest nick should start with 'Guest': %q", a.Nick)
	}
}

func TestNewGuest_VariesBySeed(t *testing.T) {
	if NewGuest(1).Nick == NewGuest(2).Nick {
		t.Fatalf("different seeds should usually yield different nicks")
	}
}
