package pagination

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderPagination(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Pagination(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestPagination_HTMXConfigRenders(t *testing.T) {
	html := renderPagination(t, Config{
		CurrentPage: 1,
		TotalPages:  2,
		BaseURL:     "/items",
		HTMX:        &HTMXConfig{Target: "#items", Swap: "innerHTML"},
	})

	assert.Contains(t, html, `hx-target="#items"`)
	assert.Contains(t, html, `hx-swap="innerHTML"`)
}
