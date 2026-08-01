// Package searchpage owns the Search component documentation page.
package searchpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Search page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/search",
	Title:   "Search",
	Active:  "search",
	Type:    "TechArticle",
	Content: searchDemoContent,
}
