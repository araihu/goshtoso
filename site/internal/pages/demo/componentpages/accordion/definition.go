// Package accordionpage owns the Accordion component documentation page.
package accordionpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Accordion page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/accordion",
	Title:   "Accordion",
	Active:  "accordion",
	Type:    "TechArticle",
	Content: accordionDemoContent,
}
