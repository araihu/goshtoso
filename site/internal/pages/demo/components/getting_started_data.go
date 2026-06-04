package components

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/components/table"
)

const gettingStartedPerPage = 5

type gettingStartedDog struct {
	Breed       string
	Image       string
	Group       string
	Origin      string
	Size        string
	Temperament string
}

var gettingStartedDogs = []gettingStartedDog{
	{Breed: "Labrador Retriever", Image: "/dog-images/labrador-retriever.webp", Group: "Sporting", Origin: "Canada", Size: "Large", Temperament: "Friendly"},
	{Breed: "German Shepherd", Image: "/dog-images/german-shepherd.webp", Group: "Herding", Origin: "Germany", Size: "Large", Temperament: "Loyal"},
	{Breed: "Golden Retriever", Image: "/dog-images/golden-retriever.webp", Group: "Sporting", Origin: "Scotland", Size: "Large", Temperament: "Gentle"},
	{Breed: "French Bulldog", Image: "/dog-images/french-bulldog.webp", Group: "Non-Sporting", Origin: "France", Size: "Small", Temperament: "Playful"},
	{Breed: "Bulldog", Image: "/dog-images/bulldog.webp", Group: "Non-Sporting", Origin: "England", Size: "Medium", Temperament: "Calm"},
	{Breed: "Poodle", Image: "/dog-images/poodle.webp", Group: "Non-Sporting", Origin: "Germany", Size: "Medium", Temperament: "Intelligent"},
	{Breed: "Beagle", Image: "/dog-images/beagle.webp", Group: "Hound", Origin: "England", Size: "Small", Temperament: "Curious"},
	{Breed: "Rottweiler", Image: "/dog-images/rottweiler.webp", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Confident"},
	{Breed: "Dachshund", Image: "/dog-images/dachshund.webp", Group: "Hound", Origin: "Germany", Size: "Small", Temperament: "Clever"},
	{Breed: "Yorkshire Terrier", Image: "/dog-images/yorkshire-terrier.webp", Group: "Toy", Origin: "England", Size: "Small", Temperament: "Spirited"},
	{Breed: "Boxer", Image: "/dog-images/boxer.webp", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Energetic"},
	{Breed: "Siberian Husky", Image: "/dog-images/siberian-husky.webp", Group: "Working", Origin: "Russia", Size: "Medium", Temperament: "Outgoing"},
	{Breed: "Shih Tzu", Image: "/dog-images/shih-tzu.webp", Group: "Toy", Origin: "China", Size: "Small", Temperament: "Affectionate"},
	{Breed: "Border Collie", Image: "/dog-images/border-collie.webp", Group: "Herding", Origin: "Scotland", Size: "Medium", Temperament: "Smart"},
	{Breed: "Doberman", Image: "/dog-images/doberman.webp", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Alert"},
	{Breed: "Corgi", Image: "/dog-images/corgi.webp", Group: "Herding", Origin: "Wales", Size: "Small", Temperament: "Happy"},
	{Breed: "Australian Shepherd", Image: "/dog-images/australian-shepherd.webp", Group: "Herding", Origin: "USA", Size: "Medium", Temperament: "Active"},
	{Breed: "Cavalier King Charles", Image: "/dog-images/cavalier-king-charles.webp", Group: "Toy", Origin: "England", Size: "Small", Temperament: "Graceful"},
	{Breed: "Great Dane", Image: "/dog-images/great-dane.webp", Group: "Working", Origin: "Germany", Size: "Large", Temperament: "Patient"},
	{Breed: "Chihuahua", Image: "/dog-images/chihuahua.webp", Group: "Toy", Origin: "Mexico", Size: "Small", Temperament: "Charming"},
}

func GettingStartedPreviewConfigFromQuery(q url.Values) table.Config {
	search := q.Get("search")
	group := q.Get("group")
	orderBy := q.Get("order_by")
	orderDir := q.Get("order_dir")
	if orderDir == "" && orderBy != "" {
		orderDir = string(table.SortAsc)
	}
	page, perPage := gettingStartedPageParams(q.Get("page"), q.Get("per_page"))

	dogs := gettingStartedFilterDogs(search, group)
	if orderBy != "" {
		gettingStartedSortDogs(dogs, orderBy, orderDir)
	}

	totalPages := max(1, (len(dogs)+perPage-1)/perPage)
	start := (page - 1) * perPage
	if start >= len(dogs) {
		start = 0
		page = 1
	}
	end := min(start+perPage, len(dogs))

	return table.Config{
		ID:         "getting-started-dogs",
		HTMX:       &table.HTMXConfig{Endpoint: "/api/getting-started/breeds"},
		Columns:    gettingStartedColumns(),
		Rows:       gettingStartedRows(dogs[start:end]),
		SortBy:     orderBy,
		SortDir:    table.SortDir(orderDir),
		Pagination: &table.PaginationConfig{CurrentPage: page, TotalPages: totalPages, PerPage: perPage},
		Filters:    gettingStartedFilters(),
	}
}

func gettingStartedInitialConfig() table.Config {
	return GettingStartedPreviewConfigFromQuery(url.Values{
		"order_by":  []string{"breed"},
		"order_dir": []string{string(table.SortAsc)},
	})
}

func gettingStartedColumns() []table.Column {
	return []table.Column{
		{Key: "breed", Label: "Breed", Sortable: true},
		{Key: "group", Label: "Group", Sortable: true},
		{Key: "origin", Label: "Origin", Sortable: true},
		{Key: "size", Label: "Size", Sortable: true},
		{Key: "temperament", Label: "Temperament"},
	}
}

func gettingStartedFilters() *table.FilterConfig {
	return &table.FilterConfig{
		Collapsible:       true,
		InitiallyExpanded: true,
		Filters: []table.Filter{
			{Key: "search", Label: "Search", Type: table.FilterSearch, Placeholder: "Search breeds, origins, temperaments..."},
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

func gettingStartedRows(dogs []gettingStartedDog) []table.Row {
	rows := make([]table.Row, len(dogs))
	for i, dog := range dogs {
		rows[i] = table.Row{
			ID: dog.Breed,
			Cells: map[string]table.Cell{
				"breed":       {Component: table.ImageCell(dog.Image, dog.Breed, dog.Group)},
				"group":       {Text: dog.Group},
				"origin":      {Text: dog.Origin},
				"size":        {Text: dog.Size},
				"temperament": {Text: dog.Temperament},
			},
		}
	}
	return rows
}

func gettingStartedFilterDogs(search, group string) []gettingStartedDog {
	search = strings.ToLower(search)
	out := make([]gettingStartedDog, 0, len(gettingStartedDogs))
	for _, dog := range gettingStartedDogs {
		if group != "" && dog.Group != group {
			continue
		}
		if search != "" &&
			!strings.Contains(strings.ToLower(dog.Breed), search) &&
			!strings.Contains(strings.ToLower(dog.Origin), search) &&
			!strings.Contains(strings.ToLower(dog.Temperament), search) {
			continue
		}
		out = append(out, dog)
	}
	return out
}

func gettingStartedSortDogs(dogs []gettingStartedDog, orderBy, orderDir string) {
	sort.SliceStable(dogs, func(i, j int) bool {
		var a, b string
		switch orderBy {
		case "breed":
			a, b = dogs[i].Breed, dogs[j].Breed
		case "group":
			a, b = dogs[i].Group, dogs[j].Group
		case "origin":
			a, b = dogs[i].Origin, dogs[j].Origin
		case "size":
			a, b = dogs[i].Size, dogs[j].Size
		default:
			return false
		}
		if orderDir == string(table.SortDesc) {
			return a > b
		}
		return a < b
	})
}

func gettingStartedPageParams(pageStr, perPageStr string) (page, perPage int) {
	page, perPage = 1, gettingStartedPerPage
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
