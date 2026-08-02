package server

import (
	"fmt"
	"net/http"

	"github.com/araihu/goshtoso/components/toast"
)

func (s *Server) handleFormExternalSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	version := r.FormValue("version")
	if version == "" {
		version = "selected version"
	} else {
		version = "v" + version
	}
	_ = toast.OOBToast(toast.Config{
		Tone:    toast.ToneSuccess,
		Title:   "Upgrade request submitted",
		Message: fmt.Sprintf("Target version: %s", version),
	}).Render(r.Context(), w)
}
