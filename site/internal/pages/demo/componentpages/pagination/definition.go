// Package paginationpage owns the Pagination component documentation page.
package paginationpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Pagination page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/pagination",
	Title:   "Pagination",
	Active:  "pagination",
	Type:    "TechArticle",
	Content: paginationDemoContent,
}
