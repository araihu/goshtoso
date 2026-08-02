// Package cardpage owns the Card component documentation page.
package cardpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Card page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/card",
	Title:   "Card",
	Active:  "card",
	Type:    "TechArticle",
	Content: cardDemoContent,
}
