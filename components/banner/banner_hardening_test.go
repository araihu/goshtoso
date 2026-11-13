package banner

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderBanner(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Banner(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestBannerRootAttrsNormalizeNamesAndKeepOwnedAttributes(t *testing.T) {
	rendered := renderBanner(t, Config{
		RootClass: "caller-class",
		RootAttrs: templ.Attributes{
			"ROLE": "region", "CLASS": "ignored-class", "X-DATA": "ignored-data",
			"X-SHOW": "ignored-show", "x-show.important": "ignored-modifier", "X-CLOAK": false,
			"ID": "mixed-case", "id": "lowercase", "ARIA-LABEL": "Invitation",
		},
	})
	assert.Contains(t, rendered, ` role="region"`)
	assert.Contains(t, rendered, ` aria-label="Invitation"`)
	assert.Contains(t, rendered, ` id="lowercase"`)
	assert.Contains(t, rendered, "caller-class")
	assert.NotContains(t, rendered, "ignored-")
	assert.NotContains(t, rendered, "mixed-case")
	for _, name := range []string{" class=", " role=", " x-data=", " x-show=", " x-cloak"} {
		assert.Equal(t, 1, strings.Count(strings.ToLower(rendered[:strings.Index(rendered, ">")]), name), name)
	}
}

func TestBannerRootAttrsKeepDefaultRoleForEmptyOverrides(t *testing.T) {
	for _, role := range []any{nil, false, "  "} {
		rendered := renderBanner(t, Config{RootAttrs: templ.Attributes{"ROLE": role}})
		assert.Contains(t, rendered, ` role="banner"`)
	}
}

func TestBannerCTA_SanitizesUnsafeHref(t *testing.T) {
	rendered := renderBanner(t, Config{
		Description: "Notice",
		CTA:         &CTAConfig{ActionLabel: "Open", Href: "javascript:alert(1)"},
	})

	assert.NotContains(t, rendered, `href="javascript:alert`)
	assert.Contains(t, rendered, `href="about:invalid#TemplFailedSanitizationURL"`)
}

func TestBannerCTA_PreservesRelativeHref(t *testing.T) {
	rendered := renderBanner(t, Config{
		Description: "Notice",
		CTA:         &CTAConfig{ActionLabel: "Open", Href: "/docs?tab=api"},
	})

	assert.Contains(t, rendered, `href="/docs?tab=api"`)
}

func TestBannerCustomContentOwnsItsActions(t *testing.T) {
	rendered := renderBanner(t, Config{
		Description: "Default description",
		CTA:         &CTAConfig{ActionLabel: "Default CTA"},
		Persistent:  true,
		RootAttrs:   templ.Attributes{"role": "group", "aria-labelledby": "notice"},
		Content: templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
			_, err := io.WriteString(w, `<h2 id="notice">Pending request</h2><form method="post"><button>Accept</button></form>`)
			return err
		}),
	})
	assert.Contains(t, rendered, `role="group"`)
	assert.Contains(t, rendered, `aria-labelledby="notice"`)
	assert.Contains(t, rendered, `<form method="post"><button>Accept</button></form>`)
	assert.NotContains(t, rendered, `role="banner"`)
	assert.NotContains(t, rendered, "Default description")
	assert.NotContains(t, rendered, "Default CTA")
}
