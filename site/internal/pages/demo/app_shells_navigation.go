package demo

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	searchfield "github.com/araihu/goshtoso/components/search"
	"github.com/araihu/goshtoso/components/sidebar"
	sidebaricons "github.com/araihu/goshtoso/site/internal/demoicons/heroicons"
)

type appShellsSidebarLink struct {
	ID          string
	Label       string
	Href        string
	Description string
}

type appShellsSidebarGroup struct {
	Title string
	Items []appShellsSidebarLink
}

func appShellsDocsNavigation(active string) componentdocshell.Navigation {
	items := make([]sidebar.Item, 0, len(appShellsSidebarTopLinks()))
	for _, link := range appShellsSidebarTopLinks() {
		items = append(items, sidebar.Item{
			ID:        link.ID,
			Label:     link.Label,
			Href:      link.Href,
			Icon:      appShellsSidebarIcon(link.ID),
			Active:    active == link.ID,
			LinkAttrs: navHxAttrs(link.Href, link.Label),
		})
	}

	sections := make([]sidebar.Section, 0, len(appShellsSidebarGroups()))
	for _, group := range appShellsSidebarGroups() {
		section := sidebar.Section{Title: group.Title}
		for _, link := range group.Items {
			section.Items = append(section.Items, sidebar.Item{
				ID:        link.ID,
				Label:     link.Label,
				Href:      link.Href,
				Icon:      appShellsSidebarIcon(link.ID),
				Active:    active == link.ID,
				LinkAttrs: navHxAttrs(link.Href, link.Label),
			})
		}
		sections = append(sections, section)
	}

	return componentdocshell.Navigation{
		Items:         items,
		SectionsTitle: "App Shells",
		Sections:      sections,
		Scope:         appShellsScope(),
		DisableSearch: true,
	}
}

func appShellsSidebarTopLinks() []appShellsSidebarLink {
	return []appShellsSidebarLink{
		{
			ID:          "module-app-shells",
			Label:       "Overview",
			Href:        "/modules/app-shells",
			Description: "Choose a foundational frame or a complete application shell.",
		},
	}
}

func appShellsSidebarGroups() []appShellsSidebarGroup {
	return []appShellsSidebarGroup{
		{
			Title: "Frames",
			Items: []appShellsSidebarLink{
				{
					ID:          "app-shells-frame-component-page",
					Label:       "Component Page",
					Href:        "/modules/app-shells/frames/component-page",
					Description: "Compose a consistent reference page from a preview, source, sections, and guidance.",
				},
			},
		},
		{
			Title: "Shells",
			Items: []appShellsSidebarLink{
				{
					ID:          "app-shells-shell-component-docs",
					Label:       "Component Docs Shell",
					Href:        "/modules/app-shells/shells/component-docs-shell",
					Description: "Frame documentation sites with scoped navigation, search, TOC, and responsive behavior.",
				},
				{
					ID:          "app-shells-shell-console",
					Label:       "Console Shell",
					Href:        "/modules/app-shells/shells/console-shell",
					Description: "Run HTMX applications with a persistent header, sidebar, drawer, and main fragment lifecycle.",
				},
				{
					ID:          "app-shells-shell-landing",
					Label:       "Landing Shell",
					Href:        "/modules/app-shells/shells/landing-shell",
					Description: "Build public product and organization pages with responsive navigation and structured footer ownership.",
				},
			},
		},
	}
}

func appShellsSidebarIcon(id string) templ.Component {
	symbol := sidebaricons.IconHeroiconsOptimized24OutlineSquares2x2
	switch id {
	case "module-app-shells", "app-shells-frame-component-page":
		symbol = sidebaricons.IconHeroiconsOptimized24OutlineCube
	case "app-shells-shell-component-docs":
		symbol = sidebaricons.IconHeroiconsOptimized24OutlineQueueList
	case "app-shells-shell-landing":
		symbol = sidebaricons.IconHeroiconsOptimized24OutlineSwatch
	}
	return sidebaricons.Icon(sidebaricons.Config{Symbol: symbol, Decorative: true})
}

func appShellsScope() *componentdocshell.ScopeMetadata {
	return &componentdocshell.ScopeMetadata{
		ModulePath:  "github.com/araihu/goshtoso-app-shells",
		ModuleLabel: "araihu/goshtoso-app-shells",
		ModuleURL:   "https://github.com/araihu/goshtoso-app-shells",
		Version:     "v0.1.6",
		VersionURL:  "https://github.com/araihu/goshtoso-app-shells/releases/tag/v0.1.6",
	}
}

func appShellsSearchItems() []searchfield.Item {
	items := make([]searchfield.Item, 0, 1+len(appShellsSidebarGroups()))
	for _, link := range appShellsSidebarTopLinks() {
		items = append(items, appShellsSearchItem(link))
	}
	for _, group := range appShellsSidebarGroups() {
		for _, link := range group.Items {
			items = append(items, appShellsSearchItem(link))
		}
	}
	return items
}

func appShellsSearchItem(link appShellsSidebarLink) searchfield.Item {
	return searchfield.Item{
		ID:          "app-shells-search-" + link.ID,
		Title:       link.Label,
		Description: link.Description,
		Href:        link.Href,
		Section:     "App Shells",
		Keywords:    []string{link.ID, "frame", "shell"},
		Attrs:       navHxAttrs(link.Href, link.Label),
	}
}
