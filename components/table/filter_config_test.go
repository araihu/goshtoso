package table

import (
	"bytes"
	"context"
	"html"
	"strings"
	"testing"
)

func renderFilterContract(t *testing.T, cfg Config) string {
	t.Helper()
	if len(cfg.Columns) == 0 {
		cfg.Columns = []Column{{Key: "name", Label: "Name"}}
	}
	var buf bytes.Buffer
	if err := Table(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render table: %v", err)
	}
	return buf.String()
}

// TestFilterConfig_ResolvedHxTarget pins the override contract: explicit
// HTMX.Target wins over the default "#{tbody-id}" resolution. The modal
// migration in tks-console depends on this — filter input must swap the
// full modal body, not the table tbody that lives inside it.
func TestFilterConfig_ResolvedHxTarget(t *testing.T) {
	cases := []struct {
		name   string
		filter FilterConfig
		cfg    Config
		want   string
	}{
		{
			name:   "default falls back to tbody id",
			filter: FilterConfig{},
			cfg:    Config{ID: "clusters"},
			want:   "#clusters-tbody",
		},
		{
			name:   "explicit HTMX.Target wins",
			filter: FilterConfig{HTMX: &FilterHTMXConfig{Target: "#install-modal-body"}},
			cfg:    Config{ID: "clusters"},
			want:   "#install-modal-body",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.resolvedHXTarget(tc.cfg)
			if got != tc.want {
				t.Fatalf("resolvedHXTarget = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestFilterRuntimeContractEmitsResolvedHXTarget(t *testing.T) {
	cfg := Config{
		ID:   "addon-picker",
		HTMX: &HTMXConfig{Endpoint: "/console/clusters/cid/addons/install"},
		Filters: &FilterConfig{
			HTMX: &FilterHTMXConfig{Target: "#install-modal-body"},
			Filters: []Filter{
				{Key: "q", Type: FilterSearch},
			},
		},
	}
	out := renderFilterContract(t, cfg)
	if !strings.Contains(out, `data-table-filter-target="#install-modal-body"`) {
		t.Fatalf("filter contract missing explicit target; got:\n%s", out)
	}
	if strings.Contains(out, `data-table-filter-target="#addon-picker-tbody"`) {
		t.Fatalf("filter contract still emits default tbody target despite override; got:\n%s", out)
	}
}

func TestFilterRuntimeContractKeepsKeysInInertAttributes(t *testing.T) {
	key := "q: '', injected: alert(1), q2"
	cfg := Config{
		ID:   "addons",
		HTMX: &HTMXConfig{Endpoint: "/console/addons"},
		Filters: &FilterConfig{
			Filters: []Filter{
				{Key: key, Type: FilterSearch},
			},
		},
	}

	out := renderFilterContract(t, cfg)
	want := `data-table-filter-key="` + html.EscapeString(key) + `"`
	if !strings.Contains(out, want) {
		t.Fatalf("filter key missing inert attribute %q:\n%s", want, out)
	}
	if strings.Contains(out, `x-model="filters[&#39;q:`) || strings.Contains(out, "<script") {
		t.Fatalf("filter key escaped into executable JavaScript:\n%s", out)
	}
}

func TestFilterRuntimeContractKeepsDefaultsInInertAttributes(t *testing.T) {
	defaultValue := "O'Reilly\\docs\nline two\r</script>"
	cfg := Config{
		ID:   "addons",
		HTMX: &HTMXConfig{Endpoint: "/console/addons"},
		Filters: &FilterConfig{
			Filters: []Filter{{
				Key:          "q",
				Type:         FilterSearch,
				DefaultValue: defaultValue,
			}},
		},
	}

	out := renderFilterContract(t, cfg)
	want := `data-table-filter-default="` + html.EscapeString(defaultValue) + `"`
	if !strings.Contains(out, want) {
		t.Fatalf("filter default missing inert attribute %q:\n%s", want, out)
	}
	if strings.Contains(out, "<script") {
		t.Fatalf("filter default emitted executable script:\n%s", out)
	}
}

// TestFilterAppearanceConstants keeps the enum surface honest — consumers
// import these, so renaming is a breaking change.
func TestFilterAppearanceConstants(t *testing.T) {
	if FilterAppearanceBar != "" {
		t.Fatalf("FilterAppearanceBar must be empty string (zero value); got %q", FilterAppearanceBar)
	}
	if FilterAppearanceInline != "inline" {
		t.Fatalf("FilterAppearanceInline must be %q; got %q", "inline", FilterAppearanceInline)
	}
}

// TestFilterConfig_ResolvedHxSwap mirrors the HTMX.Target override contract for
// swap strategy. Default is "innerHTML"; consumers can opt into "outerHTML"
// when the swap target is itself a wrapper that the server re-renders
// whole-cloth (catalog grid with empty-state on the wrapper).
func TestFilterConfig_ResolvedHxSwap(t *testing.T) {
	cases := []struct {
		name   string
		filter FilterConfig
		want   string
	}{
		{name: "default falls back to innerHTML", filter: FilterConfig{}, want: "innerHTML"},
		{name: "explicit outerHTML wins", filter: FilterConfig{HTMX: &FilterHTMXConfig{Swap: "outerHTML"}}, want: "outerHTML"},
		{name: "arbitrary swap mode passes through", filter: FilterConfig{HTMX: &FilterHTMXConfig{Swap: "morphdom"}}, want: "morphdom"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.resolvedHXSwap()
			if got != tc.want {
				t.Fatalf("resolvedHXSwap = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestFilterRuntimeContractEmitsHXSwap(t *testing.T) {
	cfg := Config{
		ID:   "addons-catalog-table",
		HTMX: &HTMXConfig{Endpoint: "/console/addons"},
		Filters: &FilterConfig{
			HTMX: &FilterHTMXConfig{
				Target: "#addons-catalog",
				Swap:   "outerHTML",
			},
			Filters: []Filter{{Key: "search", Type: FilterSearch}},
		},
	}
	out := renderFilterContract(t, cfg)
	if !strings.Contains(out, `data-table-filter-target="#addons-catalog"`) {
		t.Fatalf("filter contract missing target; got:\n%s", out)
	}
	if !strings.Contains(out, `data-table-filter-swap="outerHTML"`) {
		t.Fatalf("filter contract missing outerHTML swap; got:\n%s", out)
	}
}

func TestFilterRuntimeContractDefaultSwap(t *testing.T) {
	cfg := Config{
		ID:      "clusters",
		HTMX:    &HTMXConfig{Endpoint: "/console/clusters"},
		Filters: &FilterConfig{Filters: []Filter{{Key: "q", Type: FilterSearch}}},
	}
	out := renderFilterContract(t, cfg)
	if !strings.Contains(out, `data-table-filter-swap="innerHTML"`) {
		t.Fatalf("filter contract missing default innerHTML swap; got:\n%s", out)
	}
}

func TestFilterRuntimeContractPreservesExtraQueryParamsAndPerPage(t *testing.T) {
	cfg := Config{
		ID:               "cluster-picker-table",
		HTMX:             &HTMXConfig{Endpoint: "/console/addons/install"},
		ExtraQueryParams: "&addon_name=argo-cd",
		Filters:          &FilterConfig{Filters: []Filter{{Key: "q", Type: FilterSearch}}},
		Pagination:       &PaginationConfig{PerPage: 25},
	}
	out := renderFilterContract(t, cfg)
	for _, want := range []string{
		`data-table-filter-endpoint="/console/addons/install"`,
		`data-table-filter-extra-query="&amp;addon_name=argo-cd"`,
		`data-table-filter-per-page="25"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("filter contract missing %q; got:\n%s", want, out)
		}
	}
}

func TestFilterBar_NonCollapsibleBodyHasTopPadding(t *testing.T) {
	cfg := Config{
		ID:      "ticker-table",
		HTMX:    &HTMXConfig{Endpoint: "/api/examples/ticker/rows"},
		Columns: []Column{{Key: "symbol", Label: "Symbol"}},
		Rows:    []Row{{ID: "AAPL", Cells: map[string]Cell{"symbol": {Text: "AAPL"}}}},
		Filters: &FilterConfig{
			Filters: []Filter{{Key: "search", Label: "Filter", Type: FilterSearch}},
		},
	}

	var buf bytes.Buffer
	if err := Table(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render table: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `class="px-4 pb-4" class="px-4 py-4"`) {
		t.Fatalf("non-collapsible filter body rendered duplicate classes without top padding:\n%s", out)
	}
	if !strings.Contains(out, `x-collapse class="px-4 py-4"`) {
		t.Fatalf("non-collapsible filter body missing top padding class; got:\n%s", out)
	}
}
