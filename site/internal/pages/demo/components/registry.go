package components

import (
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/site/internal/examples/chat"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/araihu/goshtoso/site/internal/pages/demo/examples"
)

// DemoEntry describes a single navigable demo page: its title for <title>
// and the sidebar "active" key used by Layout/sidebar to highlight the
// current entry.
type DemoEntry struct {
	Title   string
	Active  string
	Content func() templ.Component
	API     []demo.APISection
}

// Demos is the single registry of every page that can be loaded either
// as a full document (initial load / direct nav / no-JS) or as an HTMX
// fragment swap into #main-content.
//
// Keys are the canonical route path *without* a leading slash, e.g.
// "components/button", "docs/theme", "getting-started". Server handlers
// translate their URL into a registry key.
var Demos = map[string]DemoEntry{
	"components/accordion":        {Title: "Accordion", Active: "accordion", Content: accordionDemoContent, API: accordionAPISections},
	"components/alert":            {Title: "Alert", Active: "alert", Content: alertDemoContent, API: alertAPISections},
	"components/avatar":           {Title: "Avatar", Active: "avatar", Content: avatarDemoContent, API: avatarAPISections},
	"components/badge":            {Title: "Badge", Active: "badge", Content: badgeDemoContent, API: badgeAPISections},
	"components/banner":           {Title: "Banner", Active: "banner", Content: bannerDemoContent, API: bannerAPISections},
	"components/breadcrumbs":      {Title: "Breadcrumbs", Active: "breadcrumbs", Content: breadcrumbsDemoContent, API: breadcrumbsAPISections},
	"components/button":           {Title: "Buttons", Active: "button", Content: buttonDemoContent, API: buttonAPISections},
	"components/card":             {Title: "Card", Active: "card", Content: cardDemoContent, API: cardAPISections},
	"components/carousel":         {Title: "Carousel", Active: "carousel", Content: carouselDemoContent, API: carouselAPISections},
	"components/chatbubble":       {Title: "Chat Bubble", Active: "chatbubble", Content: chatBubbleDemoContent, API: chatBubbleAPISections},
	"components/checkbox":         {Title: "Checkbox", Active: "checkbox", Content: checkboxDemoContent, API: checkboxAPISections},
	"components/codeblock":        {Title: "Code Block", Active: "codeblock", Content: codeBlockDemoContent, API: codeBlockAPISections},
	"components/combobox":         {Title: "Combobox", Active: "combobox", Content: comboboxDemoContent, API: comboboxAPISections},
	"components/dependencies":     {Title: "Dependencies", Active: "dependencies", Content: dependenciesDemoContent, API: dependenciesAPISections},
	"components/drawer":           {Title: "Drawer", Active: "drawer", Content: drawerDemoContent, API: drawerAPISections},
	"components/dropdown":         {Title: "Dropdown", Active: "dropdown", Content: dropdownDemoContent, API: dropdownAPISections},
	"components/fileinput":        {Title: "File Input", Active: "fileinput", Content: fileInputDemoContent, API: fileInputAPISections},
	"components/form":             {Title: "Form", Active: "form", Content: formDemoContent, API: formAPISections},
	"components/kbd":              {Title: "KBD", Active: "kbd", Content: kbdDemoContent, API: kbdAPISections},
	"components/link":             {Title: "Link", Active: "link", Content: linkDemoContent, API: linkAPISections},
	"components/modal":            {Title: "Modal", Active: "modal", Content: modalDemoContent, API: modalAPISections},
	"components/navbar":           {Title: "Navbar", Active: "navbar", Content: navbarDemoContent, API: navbarAPISections},
	"components/pagination":       {Title: "Pagination", Active: "pagination", Content: paginationDemoContent, API: paginationAPISections},
	"components/palette":          {Title: "Palette", Active: "palette", Content: paletteDemoContent, API: paletteAPISections},
	"components/radio":            {Title: "Radio", Active: "radio", Content: radioDemoContent, API: radioAPISections},
	"components/range":            {Title: "Range", Active: "range", Content: rangeDemoContent, API: rangeAPISections},
	"components/rating":           {Title: "Rating", Active: "rating", Content: ratingDemoContent, API: ratingAPISections},
	"components/select":           {Title: "Select", Active: "select", Content: selectDemoContent, API: selectAPISections},
	"components/schema-form":      {Title: "Schema Form", Active: "schema-form", Content: schemaFormDemoContent, API: schemaFormAPISections},
	"components/search":           {Title: "Search", Active: "search", Content: searchDemoContent, API: searchAPISections},
	"components/sidebar":          {Title: "Sidebar", Active: "sidebar", Content: sidebarDemoContent, API: sidebarAPISections},
	"components/spinner":          {Title: "Spinner", Active: "spinner", Content: spinnerDemoContent, API: spinnerAPISections},
	"components/steps":            {Title: "Steps", Active: "steps", Content: stepsDemoContent, API: stepsAPISections},
	"components/table":            {Title: "Table", Active: "table", Content: tableDemoContent, API: tableAPISections},
	"components/tabs":             {Title: "Tabs", Active: "tabs", Content: tabsDemoContent, API: tabsAPISections},
	"components/tags-list":        {Title: "Tags List", Active: "tags-list", Content: tagsListDemoContent, API: tagsListAPISections},
	"components/text-input":       {Title: "Text Input", Active: "text-input", Content: textInputDemoContent, API: textInputAPISections},
	"components/textarea":         {Title: "Textarea", Active: "textarea", Content: textareaDemoContent, API: textareaAPISections},
	"components/toast":            {Title: "Toast", Active: "toast", Content: toastDemoContent, API: toastAPISections},
	"components/toggle":           {Title: "Toggle", Active: "toggle", Content: toggleDemoContent, API: toggleAPISections},
	"components/structured-input": {Title: "Structured Input", Active: "structured-input", Content: structuredInputDemoContent, API: structuredInputAPISections},
	"components/tooltip":          {Title: "Tooltip", Active: "tooltip", Content: tooltipDemoContent, API: tooltipAPISections},
	"docs/agents":                 {Title: "AI Agents", Active: "agents", Content: agentsContent},
	"docs/component-model":        {Title: "Component Model", Active: "component-model", Content: componentModelContent},
	"docs/theme":                  {Title: "Theme", Active: "theme", Content: themeDemoContent},
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

// LookupDemo returns the entry for a given canonical key (no leading slash).
func LookupDemo(key string) (DemoEntry, bool) {
	e, ok := Demos[key]
	return e, ok
}

// componentCount uses registered component pages so the homepage reflects the
// public catalog rather than package directories that may not be user-facing.
func componentCount() int {
	return len(catalog.ComponentPages())
}

func DemoMeta(key string, entry DemoEntry) demo.PageMeta {
	path := "/" + strings.TrimPrefix(key, "/")
	if key == "" {
		path = "/"
	}
	title := entry.Title
	metaType := "TechArticle"
	switch {
	case strings.HasPrefix(key, "components/"):
		componentTitle := entry.Title
		title = componentTitle + " Component - Goshtoso UI Library for Go"
	case strings.HasPrefix(key, "examples/"):
		title = entry.Title + " Example - Goshtoso Go UI Components"
		metaType = "SoftwareSourceCode"
	case key == "examples":
		title = "Example Apps - Goshtoso Go UI Components"
	case key == "getting-started":
		title = "Getting Started with Goshtoso Go UI Components"
	case key == "docs/agents":
		title = "Using Goshtoso With AI Agents"
	case key == "docs/component-model":
		title = "Goshtoso Component Model"
	case key == "docs/theme":
		title = "Themes - Goshtoso UI Library for Go"
	}
	return demo.PageMeta{
		Title:       title,
		Description: demoDescription(key, entry),
		Path:        path,
		Type:        metaType,
	}
}

func MetaForKey(key string) demo.PageMeta {
	entry, ok := LookupDemo(key)
	if !ok {
		return demo.DefaultMeta("Goshtoso")
	}
	return DemoMeta(key, entry)
}

func AllPublicMeta() []demo.PageMeta {
	pages := make([]demo.PageMeta, 0, len(Demos)+1)
	pages = append(pages, demo.HomeMeta())
	keys := make([]string, 0, len(Demos))
	for key := range Demos {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		pages = append(pages, MetaForKey(key))
	}
	return pages
}

func demoDescription(key string, entry DemoEntry) string {
	if component, ok := catalog.Lookup(key); ok {
		return component.Description
	}

	descriptions := map[string]string{
		"docs/agents":          "Install the Goshtoso consumer agent skill for AI coding tools and verify npx skills distribution.",
		"docs/component-model": "Understand Goshtoso themes, primitives, stable Kind identity, configuration dimensions, and component boundaries.",
		"docs/theme":           "Customize Goshtoso themes with Tailwind CSS tokens, dark mode, live previews, and server-rendered component examples.",
		"getting-started":      "Start a Go HTMX app with Goshtoso, templ, Tailwind CSS, local runtime assets, and copy-pasteable setup code.",
		"attributions":         "Review third-party licenses and asset attributions for the Goshtoso documentation site and component library.",
		"license":              "Read the Goshtoso license terms for using the Go UI component library in personal and commercial projects.",
		"privacy":              "Understand how the Goshtoso demo site uses browser storage for preferences and local example state without analytics.",
		"examples":             "Explore full runnable Go apps built with Goshtoso components, HTMX fragments, Alpine.js state, and server-rendered templ pages.",
		"examples/todo":        "Run a cookie-backed HTMX todo list example built with Goshtoso Go components and server-rendered fragments.",
		"examples/expense":     "Track expenses in a cookie-backed HTMX example with search, category filters, pagination, a running total, and Goshtoso combobox, modal, and toast components.",
		"examples/logs":        "Explore a live log feed example using Go, SSE, htmx-ext-sse, filters, and Goshtoso interface components.",
		"examples/profile":     "Edit a profile settings screen composed from Goshtoso form, avatar, modal, toggle, and validation components.",
		"examples/ticker":      "Watch a live ticker table update with Go server-sent events, HTMX row swaps, and Goshtoso table components.",
		"examples/chat":        "Try a real-time Go chat interface with WebSockets, server-rendered message bubbles, and Goshtoso components.",
		"examples/wizard":      "Step through a cookie-backed onboarding wizard with per-step server-side validation, HTMX fragment swaps, and Goshtoso steps, form, and toast components.",
	}
	if description, ok := descriptions[key]; ok {
		return description
	}
	return entry.Title + " documentation for building interactive server-rendered Go interfaces with Goshtoso, templ, HTMX, Alpine.js, and Tailwind CSS."
}
