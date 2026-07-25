package catalog

import (
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

func TestComponentCatalogHasEveryPageOnce(t *testing.T) {
	expected := []struct {
		path    string
		section string
	}{
		{"/components/accordion", "Display"},
		{"/components/avatar", "Display"},
		{"/components/badge", "Display"},
		{"/components/banner", "Display"},
		{"/components/card", "Display"},
		{"/components/carousel", "Display"},
		{"/components/chatbubble", "Display"},
		{"/components/codeblock", "Display"},
		{"/components/dependencies", "Display"},
		{"/components/kbd", "Display"},
		{"/components/table", "Display"},
		{"/components/button", "Input"},
		{"/components/checkbox", "Input"},
		{"/components/combobox", "Input"},
		{"/components/fileinput", "Input"},
		{"/components/form", "Input"},
		{"/components/radio", "Input"},
		{"/components/range", "Input"},
		{"/components/rating", "Input"},
		{"/components/palette", "Input"},
		{"/components/search", "Input"},
		{"/components/select", "Input"},
		{"/components/schema-form", "Input"},
		{"/components/structured-input", "Input"},
		{"/components/tags-list", "Input"},
		{"/components/text-input", "Input"},
		{"/components/textarea", "Input"},
		{"/components/toggle", "Input"},
		{"/components/alert", "Feedback"},
		{"/components/toast", "Feedback"},
		{"/components/modal", "Feedback"},
		{"/components/drawer", "Feedback"},
		{"/components/spinner", "Feedback"},
		{"/components/steps", "Feedback"},
		{"/components/tooltip", "Feedback"},
		{"/components/breadcrumbs", "Navigation"},
		{"/components/dropdown", "Navigation"},
		{"/components/link", "Navigation"},
		{"/components/navbar", "Navigation"},
		{"/components/pagination", "Navigation"},
		{"/components/sidebar", "Navigation"},
		{"/components/tabs", "Navigation"},
	}

	pages := ComponentPages()
	require.Len(t, pages, 42)
	require.Len(t, pages, len(expected))

	seen := map[string]bool{}
	for i, page := range pages {
		require.Equal(t, expected[i].path, page.Path)
		require.Equal(t, expected[i].section, page.Section)
		require.Equal(t, i, page.Order)
		require.False(t, seen[page.Path], "duplicate component path %q", page.Path)
		require.NotEmpty(t, page.Key)
		require.NotEmpty(t, page.Title)
		require.NotEmpty(t, page.Active)
		require.NotEmpty(t, page.Description)
		seen[page.Path] = true
	}
}

func TestComponentCatalogMapsEveryKindExactlyOnce(t *testing.T) {
	var got []components.Kind
	seen := map[components.Kind]bool{}
	for _, page := range ComponentPages() {
		for _, kind := range page.Kinds {
			require.False(t, seen[kind], "duplicate component Kind %q", kind)
			seen[kind] = true
			got = append(got, kind)
		}
	}

	require.Len(t, got, 74)
	require.ElementsMatch(t, components.AllKinds(), got)
}

func TestComponentCatalogReturnsDefensiveCopies(t *testing.T) {
	pages := ComponentPages()
	originalPath := pages[0].Path
	originalKind := pages[0].Kinds[0]

	pages[0].Path = "/mutated"
	pages[0].Kinds[0] = "mutated"

	fresh := ComponentPages()
	require.Equal(t, originalPath, fresh[0].Path)
	require.Equal(t, originalKind, fresh[0].Kinds[0])

	entry, ok := Lookup(fresh[0].Key)
	require.True(t, ok)
	entry.Kinds[0] = "mutated"

	freshEntry, ok := Lookup(fresh[0].Key)
	require.True(t, ok)
	require.Equal(t, originalKind, freshEntry.Kinds[0])
}
