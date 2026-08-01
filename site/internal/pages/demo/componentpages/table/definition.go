// Package tablepage owns the Table component documentation page.
package tablepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Table page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/table",
	Title:   "Table",
	Active:  "table",
	Type:    "TechArticle",
	Content: tableDemoContent,
}
