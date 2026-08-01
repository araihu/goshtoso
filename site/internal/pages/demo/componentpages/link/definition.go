// Package linkpage owns the Link component documentation page.
package linkpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Link page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/link",
	Title:   "Link",
	Active:  "link",
	Type:    "TechArticle",
	Content: linkDemoContent,
}
