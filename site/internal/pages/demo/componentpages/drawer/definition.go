// Package drawerpage owns the Drawer component documentation page.
package drawerpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Drawer page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/drawer",
	Title:   "Drawer",
	Active:  "drawer",
	Type:    "TechArticle",
	Content: drawerDemoContent,
}
