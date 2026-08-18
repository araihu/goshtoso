package demo

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestComponentDocsSecondaryNavigationUsesSiteFamilies(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("Navbar"),
		"navbar",
		templ.NopComponent,
		false,
	).Render(context.Background(), &page))

	html := page.String()
	require.Contains(t, html, `id="goshtoso-site-secondary-navigation"`)
	require.Contains(t, html, `data-site-secondary-family="core"`)
	require.Contains(t, html, `window.addEventListener("componentdocshell:navigated"`)
	for _, link := range []string{
		`href="/getting-started"`,
		`href="/docs/iconpack"`,
		`href="/modules/charts"`,
		`href="/modules/app-shells"`,
		`href="/examples"`,
	} {
		require.Contains(t, html, link)
	}
	require.Contains(t, html, `aria-current="location"`)
	require.Contains(t, html, `>Core</a>`)
	require.Contains(t, html, `>Icon Packs</a>`)
	require.Contains(t, html, `>Charts</a>`)
	require.Contains(t, html, `>App Shells</a>`)
	require.Contains(t, html, `>Examples</a>`)
}

func TestComponentDocsSecondaryNavigationTracksFamily(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("Charts"),
		"module-charts",
		templ.NopComponent,
		false,
	).Render(context.Background(), &page))

	html := page.String()
	chartsStart := strings.Index(html, `href="/modules/charts"`)
	require.NotEqual(t, -1, chartsStart)
	chartsEnd := strings.Index(html[chartsStart:], `</a>`)
	require.NotEqual(t, -1, chartsEnd)
	require.Contains(t, html[chartsStart:chartsStart+chartsEnd], `aria-current="location"`)
}
