package pagination

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderPaginationHTMX(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Pagination(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestPagination_HTMXConfigRenders(t *testing.T) {
	html := renderPaginationHTMX(t, Config{
		CurrentPage: 1,
		TotalPages:  2,
		BaseURL:     "/items",
		HTMX:        &HTMXConfig{Target: "#items", Swap: "innerHTML"},
	})

	assert.Contains(t, html, `hx-target="#items"`)
	assert.Contains(t, html, `hx-swap="innerHTML"`)
}

func TestPagination_CurrentPageKeepsHTMXNavigation(t *testing.T) {
	html := renderPaginationHTMX(t, Config{
		CurrentPage: 1,
		TotalPages:  2,
		BaseURL:     "/items",
		HTMX:        &HTMXConfig{Target: "#items", Swap: "innerHTML"},
	})

	currentAt := strings.Index(html, `aria-current="page"`)
	require.NotEqual(t, -1, currentAt)
	anchorStart := strings.LastIndex(html[:currentAt], "<a")
	anchorEnd := strings.Index(html[currentAt:], ">")
	require.NotEqual(t, -1, anchorStart)
	require.NotEqual(t, -1, anchorEnd)
	currentAnchor := html[anchorStart : currentAt+anchorEnd]

	assert.Contains(t, currentAnchor, `hx-get="/items?page=1"`)
	assert.Contains(t, currentAnchor, `hx-target="#items"`)
	assert.Contains(t, currentAnchor, `hx-swap="innerHTML"`)
}
