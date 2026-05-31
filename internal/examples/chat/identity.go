package chat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CookieName is the cookie that persists a visitor's chat identity. base64url-JSON,
// mirroring the todo example's cookie layering.
const CookieName = "gt_chat"

// palette is a fixed set of avatar/background colors. Indexed deterministically
// by the guest seed so a given seed always maps to the same color.
var palette = []string{
	"#ef4444", "#f97316", "#eab308", "#22c55e",
	"#14b8a6", "#3b82f6", "#8b5cf6", "#ec4899",
}

// Identity is a visitor's display name and avatar color. Stored in the gt_chat
// cookie; also sent in each message frame so renames take effect live.
type Identity struct {
	Nick  string `json:"n"`
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
	if seed < 0 {
		seed = -seed
	}
	return Identity{
		Nick:  fmt.Sprintf("Guest-%04x", uint16(seed*2654435761)),
		Color: palette[seed%int64(len(palette))],
	}
}
