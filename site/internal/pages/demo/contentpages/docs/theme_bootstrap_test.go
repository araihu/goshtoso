package docspages

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestThemePageRendersInertBootstrapData(t *testing.T) {
	var output bytes.Buffer
	require.NoError(t, themeDemoContent().Render(context.Background(), &output))
	html := output.String()

	require.Contains(t, html, `x-data="themePage"`)
	require.NotContains(t, html, "Alpine.data(\"themePage\"")
	require.NotContains(t, html, "Alpine.data('themePage'")

	const open = `<script id="theme-page-data" type="application/json">`
	start := strings.Index(html, open)
	require.NotEqual(t, -1, start, "theme page data script missing")
	start += len(open)
	end := strings.Index(html[start:], "</script>")
	require.NotEqual(t, -1, end, "theme page data script is not closed")
	require.Equal(t, 1, strings.Count(html, `id="theme-page-data"`))

	var data themePageBootstrap
	require.NoError(t, json.Unmarshal([]byte(html[start:start+end]), &data))
	require.Len(t, data.AllThemes, 16)
	require.Equal(t, "araihu", data.AllThemes[0])
	require.Equal(t, "1rem", data.RadiusMap["2xl"])
	require.Equal(t, "Inter", data.GoogleFontMap["Inter"])
	require.Contains(t, data.Blocks["goshtoso"], "[data-theme=goshtoso]")
	require.Contains(t, data.Blocks["araihu"], "[data-theme=araihu]")
	require.NotEmpty(t, data.ThemeClassMap["minimal"]["primary"])
	require.Greater(t, len(data.AllTokens), 10)
	require.Equal(t, "Primary", data.TokenLabels["primary"])
	require.Equal(t, "dracula", data.CSSOrder[len(data.CSSOrder)-1])
}

func TestThemePageBootstrapJSONCannotBreakOutOfDataScript(t *testing.T) {
	const sentinel = `</script><script>window.injected = true</script>`
	data := themePageBootstrap{Blocks: map[string]string{"hostile": sentinel}}

	var output bytes.Buffer
	require.NoError(t, templ.JSONScript("theme-page-data", data).Render(context.Background(), &output))
	html := output.String()
	require.Equal(t, 1, strings.Count(html, "</script>"), "JSON data must not close its own script element")
	require.NotContains(t, html, `<script>window.injected = true</script>`)

	const open = `<script id="theme-page-data" type="application/json">`
	start := strings.Index(html, open)
	require.NotEqual(t, -1, start)
	start += len(open)
	end := strings.Index(html[start:], "</script>")
	require.NotEqual(t, -1, end)
	var decoded themePageBootstrap
	require.NoError(t, json.Unmarshal([]byte(html[start:start+end]), &decoded))
	require.Equal(t, sentinel, decoded.Blocks["hostile"])
}
