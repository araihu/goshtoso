// Package toastpage owns the Toast component documentation page.
package toastpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Toast page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/toast",
	Title:   "Toast",
	Active:  "toast",
	Type:    "TechArticle",
	Content: toastDemoContent,
}
