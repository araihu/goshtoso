package docspages

import (
	"bytes"
	"context"
	"github.com/araihu/goshtoso/site/internal/themes"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestThemePageConsumesPublicTools(t *testing.T) {
	var output bytes.Buffer
	require.NoError(t, themeDemoContent().Render(context.Background(), &output))
	html := output.String()
	require.Contains(t, html, "data-theme-editor=")
	require.Contains(t, html, "data-theme-playground")
	require.Contains(t, html, "/playground/theme")
	require.NotContains(t, html, `x-data="themePage"`)
	require.NotContains(t, html, "theme-page-data")
	for _, theme := range themes.All() {
		require.True(t, strings.Contains(html, theme.Label) || strings.Contains(html, strings.ReplaceAll(theme.Label, "û", "&#251;")), theme.Key)
	}
}
