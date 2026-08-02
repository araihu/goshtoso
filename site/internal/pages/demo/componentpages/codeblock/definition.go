// Package codeblockpage owns the Code Block component documentation page.
package codeblockpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Code Block page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/codeblock",
	Title:   "Code Block",
	Active:  "codeblock",
	Type:    "TechArticle",
	Content: codeBlockDemoContent,
}
