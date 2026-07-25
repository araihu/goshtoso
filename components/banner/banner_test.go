package banner

import (
	"strings"
	"testing"
)

func TestCookieBannerDefaultsToViewportFixed(t *testing.T) {
	fixedHTML := renderStructuralBanner(t, CookieBanner(CookieBannerConfig{
		Description: "Cookies",
	}))
	if !strings.Contains(fixedHTML, "fixed bottom-4") {
		t.Fatalf("default cookie banner should stay viewport fixed; got %q", fixedHTML)
	}
}
