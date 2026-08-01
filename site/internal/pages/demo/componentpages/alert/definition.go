// Package alertpage owns the Alert component documentation page.
package alertpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Alert page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/alert",
	Title:   "Alert",
	Active:  "alert",
	Type:    "TechArticle",
	Content: alertDemoContent,
}
