// Package ratingpage owns the Rating component documentation page.
package ratingpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Rating page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/rating",
	Title:   "Rating",
	Active:  "rating",
	Type:    "TechArticle",
	Content: ratingDemoContent,
}
