package profile

import "net/http"

// CookieName is the cookie that carries the encoded profile State.
const CookieName = "gt_profile"

// cookieMaxAge is ~30 days in seconds.
const cookieMaxAge = 30 * 24 * 60 * 60

// FromRequest reads and decodes State from the request cookie. A missing or
// corrupt cookie yields the zero State.
func FromRequest(r *http.Request) State {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return State{}
	}
	s, err := Decode([]byte(c.Value))
	if err != nil {
		return State{}
	}
	return s.Sanitize()
}

// SetCookie writes the encoded State as a cookie. Path "/" so it reaches both
// /examples/* and /api/examples/profile/*.
func SetCookie(w http.ResponseWriter, s State) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    Encode(s),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   cookieMaxAge,
	})
}
