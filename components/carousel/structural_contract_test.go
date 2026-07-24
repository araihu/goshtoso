package carousel

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestCarouselInfersOverlayFromSlideContent(t *testing.T) {
	html := renderStructuralCarousel(t, Carousel(Config{Slides: []Slide{{
		ImgSrc:   "/x.webp",
		Title:    "Release",
		CTALabel: "Read",
		CTAHref:  "/release",
	}}}))

	require.Contains(t, html, "Release")
	require.Contains(t, html, "/release")
	require.Contains(t, html, `x-bind:href="slide.ctaUrl"`)
}

func TestCardCarouselOwnsCardWrapper(t *testing.T) {
	html := renderStructuralCarousel(t, CardCarousel(CardConfig{
		ID:     "featured",
		Slides: []Slide{{ImgSrc: "/x.webp", Title: "Release"}},
	}))

	require.Contains(t, html, "<article")
	require.Contains(t, html, `id="featured"`)
	require.Contains(t, html, "Release")
}

func renderStructuralCarousel(t *testing.T, component templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buf))
	return buf.String()
}
