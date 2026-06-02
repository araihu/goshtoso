package server

import (
	"net/http"
	"strings"

	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/site/internal/examples/profile"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/araihu/goshtoso/site/internal/pages/demo/examples"
)

// registerProfileRoutes wires the /api/examples/profile/* endpoints.
func (s *Server) registerProfileRoutes() {
	s.mux.HandleFunc("/api/examples/profile/identity", s.handleProfileIdentity)
}

// renderProfilePage is the first-load handler for /examples/profile. Seeds a
// sample profile on first visit (no cookie) unless ?seed=0 (used by e2e).
func (s *Server) renderProfilePage(w http.ResponseWriter, r *http.Request) {
	var st profile.State
	if _, err := r.Cookie(profile.CookieName); err != nil && r.URL.Query().Get("seed") != "0" {
		st = profile.Sample()
		profile.SetCookie(w, st)
	} else {
		st = profile.FromRequest(r)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.ProfileApp(st)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.Fragment("Profile", "profile", content).Render(r.Context(), w)
		return
	}
	_ = demo.Layout("Profile", "profile", content).Render(r.Context(), w)
}

// handleProfileIdentity saves name + bio to the cookie and re-renders the
// identity fields (so server-capped values reflect back) plus a success toast.
func (s *Server) handleProfileIdentity(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	var st profile.State
	st.SetName(strings.TrimSpace(r.FormValue("name")))
	st.SetBio(strings.TrimSpace(r.FormValue("bio")))
	profile.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.IdentityFields(st).Render(r.Context(), w)
	_ = toast.OOBToast(toast.Config{
		Variant: toast.Success,
		Title:   "Saved",
		Message: "Your profile was updated.",
	}).Render(r.Context(), w)
}
