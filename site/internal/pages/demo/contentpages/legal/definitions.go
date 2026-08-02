package legalpages

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definitions contains the grouped legal and attribution pages.
var Definitions = []demo.PageDefinition{
	{Key: "attributions", Title: "Attributions", Active: "attributions", Description: "Review third-party licenses and asset attributions for the Goshtoso documentation site and component library.", Type: "TechArticle", Content: attributionsContent},
	{Key: "license", Title: "License", Active: "license", Description: "Read the Goshtoso license terms for using the Go UI component library in personal and commercial projects.", Type: "TechArticle", Content: licenseContent},
	{Key: "privacy", Title: "Privacy Policy", Active: "privacy", Description: "Understand how the Goshtoso demo site uses browser storage for preferences and local example state without analytics.", Type: "TechArticle", Content: privacyContent},
}
