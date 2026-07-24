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

func TestPageURL_PreservesQueryAndFragment(t *testing.T) {
	got := Config{BaseURL: "/items?filter=active&page=1#results"}.PageURL(3)

	assert.Equal(t, "/items?filter=active&page=3#results", got)
}

func TestPageURL_EncodesPageParamWithoutStringConcatenation(t *testing.T) {
	got := Config{BaseURL: "/items?search=two words&tag=a/b"}.PageURL(12)

	assert.Equal(t, "/items?page=12&search=two+words&tag=a%2Fb", got)
}

func TestPagination_RenderedHrefSanitizesUnsafeBaseURL(t *testing.T) {
	rendered := renderPagination(t, Config{
		Mode:        ModeSimple,
		CurrentPage: 1,
		TotalPages:  2,
		BaseURL:     "javascript:alert(1)",
	})

	assert.NotContains(t, rendered, `href="javascript:alert`)
	assert.Contains(t, rendered, `href="about:invalid#TemplFailedSanitizationURL"`)
}
