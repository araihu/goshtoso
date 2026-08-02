// Package iconpage owns the Icon component documentation page.
package iconpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Icon page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/icon",
	Title:   "Icon",
	Active:  "icon",
	Type:    "TechArticle",
	Content: iconDemoContent,
}
