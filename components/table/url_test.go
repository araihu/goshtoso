package table

import (
	"net/url"
	"testing"
)

func queryValue(t *testing.T, rawURL, key string) string {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse %q: %v", rawURL, err)
	}
	return u.Query().Get(key)
}

func TestSortURLPreservesEndpointQueryAndEncodesValues(t *testing.T) {
	cfg := Config{
		ID:         "orders table",
		HTMX:       &HTMXConfig{Endpoint: "/api/components/table/rows?variant=compact"},
		SortBy:     "",
		SortDir:    SortNone,
		Pagination: &PaginationConfig{PerPage: 25},
	}

	got := cfg.SortURL("user name")

	if queryValue(t, got, "variant") != "compact" {
		t.Fatalf("SortURL lost endpoint query: %q", got)
	}
	if queryValue(t, got, "table_id") != "orders table" {
		t.Fatalf("SortURL table_id = %q", queryValue(t, got, "table_id"))
	}
	if queryValue(t, got, "order_by") != "user name" {
		t.Fatalf("SortURL order_by = %q", queryValue(t, got, "order_by"))
	}
	if queryValue(t, got, "per_page") != "25" {
		t.Fatalf("SortURL per_page = %q", queryValue(t, got, "per_page"))
	}
}

func TestSortURLResetOmitsSortParams(t *testing.T) {
	cfg := Config{
		ID:         "orders",
		HTMX:       &HTMXConfig{Endpoint: "/api/components/table/rows"},
		SortBy:     "name",
		SortDir:    SortDesc,
		Pagination: &PaginationConfig{PerPage: 10},
	}

	got := cfg.SortURL("name")

	if queryValue(t, got, "order_by") != "" {
		t.Fatalf("SortURL reset kept order_by: %q", got)
	}
	if queryValue(t, got, "order_dir") != "" {
		t.Fatalf("SortURL reset kept order_dir: %q", got)
	}
}

func TestPageURLMergesExtraQueryParams(t *testing.T) {
	cfg := Config{
		HTMX:             &HTMXConfig{Endpoint: "/api/components/table/rows?variant=static"},
		ExtraQueryParams: "&addon_name=argo cd&membership=admin",
		Pagination:       &PaginationConfig{PerPage: 50},
		SortBy:           "created at",
		SortDir:          SortAsc,
	}

	got := cfg.PageURL(3)

	for key, want := range map[string]string{
		"variant":    "static",
		"addon_name": "argo cd",
		"membership": "admin",
		"page":       "3",
		"per_page":   "50",
		"order_by":   "created at",
		"order_dir":  "asc",
	} {
		if gotVal := queryValue(t, got, key); gotVal != want {
			t.Fatalf("PageURL query %s = %q; want %q in %q", key, gotVal, want, got)
		}
	}
}

func TestNextPageURLForPaginationInfiniteScroll(t *testing.T) {
	cfg := Config{
		HTMX: &HTMXConfig{Endpoint: "/api/components/table/rows"},
		Pagination: &PaginationConfig{
			Mode:        PaginationInfiniteScroll,
			CurrentPage: 2,
			PerPage:     15,
			HasMore:     true,
		},
	}

	got := cfg.NextPageURL()

	if queryValue(t, got, "page") != "3" {
		t.Fatalf("NextPageURL page = %q in %q", queryValue(t, got, "page"), got)
	}
	if queryValue(t, got, "variant") != "infinite" {
		t.Fatalf("NextPageURL variant = %q in %q", queryValue(t, got, "variant"), got)
	}
}
