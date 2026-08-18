package demo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppShellsDocsNavigationPutsFramesBeforeShells(t *testing.T) {
	t.Parallel()

	navigation := appShellsDocsNavigation("app-shells-shell-console")

	require.Len(t, navigation.Items, 1)
	require.Equal(t, "module-app-shells", navigation.Items[0].ID)
	require.Equal(t, "/modules/app-shells", navigation.Items[0].Href)
	require.True(t, navigation.DisableSearch)
	require.NotNil(t, navigation.Scope)
	require.Equal(t, "v0.1.6", navigation.Scope.Version)
	require.Len(t, navigation.Sections, 2)
	require.Equal(t, "Frames", navigation.Sections[0].Title)
	require.Equal(t, "Component Page", navigation.Sections[0].Items[0].Label)
	require.Equal(t, "Shells", navigation.Sections[1].Title)
	require.Equal(t, "Console Shell", navigation.Sections[1].Items[1].Label)
	require.True(t, navigation.Sections[1].Items[1].Active)
}

func TestAppShellsDocsNavigationKeepsLinksWithinModuleNamespace(t *testing.T) {
	t.Parallel()

	navigation := appShellsDocsNavigation("")
	for _, item := range navigation.Items {
		require.Contains(t, item.Href, "/modules/app-shells")
	}
	for _, section := range navigation.Sections {
		for _, item := range section.Items {
			require.Contains(t, item.Href, "/modules/app-shells/")
		}
	}
}

func TestAppShellsSearchItemsMirrorSidebarRoutes(t *testing.T) {
	t.Parallel()

	items := appShellsSearchItems()
	require.Len(t, items, 5)
	require.Equal(t, "Overview", items[0].Title)
	require.Equal(t, "/modules/app-shells", items[0].Href)
	require.Equal(t, "Component Page", items[1].Title)
	require.Equal(t, "/modules/app-shells/frames/component-page", items[1].Href)
	require.Equal(t, "Landing Shell", items[4].Title)
	require.Equal(t, "/modules/app-shells/shells/landing-shell", items[4].Href)
}
