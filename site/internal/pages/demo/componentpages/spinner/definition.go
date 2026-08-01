// Package spinnerpage owns the Spinner component documentation page.
package spinnerpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Spinner page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/spinner",
	Title:   "Spinner",
	Active:  "spinner",
	Type:    "TechArticle",
	Content: spinnerDemoContent,
}
