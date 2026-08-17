// Package popoverpage owns the Popover component documentation page.
package popoverpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Popover page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/popover",
	Title:   "Popover",
	Active:  "popover",
	Type:    "TechArticle",
	Content: popoverDemoContent,
}
