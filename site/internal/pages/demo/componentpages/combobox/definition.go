// Package comboboxpage owns the Combobox component documentation page.
package comboboxpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Combobox page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/combobox",
	Title:   "Combobox",
	Active:  "combobox",
	Type:    "TechArticle",
	Content: comboboxDemoContent,
}
