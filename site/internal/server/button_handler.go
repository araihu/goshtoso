package server

import (
	"net/http"

	buttonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/button"
)

func (s *Server) handleButtonFragment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get disabled state from query params
	disabled := r.URL.Query().Get("disabled") == "true"

	// Render just the button grid fragment
	_ = buttonpage.ButtonFragment(disabled).Render(r.Context(), w)
}
