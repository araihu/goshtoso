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

	variant := r.FormValue("variant")
	title := r.FormValue("title")
	message := r.FormValue("message")

	var v toast.Variant
	switch variant {
	case "success":
		v = toast.Success
	case "warning":
		v = toast.Warning
	case "danger":
		v = toast.Danger
	case "message":
		v = toast.Message
	default:
		v = toast.Info
	}

	cfg := toast.Config{
		Variant: v,
		Title:   title,
		Message: message,
	}

	toast.OOBToast(cfg).Render(r.Context(), w)
}
