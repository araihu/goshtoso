// Package profile holds the pure, HTTP-free domain model for the /examples/profile
// app. Only the display name and bio live here (in a cookie); avatar/banner images
// live client-side in IndexedDB and theme/dark in the app's existing theme engine,
// so the server keeps no per-user memory and never sees image bytes.
package profile

import (
	"encoding/base64"
	"encoding/json"
)

const (
	// MaxNameLen bounds the display name's stored length in runes.
	MaxNameLen = 60
	// MaxBioLen bounds the bio's stored length in runes.
	MaxBioLen = 280
	// maxCookieBytes bounds the encoded cookie value so the browser never silently
	// drops it (browsers cap a cookie near 4KB).
	maxCookieBytes = 3800
)

// State is the per-user profile text. Avatar/banner images and theme/dark are
// intentionally absent — they are client-side concerns.
type State struct {
	Name string `json:"n"`
	Bio  string `json:"b"`
}

// Encode serializes State to a base64url(JSON) string for cookie storage.
// State is always serializable; a marshal error is a programmer error and panics.
func Encode(s State) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic("profile.Encode: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a cookie value back into State. Empty input yields the zero
// State; malformed input returns an error so callers can fall back to a default.
func Decode(raw []byte) (State, error) {
	var s State
	if len(raw) == 0 {
		return s, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(string(raw))
	if err != nil {
		return State{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

// Sanitize clamps Name and Bio to their rune caps. Applied to any state decoded
// from an untrusted cookie so a hand-crafted cookie cannot exceed the limits.
func (s State) Sanitize() State {
	return State{Name: capRunes(s.Name, MaxNameLen), Bio: capRunes(s.Bio, MaxBioLen)}
}
