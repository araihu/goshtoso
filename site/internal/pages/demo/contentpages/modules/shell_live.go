package modulespages

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso-app-shells/consoleshell"
	"github.com/araihu/goshtoso-app-shells/landingshell"
	"github.com/araihu/goshtoso/components/sidebar"
)

// ShellLive serves independent documents so each shell owns its viewport and runtime.
func ShellLive(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/modules/app-shells/live/"), "/")
	if len(parts) != 2 || (parts[1] != "overview" && parts[1] != "activity") {
		http.NotFound(w, r)
		return
	}
	family, active := parts[0], parts[1]
	base := "/modules/app-shells/live/" + family + "/"
	title := "Overview"
	if active == "activity" {
		title = "Activity"
	}
	nav := []sidebar.Item{{ID: "overview", Label: "Overview", Href: base + "overview"}, {ID: "activity", Label: "Activity", Href: base + "activity"}}
	partial := r.Header.Get("HX-Request-Type") == "partial"
	var view templ.Component
	switch family {
	case "componentdocshell":
		cfg := componentdocshell.Config{Brand: componentdocshell.Brand{Name: "Atlas Docs", HomeURL: base + "overview"}, Navigation: componentdocshell.Navigation{Items: nav}, Interactions: componentdocshell.InteractionConfig{EnableHTMX: true, LocalRuntime: true}, Appearance: componentdocshell.AppearanceConfig{DefaultTheme: "araihu"}}
		page := componentdocshell.Page{Title: title, Active: active, Content: shellDocsContent(active), EnableTOC: true}
		view = componentdocshell.Layout(cfg, page)
		if partial {
			view = componentdocshell.Fragment(cfg, page)
		}
	case "consoleshell":
		cfg := consoleshell.Config{Brand: consoleshell.Brand{Name: "Atlas Cloud", HomeURL: base + "overview"}, Navigation: consoleshell.Navigation{Items: nav}, Interactions: consoleshell.InteractionConfig{EnableHTMX: true, NavigationOOB: true, LocalRuntime: true}, Appearance: consoleshell.AppearanceConfig{DefaultTheme: "araihu"}}
		page := consoleshell.Page{Title: title, Active: active, Content: shellConsoleContent(active)}
		view = consoleshell.Layout(cfg, page)
		if partial {
			view = consoleshell.Fragment(cfg, page)
		}
	case "landingshell":
		cfg := landingshell.Config{Brand: landingshell.Brand{Name: "Atlas", HomeURL: base + "overview", Tagline: "A home for your next idea"}, Navigation: []landingshell.Link{{Label: "Overview", Href: base + "overview"}, {Label: "Activity", Href: base + "activity", Primary: true}}, Appearance: landingshell.AppearanceConfig{DefaultTheme: "araihu"}, Interactions: landingshell.InteractionConfig{LocalRuntime: true}, MobileNavigation: &landingshell.MobileNavigationConfig{Title: "Explore Atlas"}, Footer: landingshell.Footer{Name: "Atlas", Meta: []string{"Built for thoughtful teams"}, Links: []landingshell.Link{{Label: "Overview", Href: base + "overview"}, {Label: "Activity", Href: base + "activity"}}}}
		view = landingshell.Layout(cfg, landingshell.Page{Title: title, Hero: shellLandingHero(base, active), Content: shellLandingContent()})
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex")
	if err := view.Render(r.Context(), w); err != nil {
		http.Error(w, "Could not render example", http.StatusInternalServerError)
	}
}
