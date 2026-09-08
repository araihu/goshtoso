// Package schematreepage owns the Schema Tree component documentation page.
package schematreepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Schema Tree documentation registry entry.
var Definition = demo.PageDefinition{
	Key: "components/schema-tree", Title: "Schema Tree", Active: "schema-tree",
	Type: "TechArticle", Content: schemaTreeDemoContent,
}
