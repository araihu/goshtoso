// Package scrollregionpage owns the Scroll Region component documentation page.
package scrollregionpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Scroll Region page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:         "components/scroll-region",
	Title:       "Scroll Region",
	Active:      "scroll-region",
	Description: "Keep bounded content keyboard- and touch-scrollable with automatic start and end boundary cues.",
	Type:        "TechArticle",
	Content:     scrollRegionDemoContent,
}
