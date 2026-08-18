package demo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChartsDocsNavigationBuildsChartsSidebar(t *testing.T) {
	t.Parallel()

	navigation := chartsDocsNavigation("charts-interactive-line")

	require.Len(t, navigation.Items, 5)
	require.Equal(t, "module-charts", navigation.Items[0].ID)
	require.Equal(t, "/modules/charts", navigation.Items[0].Href)
	require.True(t, navigation.DisableSearch)
	require.Empty(t, navigation.SearchPlaceholder)
	require.Nil(t, navigation.SearchSlot)
	require.Len(t, navigation.Sections, 7)
	require.Equal(t, "Static / Vector", navigation.Sections[0].Title)
	require.Equal(t, "Interactive / Relationships", navigation.Sections[5].Title)
	require.Equal(t, "Examples", navigation.Sections[6].Title)
	require.Equal(t, "charts-interactive-line", navigation.Sections[1].Items[1].ID)
	require.True(t, navigation.Sections[1].Items[1].Active)
}

func TestChartsDocsNavigationKeepsChartLinksWithinChartsNamespace(t *testing.T) {
	t.Parallel()

	navigation := chartsDocsNavigation("")
	for _, item := range navigation.Items {
		require.Contains(t, item.Href, "/modules/charts")
	}
	for _, section := range navigation.Sections {
		for _, item := range section.Items {
			require.Contains(t, item.Href, "/modules/charts")
		}
	}
}
