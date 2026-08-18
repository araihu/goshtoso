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
	return `cfg := componentdocshell.Config{
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

func appShellsConsoleShellCode() string {
	return `cfg := consoleshell.Config{
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

func appShellsLandingShellCode() string {
	return `cfg := landingshell.Config{
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
