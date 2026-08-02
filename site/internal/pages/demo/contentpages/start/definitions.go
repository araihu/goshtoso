package startpages

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definitions contains the grouped landing and getting-started pages.
var Definitions = []demo.PageDefinition{
	{Key: "getting-started", Title: "Getting Started", Description: "Start a Go HTMX app with Goshtoso, templ, Tailwind CSS, local runtime assets, and copy-pasteable setup code.", Type: "TechArticle", Content: gettingStartedContent},
}
