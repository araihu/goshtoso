// Package appshellpage owns the App Shell component documentation page.
package appshellpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the App Shell page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/app-shell",
	Title:   "App Shell",
	Active:  "app-shell",
	Type:    "TechArticle",
	Content: appShellDemoContent,
}
