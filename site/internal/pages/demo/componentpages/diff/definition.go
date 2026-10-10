// Package diffpage owns the generic text Diff documentation page.
package diffpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Diff page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key: "components/diff", Title: "Diff", Active: "diff",
	Type: "TechArticle", Content: diffDemoContent,
}
