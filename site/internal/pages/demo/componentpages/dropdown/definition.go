// Package dropdownpage owns the Dropdown component documentation page.
package dropdownpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Dropdown page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/dropdown",
	Title:   "Dropdown",
	Active:  "dropdown",
	Type:    "TechArticle",
	Content: dropdownDemoContent,
}
