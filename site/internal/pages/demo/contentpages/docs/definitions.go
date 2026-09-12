package docspages

import (
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	iconpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/icon"
)

// Definitions contains the grouped documentation pages.
var Definitions = []demo.PageDefinition{
	{Key: "docs/internationalization/files", Title: "Expression files", Active: "internationalization-files", Description: "Load JSON and YAML expressions from bytes, disk, and embedded files, with JSON Schema validation and a complete property reference.", Type: "TechArticle", Content: internationalizationFiles},
	{Key: "docs/internationalization/examples", Title: "Internationalization Examples", Active: "internationalization-examples", Description: "Examples of system defaults, user language preferences, HTMX request contexts, and component expression overrides.", Type: "TechArticle", Content: internationalizationExamplesContent},
	{Key: "docs/internationalization", Title: "Internationalization", Active: "internationalization", Description: "Supply application-owned translations with English defaults, request-scoped expressions, and component overrides.", Type: "TechArticle", Content: internationalizationContent},
	{Key: "docs/agents", Title: "AI Agents", Active: "agents", Description: "Install the Goshtoso consumer skill for AI coding tools and follow the supported integration path.", Type: "TechArticle", Content: agentsContent},
	{Key: "docs/component-model", Title: "Component Model", Active: "component-model", Description: "Understand Goshtoso's common component interface, concrete return values, constructor styles, stable Kind identity, and rendered defaults.", Type: "TechArticle", Content: componentModelContent},
	{Key: "docs/icon-catalog", Title: "Icon Catalog", Active: "icon-catalog", Description: "Browse bundled Heroicons symbols, inspect accessible states, and copy ready-to-use templ examples.", Type: "TechArticle", Content: iconpage.IconCatalogContent},
	{Key: "docs/iconpack", Title: "Icon Packs", Active: "iconpack", Description: "Generate a consumer-owned icon package from verified Arai Hu Assets releases, GitHub trees, remote SVGs, or multiple sources with an explicit .iconpack.yaml lock.", Type: "TechArticle", Content: iconpackContent},
	{Key: "docs/theme", Title: "Theme", Active: "theme", Description: "Customize Goshtoso themes with Tailwind CSS tokens, dark mode, live previews, and server-rendered component examples.", Type: "TechArticle", Content: themeDemoContent},
}
