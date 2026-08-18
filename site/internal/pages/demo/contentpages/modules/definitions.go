package modulespages

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definitions contains the grouped optional-module pages.
var Definitions = []demo.PageDefinition{
	{Key: "modules/charts", Title: "Charts", Active: "module-charts", Description: "Explore static, interactive, and interactive 3D chart components from the optional Goshtoso Charts module.", Type: "TechArticle", Content: chartsModuleContent},
	{Key: "modules/app-shells", Title: "App Shells", Active: "module-app-shells", Description: "Explore foundational frames and reusable documentation, console, and landing shells from Goshtoso App Shells v0.1.6.", Type: "TechArticle", Content: appShellsModuleContent},
	{Key: "modules/app-shells/frames/component-page", Title: "Component Page", Active: "app-shells-frame-component-page", Description: "Compose consistent component reference pages from previews, source code, sections, and consumer-owned guidance.", Type: "TechArticle", Content: appShellsComponentPageContent},
	{Key: "modules/app-shells/shells/component-docs-shell", Title: "Component Docs Shell", Active: "app-shells-shell-component-docs", Description: "Frame documentation sites with scoped navigation, search, table of contents, and responsive behavior.", Type: "TechArticle", Content: appShellsComponentDocsShellContent},
	{Key: "modules/app-shells/shells/console-shell", Title: "Console Shell", Active: "app-shells-shell-console", Description: "Frame HTMX applications with a persistent header, sidebar, drawer, and main fragment lifecycle.", Type: "TechArticle", Content: appShellsConsoleShellContent},
	{Key: "modules/app-shells/shells/landing-shell", Title: "Landing Shell", Active: "app-shells-shell-landing", Description: "Frame public product and organization pages with responsive navigation and structured footer ownership.", Type: "TechArticle", Content: appShellsLandingShellContent},
}
