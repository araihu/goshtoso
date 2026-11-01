// Package layoutpage owns the Layout component documentation.
package layoutpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition registers the public mixed-content layout demo.
var Definition = demo.PageDefinition{Key: "components/layout", Title: "Layout", Active: "layout", Type: "TechArticle", Content: layoutDemoContent}
