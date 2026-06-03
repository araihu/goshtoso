package breadcrumbs

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderBreadcrumbs(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Breadcrumbs(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestBreadcrumbs_SanitizesUnsafeHref(t *testing.T) {
	rendered := renderBreadcrumbs(t, Config{
		Items: []Item{{Label: "Bad", Href: "javascript:alert(1)"}},
	})

	assert.NotContains(t, rendered, `href="javascript:alert`)
	assert.Contains(t, rendered, `href="about:invalid#TemplFailedSanitizationURL"`)
}

func TestBreadcrumbs_PreservesRelativeHref(t *testing.T) {
	rendered := renderBreadcrumbs(t, Config{
		Items: []Item{{Label: "Docs", Href: "/docs?tab=api"}},
	})

	assert.Contains(t, rendered, `href="/docs?tab=api"`)
}
