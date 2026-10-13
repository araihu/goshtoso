package deploymentspage

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/araihu/goshtoso/site/internal/examples/deployments"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
)

var Definition = demo.PageDefinition{Key: "examples/deployments", Title: "Deployment Console", Active: "deployments", Description: "A complete application layout with navigation, searchable deployments, a validated creation form, and approval actions.", Type: "SoftwareSourceCode", Content: Content}

type View struct {
	State                         deployments.State
	Screen, Query                 string
	Record                        deployments.Record
	Errors                        map[string]string
	Service, Version, Environment string
	Created                       bool
}

func navigation(v View) sidebar.Config {
	active := v.Screen
	if active == "detail" {
		active = "list"
	}
	items := []sidebar.Item{
		{ID: "overview", Label: "Overview", Href: v.State.URL("overview", 0)},
		{ID: "list", Label: "Deployments", Href: v.State.URL("list", 0)},
		{ID: "new", Label: "New deployment", Href: v.State.URL("new", 0)},
	}
	for i := range items {
		items[i].Active = items[i].ID == active
		items[i].LinkAttrs = navigationAttrs(items[i].Href)
	}
	return sidebar.Config{Items: items, DisableSkipLink: true}
}

func navigationAttrs(href string) templ.Attributes {
	return templ.Attributes{
		"hx-get":      href,
		"hx-target":   "#main-content",
		"hx-swap":     "innerHTML",
		"hx-push-url": "true",
	}
}
