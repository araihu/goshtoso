package demo

import (
	"testing"

	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/stretchr/testify/require"
)

func TestExamplesDocsNavigationContainsEveryExample(t *testing.T) {
	t.Parallel()

	navigation := examplesDocsNavigation("chat")

	require.Empty(t, navigation.Items)
	require.Len(t, navigation.Sections, 2)
	require.Equal(t, "Modules", navigation.Sections[0].Title)
	require.Equal(t, "Examples", navigation.Sections[1].Title)
	require.Equal(t, []string{"Charts", "App Shells"}, sidebarItemLabels(navigation.Sections[0].Items))
	require.Equal(t, []string{
		"Live Ticker",
		"Todo List",
		"Expense Tracker",
		"Chat",
		"Live Log Feed",
		"Profile",
		"Onboarding Wizard",
	}, sidebarItemLabels(navigation.Sections[1].Items))
	require.Equal(t, "/examples/chat", navigation.Sections[1].Items[3].Href)
	require.True(t, navigation.Sections[1].Items[3].Active)
	require.False(t, navigation.Sections[1].Items[0].Active)
	require.True(t, navigation.DisableSearch)

	for _, section := range navigation.Sections {
		for _, item := range section.Items {
			require.Equal(t, item.Href, item.LinkAttrs["hx-get"])
			require.Equal(t, "#main-content", item.LinkAttrs["hx-target"])
		}
	}
}

func sidebarItemLabels(items []sidebar.Item) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}
