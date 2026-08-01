// Package actiongrouppage owns the Action Group component documentation page.
package actiongrouppage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Action Group page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:         "components/action-group",
	Title:       "Action Group",
	Active:      "action-group",
	Description: "Keep a primary action visible while lower-priority actions collapse into an accessible overflow menu.",
	Type:        "TechArticle",
	Content:     actionGroupDemoContent,
}
