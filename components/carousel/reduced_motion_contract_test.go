package carousel

import (
	"strings"
	"testing"
)

func TestCarouselAndCardCarouselRenderReducedMotionTransitions(t *testing.T) {
	slides := []Slide{
		{ImgSrc: "/one.webp", ImgAlt: "One"},
		{ImgSrc: "/two.webp", ImgAlt: "Two"},
	}

	carouselHTML := renderCoverageCarousel(t, Config{ID: "motion-carousel", Slides: slides, Touch: true})
	if !strings.Contains(carouselHTML, "motion-reduce:transition-none") {
		t.Fatalf("Carousel slide transition lacks reduced-motion override: %s", carouselHTML)
	}

	cardHTML := renderStructuralCarousel(t, CardCarousel(CardConfig{ID: "motion-card", Slides: slides, Touch: true}))
	if !strings.Contains(cardHTML, "motion-reduce:transition-none") {
		t.Fatalf("Card Carousel slide transition lacks reduced-motion override: %s", cardHTML)
	}
}
