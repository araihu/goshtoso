package banner

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestCookieBannerOwnsDialogContract(t *testing.T) {
	html := renderStructuralBanner(t, CookieBanner(CookieBannerConfig{
		Description: "We use local storage.",
	}))

	require.Contains(t, html, `role="dialog"`)
	require.Contains(t, html, "Cookie Consent")
}

func renderStructuralBanner(t *testing.T, component templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buf))
	return buf.String()
}
