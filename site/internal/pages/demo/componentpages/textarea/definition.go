// Package textareapage owns the Textarea component documentation page.
package textareapage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Textarea page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/textarea",
	Title:   "Textarea",
	Active:  "textarea",
	Type:    "TechArticle",
	Content: textareaDemoContent,
}
