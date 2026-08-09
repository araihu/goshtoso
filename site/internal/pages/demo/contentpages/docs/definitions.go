package docspages

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definitions contains the grouped documentation pages.
var Definitions = []demo.PageDefinition{
	{Key: "docs/agents", Title: "AI Agents", Active: "agents", Description: "Install the Goshtoso consumer agent skill for AI coding tools and verify npx skills distribution.", Type: "TechArticle", Content: agentsContent},
	{Key: "docs/application-patterns", Title: "Application Patterns", Active: "application-patterns", Description: "Compose App Shell, Operations List, Detail Workspace, and Multi-step Workflow product surfaces from server-rendered Goshtoso components.", Type: "TechArticle", Content: applicationPatternsContent},
	{Key: "docs/component-model", Title: "Component Model", Active: "component-model", Description: "Understand Goshtoso's common component interface, concrete return values, constructor styles, stable Kind identity, and rendered defaults.", Type: "TechArticle", Content: componentModelContent},
	{Key: "docs/iconpack", Title: "Icon Packs", Active: "iconpack", Description: "Generate verified consumer-local Heroicons and Developer Icons packages that render through Goshtoso's core icon component.", Type: "TechArticle", Content: iconpackContent},
	{Key: "docs/theme", Title: "Theme", Active: "theme", Description: "Customize Goshtoso themes with Tailwind CSS tokens, dark mode, live previews, and server-rendered component examples.", Type: "TechArticle", Content: themeDemoContent},
}
