package components

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/site/internal/examples/chat"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	accordionpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/accordion"
	actiongrouppage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/actiongroup"
	alertpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/alert"
	appshellpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/appshell"
	avatarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/avatar"
	badgepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/badge"
	bannerpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/banner"
	breadcrumbspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/breadcrumbs"
	buttonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/button"
	cardpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/card"
	carouselpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/carousel"
	chatbubblepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/chatbubble"
	checkboxpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/checkbox"
	codeblockpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/codeblock"
	comboboxpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/combobox"
	drawerpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/drawer"
	dropdownpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/dropdown"
	emptystatepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/emptystate"
	fileinputpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/fileinput"
	formpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/form"
	headpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/head"
	iconpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/icon"
	kbdpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/kbd"
	linkpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/link"
	modalpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/modal"
	navbarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/navbar"
	pageheaderpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/pageheader"
	paginationpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/pagination"
	palettepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/palette"
	panelpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/panel"
	radiopage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/radio"
	rangepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/range"
	ratingpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/rating"
	schemaformpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/schemaform"
	searchpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/search"
	selectpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/select"
	sidebarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/sidebar"
	skeletonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/skeleton"
	spinnerpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/spinner"
	stepspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/steps"
	structuredinputpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/structuredinput"
	tablepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/table"
	tabspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/tabs"
	tagslistpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/tagslist"
	textareapage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/textarea"
	textinputpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/textinput"
	toastpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/toast"
	togglepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/toggle"
	toolbarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/toolbar"
	tooltippage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/tooltip"
	"github.com/araihu/goshtoso/site/internal/pages/demo/examples"
	demoregistry "github.com/araihu/goshtoso/site/internal/pages/demo/registry"
)

// DemoEntry describes a single navigable demo page: its title for <title>
// and the sidebar "active" key used by Layout/sidebar to highlight the
// current entry.
type DemoEntry = demo.PageDefinition

// Demos is the single registry of every page that can be loaded either
// as a full document (initial load / direct nav / no-JS) or as an HTMX
// fragment swap into #main-content.
//
// Keys are the canonical route path *without* a leading slash, e.g.
// "components/button", "docs/theme", "getting-started". Server handlers
// translate their URL into a registry key.
var Demos = legacyDemos()

var defaultRegistry = mustLegacyRegistry(Demos)

func legacyDemos() map[string]DemoEntry {
	demos := map[string]DemoEntry{
		"components/app-shell":        appshellpage.Definition,
		"components/page-header":      pageheaderpage.Definition,
		"components/toolbar":          toolbarpage.Definition,
		"components/action-group":     actiongrouppage.Definition,
		"components/panel":            panelpage.Definition,
		"components/empty-state":      emptystatepage.Definition,
		"components/skeleton":         skeletonpage.Definition,
		"components/accordion":        accordionpage.Definition,
		"components/alert":            alertpage.Definition,
		"components/avatar":           avatarpage.Definition,
		"components/badge":            badgepage.Definition,
		"components/banner":           bannerpage.Definition,
		"components/breadcrumbs":      breadcrumbspage.Definition,
		"components/button":           buttonpage.Definition,
		"components/card":             cardpage.Definition,
		"components/carousel":         carouselpage.Definition,
		"components/chatbubble":       chatbubblepage.Definition,
		"components/checkbox":         checkboxpage.Definition,
		"components/codeblock":        codeblockpage.Definition,
		"components/combobox":         comboboxpage.Definition,
		"components/dependencies":     headpage.Definition,
		"components/drawer":           drawerpage.Definition,
		"components/dropdown":         dropdownpage.Definition,
		"components/fileinput":        fileinputpage.Definition,
		"components/form":             formpage.Definition,
		"components/kbd":              kbdpage.Definition,
		"components/link":             linkpage.Definition,
		"components/modal":            modalpage.Definition,
		"components/navbar":           navbarpage.Definition,
		"components/pagination":       paginationpage.Definition,
		"components/palette":          palettepage.Definition,
		"components/radio":            radiopage.Definition,
		"components/range":            rangepage.Definition,
		"components/rating":           ratingpage.Definition,
		"components/select":           selectpage.Definition,
		"components/schema-form":      schemaformpage.Definition,
		"components/search":           searchpage.Definition,
		"components/sidebar":          sidebarpage.Definition,
		"components/spinner":          spinnerpage.Definition,
		"components/steps":            stepspage.Definition,
		"components/table":            tablepage.Definition,
		"components/tabs":             tabspage.Definition,
		"components/tags-list":        tagslistpage.Definition,
		"components/text-input":       textinputpage.Definition,
		"components/textarea":         textareapage.Definition,
		"components/toast":            toastpage.Definition,
		"components/toggle":           togglepage.Definition,
		"components/structured-input": structuredinputpage.Definition,
		"components/tooltip":          tooltippage.Definition,
		"docs/agents":                 {Title: "AI Agents", Active: "agents", Content: agentsContent},
		"docs/application-patterns":   {Title: "Application Patterns", Active: "application-patterns", Content: applicationPatternsContent},
		"docs/component-model":        {Title: "Component Model", Active: "component-model", Content: componentModelContent},
		"docs/theme":                  {Title: "Theme", Active: "theme", Content: themeDemoContent},
		"modules/charts":              {Title: "Charts", Active: "module-charts", Content: chartsModuleContent},
		"modules/app-shells":          {Title: "App Shells", Active: "module-app-shells", Content: appShellsModuleContent},
		"getting-started":             {Title: "Getting Started", Content: gettingStartedContent},
		"attributions":                {Title: "Attributions", Active: "attributions", Content: attributionsContent},
		"license":                     {Title: "License", Active: "license", Content: licenseContent},
		"privacy":                     {Title: "Privacy Policy", Active: "privacy", Content: privacyContent},
		"examples":                    {Title: "Examples", Active: "examples", Content: examples.IndexContent},
		"examples/todo":               {Title: "Todo List", Active: "todo", Content: examples.TodoContent},
		"examples/expense":            {Title: "Expense Tracker", Active: "expense", Content: examples.ExpenseContent},
		"examples/chat":               {Title: "Chat", Active: "chat", Content: func() templ.Component { return examples.ChatApp(chat.NewGuest(0)) }},
		"examples/logs":               {Title: "Live Log Feed", Active: "logs", Content: examples.LogsContent},
		"examples/profile":            {Title: "Profile", Active: "profile", Content: examples.ProfileContent},
		"examples/ticker":             {Title: "Live Ticker", Active: "ticker", Content: examples.TickerContent},
		"examples/wizard":             {Title: "Onboarding Wizard", Active: "wizard", Content: examples.WizardContent},
	}
	if _, ok := catalog.Lookup("components/icon"); ok {
		demos["components/icon"] = iconpage.Definition
	}
	return demos
}

