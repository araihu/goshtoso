package server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/araihu/goshtoso/components/table"
)

// tableRecord is the server-side data model for demo table rows
type tableRecord struct {
	ID         string
	Name       string
	Email      string
	Membership string
}

// allRecords returns the full dataset for table demos
func allRecords() []tableRecord {
	return []tableRecord{
		{ID: "2335", Name: "Alice Brown", Email: "alice.brown@penguinui.com", Membership: "Silver"},
		{ID: "2338", Name: "Bob Johnson", Email: "johnson.bob@penguinui.com", Membership: "Gold"},
		{ID: "2342", Name: "Sarah Adams", Email: "s.adams@penguinui.com", Membership: "Gold"},
		{ID: "2345", Name: "Alex Martinez", Email: "alex.martinez@penguinui.com", Membership: "Gold"},
		{ID: "2346", Name: "Ryan Thompson", Email: "ryan.thompson@penguinui.com", Membership: "Silver"},
		{ID: "2349", Name: "Emily Rodriguez", Email: "emily.rodriguez@penguinui.com", Membership: "Gold"},
		{ID: "2350", Name: "James Wilson", Email: "james.wilson@penguinui.com", Membership: "Silver"},
		{ID: "2351", Name: "Sophia Chen", Email: "sophia.chen@penguinui.com", Membership: "Gold"},
		{ID: "2352", Name: "Michael Davis", Email: "m.davis@penguinui.com", Membership: "Silver"},
		{ID: "2353", Name: "Olivia Taylor", Email: "olivia.taylor@penguinui.com", Membership: "Gold"},
		{ID: "2354", Name: "Daniel Lee", Email: "daniel.lee@penguinui.com", Membership: "Silver"},
		{ID: "2355", Name: "Emma Harris", Email: "emma.harris@penguinui.com", Membership: "Gold"},
	}
}

func recordToRow(rec tableRecord) table.Row {
	return table.Row{
		ID: rec.ID,
		Cells: map[string]table.Cell{
			"id":         {Text: rec.ID},
			"name":       {Text: rec.Name},
			"email":      {Text: rec.Email},
			"membership": {Text: rec.Membership},
		},
	}
}

func recordsToRows(recs []tableRecord) []table.Row {
	rows := make([]table.Row, len(recs))
	for i, rec := range recs {
		rows[i] = recordToRow(rec)
	}
	return rows
}

func sortRecords(recs []tableRecord, orderBy string, orderDir string) {
	sort.SliceStable(recs, func(i, j int) bool {
		var a, b string
		switch orderBy {
		case "id":
			a, b = recs[i].ID, recs[j].ID
		case "name":
			a, b = strings.ToLower(recs[i].Name), strings.ToLower(recs[j].Name)
		case "email":
			a, b = strings.ToLower(recs[i].Email), strings.ToLower(recs[j].Email)
		case "membership":
			a, b = strings.ToLower(recs[i].Membership), strings.ToLower(recs[j].Membership)
		default:
			return false
		}
		if orderDir == "desc" {
			return a > b
		}
		return a < b
	})
}

func filterRecords(recs []tableRecord, search string, membership string) []tableRecord {
	if search == "" && membership == "" {
		return recs
	}
	search = strings.ToLower(search)
	var filtered []tableRecord
	for _, rec := range recs {
		if membership != "" && !strings.EqualFold(rec.Membership, membership) {
			continue
		}
		if search != "" &&
			!strings.Contains(strings.ToLower(rec.Name), search) &&
			!strings.Contains(strings.ToLower(rec.Email), search) &&
			!strings.Contains(rec.ID, search) {
			continue
		}
		filtered = append(filtered, rec)
	}
	return filtered
}

func (s *Server) handleTableRows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	q := r.URL.Query()
	variant := q.Get("variant")
	orderBy := q.Get("order_by")
	orderDir := q.Get("order_dir")
	pageStr := q.Get("page")
	perPageStr := q.Get("per_page")
	search := q.Get("search")
	membership := q.Get("membership")
	tableID := q.Get("table_id")

	records := allRecords()

	// Apply filtering
	records = filterRecords(records, search, membership)

	// Apply sorting
	if orderBy != "" {
		if orderDir == "" {
			orderDir = "asc"
		}
		sortRecords(records, orderBy, orderDir)
	}

	// Simulate server latency for lazy load / infinite scroll demos
	if variant == "lazy" || variant == "infinite" {
		time.Sleep(500 * time.Millisecond)
	}

	// Parse pagination params
	page := 1
	perPage := 3
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	totalPages := (len(records) + perPage - 1) / perPage

	// Paginate
	start := (page - 1) * perPage
	if start >= len(records) {
		start = 0
		page = 1
	}
	end := start + perPage
	if end > len(records) {
		end = len(records)
	}
	pageRecords := records[start:end]
	rows := recordsToRows(pageRecords)

	hasMore := end < len(records)
	nextPage := page + 1

	cfg := table.Config{
		Columns: []table.Column{
			{Key: "id", Label: "CustomerID", Sortable: true},
			{Key: "name", Label: "Name", Sortable: true},
			{Key: "email", Label: "Email"},
			{Key: "membership", Label: "Membership", Sortable: true},
		},
		Rows:    rows,
		SortBy:  orderBy,
		SortDir: table.SortDir(orderDir),
	}

	// For infinite scroll, render rows without tbody wrapper (appended to existing tbody)
	if variant == "infinite" {
		if hasMore {
			cfg.HTMXEndpoint = "/api/components/table/rows?variant=infinite"
			cfg.InfiniteScroll = &table.InfiniteScrollConfig{
				NextPage: nextPage,
				HasMore:  true,
			}
		}
		table.TableRows(cfg).Render(r.Context(), w)
		return
	}

	cfg.HTMXEndpoint = "/api/components/table/rows"

	if tableID != "" {
		cfg.ID = tableID
	}

	// For pagination, render rows as tbody inner HTML + OOB pagination update
	if pageStr != "" || variant == "" {
		cfg.Pagination = &table.PaginationConfig{
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     perPage,
		}
		if tableID == "" {
			cfg.ID = "paginated-table"
		}
	}

	// Render just the table rows (tbody inner content)
	for _, row := range rows {
		table.TableRow(cfg, row).Render(r.Context(), w)
	}

	// OOB swap: update sort headers so icons and next-sort URLs reflect current state.
	// Wrapped in <template> so the HTML parser doesn't strip <thead>/<tr> elements
	// when they appear alongside tbody <tr> rows in the response.
	if tableID != "" {
		fmt.Fprintf(w, `<template><thead id="%s" hx-swap-oob="outerHTML" class="%s">`,
			cfg.TheadID(), cfg.TheadClasses())
		table.TableHeadContent(cfg).Render(r.Context(), w)
		fmt.Fprintf(w, `</thead></template>`)
	}

	// OOB swap: update pagination controls so active page, prev/next states refresh
	if cfg.Pagination != nil && cfg.Pagination.TotalPages > 1 {
		fmt.Fprintf(w, `<div id="%s" hx-swap-oob="true" class="flex items-center justify-between border-t border-outline px-4 py-3 dark:border-outline-dark">`, cfg.PaginationID())
		fmt.Fprintf(w, `<div class="text-sm text-on-surface/70 dark:text-on-surface-dark/70">Page %d of %d</div>`, page, totalPages)
		table.TablePaginationNav(cfg).Render(r.Context(), w)
		fmt.Fprintf(w, `</div>`)
	}
}
