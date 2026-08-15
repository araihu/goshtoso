package conformanceledger

import (
	"strings"
	"testing"
)

func TestEveryCurrentRouteHasExplicitChecklistMapping(t *testing.T) {
	inventory, err := DeriveInventory(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range inventory.Routes {
		mappings, err := ChecklistMappingsForRoute(route.Value)
		if err != nil {
			t.Errorf("%s: %v", route.Value, err)
			continue
		}
		if len(mappings) == 0 {
			t.Errorf("%s: empty checklist mappings", route.Value)
		}
		for _, mapping := range mappings {
			if !strings.HasPrefix(mapping.URL, "https://www.checklist.design/design-system/") {
				t.Errorf("%s: noncanonical checklist URL %q", route.Value, mapping.URL)
			}
		}
	}
}

func TestChecklistSynonymMappingsAreExplicit(t *testing.T) {
	tests := []struct {
		route string
		url   string
	}{
		{route: "/components/range", url: "https://www.checklist.design/design-system/slider"},
		{route: "/components/search", url: "https://www.checklist.design/design-system/searchbar"},
		{route: "/components/spinner", url: "https://www.checklist.design/design-system/loading"},
		{route: "/components/text-input", url: ChecklistInput},
	}
	for _, test := range tests {
		mappings, err := ChecklistMappingsForRoute(test.route)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, mapping := range mappings {
			if mapping.URL == test.url && mapping.Kind == ChecklistSynonym && mapping.Rationale != "" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s missing explicit synonym mapping to %s", test.route, test.url)
		}
	}
}

func TestUnknownRouteCannotInheritFoundationSilently(t *testing.T) {
	if _, err := ChecklistMappingsForRoute("/components/not-real"); err == nil {
		t.Fatal("unknown route unexpectedly mapped")
	}
}
