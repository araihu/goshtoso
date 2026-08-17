// Package splitbuttonpage owns the SplitButton component documentation page.
package splitbuttonpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the SplitButton page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/splitbutton",
	Title:   "Split Button",
	Active:  "splitbutton",
	Type:    "TechArticle",
	Content: splitButtonDemoContent,
}
