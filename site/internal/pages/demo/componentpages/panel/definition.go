// Package panelpage owns the Panel component documentation page.
package panelpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Panel page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/panel",
	Title:   "Panel",
	Active:  "panel",
	Type:    "TechArticle",
	Content: panelDemoContent,
}
