// Package schemaformpage owns the Schema Form component documentation page.
package schemaformpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Schema Form page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/schema-form",
	Title:   "Schema Form",
	Active:  "schema-form",
	Type:    "TechArticle",
	Content: schemaFormDemoContent,
}
