package server

import (
	"fmt"
	"net/http"
)

func (s *Server) handleBannerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<p class="text-sm font-medium text-success">Banner action received</p>`)
}
