// Package structuredinputpage owns the Structured Input component documentation page.
package structuredinputpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Structured Input page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/structured-input",
	Title:   "Structured Input",
	Active:  "structured-input",
	Type:    "TechArticle",
	Content: structuredInputDemoContent,
}