func mustLegacyRegistry(entries map[string]DemoEntry) *demoregistry.Registry {
	definitions := make([]demo.PageDefinition, 0, len(entries))
	for key, entry := range entries {
		entry.Key = key
		entry.Description = demoDescription(key, entry)
		definitions = append(definitions, entry)
	}
	registry, err := demoregistry.New(definitions, catalog.ComponentPages())
	if err != nil {
		panic(err)
	}
	return registry
}

// LookupDemo returns the entry for a given canonical key (no leading slash).
func LookupDemo(key string) (DemoEntry, bool) {
	return defaultRegistry.Lookup(key)
}

// componentCount uses registered component pages so the homepage reflects the
// public catalog rather than package directories that may not be user-facing.
func componentCount() int {
	return len(catalog.ComponentPages())
}

func DemoMeta(key string, entry DemoEntry) demo.PageMeta {
	entry.Key = key
	entry.Description = demoDescription(key, entry)
	return demoregistry.MetaForDefinition(entry)
}

func MetaForKey(key string) demo.PageMeta {
	return defaultRegistry.MetaForKey(key)
}

func AllPublicMeta() []demo.PageMeta {
	return defaultRegistry.AllPublicMeta()
}

func demoDescription(key string, entry DemoEntry) string {
	if component, ok := catalog.Lookup(key); ok {
		return component.Description
	}

	descriptions := map[string]string{
		"docs/agents":               "Install the Goshtoso consumer agent skill for AI coding tools and verify npx skills distribution.",
		"docs/application-patterns": "Compose App Shell, Operations List, Detail Workspace, and Multi-step Workflow product surfaces from server-rendered Goshtoso components.",
		"modules/charts":            "Explore static, interactive, and interactive 3D chart components from the optional Goshtoso Charts module.",
		"modules/app-shells":        "Explore reusable documentation and console application shells from the optional Goshtoso App Shells module.",
		"docs/component-model":      "Understand Goshtoso's common component interface, concrete return values, constructor styles, stable Kind identity, and rendered defaults.",
		"docs/theme":                "Customize Goshtoso themes with Tailwind CSS tokens, dark mode, live previews, and server-rendered component examples.",
		"getting-started":           "Start a Go HTMX app with Goshtoso, templ, Tailwind CSS, local runtime assets, and copy-pasteable setup code.",
		"attributions":              "Review third-party licenses and asset attributions for the Goshtoso documentation site and component library.",
		"license":                   "Read the Goshtoso license terms for using the Go UI component library in personal and commercial projects.",
		"privacy":                   "Understand how the Goshtoso demo site uses browser storage for preferences and local example state without analytics.",
		"examples":                  "Explore full runnable Go apps built with Goshtoso components, HTMX fragments, Alpine.js state, and server-rendered templ pages.",
		"examples/todo":             "Run a cookie-backed HTMX todo list example built with Goshtoso Go components and server-rendered fragments.",
		"examples/expense":          "Track expenses in a cookie-backed HTMX example with search, category filters, pagination, a running total, and Goshtoso combobox, modal, and toast components.",
		"examples/logs":             "Explore a live log feed example using Go, SSE, htmx-ext-sse, filters, and Goshtoso interface components.",
		"examples/profile":          "Edit a profile settings screen composed from Goshtoso form, avatar, modal, toggle, and validation components.",
		"examples/ticker":           "Watch a live ticker table update with Go server-sent events, HTMX row swaps, and Goshtoso table components.",
		"examples/chat":             "Try a real-time Go chat interface with WebSockets, server-rendered message bubbles, and Goshtoso components.",
		"examples/wizard":           "Step through a cookie-backed onboarding wizard with per-step server-side validation, HTMX fragment swaps, and Goshtoso steps, form, and toast components.",
	}
	if description, ok := descriptions[key]; ok {
		return description
	}
	return entry.Title + " documentation for building interactive server-rendered Go interfaces with Goshtoso, templ, HTMX, Alpine.js, and Tailwind CSS."
}
