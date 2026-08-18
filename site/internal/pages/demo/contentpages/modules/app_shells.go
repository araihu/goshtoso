package modulespages

const appShellsVersion = "v0.1.6"

func appShellsComponentPageCode() string {
	return `@componentpage.Page(componentpage.Config{
    Title:       "Button",
    Description: "Trigger an action with a clear accessible label.",
    Primary: componentpage.Example{
        PreviewLabel: "Default",
        Preview:      button.Button(),
        Code:         "@button.Button() { Continue }",
    },
})`
}

func appShellsComponentPageStatesCode() string {
	return `Sections: []componentpage.Example{{
    Title:       "Disabled",
    Description: "Keep unavailable states explicit in the reference page.",
    Preview:     button.Button(button.Disabled()),
    Code:        "@button.Button(button.Disabled()) { Continue }",
}}`
}

func appShellsComponentDocsShellCode() string {
	return `import (
    "net/http"

    "github.com/araihu/goshtoso/assets"
    "github.com/araihu/goshtoso/components/sidebar"
    "github.com/araihu/goshtoso-app-shells/componentdocshell"
    shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
)

func registerAssets(mux *http.ServeMux) {
    mux.Handle("GET /assets/", assets.Handler())
    mux.Handle("GET /componentdocshell/assets/", shellassets.Handler())
}

cfg := componentdocshell.Config{
    Brand: componentdocshell.Brand{
        Name: "Product docs",
        HomeURL: "/",
    },
    Navigation: componentdocshell.Navigation{
        Items: []sidebar.Item{{
            ID: "overview", Label: "Overview", Href: "/",
        }},
    },
    Interactions: componentdocshell.InteractionConfig{
        EnableHTMX: true,
        LocalRuntime: true,
    },
}
page := componentdocshell.Page{
    Title: "Overview",
    Active: "overview",
    Content: overviewContent(),
}
view := componentdocshell.Layout(cfg, page)`
}

func appShellsComponentDocsShellFragmentCode() string {
	return `func renderDocs(
    writer http.ResponseWriter,
    request *http.Request,
    cfg componentdocshell.Config,
    page componentdocshell.Page,
) {
    view := componentdocshell.Layout(cfg, page)
    if request.Header.Get("HX-Request") == "true" {
        view = componentdocshell.Fragment(cfg, page)
    }
    if err := view.Render(request.Context(), writer); err != nil {
        http.Error(writer, err.Error(), http.StatusInternalServerError)
    }
}`
}

func appShellsConsoleShellCode() string {
	return `import (
    "net/http"

    "github.com/araihu/goshtoso/assets"
    "github.com/araihu/goshtoso/components/sidebar"
    "github.com/araihu/goshtoso-app-shells/consoleshell"
    shellassets "github.com/araihu/goshtoso-app-shells/consoleshell/assets"
)

func registerAssets(mux *http.ServeMux) {
    mux.Handle("GET /assets/", assets.Handler())
    mux.Handle("GET /consoleshell/assets/", shellassets.Handler())
}

cfg := consoleshell.Config{
    Brand: consoleshell.Brand{Name: "Operations", HomeURL: "/"},
    Navigation: consoleshell.Navigation{
        Items: []sidebar.Item{{
            ID: "runs", Label: "Runs", Href: "/runs",
        }},
    },
    Interactions: consoleshell.InteractionConfig{
        EnableHTMX: true,
        NavigationOOB: true,
        LocalRuntime: true,
    },
}
page := consoleshell.Page{
    Title: "Runs",
    Active: "runs",
    Content: runsPage(),
}
view := consoleshell.Layout(cfg, page)`
}

func appShellsConsoleShellFragmentCode() string {
	return `func renderConsole(
    writer http.ResponseWriter,
    request *http.Request,
    cfg consoleshell.Config,
    page consoleshell.Page,
) {
    view := consoleshell.Layout(cfg, page)
    if request.Header.Get("HX-Request") == "true" {
        view = consoleshell.Fragment(cfg, page)
    }
    if err := view.Render(request.Context(), writer); err != nil {
        http.Error(writer, err.Error(), http.StatusInternalServerError)
    }
}`
}

func appShellsLandingShellCode() string {
	return `import (
    "net/http"

    "github.com/araihu/goshtoso/assets"
    "github.com/araihu/goshtoso-app-shells/landingshell"
    shellassets "github.com/araihu/goshtoso-app-shells/landingshell/assets"
)

func registerAssets(mux *http.ServeMux) {
    mux.Handle("GET /assets/", assets.Handler())
    mux.Handle("GET /landingshell/assets/", shellassets.Handler())
}

cfg := landingshell.Config{
    Brand: landingshell.Brand{
        Name: "Product",
        HomeURL: "/",
        Tagline: "server-rendered Go UI",
    },
    Navigation: []landingshell.Link{{
        Label: "Docs", Href: "/docs", Primary: true,
    }},
    Appearance: landingshell.AppearanceConfig{
        DefaultTheme: "araihu",
        InitialColorScheme: landingshell.ColorSchemeSystem,
        PersistPreferences: true,
    },
}
page := landingshell.Page{
    Title: "Home",
    Description: "Product description",
    Hero: hero(),
    Content: content(),
}
view := landingshell.Layout(cfg, page)`
}
