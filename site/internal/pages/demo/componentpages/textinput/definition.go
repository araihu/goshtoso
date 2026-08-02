// Package textinputpage owns the Text Input component documentation page.
package textinputpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Text Input page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/text-input",
	Title:   "Text Input",
	Active:  "text-input",
	Type:    "TechArticle",
	Content: textInputDemoContent,
}
