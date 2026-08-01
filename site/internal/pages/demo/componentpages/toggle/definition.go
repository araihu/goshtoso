// Package togglepage owns the Toggle component documentation page.
package togglepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Toggle page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/toggle",
	Title:   "Toggle",
	Active:  "toggle",
	Type:    "TechArticle",
	Content: toggleDemoContent,
}
