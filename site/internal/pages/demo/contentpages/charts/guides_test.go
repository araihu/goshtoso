package charts

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChartGuideNavigationLivesInScopedSidebar(t *testing.T) {
	t.Parallel()

	var modes bytes.Buffer
	require.NoError(t, ChartModesPage(false).Render(context.Background(), &modes))
	require.NotContains(t, modes.String(), `data-chart-guide-nav`)
	require.NotContains(t, modes.String(), `Chart guide index`)

	var controls bytes.Buffer
	require.NoError(t, ChartControlsPage(false, ChartControlExamples{}).Render(context.Background(), &controls))
	require.NotContains(t, controls.String(), `data-chart-guide-nav`)
	require.NotContains(t, controls.String(), `Chart guide index`)
}

func TestThemePlaygroundFrameIncludesColorModeControl(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	require.NoError(t, ThemePlaygroundFrame().Render(context.Background(), &buffer))
	html := buffer.String()
	require.Contains(t, html, `id="theme-playground-color-mode"`)
	require.Contains(t, html, `Color mode`)
	require.Contains(t, html, `x-on:click="toggleDark()"`)
	require.Contains(t, html, `x-bind:aria-label="dark ? 'Switch to light mode' : 'Switch to dark mode'"`)
	require.Contains(t, html, `x-bind:aria-pressed="dark"`)
}
