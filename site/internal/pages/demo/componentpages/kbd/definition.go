// Package kbdpage owns the KBD component documentation page.
package kbdpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the KBD page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/kbd",
	Title:   "KBD",
	Active:  "kbd",
	Type:    "TechArticle",
	Content: kbdDemoContent,
}
