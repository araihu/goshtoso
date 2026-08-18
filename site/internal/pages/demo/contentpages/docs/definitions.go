package docspages

import (
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	iconpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/icon"
)

// Definitions contains the grouped documentation pages.
var Definitions = []demo.PageDefinition{
	{Key: "docs/agents", Title: "AI Agents", Active: "agents", Description: "Install the Goshtoso consumer skill for AI coding tools and follow the supported integration path.", Type: "TechArticle", Content: agentsContent},
	{Key: "docs/application-patterns", Title: "Application Patterns", Active: "application-patterns", Description: "Compose App Shell, Operations List, Detail Workspace, and Multi-step Workflow product surfaces from server-rendered Goshtoso components.", Type: "TechArticle", Content: applicationPatternsContent},
	{Key: "docs/component-model", Title: "Component Model", Active: "component-model", Description: "Understand Goshtoso's common component interface, concrete return values, constructor styles, stable Kind identity, and rendered defaults.", Type: "TechArticle", Content: componentModelContent},
	{Key: "docs/icon-catalog", Title: "Icon Catalog", Active: "icon-catalog", Description: "Browse bundled Heroicons symbols, inspect accessible states, and copy ready-to-use templ examples.", Type: "TechArticle", Content: iconpage.IconCatalogContent},
	{Key: "docs/iconpack", Title: "Icon Packs", Active: "iconpack", Description: "Generate a consumer-owned icon package from verified Arai Hu Assets releases, GitHub trees, remote SVGs, or multiple sources with an explicit .iconpack.yaml lock.", Type: "TechArticle", Content: iconpackContent},
	{Key: "docs/theme", Title: "Theme", Active: "theme", Description: "Customize Goshtoso themes with Tailwind CSS tokens, dark mode, live previews, and server-rendered component examples.", Type: "TechArticle", Content: themeDemoContent},
}
