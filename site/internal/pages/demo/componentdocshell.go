package demo

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
)

// ComponentDocsLayout delegates the demo frame to the reusable component
// documentation shell while retaining Goshtoso-owned routes and content.
func ComponentDocsLayout(meta PageMeta, active string, content templ.Component, persist bool) templ.Component {
	return componentdocshell.Layout(componentDocsConfig(persist), componentDocsPage(meta, active, content))
}

func ComponentDocsFragment(meta PageMeta, active string, content templ.Component, persist bool) templ.Component {
	return componentdocshell.Fragment(componentDocsConfig(persist), componentDocsPage(meta, active, content))
}

func componentDocsConfig(persist bool) componentdocshell.Config {
	sections := getSidebarSections("")
	for i := range sections {
		for j := range sections[i].Items {
			sections[i].Items[j].ID = componentDocsID(sections[i].Items[j].ID)
		}
	}
	return componentdocshell.Config{
		Brand:        componentdocshell.Brand{Name: SiteName, HomeURL: "/", Logo: templ.Raw(`<img src="/assets/images/goshtoso-mark.svg" alt="" aria-hidden="true" class="size-8 rounded-lg">`), FaviconURL: "/favicon.svg"},
		Navigation:   componentdocshell.Navigation{Items: getSidebarTopItems(""), SectionsTitle: "Components", Sections: sections, SearchPlaceholder: "Search"},
		Appearance:   componentdocshell.AppearanceConfig{Themes: getThemeOptions(), DefaultTheme: "goshtoso", PersistPreferences: persist},
		Interactions: componentdocshell.InteractionConfig{EnableHTMX: true, LocalRuntime: true}, RepositoryURL: "https://github.com/araihu/goshtoso", AssetPrefix: "/componentdocshell/assets/",
	}
}

func componentDocsPage(meta PageMeta, active string, content templ.Component) componentdocshell.Page {
	return componentdocshell.Page{Title: meta.Title, DocumentTitle: meta.TitleText(), Description: meta.Description, CanonicalURL: meta.CanonicalURL(), Active: componentDocsID(active), Content: componentDocsContent(content, active), Head: HeadMeta(meta), EnableTOC: true}
}

func componentDocsID(active string) string {
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
