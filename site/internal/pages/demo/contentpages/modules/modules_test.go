package modulespages

import (
	"bytes"
	"context"
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
