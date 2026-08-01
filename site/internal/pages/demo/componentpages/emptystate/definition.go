// Package emptystatepage owns the Empty State component documentation page.
package emptystatepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Empty State page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/empty-state",
	Title:   "Empty State",
	Active:  "empty-state",
	Type:    "TechArticle",
	Content: emptyStateDemoContent,
}
