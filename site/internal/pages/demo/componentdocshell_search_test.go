package demo

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/stretchr/testify/require"
)

func TestComponentDocsHeaderActionsUsesFamilySearchConfig(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		family string
		search string
	}{
		{family: "core", search: "docs-search"},
		{family: "charts", search: "docs-search"},
		{family: "icon-packs", search: "docs-search"},
		{family: "app-shells", search: "docs-search"},
		{family: "examples", search: "docs-search"},
	} {
		t.Run(tc.family, func(t *testing.T) {
			var buf bytes.Buffer
			err := componentDocsHeaderActions(tc.family).Render(context.Background(), &buf)
			require.NoError(t, err)

			html := buf.String()
			require.Contains(t, html, `class="component-doc-shell__global-search"`)
			require.Contains(t, html, `id="`+tc.search+`"`)
			require.Contains(t, html, `data-search-field`)
			require.Contains(t, html, `data-navbar-secondary`)
		})
	}
}

func TestComponentDocsNavigationKeepsSearchOutOfSidebars(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		nav  componentdocshell.Navigation
	}{
		{name: "core", nav: componentDocsNavigation("")},
		{name: "charts", nav: chartsDocsNavigation("module-charts")},
		{name: "icons", nav: iconsDocsNavigation("icon")},
		{name: "app shells", nav: appShellsDocsNavigation("module-app-shells")},
		{name: "examples", nav: examplesDocsNavigation("ticker")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, tc.nav.DisableSearch)
			require.Empty(t, tc.nav.SearchPlaceholder)
			require.Nil(t, tc.nav.SearchSlot)
		})
	}
}

func TestLegacyLayoutMovesSearchIntoHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := LayoutWithMeta(DefaultMeta("Legacy page"), "", templ.NopComponent).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, `class="demo-global-search"`)
	require.Contains(t, html, `id="docs-search"`)
	require.Contains(t, html, `id="docs-search-dialog"`)
	require.NotContains(t, html, `class="docs-sidebar-search`)
}
