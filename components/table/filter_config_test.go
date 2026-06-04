package table

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestFilterConfig_ResolvedHxTarget pins the override contract: explicit
// HxTarget wins over the default "#{tbody-id}" resolution. The modal
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
			name:   "explicit HxTarget wins",
			filter: FilterConfig{HxTarget: "#install-modal-body"},
			cfg:    Config{ID: "clusters"},
			want:   "#install-modal-body",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.ResolvedHxTarget(tc.cfg)
			if got != tc.want {
				t.Fatalf("ResolvedHxTarget = %q; want %q", got, tc.want)
			}
		})
	}
}

// TestFilterScriptData_EmitsHxTarget guards against regressions in the
// Alpine.data template: the emitted `applyFilters()` body must use the
// resolved target string, not the raw ID. Without this, adding the
// HxTarget override silently had no effect because the Sprintf still
// referenced tbody.
func TestFilterScriptData_EmitsHxTarget(t *testing.T) {
	cfg := Config{
		ID:   "addon-picker",
		HTMX: &HTMXConfig{Endpoint: "/console/clusters/cid/addons/install"},
		Filters: &FilterConfig{
			HxTarget: "#install-modal-body",
			Filters: []Filter{
				{Key: "q", Type: FilterSearch},
			},
		},
	}
	out := filterScriptData(cfg)
	if !strings.Contains(out, "target: '#install-modal-body'") {
		t.Fatalf("filterScriptData missing explicit target; got:\n%s", out)
	}
	// Make sure the default tbody selector does NOT leak into the same
	// applyFilters call.
	if strings.Contains(out, "target: '#addon-picker-tbody'") {
		t.Fatalf("filterScriptData still emits default tbody target despite override; got:\n%s", out)
	}
}

// TestFilterVariant_Constants keeps the enum surface honest — consumers
// import these, so renaming is a breaking change.
func TestFilterVariant_Constants(t *testing.T) {
	if FilterVariantBar != "" {
		t.Fatalf("FilterVariantBar must be empty string (zero value); got %q", FilterVariantBar)
	}
	if FilterVariantInline != "inline" {
		t.Fatalf("FilterVariantInline must be %q; got %q", "inline", FilterVariantInline)
	}
}

// TestFilterConfig_ResolvedHxSwap mirrors the HxTarget override contract for
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
		{name: "explicit outerHTML wins", filter: FilterConfig{HxSwap: "outerHTML"}, want: "outerHTML"},
		{name: "arbitrary swap mode passes through", filter: FilterConfig{HxSwap: "morphdom"}, want: "morphdom"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.ResolvedHxSwap()
			if got != tc.want {
				t.Fatalf("ResolvedHxSwap = %q; want %q", got, tc.want)
			}
		})
	}
}

// TestFilterScriptData_EmitsHxSwap guards Sprintf positional arity for the
// new swap arg. Without the test the loop ordering between hxTarget and
// hxSwap could silently swap and produce target='innerHTML' / swap='#x'.
func TestFilterScriptData_EmitsHxSwap(t *testing.T) {
	cfg := Config{
		ID:   "addons-catalog-table",
		HTMX: &HTMXConfig{Endpoint: "/console/addons"},
		Filters: &FilterConfig{
			HxTarget: "#addons-catalog",
			HxSwap:   "outerHTML",
			Filters:  []Filter{{Key: "search", Type: FilterSearch}},
		},
	}
	out := filterScriptData(cfg)
	if !strings.Contains(out, "target: '#addons-catalog'") {
		t.Fatalf("filterScriptData missing target; got:\n%s", out)
	}
	if !strings.Contains(out, "swap: 'outerHTML'") {
		t.Fatalf("filterScriptData missing outerHTML swap; got:\n%s", out)
	}
}

// TestFilterScriptData_DefaultSwap pins the default swap behavior so we don't
// regress callers that omit HxSwap.
func TestFilterScriptData_DefaultSwap(t *testing.T) {
	cfg := Config{
		ID:      "clusters",
		HTMX:    &HTMXConfig{Endpoint: "/console/clusters"},
		Filters: &FilterConfig{Filters: []Filter{{Key: "q", Type: FilterSearch}}},
	}
	out := filterScriptData(cfg)
	if !strings.Contains(out, "swap: 'innerHTML'") {
		t.Fatalf("filterScriptData missing default innerHTML swap; got:\n%s", out)
	}
}

// TestFilterScriptData_PreservesExtraQueryParams locks the contract that
// ExtraQueryParams (already prefixed with '&') is appended to the auto
// `?_filter=1` marker so static query state survives every filter request.
// Modal flows depend on this — the modal's `?addon_name=X` context must
// follow filter swaps, otherwise the BFF can't resolve which addon's
// clusters to filter.
func TestFilterScriptData_PreservesExtraQueryParams(t *testing.T) {
	cfg := Config{
		ID:               "cluster-picker-table",
		HTMX:             &HTMXConfig{Endpoint: "/console/addons/install"},
		ExtraQueryParams: "&addon_name=argo-cd",
		Filters:          &FilterConfig{Filters: []Filter{{Key: "q", Type: FilterSearch}}},
	}
	out := filterScriptData(cfg)
	if !strings.Contains(out, "?_filter=1&addon_name=argo-cd") {
		t.Fatalf("filterScriptData lost ExtraQueryParams in filter URL; got:\n%s", out)
	}
}

func TestFilterBar_NonCollapsibleBodyHasTopPadding(t *testing.T) {
	cfg := Config{
		ID:      "ticker-table",
		HTMX:    &table.HTMXConfig{Endpoint: "/api/examples/ticker/rows"},
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
