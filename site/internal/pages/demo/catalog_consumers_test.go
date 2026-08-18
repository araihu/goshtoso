package demo

import (
	"testing"

	searchfield "github.com/araihu/goshtoso/components/search"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/stretchr/testify/require"
)

func TestComponentSidebarAndNavigationFollowCatalogOrder(t *testing.T) {
	pages := make([]catalog.Entry, 0, len(catalog.ComponentPages()))
	for _, page := range catalog.ComponentPages() {
		if page.Active != "app-shell" {
			pages = append(pages, page)
		}
	}
	sections := getSidebarSections("button")

	require.Len(t, sections, 6)
	require.Equal(t, "Composition", sections[0].Title)
	require.Equal(t, "Examples", sections[5].Title)
	require.Equal(t, "Live Ticker", sections[5].Items[0].Label)
	require.Len(t, sections[5].Items, 7)

	var componentItems int
	pageIndex := 0
	for _, section := range sections[:5] {
		for _, item := range section.Items {
			page := pages[pageIndex]
			require.Equal(t, page.Section, section.Title)
			require.Equal(t, page.Active, item.ID)
			require.Equal(t, page.Title, item.Label)
			require.Equal(t, page.Path, item.Href)
			require.Equal(t, page.Active == "button", item.Active)
			pageIndex++
			componentItems++
		}
	}
	require.Len(t, pages, componentItems)

	for i, page := range pages {
		prev, next := getComponentNav(page.Active)
		if i == 0 {
			require.Nil(t, prev)
		} else {
			require.Equal(t, pages[i-1].Title, prev.Label)
			require.Equal(t, pages[i-1].Path, prev.Href)
		}
		if i == len(pages)-1 {
			require.Nil(t, next)
		} else {
			require.Equal(t, pages[i+1].Title, next.Label)
			require.Equal(t, pages[i+1].Path, next.Href)
		}
	}
}

func TestComponentSearchEntriesUseCatalogDescriptions(t *testing.T) {
	items := getSearchItems()
	byHref := make(map[string]struct {
		title       string
		description string
		section     string
	}, len(items))
	for _, item := range items {
		byHref[item.Href] = struct {
			title       string
			description string
			section     string
		}{
			title:       item.Title,
			description: item.Description,
			section:     item.Section,
		}
	}

	for _, page := range catalog.ComponentPages() {
		if page.Active == "app-shell" {
			_, exists := byHref[page.Path]
			require.False(t, exists, "legacy App Shell page should not be discoverable beside the App Shells module")
			continue
		}
		item, ok := byHref[page.Path]
		require.Truef(t, ok, "missing search entry for %s", page.Path)
		require.Equal(t, page.Title, item.title)
		require.Equal(t, page.Description, item.description)
		require.Equal(t, page.Section, item.section)
	}

	_, hasExamples := byHref["/examples/todo"]
	require.True(t, hasExamples, "non-component example search entries must remain")
	_, hasDocs := byHref["/getting-started"]
	require.True(t, hasDocs, "non-component docs search entries must remain")
	iconCatalog, hasIconCatalog := byHref["/docs/icon-catalog"]
	require.True(t, hasIconCatalog, "dedicated Icon Catalog must be discoverable")
	require.Equal(t, "Icon Catalog", iconCatalog.title)
	iconPacks, hasIconPacks := byHref["/docs/iconpack"]
	require.True(t, hasIconPacks, "Icon Packs must remain discoverable")
	require.Equal(t, "Icon Packs", iconPacks.title)
	for _, link := range appShellsSidebarTopLinks() {
		_, exists := byHref[link.Href]
		require.Truef(t, exists, "missing App Shells search entry for %s", link.Href)
	}
	for _, group := range appShellsSidebarGroups() {
		for _, link := range group.Items {
			_, exists := byHref[link.Href]
			require.Truef(t, exists, "missing App Shells search entry for %s", link.Href)
		}
	}
}

func TestGlobalSearchBoostsActiveFamilyWithoutHidingFallbacks(t *testing.T) {
	core := searchItemsByHref(getSearchItems("core"))
	charts := searchItemsByHref(getSearchItems("charts"))

	for _, href := range []string{
		"/components/button",
		"/docs/icon-catalog",
		"/modules/charts/components/line",
		"/modules/app-shells/shells/console-shell",
		"/examples/ticker",
	} {
		_, coreOK := core[href]
		_, chartsOK := charts[href]
		require.Truef(t, coreOK, "core search is missing %s", href)
		require.Truef(t, chartsOK, "charts search is missing fallback %s", href)
	}

	require.Equal(t, len(core), len(charts), "changing family should boost results, not remove them")
	require.Equal(t, 1, core["/components/button"].Priority)
	require.Equal(t, 0, core["/modules/charts/components/line"].Priority)
	require.Equal(t, 1, charts["/modules/charts/components/line"].Priority)
	require.Equal(t, 0, charts["/components/button"].Priority)
}

func searchItemsByHref(items []searchfield.Item) map[string]searchfield.Item {
	byHref := make(map[string]searchfield.Item, len(items))
	for _, item := range items {
		byHref[item.Href] = item
	}
	return byHref
}
