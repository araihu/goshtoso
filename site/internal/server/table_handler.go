package server

import (
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/araihu/goshtoso/components/table"
)

const tableHeadClasses = "border-b border-outline bg-surface-alt text-sm text-on-surface-strong dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark-strong"

var safeTableIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

func resolvedTableID(cfg table.Config) string {
	if cfg.ID != "" {
		return cfg.ID
	}
	return "table"
}

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

// parsePageParams reads page/per_page from their raw query strings, falling
// back to the demo defaults (page 1, 3 rows) when absent or invalid.
func parsePageParams(pageStr, perPageStr string) (page, perPage int) {
	page, perPage = 1, 3
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
	return page, perPage
}

func tableRowsEndpoint(variant string, perPage int) string {
	const endpoint = "/api/components/table/rows"
	if variant == "" {
		return endpoint
	}
	query := url.Values{}
	query.Set("variant", variant)
	query.Set("per_page", strconv.Itoa(perPage))
	return endpoint + "?" + query.Encode()
}

func activeTableFilterQuery(search, membership string) string {
	query := url.Values{}
	if search != "" {
		query.Set("search", search)
	}
	if membership != "" {
		query.Set("membership", membership)
	}
	return query.Encode()
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
	if tableID != "" && !safeTableIDPattern.MatchString(tableID) {
		http.Error(w, "invalid table_id", http.StatusBadRequest)
		return
	}

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
	page, perPage := parsePageParams(pageStr, perPageStr)

	totalPages := (len(records) + perPage - 1) / perPage

	// Paginate
	start := (page - 1) * perPage
	if start >= len(records) {
		start = 0
		page = 1
	}
	end := min(start+perPage, len(records))
	pageRecords := records[start:end]
	rows := recordsToRows(pageRecords)

	hasMore := end < len(records)
	nextPage := page + 1

	columns := []table.Column{
		{Key: "id", Label: "CustomerID", Sortable: true},
		{Key: "name", Label: "Name", Sortable: true},
		{Key: "email", Label: "Email"},
		{Key: "membership", Label: "Membership", Sortable: true},
	}
	if variant == "inline-filtered" {
		columns = []table.Column{
			{Key: "id", Label: "CustomerID"},
			{Key: "name", Label: "Name"},
			{Key: "email", Label: "Email"},
			{Key: "membership", Label: "Membership"},
		}
	}

	cfg := table.Config{
		Columns: columns,
		Rows:    rows,
		SortBy:  orderBy,
		SortDir: table.SortDir(orderDir),
	}
	cfg.ExtraQueryParams = activeTableFilterQuery(search, membership)

	// For infinite scroll, render rows without tbody wrapper (appended to existing tbody)
	if variant == "infinite" {
		if hasMore {
			cfg.HTMX = &table.HTMXConfig{Endpoint: "/api/components/table/rows?variant=infinite"}
			cfg.InfiniteScroll = &table.InfiniteScrollConfig{
				NextPage: nextPage,
				HasMore:  true,
			}
		}
		_ = table.TableRows(cfg).Render(r.Context(), w)
		return
	}

	cfg.HTMX = &table.HTMXConfig{Endpoint: tableRowsEndpoint(variant, perPage)}

	if tableID != "" {
		cfg.ID = tableID
	}

	// For pagination, render rows as tbody inner HTML + OOB pagination update
	if pageStr != "" || variant == "" || q.Has("_filter") {
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
		_ = table.TableRow(cfg, row).Render(r.Context(), w)
	}

	// OOB swap: update sort headers so icons and next-sort URLs reflect current state.
	// Wrapped in <template> so the HTML parser doesn't strip <thead>/<tr> elements
	// when they appear alongside tbody <tr> rows in the response.
	if tableID != "" {
		_ = tableHeadOOBFragment(
			resolvedTableID(cfg)+"-thead",
			cfg.SortBy,
			string(cfg.SortDir),
			table.TableHeadContent(cfg),
		).Render(r.Context(), w)
	}

	// OOB swap: always replace the stable pagination host. One-page and empty
	// results hide it without removing the target needed to restore controls.
	if cfg.Pagination != nil {
		_ = tablePaginationOOBFragment(
			resolvedTableID(cfg)+"-pagination",
			page,
			totalPages,
			table.TablePaginationNav(cfg),
		).Render(r.Context(), w)
	}
}
