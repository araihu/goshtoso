// Package formpage owns the Form component documentation page.
package formpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Form page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/form",
	Title:   "Form",
	Active:  "form",
	Type:    "TechArticle",
	Content: formDemoContent,
}
