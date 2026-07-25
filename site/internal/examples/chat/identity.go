package chat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CookieName is the cookie that persists a visitor's chat identity. base64url-JSON,
// mirroring the todo example's cookie layering.
const CookieName = "gt_chat"

// palette holds avatar.Tone tokens (not hex). Identity.Color stores one of
// these so the message avatar can color itself via the avatar component's
// Tone API (a hex value would be an invalid Tailwind class — no color).
var palette = []string{"primary", "secondary", "info", "success", "warning", "danger", "inverse"}

// Identity is a visitor's display name and avatar tone. Stored in the gt_chat
// cookie; also sent in each message frame so renames take effect live.
type Identity struct {
	Nick string `json:"n"`
	// Color holds an avatar.Tone token (e.g. "info", "primary") — consumed
	// as avatar.Tone in templates. A hex value would be an invalid Tailwind
	// class and produce no visible color.
	Color string `json:"c"`
}

// Encode serializes Identity to a base64url(JSON) cookie value. Identity is
// always serializable; a marshal error is a programmer error and panics.
func (i Identity) Encode() string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(fmt.Sprintf("chat: marshal identity: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a base64url(JSON) cookie value back into an Identity.
func Decode(s string) (Identity, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Identity{}, err
	}
	var id Identity
	if err := json.Unmarshal(raw, &id); err != nil {
		return Identity{}, err
	}
	return id, nil
}

// NewGuest builds a deterministic guest identity from a numeric seed (e.g. a
// connection counter). Same seed → same nick + color, so behavior is reproducible.
func NewGuest(seed int64) Identity {
	u := uint64(seed)
	if seed < 0 {
		u = uint64(-(seed + 1)) + 1 // safe abs, no overflow for math.MinInt64
	}
	return Identity{
		Nick:  fmt.Sprintf("Guest-%04x", uint16(u*2654435761)),
		Color: palette[u%uint64(len(palette))],
	}
}
