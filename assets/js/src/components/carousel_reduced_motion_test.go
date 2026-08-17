package componentruntime

import (
	"regexp"
	"testing"
)

func TestCarouselRuntimeOwnsReducedMotionLifecycle(t *testing.T) {
	t.Parallel()

	source := readRuntimeSource(t, "carousel.js")
	contracts := []struct {
		name    string
		pattern string
	}{
		{name: "media query", pattern: `matchMedia\(["']\(prefers-reduced-motion: reduce\)["']\)`},
		{name: "change subscription", pattern: `addEventListener\(["']change["']`},
		{name: "change cleanup", pattern: `removeEventListener\(["']change["']`},
		{name: "reduced state", pattern: `reducedMotion`},
		{name: "autoplay suppression", pattern: `if\s*\([^\n)]*reducedMotion[^\n)]*\)\s*return`},
		{name: "timer cancellation", pattern: `clearInterval\([^\n)]*autoplayInterval`},
		{name: "normal-mode restart", pattern: `reducedMotion[\s\S]{0,800}\.autoplay\(\)`},
	}
	for _, contract := range contracts {
		if !regexp.MustCompile(contract.pattern).MatchString(source) {
			t.Errorf("Carousel runtime missing reduced-motion %s contract (%s)", contract.name, contract.pattern)
		}
	}
}
