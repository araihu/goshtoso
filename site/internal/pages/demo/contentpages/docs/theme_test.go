package docspages

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	demothemes "github.com/araihu/goshtoso/site/internal/themes"
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

func TestThemeGridRendersReferenceStylePreviewCards(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, themeGridSection().Render(context.Background(), &buf))
	html := buf.String()

	assert.Equal(t, len(getThemeInfos()), strings.Count(html, `data-theme-key=`))
	assert.Equal(t, len(getThemeInfos()), strings.Count(html, `data-theme-ownership=`))
	assert.Equal(t, len(getThemeInfos()), strings.Count(html, `x-bind:aria-pressed=`))
	assert.Equal(t, len(getThemeInfos()), strings.Count(html, `data-theme-selected-icon`))
	assert.Contains(t, html, `h-24 border-t`)
	assert.Contains(t, html, `group-hover:gap-1.5`)
	assert.Contains(t, html, `bg-[#f8f8f2] dark:bg-[#282a36]`)
}

func TestThemeCSSExportBlocksMatchThemeSource(t *testing.T) {
	source, err := os.ReadFile("../../../../../../all-themes.css")
	require.NoError(t, err)

	blocks := getThemeCSSBlocks()
	require.Len(t, blocks, demothemes.Count())
	for _, theme := range demothemes.All() {
		exported, ok := blocks[theme.Key]
		require.True(t, ok, "catalog theme %q missing from exported CSS blocks", theme.Key)
		key := theme.Key
		t.Run(key, func(t *testing.T) {
			want := themeDeclarations(t, string(source), key)
			got := cssDeclarations(exported)
			assert.Equal(t, want, got, "exported theme CSS drifted from all-themes.css")
		})
	}
}

func TestThemeInfosCarryCatalogOwnership(t *testing.T) {
	catalog := demothemes.All()
	infos := getThemeInfos()
	require.Len(t, infos, len(catalog))
	for index := range catalog {
		require.Equal(t, catalog[index].Key, infos[index].Key)
		require.Equal(t, catalog[index].Label, infos[index].Label)
		require.Equal(t, string(catalog[index].Ownership), infos[index].Ownership)
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
