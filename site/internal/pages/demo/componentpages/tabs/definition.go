// Package tabspage owns the Tabs component documentation page.
package tabspage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Tabs page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/tabs",
	Title:   "Tabs",
	Active:  "tabs",
	Type:    "TechArticle",
	Content: tabsDemoContent,
}
