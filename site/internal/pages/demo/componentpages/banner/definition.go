// Package bannerpage owns the Banner component documentation page.
package bannerpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Banner page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/banner",
	Title:   "Banner",
	Active:  "banner",
	Type:    "TechArticle",
	Content: bannerDemoContent,
}
