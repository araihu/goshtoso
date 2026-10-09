package server

import (
	"net/http"

	selectpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/select"
)

func (s *Server) handleSelectFilter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := selectpage.FilterResult(r.URL.Query().Get("type")).Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render filter result", http.StatusInternalServerError)
	}
}
