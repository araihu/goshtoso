package catalog_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	democomponents "github.com/araihu/goshtoso/site/internal/pages/demo/components"
	"github.com/stretchr/testify/require"
)

func TestComponentCatalogHasEveryPageOnce(t *testing.T) {
	expected := []struct {
		path    string
		section string
	}{
		{"/components/app-shell", "Composition"},
		{"/components/page-header", "Composition"},
		{"/components/toolbar", "Composition"},
		{"/components/action-group", "Composition"},
		{"/components/panel", "Composition"},
		{"/components/empty-state", "Composition"},
		{"/components/skeleton", "Composition"},
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

	pages := catalog.ComponentPages()
	require.Len(t, pages, 49)
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
		expectedPackagePath := "github.com/araihu/goshtoso/" + strings.ReplaceAll(page.Key, "-", "")
		if page.Key == "components/dependencies" {
			expectedPackagePath = "github.com/araihu/goshtoso/components/head"
		}
		require.Equal(t, expectedPackagePath, page.GoPackagePath())
		activeEntry, ok := catalog.LookupActive(page.Active)
		require.True(t, ok)
		require.Equal(t, page.Key, activeEntry.Key)
		expectedDocsURL := "https://pkg.go.dev/github.com/araihu/goshtoso@v0.0.12/" + strings.ReplaceAll(page.Key, "-", "")
		if page.Key == "components/dependencies" {
			expectedDocsURL = "https://pkg.go.dev/github.com/araihu/goshtoso@v0.0.12/components/head"
		}
		require.Equal(t, expectedDocsURL, page.GoDocsURL("v0.0.12"))
		seen[page.Path] = true
	}
}

func TestDependenciesDocumentationMapsToHeadPackage(t *testing.T) {
	entry, ok := catalog.Lookup("components/dependencies")
	require.True(t, ok)
	require.Equal(t, "github.com/araihu/goshtoso/components/head", entry.GoPackagePath())
	require.Equal(
		t,
		"https://pkg.go.dev/github.com/araihu/goshtoso@v0.0.12/components/head",
		entry.GoDocsURL("v0.0.12"),
	)
}

func TestComponentCatalogMapsEveryKindExactlyOnce(t *testing.T) {
	var got []components.Kind
	seen := map[components.Kind]bool{}
	for _, page := range catalog.ComponentPages() {
		for _, kind := range page.Kinds {
			require.False(t, seen[kind], "duplicate component Kind %q", kind)
			seen[kind] = true
			got = append(got, kind)
		}
	}

	require.Len(t, got, 81)
	want := components.AllKinds()
	slices.Sort(got)
	slices.Sort(want)
	require.Equal(t, want, got)
}

func TestComponentCatalogPathsMatchDemoRegistryExactly(t *testing.T) {
	pages := catalog.ComponentPages()
	catalogKeys := make([]string, 0, len(pages))
	registryKeys := make([]string, 0, len(pages))

	for _, page := range pages {
		require.Equalf(t, "/"+page.Key, page.Path, "%s canonical component path", page.Key)
		entry, ok := democomponents.Demos[page.Key]
		require.Truef(t, ok, "missing demo registry entry for %s", page.Key)
		require.Equalf(t, page.Active, entry.Active, "%s navigation active key", page.Key)
		require.NotNilf(t, entry.Content, "%s route content", page.Key)
		catalogKeys = append(catalogKeys, strings.TrimPrefix(page.Path, "/"))
	}
	for key := range democomponents.Demos {
		if strings.HasPrefix(key, "components/") {
			registryKeys = append(registryKeys, key)
		}
	}

	slices.Sort(catalogKeys)
	slices.Sort(registryKeys)
	require.Len(t, catalogKeys, 49)
	require.Equal(t, catalogKeys, registryKeys)
}

func TestComponentCatalogReturnsDefensiveCopies(t *testing.T) {
	pages := catalog.ComponentPages()
	originalPath := pages[0].Path
	originalKind := pages[0].Kinds[0]

	pages[0].Path = "/mutated"
	pages[0].Kinds[0] = "mutated"

	fresh := catalog.ComponentPages()
	require.Equal(t, originalPath, fresh[0].Path)
	require.Equal(t, originalKind, fresh[0].Kinds[0])

	entry, ok := catalog.Lookup(fresh[0].Key)
	require.True(t, ok)
	entry.Kinds[0] = "mutated"

	freshEntry, ok := catalog.Lookup(fresh[0].Key)
	require.True(t, ok)
	require.Equal(t, originalKind, freshEntry.Kinds[0])
}
