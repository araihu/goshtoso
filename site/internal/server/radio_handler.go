package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// handleRadioEcho returns a small HTML fragment for the radio demo's HTMX showcase.
// It echoes back the radio value selected by the user.
func (s *Server) handleRadioEcho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	value := r.URL.Query().Get("value")
	if value == "" {
		value = "(empty)"
	}
	_, _ = fmt.Fprintf(w, `Server: you picked <span class="font-mono font-semibold">%s</span> at %s.`,
		htmlEscape(value), time.Now().Format("15:04:05.000"))
}

// htmlEscape applies a minimal HTML escape suitable for an attribute-free text fragment.
func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
