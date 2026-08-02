// Package profilepage owns the Profile runnable example page.
package profilepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Profile example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples/profile",
	Title:   "Profile",
	Active:  "profile",
	Type:    "SoftwareSourceCode",
	Content: ProfileContent,
}
