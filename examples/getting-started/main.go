package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/table"
)

// Dog represents a dog breed with metadata
type Dog struct {
	Breed       string
	Group       string
	Origin      string
	Size        string
	Temperament string
}

var breeds = []Dog{
	{Breed: "Labrador Retriever", Group: "Sporting", Origin: "Canada", Size: "Large", Temperament: "Friendly"},
	{Breed: "German Shepherd", Group: "Herding", Origin: "Germany", Size: "Large", Temperament: "Loyal"},
	{Breed: "Golden Retriever", Group: "Sporting", Origin: "Scotland", Size: "Large", Temperament: "Gentle"},
	{Breed: "French Bulldog", Group: "Non-Sporting", Origin: "France", Size: "Small", Temperament: "Playful"},
	{Breed: "Bulldog", Group: "Non-Sporting", Origin: "England", Size: "Medium", Temperament: "Calm"},
	{Breed: "Poodle", Group: "Non-Sporting", Origin: "Germany", Size: "Medium", Temperament: "Intelligent"},
	{Breed: "Beagle", Group: "Hound", Origin: "England", Size: "Small", Temperament: "Curious"},
	{Breed: "Rottweiler", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Confident"},
	{Breed: "Dachshund", Group: "Hound", Origin: "Germany", Size: "Small", Temperament: "Clever"},
	{Breed: "Yorkshire Terrier", Group: "Toy", Origin: "England", Size: "Small", Temperament: "Spirited"},
	{Breed: "Boxer", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Energetic"},
	{Breed: "Siberian Husky", Group: "Working", Origin: "Russia", Size: "Medium", Temperament: "Outgoing"},
	{Breed: "Shih Tzu", Group: "Toy", Origin: "China", Size: "Small", Temperament: "Affectionate"},
	{Breed: "Border Collie", Group: "Herding", Origin: "Scotland", Size: "Medium", Temperament: "Smart"},
	{Breed: "Doberman", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Alert"},
	{Breed: "Corgi", Group: "Herding", Origin: "Wales", Size: "Small", Temperament: "Happy"},
	{Breed: "Australian Shepherd", Group: "Herding", Origin: "USA", Size: "Medium", Temperament: "Active"},
	{Breed: "Cavalier King Charles", Group: "Toy", Origin: "England", Size: "Small", Temperament: "Graceful"},
	{Breed: "Great Dane", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Patient"},
	{Breed: "Chihuahua", Group: "Toy", Origin: "Mexico", Size: "Small", Temperament: "Charming"},
}

const perPage = 5

// columns defines the table headers
func columns() []table.Column {
	return []table.Column{
		{Key: "breed", Label: "Breed", Sortable: true},
		{Key: "group", Label: "Group", Sortable: true},
		{Key: "origin", Label: "Origin", Sortable: true},
		{Key: "size", Label: "Size", Sortable: true},
		{Key: "temperament", Label: "Temperament"},
	}
}

// filters defines the built-in filter bar controls
func filters() *table.FilterConfig {
	return &table.FilterConfig{
		Collapsible:       true,
		InitiallyExpanded: true,
		Filters: []table.Filter{
			{
				Key:         "search",
				Label:       "Search",
				Type:        table.FilterSearch,
				Placeholder: "Search breeds, origins, temperaments...",
			},
			{
				Key:   "group",
				Label: "Group",
				Type:  table.FilterSelect,
				Options: []table.FilterOption{
					{Value: "", Label: "All Groups"},
					{Value: "Sporting", Label: "Sporting"},
					{Value: "Herding", Label: "Herding"},
					{Value: "Hound", Label: "Hound"},
					{Value: "Working", Label: "Working"},
					{Value: "Non-Sporting", Label: "Non-Sporting"},
					{Value: "Toy", Label: "Toy"},
				},
			},
		},
	}
}

// dogsToRows converts Dogs into table.Row
func dogsToRows(dogs []Dog) []table.Row {
	rows := make([]table.Row, len(dogs))
	for i, d := range dogs {
		rows[i] = table.Row{
			ID: d.Breed,
			Cells: map[string]table.Cell{
				"breed":       {Text: d.Breed},
				"group":       {Text: d.Group},
				"origin":      {Text: d.Origin},
				"size":        {Text: d.Size},
				"temperament": {Text: d.Temperament},
			},
		}
	}
	return rows
}

// filterAndSort applies search, group filter, and sort to the breed list
func filterAndSort(search, group, orderBy, orderDir string) []Dog {
	var filtered []Dog
	search = strings.ToLower(search)
	for _, d := range breeds {
		if group != "" && d.Group != group {
			continue
		}
		if search != "" &&
			!strings.Contains(strings.ToLower(d.Breed), search) &&
			!strings.Contains(strings.ToLower(d.Origin), search) &&
			!strings.Contains(strings.ToLower(d.Temperament), search) {
			continue
		}
		filtered = append(filtered, d)
	}
	if orderBy != "" {
		sort.SliceStable(filtered, func(i, j int) bool {
			var a, b string
			switch orderBy {
			case "breed":
				a, b = filtered[i].Breed, filtered[j].Breed
			case "group":
				a, b = filtered[i].Group, filtered[j].Group
			case "origin":
				a, b = filtered[i].Origin, filtered[j].Origin
			case "size":
				a, b = filtered[i].Size, filtered[j].Size
			}
			if orderDir == "desc" {
				return a > b
			}
			return a < b
		})
	}
	return filtered
}

func appMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve embedded Goshtoso assets (CSS with themes, Alpine.js, HTMX, fonts)
	mux.Handle("/assets/", assets.Handler())

	// Main page — renders the full table with filters, sorting, and pagination
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		dogs := filterAndSort("", "", "breed", "asc")
		totalPages := max(1, (len(dogs)+perPage-1)/perPage)
		pageRows := dogsToRows(dogs[:min(perPage, len(dogs))])

		Page(table.Config{
			ID:           "breeds",
			HTMXEndpoint: "/api/breeds",
			Columns:      columns(),
			Rows:         pageRows,
			SortBy:       "breed",
			SortDir:      table.SortAsc,
			Pagination:   &table.PaginationConfig{CurrentPage: 1, TotalPages: totalPages, PerPage: perPage},
			Filters:      filters(),
		}).Render(r.Context(), w)
	})

	// HTMX endpoint — returns filtered/sorted/paginated table rows
	mux.HandleFunc("/api/breeds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		q := r.URL.Query()
		search := q.Get("search")
		group := q.Get("group")
		orderBy := q.Get("order_by")
		orderDir := q.Get("order_dir")
		if orderDir == "" {
			orderDir = "asc"
		}

		dogs := filterAndSort(search, group, orderBy, orderDir)

		// Pagination
		page := 1
		pp := perPage
		if v := q.Get("page"); v != "" {
			if p, err := strconv.Atoi(v); err == nil && p > 0 {
				page = p
			}
		}
		if v := q.Get("per_page"); v != "" {
			if p, err := strconv.Atoi(v); err == nil && p > 0 {
				pp = p
			}
		}
		totalPages := max(1, (len(dogs)+pp-1)/pp)
		start := (page - 1) * pp
		if start >= len(dogs) {
			start = 0
			page = 1
		}
		end := start + pp
		if end > len(dogs) {
			end = len(dogs)
		}

		cfg := table.Config{
			ID:           "breeds",
			Columns:      columns(),
			Rows:         dogsToRows(dogs[start:end]),
			HTMXEndpoint: "/api/breeds",
			SortBy:       orderBy,
			SortDir:      table.SortDir(orderDir),
			Pagination:   &table.PaginationConfig{CurrentPage: page, TotalPages: totalPages, PerPage: pp},
		}

		// Render table rows
		for _, row := range cfg.Rows {
			table.TableRow(cfg, row).Render(r.Context(), w)
		}

		// OOB: always update pagination controls so filtered one-page results
		// clear any stale paginator from the previous unfiltered table state.
		fmt.Fprintf(w, `<div id="%s" hx-swap-oob="true" class="flex items-center justify-between border-t border-gray-200 px-4 py-3">`, cfg.PaginationID())
		fmt.Fprintf(w, `<div class="text-sm text-gray-500">Page %d of %d</div>`, page, totalPages)
		table.TablePaginationNav(cfg).Render(r.Context(), w)
		fmt.Fprintf(w, `</div>`)
	})

	return mux
}

func main() {
	addr := ":3000"
	log.Printf("Dog breeds app running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, appMux()))
}
