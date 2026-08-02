package server

import (
	"net/http"
	"strconv"

	stepspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/steps"
)

func (s *Server) handleStepsDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	current := 2
	if raw := r.URL.Query().Get("step"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			current = parsed
		}
	}

	if current < 1 {
		current = 1
	}
	if current > 4 {
		current = 4
	}

	_ = stepspage.StepsHTMXFlow(current).Render(r.Context(), w)
}
