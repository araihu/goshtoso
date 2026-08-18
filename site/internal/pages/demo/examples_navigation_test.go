package demo

import (
	"testing"

	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/stretchr/testify/require"
)

func TestExamplesDocsNavigationContainsEveryExample(t *testing.T) {
	t.Parallel()

	navigation := examplesDocsNavigation("chat")

	require.Len(t, navigation.Items, 7)
	require.Equal(t, []string{
		"Live Ticker",
		"Todo List",
		"Expense Tracker",
		"Chat",
		"Live Log Feed",
		"Profile",
		"Onboarding Wizard",
	}, sidebarItemLabels(navigation.Items))
	require.Equal(t, "/examples/chat", navigation.Items[3].Href)
	require.True(t, navigation.Items[3].Active)
	require.False(t, navigation.Items[0].Active)
	require.True(t, navigation.DisableSearch)

	for _, item := range navigation.Items {
		require.Equal(t, item.Href, item.LinkAttrs["hx-get"])
		require.Equal(t, "#main-content", item.LinkAttrs["hx-target"])
	}
}

func sidebarItemLabels(items []sidebar.Item) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}
