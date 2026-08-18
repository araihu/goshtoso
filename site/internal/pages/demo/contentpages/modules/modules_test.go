package modulespages

import (
	"bytes"
	"context"
	"html"
	"regexp"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestChartsModuleRendersMigratedChartsGettingStartedPage(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	require.NoError(t, chartsModuleContent().Render(context.Background(), &buffer))
	html := buffer.String()
	require.Contains(t, html, `Goshtoso Charts`)
	require.Contains(t, html, `Add your first chart`)
	require.Contains(t, html, `href="/modules/charts/docs/chart-modes"`)
	for _, kind := range []string{"static", "interactive", "line-3d"} {
		require.Contains(t, html, `data-chart-module-showcase="`+kind+`"`)
		require.Contains(t, html, `data-charts-showcase-frame="`+kind+`"`)
	}
	require.NotContains(t, html, `charts.goshtoso.araihu.com`)
}

func TestChartsShowcaseVariantsUseRealModuleComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind chartsShowcaseKind
		want string
	}{
		{chartsShowcaseStatic, `data-showcase-chart="static"`},
		{chartsShowcaseInteractive, `data-showcase-chart="interactive"`},
		{chartsShowcaseLine3D, `data-showcase-chart="line-3d"`},
	}
	for _, test := range tests {
		var buffer bytes.Buffer
		require.NoError(t, ChartsShowcasePageFor(test.kind).Render(context.Background(), &buffer))
		require.Contains(t, buffer.String(), test.want)
	}
}

func TestAppShellsModuleReferencesCurrentPackagesAndCompositionRecipe(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	require.NoError(t, appShellsModuleContent().Render(context.Background(), &buffer))
	html := buffer.String()
	for _, want := range []string{
		`Goshtoso App Shells`,
		`v0.1.6`,
		`Application frames and shells`,
		`Build the frame once`,
		`Frames`,
		`Choose a frame when you need a consistent page shape`,
		`Shells`,
		`Choose a shell when the product owns the surrounding regions`,
		`componentdocshell`,
		`componentpage`,
		`consoleshell`,
		`landingshell`,
		`Compose a custom shell when you need a different boundary`,
		`href="https://github.com/araihu/goshtoso-app-shells"`,
		`href="/docs/application-patterns#app-shell"`,
	} {
		require.Contains(t, html, want)
	}
}

func TestAppShellsPackagePagesRenderCurrentAPISurface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		page func() templ.Component
		want []string
	}{
		{name: "frame", page: appShellsComponentPageContent, want: []string{"Component Page", "componentpage", "data-component-page"}},
		{name: "component docs shell", page: appShellsComponentDocsShellContent, want: []string{"Component Docs Shell", "componentdocshell", "Live proof"}},
		{name: "console shell", page: appShellsConsoleShellContent, want: []string{"Console Shell", "consoleshell", "NavigationOOB"}},
		{name: "landing shell", page: appShellsLandingShellContent, want: []string{"Landing Shell", "landingshell", "Hero slot"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buffer bytes.Buffer
			require.NoError(t, test.page().Render(context.Background(), &buffer))
			for _, want := range test.want {
				require.Contains(t, buffer.String(), want)
			}
		})
	}
}

func TestAppShellsPackagePagesDocumentRequiredSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		page func() templ.Component
		want []string
	}{
		{
			name: "component page",
			page: appShellsComponentPageContent,
			want: []string{
				`go get github.com/araihu/goshtoso-app-shells/componentpage@v0.1.6`,
				`document shell owns runtime and asset mounting`,
			},
		},
		{
			name: "component docs shell",
			page: appShellsComponentDocsShellContent,
			want: []string{
				`go get github.com/araihu/goshtoso-app-shells/componentdocshell@v0.1.6`,
				`github.com/araihu/goshtoso-app-shells/componentdocshell/assets`,
				`mux.Handle("GET /componentdocshell/assets/", shellassets.Handler())`,
			},
		},
		{
			name: "console shell",
			page: appShellsConsoleShellContent,
			want: []string{
				`go get github.com/araihu/goshtoso-app-shells/consoleshell@v0.1.6`,
				`github.com/araihu/goshtoso-app-shells/consoleshell/assets`,
				`mux.Handle("GET /consoleshell/assets/", shellassets.Handler())`,
			},
		},
		{
			name: "landing shell",
			page: appShellsLandingShellContent,
			want: []string{
				`go get github.com/araihu/goshtoso-app-shells/landingshell@v0.1.6`,
				`github.com/araihu/goshtoso-app-shells/landingshell/assets`,
				`mux.Handle("GET /landingshell/assets/", shellassets.Handler())`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buffer bytes.Buffer
			require.NoError(t, test.page().Render(context.Background(), &buffer))
			text := renderedPlainText(buffer.String())
			for _, want := range test.want {
				require.Contains(t, text, want)
			}
		})
	}
}

func TestAppShellHTMXPagesDocumentFragmentResponseContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		page func() templ.Component
		want []string
	}{
		{
			name: "component docs shell",
			page: appShellsComponentDocsShellContent,
			want: []string{
				`Full documents and HTMX fragments`,
				`request.Header.Get("HX-Request") == "true"`,
				`componentdocshell.Fragment(cfg, page)`,
				`main content, scoped sidebar, and family navigation`,
			},
		},
		{
			name: "console shell",
			page: appShellsConsoleShellContent,
			want: []string{
				`Full documents and HTMX fragments`,
				`request.Header.Get("HX-Request") == "true"`,
				`consoleshell.Fragment(cfg, page)`,
				`NavigationOOB`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buffer bytes.Buffer
			require.NoError(t, test.page().Render(context.Background(), &buffer))
			text := renderedPlainText(buffer.String())
			for _, want := range test.want {
				require.Contains(t, text, want)
			}
		})
	}
}

func TestComponentDocsShellPageKeepsGlobalSearchConsumerOwned(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	require.NoError(t, appShellsComponentDocsShellContent().Render(context.Background(), &buffer))
	html := buffer.String()
	require.Contains(t, html, `Consumer-owned global search`)
	require.NotContains(t, html, `Brand, global search, appearance controls, and family navigation.`)
}

func TestAppShellsOverviewExplainsFragmentAvailability(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	require.NoError(t, appShellsModuleContent().Render(context.Background(), &buffer))
	html := buffer.String()
	require.Contains(t, html, `Fragment support`)
	require.Contains(t, html, `Only the documentation and console shells expose HTMX fragment responses in v0.1.6.`)
	require.Contains(t, html, `Component Page is embedded inside a document; Landing Shell renders complete public pages.`)
}

var renderedTagPattern = regexp.MustCompile(`<[^>]+>`)

func renderedPlainText(markup string) string {
	return html.UnescapeString(renderedTagPattern.ReplaceAllString(markup, ""))
}
