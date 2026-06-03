package banner

import (
	"context"
	"strings"
	"testing"
)

func renderBanner(t *testing.T, cfg Config) string {
	t.Helper()

	var html strings.Builder
	if err := Banner(cfg).Render(context.Background(), &html); err != nil {
		t.Fatalf("render banner: %v", err)
	}
	return html.String()
}

func TestCookieBannerPositionClasses(t *testing.T) {
	fixedHTML := renderBanner(t, Config{
		CookieBanner: true,
		Text:         "Cookies",
	})
	if !strings.Contains(fixedHTML, "fixed bottom-4") {
		t.Fatalf("default cookie banner should stay viewport fixed; got %q", fixedHTML)
	}

	relativeHTML := renderBanner(t, Config{
		CookieBanner: true,
		Position:     PositionRelative,
		Text:         "Cookies",
	})
	if strings.Contains(relativeHTML, "fixed bottom-4") {
		t.Fatalf("relative cookie banner should not stay viewport fixed; got %q", relativeHTML)
	}
	if !strings.Contains(relativeHTML, "absolute bottom-4") {
		t.Fatalf("relative cookie banner should be absolutely positioned inside its container; got %q", relativeHTML)
	}
}
