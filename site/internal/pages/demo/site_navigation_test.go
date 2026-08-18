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
	require.NotContains(t, html, `window.addEventListener("componentdocshell:navigated"`)
	for _, link := range []string{
		`href="/getting-started"`,
		`href="/docs/agents"`,
		`href="/components/icon"`,
		`href="/modules/charts"`,
		`href="/modules/app-shells"`,
		`href="/examples/ticker"`,
	} {
		require.Contains(t, html, link)
	}
	require.Contains(t, html, `aria-current="location"`)
	require.Contains(t, html, `>Core</a>`)
	require.Contains(t, html, `>AI Agents</a>`)
	require.Contains(t, html, `>Icons</a>`)
	require.Contains(t, html, `>Charts</a>`)
	require.Contains(t, html, `>App Shells</a>`)
	require.Contains(t, html, `>Examples</a>`)
}

func TestComponentDocsSecondaryNavigationTracksAgents(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("AI Agents"),
		"agents",
		templ.NopComponent,
		false,
	).Render(context.Background(), &page))

	html := page.String()
	agentsStart := strings.Index(html, `href="/docs/agents"`)
	require.NotEqual(t, -1, agentsStart)
	agentsEnd := strings.Index(html[agentsStart:], `</a>`)
	require.NotEqual(t, -1, agentsEnd)
	require.Contains(t, html[agentsStart:agentsStart+agentsEnd], `aria-current="location"`)
	for _, sidebarItem := range []string{"AI Agents", "Attributions", "License"} {
		require.NotContains(t, html, `data-sidebar-item="`+sidebarItem+`"`)
	}
}

func TestComponentDocsSecondaryNavigationFitsHeaderSlot(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("Navbar"),
		"navbar",
		templ.NopComponent,
		false,
	).Render(context.Background(), &page))

	html := page.String()
	require.Contains(t, html, `.component-doc-shell__site-header-actions [data-navbar-secondary] > nav > div`)
	require.Contains(t, html, `padding-block: 0;`)
	require.Contains(t, html, `overflow-y: hidden;`)
	require.Contains(t, html, `height: 100%;
			min-height: 0;
			padding-block: 0.5rem;`)
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

func TestComponentDocsSecondaryNavigationTracksAppShellsSubpage(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("Console Shell"),
		"app-shells-shell-console",
		templ.NopComponent,
		false,
	).Render(context.Background(), &page))

	html := page.String()
	appShellsStart := strings.Index(html, `href="/modules/app-shells"`)
	require.NotEqual(t, -1, appShellsStart)
	appShellsEnd := strings.Index(html[appShellsStart:], `</a>`)
	require.NotEqual(t, -1, appShellsEnd)
	require.Contains(t, html[appShellsStart:appShellsStart+appShellsEnd], `aria-current="location"`)
	for _, href := range []string{
		`href="/modules/app-shells/frames/component-page"`,
		`href="/modules/app-shells/shells/component-docs-shell"`,
		`href="/modules/app-shells/shells/console-shell"`,
		`href="/modules/app-shells/shells/landing-shell"`,
	} {
		require.Contains(t, html, href)
	}
}

func TestComponentDocsSecondaryNavigationTracksExamplesDefault(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("Live Ticker"),
		"ticker",
		templ.NopComponent,
		false,
	).Render(context.Background(), &page))

	html := page.String()
	examplesStart := strings.Index(html, `href="/examples/ticker"`)
	require.NotEqual(t, -1, examplesStart)
	examplesEnd := strings.Index(html[examplesStart:], `</a>`)
	require.NotEqual(t, -1, examplesEnd)
	require.Contains(t, html[examplesStart:examplesStart+examplesEnd], `aria-current="location"`)
	for _, href := range []string{
		`href="/examples/ticker"`,
		`href="/examples/todo"`,
		`href="/examples/expense"`,
		`href="/examples/chat"`,
		`href="/examples/logs"`,
		`href="/examples/profile"`,
		`href="/examples/wizard"`,
	} {
		require.Contains(t, html, href)
	}
}

func TestSiteFooterLinksGeneratedResources(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, siteFooter().Render(context.Background(), &page))

	html := page.String()
	require.Contains(t, html, `aria-label="Site links"`)
	require.Contains(t, html, `href="/sitemap.xml"`)
	require.Contains(t, html, `>Sitemap XML</a>`)
	require.Contains(t, html, `href="/llms.txt"`)
	require.Contains(t, html, `>llms.txt</a>`)
}
