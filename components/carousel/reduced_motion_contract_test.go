package carousel

import (
	"strings"
	"testing"
)

func TestCarouselAndCardCarouselRenderReducedMotionContracts(t *testing.T) {
	slides := []Slide{
		{ImgSrc: "/one.webp", ImgAlt: "One"},
		{ImgSrc: "/two.webp", ImgAlt: "Two"},
	}

	carouselHTML := renderCoverageCarousel(t, Config{ID: "motion-carousel", Slides: slides, Touch: true})
	if !strings.Contains(carouselHTML, "motion-reduce:transition-none!") {
		t.Fatalf("Carousel slide transition lacks reduced-motion override: %s", carouselHTML)
	}

	cardHTML := renderStructuralCarousel(t, CardCarousel(CardConfig{ID: "motion-card", Slides: slides, Touch: true}))
	if !strings.Contains(cardHTML, "motion-reduce:transition-none!") {
		t.Fatalf("Card Carousel slide transition lacks reduced-motion override: %s", cardHTML)
	}

	htmxHTML := renderCoverageCarousel(t, Config{
		ID: "motion-htmx",
		HTMX: &HTMXConfig{
			Get: "/api/carousel",
		},
	})
	if !strings.Contains(htmxHTML, "animate-spin motion-reduce:animate-none") {
		t.Fatalf("HTMX Carousel loading spinner lacks reduced-motion override: %s", htmxHTML)
	}
}

func TestCarouselControlsHonorReducedMotion(t *testing.T) {
	if got := navButtonClasses(); !strings.Contains(got, "transition motion-reduce:transition-none") {
		t.Fatalf("carousel navigation buttons lack reduced-motion transition suppression: %s", got)
	}

	html := renderCoverageCarousel(t, Config{
		ID:       "motion-controls",
		Autoplay: &AutoplayConfig{Interval: 2500},
		Slides:   []Slide{{ImgSrc: "/one.webp", ImgAlt: "One", Title: "One", CTAHref: "/one", CTALabel: "Open"}},
	})
	for _, want := range []string{
		`class="size-2 rounded-full transition motion-reduce:transition-none"`,
		"transition motion-reduce:transition-none hover:opacity-80",
		"transition motion-reduce:transition-none hover:opacity-75",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("carousel controls missing reduced-motion class %q: %s", want, html)
		}
	}
}
