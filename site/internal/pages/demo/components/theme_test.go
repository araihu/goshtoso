package components

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThemeCSSExportRendersSingleDynamicOutput(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, themeDemoContent().Render(context.Background(), &buf))
	html := buf.String()

	assert.Contains(t, html, `id="theme-css-output"`)
	assert.NotContains(t, html, `theme-css-single-`)
	assert.NotContains(t, html, `theme-css-multi-`)
	assert.Equal(t, 0, strings.Count(html, `<span class="ch-`))
}

func TestThemeCSSExportBlocksMatchThemeSource(t *testing.T) {
	source, err := os.ReadFile("../../../../../all-themes.css")
	require.NoError(t, err)

	for key, exported := range getThemeCSSBlocks() {
		t.Run(key, func(t *testing.T) {
			want := themeDeclarations(t, string(source), key)
			got := cssDeclarations(exported)
			assert.Equal(t, want, got, "exported theme CSS drifted from all-themes.css")
		})
	}
}

func themeDeclarations(t *testing.T, source, key string) map[string]string {
	t.Helper()
	needle := "[data-theme=" + key + "] {"
	start := strings.Index(source, needle)
	require.NotEqual(t, -1, start, "theme %q missing from all-themes.css", key)
	rest := source[start:]
	end := strings.Index(rest, "}")
	require.NotEqual(t, -1, end, "theme %q block is not closed", key)
	return cssDeclarations(rest[:end+1])
}

func cssDeclarations(block string) map[string]string {
	declarations := make(map[string]string)
	for line := range strings.SplitSeq(block, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "--") {
			continue
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if comment := strings.Index(value, "/*"); comment >= 0 {
			value = value[:comment]
		}
		value = strings.TrimSpace(value)
		declarations[name] = strings.TrimSuffix(value, ";")
	}
	return declarations
}
