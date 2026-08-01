// Package radiopage owns the Radio component documentation page.
package radiopage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Radio page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/radio",
	Title:   "Radio",
	Active:  "radio",
	Type:    "TechArticle",
	Content: radioDemoContent,
}
