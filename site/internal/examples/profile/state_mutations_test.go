package profile

import (
	"strings"
	"testing"
)

func TestSetNameTrimsAndCaps(t *testing.T) {
	var s State
	s.SetName("  Ada  ")
	if s.Name != "Ada" {
		t.Fatalf("trim failed: %q", s.Name)
	}
	s.SetName(strings.Repeat("x", MaxNameLen+50))
	if len([]rune(s.Name)) != MaxNameLen {
		t.Fatalf("cap failed: got %d runes", len([]rune(s.Name)))
	}
}

func TestSetBioCaps(t *testing.T) {
	var s State
	s.SetBio(strings.Repeat("y", MaxBioLen+100))
	if len([]rune(s.Bio)) != MaxBioLen {
		t.Fatalf("bio cap failed: got %d runes", len([]rune(s.Bio)))
	}
}

func TestSetWithinBudget(t *testing.T) {
	var s State
	s.SetBio(strings.Repeat("z", MaxBioLen))
	s.SetName(strings.Repeat("n", MaxNameLen))
	if len(Encode(s)) > maxCookieBytes {
		t.Fatalf("encoded %d exceeds budget %d", len(Encode(s)), maxCookieBytes)
	}
}
