// Package modalpage owns the Modal component documentation page.
package modalpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Modal page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/modal",
	Title:   "Modal",
	Active:  "modal",
	Type:    "TechArticle",
	Content: modalDemoContent,
}
