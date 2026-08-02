// Package fileinputpage owns the File Input component documentation page.
package fileinputpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the File Input page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/fileinput",
	Title:   "File Input",
	Active:  "fileinput",
	Type:    "TechArticle",
	Content: fileInputDemoContent,
}
