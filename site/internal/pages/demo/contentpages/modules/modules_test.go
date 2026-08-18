package modulespages

import (
	"bytes"
	"context"
	"testing"

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
		`This documentation site uses Component Docs Shell`,
		`componentdocshell`,
		`componentpage`,
		`consoleshell`,
		`href="https://github.com/araihu/goshtoso-app-shells"`,
		`href="/docs/application-patterns#app-shell"`,
	} {
		require.Contains(t, html, want)
	}
}
