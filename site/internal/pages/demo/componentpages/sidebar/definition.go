// Package sidebarpage owns the Sidebar component documentation page.
package sidebarpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Sidebar page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/sidebar",
	Title:   "Sidebar",
	Active:  "sidebar",
	Type:    "TechArticle",
	Content: sidebarDemoContent,
}
