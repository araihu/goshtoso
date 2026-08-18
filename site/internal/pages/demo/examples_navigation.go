package demo

import (
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/sidebar"
)

type examplesSidebarLink struct {
	ID    string
	Label string
	Href  string
}

func examplesDocsNavigation(active string) componentdocshell.Navigation {
	return componentdocshell.Navigation{
		Sections: []sidebar.Section{
			{
				Title: "Modules",
				Items: []sidebar.Item{
					{ID: "module-charts", Label: "Charts", Href: "/modules/charts", LinkAttrs: navHxAttrs("/modules/charts", "Charts")},
					{ID: "module-app-shells", Label: "App Shells", Href: "/modules/app-shells", LinkAttrs: navHxAttrs("/modules/app-shells", "App Shells")},
				},
			},
			{
				Title: "Examples",
				Items: examplesSidebarItems(active),
			},
		},
		DisableSearch: true,
	}
}

func examplesSidebarItems(active string) []sidebar.Item {
	items := make([]sidebar.Item, 0, len(examplesSidebarLinks()))
	for _, link := range examplesSidebarLinks() {
		items = append(items, sidebar.Item{
			ID:        link.ID,
			Label:     link.Label,
			Href:      link.Href,
			Active:    active == link.ID,
			LinkAttrs: navHxAttrs(link.Href, link.Label),
		})
	}
	return items
}

func examplesSidebarLinks() []examplesSidebarLink {
	return []examplesSidebarLink{
		{ID: "ticker", Label: "Live Ticker", Href: "/examples/ticker"},
		{ID: "todo", Label: "Todo List", Href: "/examples/todo"},
		{ID: "expense", Label: "Expense Tracker", Href: "/examples/expense"},
		{ID: "chat", Label: "Chat", Href: "/examples/chat"},
		{ID: "logs", Label: "Live Log Feed", Href: "/examples/logs"},
		{ID: "profile", Label: "Profile", Href: "/examples/profile"},
		{ID: "wizard", Label: "Onboarding Wizard", Href: "/examples/wizard"},
	}
}
