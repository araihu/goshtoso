// Package buttonpage owns the Button component documentation page.
package buttonpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Button page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:         "components/button",
	Title:       "Buttons",
	Active:      "button",
	Description: "Trigger actions with semantic tones, sizes, native form behavior, and first-class HTMX wiring.",
	Type:        "TechArticle",
	Content:     buttonDemoContent,
}
