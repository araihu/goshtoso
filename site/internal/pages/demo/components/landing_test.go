package components

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"

	siteassets "github.com/araihu/goshtoso/site/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLandingPageRendersRegistryComponentCount(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &buf))

	expected := strconv.Itoa(componentCount()) + " components"
	assert.Equal(t, 2, strings.Count(buf.String(), expected))
	assert.NotContains(t, buf.String(), "local assets")
}

func TestLandingPageUsesSiteOwnedBootstrapAndProviders(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &buf))

	html := buf.String()
	assert.Contains(t, html, "JavaScript build step optional")
	assert.NotContains(t, html, "No JavaScript build step")
	assert.Contains(t, html, `data-demo-storage-policy="strict"`)
	assert.Contains(t, html, `data-demo-theme-bootstrap`)
	assert.Contains(t, html, `<script src="`+siteassets.DemoBundleURL+`"></script>`)
	assert.Contains(t, html, `x-data="demoStorageConsent"`)
	assert.NotContains(t, html, "window.goshtosoStorageConsent={")
	assert.NotContains(t, html, "localStorage.getItem('darkMode')")
}

func TestLandingInstallCommandUsesCopyableCodeBlock(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &buf))

	html := buf.String()
	assert.Contains(t, html, `aria-label="Install Goshtoso"`)
	assert.Contains(t, html, `class="codeblock overflow-x-auto`)
	assert.Contains(t, html, `aria-label="Copy Install Goshtoso code"`)
	assert.Contains(t, html, `go get github.com/araihu/goshtoso@latest`)
}

func TestLandingPlaygroundIsIsolatedFromHomepageTheme(t *testing.T) {
	t.Parallel()

	var landing bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &landing))
	landingHTML := landing.String()
	assert.Contains(t, landingHTML, `id="theme-playground-frame"`)
	assert.Contains(t, landingHTML, `src="/playground/theme"`)
	assert.NotContains(t, landingHTML, `id="home-theme-picker"`)
	assert.NotContains(t, landingHTML, `Every control below is a real Goshtoso component`)
	assert.Contains(t, landingHTML, `Publish static sites`)

	var playground bytes.Buffer
	require.NoError(t, LandingPlaygroundPage().Render(context.Background(), &playground))
	html := playground.String()
	assert.Contains(t, html, `<html lang="en" data-theme="araihu" data-landing-playground class="overflow-hidden">`)
	assert.Contains(t, html, `<script src="`+siteassets.LandingPlaygroundBundleURL+`"></script>`)
	assert.NotContains(t, html, `<script src="`+siteassets.DemoBundleURL+`"></script>`)
	assert.Contains(t, html, `class="m-0 overflow-hidden`)
	assert.Contains(t, html, `id="home-theme-picker"`)
	assert.Contains(t, html, `document.documentElement.dataset.theme = theme`)
	assert.Contains(t, html, `Live theme preview`)
	assert.NotContains(t, html, `Customer workspace`)
	assert.NotContains(t, html, `Explore all`)
	assert.NotContains(t, html, `localStorage.theme`)
}

func TestLandingExtensionsPreloadIsolatedCharts(t *testing.T) {
	t.Parallel()

	var landing bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &landing))
	html := landing.String()
	for _, want := range []string{
		`id="extensions"`,
		`Goshtoso Charts`,
		`github.com/araihu/goshtoso-charts`,
		`href="https://charts.goshtoso.araihu.com"`,
		`id="charts-showcase-loader"`,
		`hx-get="/playground/extensions/charts/frame?variant=line-3d"`,
		`hx-trigger="load"`,
		`hx-swap="outerHTML"`,
	} {
		assert.Contains(t, html, want)
	}
	assert.NotContains(t, html, `Goshtoso App Shells`)
	assert.NotContains(t, html, `href="/modules/app-shells"`)
	assert.NotContains(t, html, `Visit Arai Hû`)
	assert.Equal(t, 3, strings.Count(html, `Work in progress`))
	assert.Contains(t, html, `agent-driven code changes`)
	assert.NotContains(t, html, `agent-piloted`)
	assert.NotContains(t, html, `id="charts-showcase-frame"`)
	assert.NotContains(t, html, `/charts/assets/js/runtime/echarts/`)

	var frame bytes.Buffer
	require.NoError(t, ChartsShowcaseFrame().Render(context.Background(), &frame))
	frameHTML := frame.String()
	assert.Contains(t, frameHTML, `id="charts-showcase-frame-line-3d"`)
	assert.Contains(t, frameHTML, `src="/playground/extensions/charts?variant=line-3d"`)
	assert.Contains(t, frameHTML, `loading="eager"`)
	assert.Contains(t, frameHTML, `scrolling="no"`)
}

func TestChartsShowcasePageUsesAutoRotatingLine3DWithoutOuterShell(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	require.NoError(t, ChartsShowcasePage().Render(context.Background(), &buffer))
	html := buffer.String()
	assert.Contains(t, html, `data-charts-showcase`)
	assert.Contains(t, html, `/charts/assets/js/runtime/echarts/`)
	assert.Contains(t, html, `/charts/assets/js/runtime/three-d/`)
	assert.Contains(t, html, `data-showcase-chart="line-3d"`)
	assert.Contains(t, html, `Charts for every use case`)
	assert.Equal(t, 1, strings.Count(html, `data-showcase-component="line-3d"`))
	assert.NotContains(t, html, `data-chart-carousel-slide`)
	assert.NotContains(t, html, `id="charts-use-cases-carousel"`)
	assert.NotContains(t, html, `data-showcase-chart="surface-3d"`)
	assert.NotContains(t, html, `data-showcase-chart="scatter-3d"`)
	assert.NotContains(t, html, `component-doc-shell__header`)
	assert.Contains(t, html, `data-chart-fallback`)
	assert.NotContains(t, html, `Typed Go configs, local runtime, theme-aware output.`)
	assert.NotContains(t, html, `<figcaption`)
	assert.Contains(t, html, `"legend":{"show":false}`)
	assert.Contains(t, html, `"autoRotate":true`)
}
