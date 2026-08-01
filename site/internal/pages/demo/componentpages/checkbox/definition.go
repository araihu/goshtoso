// Package checkboxpage owns the Checkbox component documentation page.
package checkboxpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Checkbox page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/checkbox",
	Title:   "Checkbox",
	Active:  "checkbox",
	Type:    "TechArticle",
	Content: checkboxDemoContent,
}
