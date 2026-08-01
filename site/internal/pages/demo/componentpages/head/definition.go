// Package headpage owns the Dependencies component documentation page.
package headpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Dependencies page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/dependencies",
	Title:   "Dependencies",
	Active:  "dependencies",
	Type:    "TechArticle",
	Content: dependenciesDemoContent,
}
