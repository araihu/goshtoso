// Package navbarpage owns the Navbar component documentation page.
package navbarpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Navbar page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/navbar",
	Title:   "Navbar",
	Active:  "navbar",
	Type:    "TechArticle",
	Content: navbarDemoContent,
}
