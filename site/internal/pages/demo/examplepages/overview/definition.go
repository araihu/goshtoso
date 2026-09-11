package overviewpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

var Definition = demo.PageDefinition{
	Key: "examples", Title: "Examples", Active: "examples",
	Description: "Try working Goshtoso applications: manage expenses, edit a profile, complete onboarding, and explore live updates.",
	Type:        "CollectionPage", Content: Content,
}
