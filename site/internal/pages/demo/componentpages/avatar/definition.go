// Package avatarpage owns the Avatar component documentation page.
package avatarpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Avatar page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/avatar",
	Title:   "Avatar",
	Active:  "avatar",
	Type:    "TechArticle",
	Content: avatarDemoContent,
}
