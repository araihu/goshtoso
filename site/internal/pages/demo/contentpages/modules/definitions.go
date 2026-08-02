package modulespages

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definitions contains the grouped optional-module pages.
var Definitions = []demo.PageDefinition{
	{Key: "modules/charts", Title: "Charts", Active: "module-charts", Description: "Explore static, interactive, and interactive 3D chart components from the optional Goshtoso Charts module.", Type: "TechArticle", Content: chartsModuleContent},
	{Key: "modules/app-shells", Title: "App Shells", Active: "module-app-shells", Description: "Explore reusable documentation and console application shells from the optional Goshtoso App Shells module.", Type: "TechArticle", Content: appShellsModuleContent},
}
