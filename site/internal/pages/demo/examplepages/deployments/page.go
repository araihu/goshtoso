package deploymentspage

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/consoleshell"
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

func Layout(v View) templ.Component {
	active := v.Screen
	if active == "detail" {
		active = "list"
	}
	title := map[string]string{"overview": "Overview", "list": "Deployments", "new": "New deployment", "detail": "Deployment details"}[v.Screen]
	cfg := consoleshell.Config{
		Brand: consoleshell.Brand{Name: "Northstar", HomeURL: v.State.URL("overview", 0)},
		Navigation: consoleshell.Navigation{DisableSearch: true, Items: []sidebar.Item{
			{ID: "overview", Label: "Overview", Href: v.State.URL("overview", 0)},
			{ID: "list", Label: "Deployments", Href: v.State.URL("list", 0)},
			{ID: "new", Label: "New deployment", Href: v.State.URL("new", 0)},
		}},
		HeaderActions: HeaderActions(), Footer: Footer(),
		Interactions: consoleshell.InteractionConfig{LocalRuntime: true},
		Appearance:   consoleshell.AppearanceConfig{DefaultTheme: "araihu"},
	}
	return consoleshell.Layout(cfg, consoleshell.Page{Title: title, Active: active, Content: Body(v), Head: Styles()})
}
