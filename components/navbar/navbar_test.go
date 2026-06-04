package navbar

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderNavbar(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Navbar(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestNavbar_RenderTargets(t *testing.T) {
	html := renderNavbar(t, Config{
		NavClass: "nav-extra",
		NavAttrs: templ.Attributes{"data-test": "nav"},
	})

	assert.Contains(t, html, "nav-extra")
	assert.Contains(t, html, `data-test="nav"`)
}
