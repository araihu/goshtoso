// Package carouselpage owns the Carousel component documentation page.
package carouselpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Carousel page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/carousel",
	Title:   "Carousel",
	Active:  "carousel",
	Type:    "TechArticle",
	Content: carouselDemoContent,
}
