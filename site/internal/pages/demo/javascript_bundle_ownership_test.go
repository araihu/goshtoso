package demo

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	siteassets "github.com/araihu/goshtoso/site/assets"
	"github.com/stretchr/testify/require"
)

func TestComponentDocsLayoutOwnsDemoBundleOutsideHeadDependencies(t *testing.T) {
	t.Parallel()

	cfg := componentDocsConfig(false)
	require.Equal(t, []string{
		siteassets.DemoBundleURL,
		assets.HTMXExtWSURL,
		assets.HTMXExtSSEURL,
	}, cfg.Interactions.RuntimeScripts)

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("Bundle ownership"),
		"",
		templ.Raw("<p>content</p>"),
		false,
	).Render(context.Background(), &page))
	require.Equal(t, 1, strings.Count(page.String(), siteassets.DemoBundleURL))
	require.Contains(t, page.String(), `<script src="`+siteassets.DemoBundleURL+`"></script>`)
	require.Contains(t, page.String(), `defer src="`+assets.FirstPartyBundleURL+`"`)
	require.Contains(t, page.String(), `defer src="`+assets.AlpineJSURL+`"`)

	var publicHead strings.Builder
	require.NoError(t, head.Dependencies(head.WithLocalRuntime()).Render(context.Background(), &publicHead))
	require.NotContains(t, publicHead.String(), siteassets.DemoBundleURL,
		"public head dependencies must never load demo-site JavaScript")
}

func TestDemoRuntimeProvidersStayExternalAndLifecycleAware(t *testing.T) {
	t.Parallel()

	var layout strings.Builder
	require.NoError(t, LayoutWithMeta(DefaultMeta("External runtime"), "", templ.NopComponent).Render(context.Background(), &layout))
	require.Contains(t, layout.String(), `x-data="demoLayout"`)
	require.Contains(t, layout.String(), `x-data="demoStorageConsent"`)
	require.NotContains(t, layout.String(), "Alpine.store('nav'", "navigation runtime must stay in site bundle source")
	require.NotContains(t, layout.String(), "window.buildTOC = function", "TOC runtime must stay in site bundle source")

	var tabs strings.Builder
	require.NoError(t, TabView(TabViewProps{}, templ.NopComponent, "", "").Render(context.Background(), &tabs))
	require.Contains(t, tabs.String(), `x-data="demoTabView"`)
	require.NotContains(t, tabs.String(), "navigator.clipboard.writeText", "tab behavior must stay in site bundle source")

	for _, source := range []struct {
		path  string
		wants []string
	}{
		{path: "../../../assets/js/src/site-bootstrap.js", wants: []string{"goshtosoStorageConsent", "data-demo-theme-bootstrap"}},
		{path: "../../../assets/js/src/demo-layout.js", wants: []string{"Alpine.data(\"demoLayout\"", "Alpine.data(\"demoStorageConsent\"", "removeEventListener", "disconnect"}},
		{path: "../../../assets/js/src/tab-view.js", wants: []string{"Alpine.data(\"demoTabView\"", "destroy"}},
		{path: "../../../assets/js/src/select-demo.js", wants: []string{"goshtosoRestoreSelectDraft"}},
	} {
		body, err := os.ReadFile(source.path)
		require.NoError(t, err)
		for _, want := range source.wants {
			require.Contains(t, string(body), want, source.path)
		}
	}
}
