// Package tagslistpage owns the Tags List component documentation page.
package tagslistpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Tags List page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/tags-list",
	Title:   "Tags List",
	Active:  "tags-list",
	Type:    "TechArticle",
	Content: tagsListDemoContent,
}
