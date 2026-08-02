package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/araihu/goshtoso/components/carousel"
)

func (s *Server) handleCarouselSlides(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Simulate server processing delay
	time.Sleep(500 * time.Millisecond)

	// Return a rendered static carousel with sample slides
	cfg := carousel.Config{
		Slides: []carousel.Slide{
			{
				ImgSrc:      "/assets/images/carousel/slide-1.webp",
				ImgAlt:      "Vibrant abstract painting with swirling blue and light pink hues on a canvas.",
				Title:       "Loaded via HTMX",
				Description: fmt.Sprintf("This carousel was fetched from the server at %s.", time.Now().Format("15:04:05")),
			},
			{
				ImgSrc:      "/assets/images/carousel/slide-2.webp",
				ImgAlt:      "Vibrant abstract painting with swirling red, yellow, and pink hues on a canvas.",
				Title:       "Dynamic Content",
				Description: "Slides can come from a database, API, or any backend source.",
			},
			{
				ImgSrc:      "/assets/images/carousel/slide-3.webp",
				ImgAlt:      "Vibrant abstract painting with swirling blue and purple hues on a canvas.",
				Title:       "HTMX + Alpine.js",
				Description: "Server delivers HTML fragments, Alpine handles interactivity.",
			},
		},
	}
	_ = carousel.Carousel(cfg).Render(r.Context(), w)
}
