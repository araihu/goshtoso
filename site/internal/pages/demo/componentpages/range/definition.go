// Package rangepage owns the Range component documentation page.
package rangepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Range page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/range",
	Title:   "Range",
	Active:  "range",
	Type:    "TechArticle",
	Content: rangeDemoContent,
}
