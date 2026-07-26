package demo

import (
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/stretchr/testify/require"
)

func TestComponentSidebarAndNavigationFollowCatalogOrder(t *testing.T) {
	pages := catalog.ComponentPages()
	sections := getSidebarSections("button")

	require.Len(t, sections, 6)
	require.Equal(t, "Example Apps", sections[5].Title)
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
}
