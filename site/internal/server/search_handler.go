package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
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

// handleRemoteSearch demonstrates the bounded JSON contract used by
// search.RemoteSource. Production handlers should apply authorization and
// ranking before returning this caller-owned Item-shaped payload.
func (s *Server) handleRemoteSearch(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/components/search/remote" {
		http.NotFound(w, r)
		return
	}
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if query == "error" {
		http.Error(w, "search unavailable", http.StatusServiceUnavailable)
		return
	}
	if query == "slow" {
		time.Sleep(350 * time.Millisecond)
	}
	items := []searchItemResponse{
		{ID: "search-remote-fast", Title: "Fast resource", Description: "Server-ranked result for cancellation and stale-response coverage.", Href: "/components/search", Section: "Server"},
		{ID: "search-remote-slow", Title: "Slow stale resource", Description: "Delayed result that must not replace a newer query.", Href: "/components/search", Section: "Server"},
		{ID: "search-remote-table", Title: "Table resource", Description: "Returned in server order after server-side filtering.", Href: "/components/table", Section: "Server"},
	}
	filtered := make([]searchItemResponse, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Title+" "+item.Description), query) {
			filtered = append(filtered, item)
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": filtered})
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
