// Package toolbarpage owns the Toolbar component documentation page.
package toolbarpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Toolbar page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/toolbar",
	Title:   "Toolbar",
	Active:  "toolbar",
	Type:    "TechArticle",
	Content: toolbarDemoContent,
}
