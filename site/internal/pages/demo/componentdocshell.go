package demo

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/navbar"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/araihu/goshtoso/site/internal/buildinfo"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
)

// ComponentDocsLayout delegates the demo frame to the reusable component
// documentation shell while retaining Goshtoso-owned routes and content.
func ComponentDocsLayout(meta PageMeta, active string, content templ.Component, persist bool) templ.Component {
	cfg := componentDocsConfig(persist, active)
	return componentdocshell.Layout(cfg, componentDocsPage(cfg, meta, active, content))
}

func ComponentDocsFragment(meta PageMeta, active string, content templ.Component, persist bool) templ.Component {
	cfg := componentDocsConfig(persist, active)
	return componentdocshell.Fragment(cfg, componentDocsPage(cfg, meta, active, content))
}

func componentDocsConfig(persist bool, active string) componentdocshell.Config {
	navigation := componentDocsNavigation(active)
	return componentdocshell.Config{
		Brand:      componentdocshell.Brand{Name: SiteName, HomeURL: "/", Logo: templ.Raw(`<img src="/assets/images/goshtoso-logo.svg" alt="" aria-hidden="true" class="h-12 w-auto">`), HideName: true, FaviconURL: "/favicon.svg", Badge: componentDocsBuildBadge(buildinfo.GoDocsVersion())},
		Navigation: navigation,
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
			RuntimeScripts: componentDocsAdditionalRuntimeScripts(),
		},
		TOC:           componentdocshell.TOCConfig{RailID: "toc-rail", ListID: "toc-list"},
		HeaderActions: componentDocsHeaderActions(componentDocsFamily(active)),
		BodyEnd:       componentDocsBodyEndFor(componentDocsFamily(active)), RepositoryURL: "https://github.com/araihu/goshtoso", AssetPrefix: "/componentdocshell/assets/",
	}
}

func componentDocsNavigation(active string) componentdocshell.Navigation {
	switch componentDocsFamily(active) {
	case "charts":
		return chartsDocsNavigation(active)
	case "icon-packs":
		return iconsDocsNavigation(active)
	case "app-shells":
		return appShellsDocsNavigation(active)
	case "examples":
		return examplesDocsNavigation(active)
	}

	sections := getSidebarSections("")
	for i := range sections {
		for j := range sections[i].Items {
			sections[i].Items[j].ID = componentDocsID(sections[i].Items[j].ID)
		}
	}
	return componentdocshell.Navigation{
		Items:         getSidebarTopItems(""),
		SectionsTitle: "Components",
		Sections:      sections,
		DisableSearch: true,
	}
}

func componentDocsSecondaryConfig(activeFamily string) navbar.SecondaryConfig {
	return navbar.SecondaryConfig{
		Links: []navbar.SecondaryLink{
			{Label: "Core", Href: "/getting-started", Current: componentDocsSecondaryCurrent("core", activeFamily), LinkAttrs: componentDocsSecondaryLinkAttrs("core")},
			{Label: "AI Agents", Href: "/docs/agents", Current: componentDocsSecondaryCurrent("agents", activeFamily), LinkAttrs: componentDocsSecondaryLinkAttrs("agents")},
			{Label: "Icons", Href: "/components/icon", Current: componentDocsSecondaryCurrent("icon-packs", activeFamily), LinkAttrs: componentDocsSecondaryLinkAttrs("icon-packs")},
			{Label: "Charts", Href: "/modules/charts", Current: componentDocsSecondaryCurrent("charts", activeFamily), LinkAttrs: componentDocsSecondaryLinkAttrs("charts")},
			{Label: "App Shells", Href: "/modules/app-shells", Current: componentDocsSecondaryCurrent("app-shells", activeFamily), LinkAttrs: componentDocsSecondaryLinkAttrs("app-shells")},
			{Label: "Examples", Href: "/examples/ticker", Current: componentDocsSecondaryCurrent("examples", activeFamily), LinkAttrs: componentDocsSecondaryLinkAttrs("examples")},
		},
		AriaLabel:  "Goshtoso documentation",
		Scrollable: true,
		RootClass:  "component-doc-shell__site-secondary-row",
		RootAttrs:  templ.Attributes{"id": "goshtoso-site-secondary-navigation"},
	}
}

func componentDocsSecondaryLinkAttrs(familyID string) templ.Attributes {
	return templ.Attributes{"data-site-secondary-family": familyID}
}

func componentDocsSecondaryCurrent(familyID, activeFamily string) navbar.SecondaryCurrent {
	if familyID == activeFamily {
		return navbar.SecondaryCurrentLocation
	}
	return navbar.SecondaryCurrentNone
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
	return componentdocshell.Page{
		Title:         meta.Title,
		DocumentTitle: meta.TitleText(),
		Description:   meta.Description,
		CanonicalURL:  meta.CanonicalURL(),
		SiteName:      SiteName,
		Locale:        "en_US",
		SocialImage: componentdocshell.SocialImage{
			URL:      meta.OGImageURL(),
			MIMEType: OGImageMIMEType,
			Width:    1200,
			Height:   630,
			Alt:      meta.OGImageAlt(),
		},
		Active:    configuredComponentDocsActive(cfg.Navigation, componentDocsID(active)),
		Content:   componentDocsContent(content, active),
		Head:      componentDocsHead(meta, componentDocsFamily(active)),
		EnableTOC: true,
	}
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
