// Package indexpage owns the runnable examples index.
package indexpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the examples index's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples",
	Title:   "Examples",
	Active:  "examples",
	Type:    "TechArticle",
	Content: IndexContent,
}
