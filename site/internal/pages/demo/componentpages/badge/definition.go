// Package badgepage owns the Badge component documentation page.
package badgepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Badge page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/badge",
	Title:   "Badge",
	Active:  "badge",
	Type:    "TechArticle",
	Content: badgeDemoContent,
}
