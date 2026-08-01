// Package skeletonpage owns the Skeleton component documentation page.
package skeletonpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Skeleton page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/skeleton",
	Title:   "Skeleton",
	Active:  "skeleton",
	Type:    "TechArticle",
	Content: skeletonDemoContent,
}
