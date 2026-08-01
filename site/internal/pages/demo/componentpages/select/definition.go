// Package selectpage owns the Select component documentation page.
package selectpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Select page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/select",
	Title:   "Select",
	Active:  "select",
	Type:    "TechArticle",
	Content: selectDemoContent,
}
