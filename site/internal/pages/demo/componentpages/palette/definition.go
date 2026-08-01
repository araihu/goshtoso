// Package palettepage owns the Palette component documentation page.
package palettepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Palette page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/palette",
	Title:   "Palette",
	Active:  "palette",
	Type:    "TechArticle",
	Content: paletteDemoContent,
}
