package demo

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/icon"
	"github.com/araihu/goshtoso/components/sidebar"
	sidebaricons "github.com/araihu/goshtoso/site/internal/demoicons/heroicons"
)

func iconsDocsNavigation(active string) componentdocshell.Navigation {
	activeID := componentDocsID(active)
	return componentdocshell.Navigation{
		Items: []sidebar.Item{
			{
				ID:        "component-icon",
				Label:     "Icon",
				Href:      "/components/icon",
				Icon:      iconsSidebarIcon(sidebaricons.IconHeroiconsOptimized24OutlineSquares2x2),
				Active:    activeID == "component-icon",
				LinkAttrs: navHxAttrs("/components/icon", "Icon"),
			},
			{
				ID:        "icon-catalog",
				Label:     "Icon Catalog",
				Href:      "/docs/icon-catalog",
				Icon:      iconsSidebarIcon(sidebaricons.IconHeroiconsOptimized24OutlineQueueList),
				Active:    activeID == "icon-catalog",
				LinkAttrs: navHxAttrs("/docs/icon-catalog", "Icon Catalog"),
			},
			{
				ID:        "iconpack",
				Label:     "Icon Packs",
				Href:      "/docs/iconpack",
				Icon:      iconsSidebarIcon(sidebaricons.IconHeroiconsOptimized24OutlineClipboardDocumentList),
				Active:    activeID == "iconpack",
				LinkAttrs: navHxAttrs("/docs/iconpack", "Icon Packs"),
			},
		},
		DisableSearch: true,
	}
}

func iconsSidebarIcon(symbol icon.Symbol) templ.Component {
	return sidebaricons.Icon(sidebaricons.Config{Symbol: symbol, Decorative: true})
}
