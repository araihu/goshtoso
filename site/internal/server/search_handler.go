package server

import (
	"encoding/json"
	"net/http"
)

type searchItemResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Href        string   `json:"href,omitempty"`
	Kind        string   `json:"kind,omitempty"`
	Method      string   `json:"method,omitempty"`
	Path        string   `json:"path,omitempty"`
	Section     string   `json:"section,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

func (s *Server) handleSearchItems(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/components/search/items" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode([]searchItemResponse{
		{
			ID:          "search-remote-table",
			Title:       "Fetched table",
			Description: "Loaded from a JSON endpoint and filtered in the browser.",
			Href:        "/components/table",
			Section:     "Remote",
			Keywords:    []string{"fetched remote json"},
		},
		{
			ID:          "search-remote-list-teams",
			Title:       "List teams",
			Description: "Operation result with method and route metadata.",
			Href:        "/api/teams",
			Kind:        "Operation",
			Method:      "GET",
			Path:        "/teams",
			Section:     "Remote API",
			Keywords:    []string{"teams operation"},
		},
		{
			ID:          "search-remote-combobox",
			Title:       "Remote combobox",
			Description: "Another result from the same client-side item source.",
			Href:        "/components/combobox",
			Section:     "Remote",
			Keywords:    []string{"async source"},
		},
	})
}
