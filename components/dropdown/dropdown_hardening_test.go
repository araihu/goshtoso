package dropdown

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderDropdown(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Dropdown(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestDropdownLink_SanitizesUnsafeHref(t *testing.T) {
	rendered := renderDropdown(t, Config{
		ID: "menu",
		Sections: []Section{{Items: []Item{
			{Label: "Bad", Href: "javascript:alert(1)"},
		}}},
	})

	assert.NotContains(t, rendered, `href="javascript:alert`)
	assert.Contains(t, rendered, `href="about:invalid#TemplFailedSanitizationURL"`)
}

func TestDropdownLink_PreservesRelativeHref(t *testing.T) {
	rendered := renderDropdown(t, Config{
		ID: "menu",
		Sections: []Section{{Items: []Item{
			{Label: "Docs", Href: "/docs?tab=api"},
		}}},
	})

	assert.Contains(t, rendered, `href="/docs?tab=api"`)
}
