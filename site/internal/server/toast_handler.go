package server

import (
	"net/http"

	"github.com/araihu/goshtoso/components/toast"
)

func (s *Server) handleToastOOB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	tone := r.FormValue("tone")
	title := r.FormValue("title")
	message := r.FormValue("message")

	var selectedTone toast.Tone
	switch tone {
	case "success":
		selectedTone = toast.ToneSuccess
	case "warning":
		selectedTone = toast.ToneWarning
	case "danger":
		selectedTone = toast.ToneDanger
	default:
		selectedTone = toast.ToneInfo
	}

	cfg := toast.Config{
		Tone:    selectedTone,
		Title:   title,
		Message: message,
	}

	_ = toast.OOBToast(cfg).Render(r.Context(), w)
}
