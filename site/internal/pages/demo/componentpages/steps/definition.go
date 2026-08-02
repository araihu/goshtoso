// Package stepspage owns the Steps component documentation page.
package stepspage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Steps page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/steps",
	Title:   "Steps",
	Active:  "steps",
	Type:    "TechArticle",
	Content: stepsDemoContent,
}
