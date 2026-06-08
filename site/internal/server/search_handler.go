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
			ID:          "search-remote-combobox",
			Title:       "Remote combobox",
			Description: "Another result from the same client-side item source.",
			Href:        "/components/combobox",
			Section:     "Remote",
			Keywords:    []string{"async source"},
		},
	})
}
