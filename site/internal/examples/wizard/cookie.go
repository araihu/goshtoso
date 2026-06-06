// internal/examples/wizard/cookie.go
package wizard

import "net/http"

// CookieName is the cookie that carries the encoded wizard WizardState.
const CookieName = "gt_wizard"

// cookieMaxAge is ~30 days in seconds.
const cookieMaxAge = 30 * 24 * 60 * 60

// FromRequest reads and decodes WizardState from the request cookie. A missing
// or corrupt cookie yields the zero state.
func FromRequest(r *http.Request) WizardState {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return WizardState{}
	}
	s, err := Decode([]byte(c.Value))
	if err != nil {
		return WizardState{}
	}
	return s
}

// SetCookie writes the encoded WizardState as a cookie. Path is "/" so it is
// sent to both /examples/* and /api/examples/wizard/* (a narrower path would not
// reach the API endpoints).
func SetCookie(w http.ResponseWriter, s WizardState) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    Encode(s),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   cookieMaxAge,
	})
}
