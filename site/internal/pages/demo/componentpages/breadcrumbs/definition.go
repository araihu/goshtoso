// Package breadcrumbspage owns the Breadcrumbs component documentation page.
package breadcrumbspage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Breadcrumbs page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/breadcrumbs",
	Title:   "Breadcrumbs",
	Active:  "breadcrumbs",
	Type:    "TechArticle",
	Content: breadcrumbsDemoContent,
}
