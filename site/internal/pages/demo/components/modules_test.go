package components

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChartsModuleRendersThreeHTMXLazyShowcases(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	require.NoError(t, chartsModuleContent().Render(context.Background(), &buffer))
	html := buffer.String()
	for _, kind := range []string{"static", "interactive", "line-3d"} {
		require.Contains(t, html, `data-chart-module-showcase="`+kind+`"`)
		require.Contains(t, html, `hx-get="/playground/extensions/charts/frame?variant=`+kind+`"`)
	}
	require.Equal(t, 3, bytes.Count(buffer.Bytes(), []byte(`hx-trigger="intersect once"`)))
	require.Contains(t, html, `https://charts.goshtoso.araihu.com`)
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
