package components

import (
	"bytes"
	"context"
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
