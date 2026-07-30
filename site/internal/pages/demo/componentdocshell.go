package demo

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/sidebar"
	siteassets "github.com/araihu/goshtoso/site/assets"
	"github.com/araihu/goshtoso/site/internal/buildinfo"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
)

// ComponentDocsLayout delegates the demo frame to the reusable component
// documentation shell while retaining Goshtoso-owned routes and content.
func ComponentDocsLayout(meta PageMeta, active string, content templ.Component, persist bool) templ.Component {
	cfg := componentDocsConfig(persist)
	return componentdocshell.Layout(cfg, componentDocsPage(cfg, meta, active, content))
}

func ComponentDocsFragment(meta PageMeta, active string, content templ.Component, persist bool) templ.Component {
	cfg := componentDocsConfig(persist)
	return componentdocshell.Fragment(cfg, componentDocsPage(cfg, meta, active, content))
}

func componentDocsConfig(persist bool) componentdocshell.Config {
	sections := getSidebarSections("")
	for i := range sections {
		for j := range sections[i].Items {
			sections[i].Items[j].ID = componentDocsID(sections[i].Items[j].ID)
		}
	}
	return componentdocshell.Config{
		Brand:      componentdocshell.Brand{Name: SiteName, HomeURL: "/", Logo: templ.Raw(`<img src="/assets/images/goshtoso-logo.svg" alt="" aria-hidden="true" class="h-12 w-auto">`), HideName: true, FaviconURL: "/favicon.svg", Badge: componentDocsBuildBadge(buildinfo.GoDocsVersion())},
		Navigation: componentdocshell.Navigation{Items: getSidebarTopItems(""), SectionsTitle: "Components", Sections: sections, SearchPlaceholder: "Search", SearchSlot: sidebarSearchSlot()},
		Appearance: componentdocshell.AppearanceConfig{
			Themes:               getThemeOptions(),
			DefaultTheme:         "araihu",
			DisableThemeSelector: true,
			PersistPreferences:   persist,
			DarkModeBinding: &componentdocshell.DarkModeBinding{
				ButtonID:         "darkModeToggleBtn",
				StateExpression:  "$store.darkMode.on",
				ToggleExpression: "$store.darkMode.toggle()",
			},
		},
		Interactions: componentdocshell.InteractionConfig{
			EnableHTMX:   true,
			LocalRuntime: true,
			// componentdocshell renders RuntimeScripts synchronously after its deferred
			// Goshtoso/Alpine tags. The site providers therefore subscribe to
			// alpine:init before Alpine executes, without entering head.Dependencies.
			RuntimeScripts: []string{siteassets.DemoBundleURL, assets.HTMXExtWSURL, assets.HTMXExtSSEURL},
		},
		TOC:     componentdocshell.TOCConfig{RailID: "toc-rail", ListID: "toc-list"},
		BodyEnd: componentDocsBodyEnd(), RepositoryURL: "https://github.com/araihu/goshtoso", AssetPrefix: "/componentdocshell/assets/",
	}
}

func componentDocsBuildBadge(version string) *componentdocshell.BrandBadge {
	if version == "development" {
		return &componentdocshell.BrandBadge{Label: "dev", AriaLabel: "Development build"}
	}
	if !goModuleVersionPattern.MatchString(version) {
		return nil
	}
	return &componentdocshell.BrandBadge{
		Label:     version,
		AriaLabel: "Goshtoso release " + version,
		Href:      "https://github.com/araihu/goshtoso/releases/tag/" + version,
	}
}

func componentDocsPage(cfg componentdocshell.Config, meta PageMeta, active string, content templ.Component) componentdocshell.Page {
	return componentdocshell.Page{Title: meta.Title, DocumentTitle: meta.TitleText(), Description: meta.Description, CanonicalURL: meta.CanonicalURL(), Active: configuredComponentDocsActive(cfg.Navigation, componentDocsID(active)), Content: componentDocsContent(content, active), Head: componentDocsHead(meta), EnableTOC: true}
}

func configuredComponentDocsActive(navigation componentdocshell.Navigation, active string) string {
	if componentDocsItemsContain(navigation.Items, active) {
		return active
	}
	for _, section := range navigation.Sections {
		if componentDocsItemsContain(section.Items, active) {
			return active
		}
	}
	return ""
}

func componentDocsItemsContain(items []sidebar.Item, active string) bool {
	for _, item := range items {
		if item.ID == active || componentDocsItemsContain(item.Items, active) {
			return true
		}
	}
	return false
}

func componentDocsID(active string) string {
	// Getting Started retains the legacy empty active key used by its registry
	// entry, while the reusable shell needs the concrete navigation item ID.
	if active == "" || active == "getting-started" {
		return "home"
	}
	for _, p := range catalog.ComponentPages() {
		if p.Active == active {
			return "component-" + active
		}
	}
	return active
}

func componentDocsContent(content templ.Component, active string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, c := range []templ.Component{content, componentGoAPIReference(active), componentNavFooter(active), siteFooter()} {
			if c != nil {
				if err := c.Render(ctx, w); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
