package banner

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderBanner(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Banner(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestBannerCTA_SanitizesUnsafeHref(t *testing.T) {
	rendered := renderBanner(t, Config{
		Text: "Notice",
		CTA:  &CTAConfig{Text: "Open", Href: "javascript:alert(1)"},
	})

	assert.NotContains(t, rendered, `href="javascript:alert`)
	assert.Contains(t, rendered, `href="about:invalid#TemplFailedSanitizationURL"`)
}

func TestBannerCTA_PreservesRelativeHref(t *testing.T) {
	rendered := renderBanner(t, Config{
		Text: "Notice",
		CTA:  &CTAConfig{Text: "Open", Href: "/docs?tab=api"},
	})

	assert.Contains(t, rendered, `href="/docs?tab=api"`)
}
