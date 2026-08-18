package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChartsPageRoutesHaveUniquePathsAndActiveItems(t *testing.T) {
	t.Parallel()

	routes := chartsPageRoutes()
	paths := make(map[string]struct{}, len(routes))
	activeItems := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		require.NotEmpty(t, route.Path)
		require.NotEmpty(t, route.Active)
		require.NotEmpty(t, route.Title)
		require.NotNil(t, route.Render)
		require.Truef(t, len(route.Path) >= len("/modules/charts") && route.Path[:len("/modules/charts")] == "/modules/charts", "route %q is outside the Charts namespace", route.Path)
		_, pathExists := paths[route.Path]
		require.False(t, pathExists, "duplicate Charts route %q", route.Path)
		paths[route.Path] = struct{}{}
		_, activeExists := activeItems[route.Active]
		require.False(t, activeExists, "duplicate Charts sidebar active item %q", route.Active)
		activeItems[route.Active] = struct{}{}
	}
	require.GreaterOrEqual(t, len(routes), 40)
}
