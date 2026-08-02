// Package tooltippage owns the Tooltip component documentation page.
package tooltippage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Tooltip page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/tooltip",
	Title:   "Tooltip",
	Active:  "tooltip",
	Type:    "TechArticle",
	Content: tooltipDemoContent,
}
