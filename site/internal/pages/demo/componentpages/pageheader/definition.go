// Package pageheaderpage owns the Page Header component documentation page.
package pageheaderpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Page Header page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/page-header",
	Title:   "Page Header",
	Active:  "page-header",
	Type:    "TechArticle",
	Content: pageHeaderDemoContent,
}
