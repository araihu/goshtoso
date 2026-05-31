package selectfield

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderSelect(t *testing.T, cfg Config, children templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	ctx := context.Background()
	if children != nil {
		ctx = templ.WithChildren(ctx, children)
	}
	require.NoError(t, Select(cfg).Render(ctx, &buf))
	return buf.String()
}

func TestSelect_ShellMode_RendersValueExprAndChildren(t *testing.T) {
	body := templ.Raw(`<div data-test-body>palette here</div>`)
	html := renderSelect(t, Config{
		ID:        "color-surface",
		Shell:     true,
		ValueExpr: "classLabel('surface')",
	}, body)

	assert.Contains(t, html, `x-data="{ isOpen: false, openedWithKeyboard: false }"`)
	assert.NotContains(t, html, "allOptions")
	// templ HTML-escapes single quotes in dynamic { } attribute values; the
	// browser decodes them before Alpine reads the attribute, so this works at
	// runtime — the same path the original colorRow used in production.
	assert.Contains(t, html, `x-text="classLabel(&#39;surface&#39;)"`)
	assert.Contains(t, html, "select-close")
	assert.Contains(t, html, "data-test-body")
	assert.Equal(t, 0, strings.Count(html, `role="option"`))
	assert.Equal(t, 0, strings.Count(html, `type="text"`))
}

func TestSelect_DataMode_Unchanged(t *testing.T) {
	html := renderSelect(t, Config{
		ID:      "fruit",
		Options: []Option{{Value: "a", Label: "Apple"}},
	}, nil)
	assert.Contains(t, html, "allOptions")
	assert.Contains(t, html, `role="listbox"`)
}
