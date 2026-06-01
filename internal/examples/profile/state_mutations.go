package profile

import "strings"

// capRunes trims s and caps it to max runes without splitting a multibyte rune.
func capRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// SetName sets the trimmed, rune-capped display name. The change is refused
// (name kept unchanged) if the resulting encoded cookie would exceed the budget.
func (s *State) SetName(name string) {
	candidate := State{Name: capRunes(name, MaxNameLen), Bio: s.Bio}
	if len(Encode(candidate)) > maxCookieBytes {
		return
	}
	s.Name = candidate.Name
}

// SetBio sets the trimmed, rune-capped bio. Refused if over the cookie budget.
func (s *State) SetBio(bio string) {
	candidate := State{Name: s.Name, Bio: capRunes(bio, MaxBioLen)}
	if len(Encode(candidate)) > maxCookieBytes {
		return
	}
	s.Bio = candidate.Bio
}
