package components

import (
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/site/internal/examples/chat"
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
}

// Demos is the single registry of every page that can be loaded either
// as a full document (initial load / direct nav / no-JS) or as an HTMX
// fragment swap into #main-content.
//
// Keys are the canonical route path *without* a leading slash, e.g.
// "components/button", "docs/theme", "getting-started". Server handlers
// translate their URL into a registry key.
var Demos = map[string]DemoEntry{
	"components/accordion":        {"Accordion", "accordion", accordionDemoContent},
	"components/alert":            {"Alert", "alert", alertDemoContent},
	"components/avatar":           {"Avatar", "avatar", avatarDemoContent},
	"components/badge":            {"Badge", "badge", badgeDemoContent},
	"components/banner":           {"Banner", "banner", bannerDemoContent},
	"components/breadcrumbs":      {"Breadcrumbs", "breadcrumbs", breadcrumbsDemoContent},
	"components/button":           {"Buttons", "button", buttonDemoContent},
	"components/card":             {"Card", "card", cardDemoContent},
	"components/carousel":         {"Carousel", "carousel", carouselDemoContent},
	"components/chatbubble":       {"Chat Bubble", "chatbubble", chatBubbleDemoContent},
	"components/checkbox":         {"Checkbox", "checkbox", checkboxDemoContent},
	"components/codeblock":        {"Code Block", "codeblock", codeBlockDemoContent},
	"components/combobox":         {"Combobox", "combobox", comboboxDemoContent},
	"components/dependencies":     {"Dependencies", "dependencies", dependenciesDemoContent},
	"components/drawer":           {"Drawer", "drawer", drawerDemoContent},
	"components/dropdown":         {"Dropdown", "dropdown", dropdownDemoContent},
	"components/fileinput":        {"File Input", "fileinput", fileInputDemoContent},
	"components/form":             {"Form", "form", formDemoContent},
	"components/kbd":              {"KBD", "kbd", kbdDemoContent},
	"components/link":             {"Link", "link", linkDemoContent},
	"components/modal":            {"Modal", "modal", modalDemoContent},
	"components/navbar":           {"Navbar", "navbar", navbarDemoContent},
	"components/pagination":       {"Pagination", "pagination", paginationDemoContent},
	"components/palette":          {"Palette", "palette", paletteDemoContent},
	"components/radio":            {"Radio", "radio", radioDemoContent},
	"components/range":            {"Range", "range", rangeDemoContent},
	"components/rating":           {"Rating", "rating", ratingDemoContent},
	"components/select":           {"Select", "select", selectDemoContent},
	"components/schema-field":     {"Schema Field", "schema-field", schemaFieldDemoContent},
	"components/search":           {"Search", "search", searchDemoContent},
	"components/sidebar":          {"Sidebar", "sidebar", sidebarDemoContent},
	"components/spinner":          {"Spinner", "spinner", spinnerDemoContent},
	"components/steps":            {"Steps", "steps", stepsDemoContent},
	"components/table":            {"Table", "table", tableDemoContent},
	"components/tabs":             {"Tabs", "tabs", tabsDemoContent},
	"components/tags-list":        {"Tags List", "tags-list", tagsListDemoContent},
	"components/text-input":       {"Text Input", "text-input", textInputDemoContent},
	"components/textarea":         {"Textarea", "textarea", textareaDemoContent},
	"components/toast":            {"Toast", "toast", toastDemoContent},
	"components/toggle":           {"Toggle", "toggle", toggleDemoContent},
	"components/structured-input": {"Structured Input", "structured-input", structuredInputDemoContent},
	"components/tooltip":          {"Tooltip", "tooltip", tooltipDemoContent},
	"docs/theme":                  {"Theme", "theme", themeDemoContent},
	"getting-started":             {"Getting Started", "", gettingStartedContent},
	"attributions":                {"Attributions", "attributions", attributionsContent},
	"license":                     {"License", "license", licenseContent},
	"privacy":                     {"Privacy Policy", "privacy", privacyContent},
	"examples":                    {"Examples", "examples", examples.IndexContent},
	"examples/todo":               {"Todo List", "todo", examples.TodoContent},
	"examples/expense":            {"Expense Tracker", "expense", examples.ExpenseContent},
	"examples/chat":               {"Chat", "chat", func() templ.Component { return examples.ChatApp(chat.NewGuest(0)) }},
	"examples/logs":               {"Live Log Feed", "logs", examples.LogsContent},
	"examples/profile":            {"Profile", "profile", examples.ProfileContent},
	"examples/ticker":             {"Live Ticker", "ticker", examples.TickerContent},
	"examples/wizard":             {"Onboarding Wizard", "wizard", examples.WizardContent},
}

// LookupDemo returns the entry for a given canonical key (no leading slash).
func LookupDemo(key string) (DemoEntry, bool) {
	e, ok := Demos[key]
	return e, ok
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
	descriptions := map[string]string{
		"components/accordion":        "Build accessible accordion interfaces in Go with server-rendered templ markup, HTMX-friendly content, and Alpine.js disclosure behavior.",
		"components/avatar":           "Render Go avatar components with images, initials, placeholders, status indicators, borders, and stacked group layouts.",
		"components/badge":            "Use Goshtoso badge components for statuses, counts, semantic colors, icon labels, and notification dots in Go applications.",
		"components/banner":           "Create inline announcement and alert banners with tones, icons, actions, and dismiss patterns for server-rendered Go UIs.",
		"components/breadcrumbs":      "Add hierarchical breadcrumb navigation to Go documentation routes, nested pages, and HTMX-enhanced server-rendered apps.",
		"components/button":           "Build Go button components with variants, sizes, icons, loading states, and HTMX request integrations.",
		"components/card":             "Compose structured card layouts in Go with media, headings, body content, actions, and responsive variants.",
		"components/carousel":         "Render carousel content rails for images, feature panels, and grouped cards with Goshtoso, templ, and Alpine.js.",
		"components/chatbubble":       "Create conversation message bubbles with alignment, avatars, metadata, and grouped thread layouts for Go chat interfaces.",
		"components/checkbox":         "Build checkbox fields and groups with accessible labels, helper text, disabled states, and validation in Go.",
		"components/codeblock":        "Show copy-pasteable code snippets with semantic pre and code markup, labels, scroll bounds, copy actions, and themes.",
		"components/combobox":         "Add searchable single and multi-select combobox controls with keyboard navigation to server-rendered Go forms.",
		"components/dependencies":     "Load Goshtoso's local CSS, Alpine.js, HTMX, and runtime helpers from embedded assets without CDN dependencies.",
		"components/drawer":           "Create slide-over panels for details, filters, and HTMX workflows with Alpine.js state and focus trapping.",
		"components/dropdown":         "Render dropdown menus with grouped actions, icons, alignment, and keyboard-friendly behavior for Go UI workflows.",
		"components/fileinput":        "Create file upload controls with labels, helper text, selected-file display, and validation states in Goshtoso.",
		"components/form":             "Build form layout patterns, field groups, server validation messages, and HTMX submit behavior with Go and templ.",
		"components/kbd":              "Display semantic keyboard shortcut hints for command palettes, forms, and toolbar controls in Go documentation.",
		"components/link":             "Render accessible link components for inline navigation, external links, icon links, and disabled states in Go UIs.",
		"components/modal":            "Create modal dialogs with focus management, actions, dismissal, and scroll handling using Goshtoso and Alpine.js.",
		"components/navbar":           "Build top navigation bars with links, actions, responsive menus, and brand areas for server-rendered Go sites.",
		"components/pagination":       "Add pagination controls for tables, lists, and HTMX-powered result sets in Go applications.",
		"components/palette":          "Render color picking surfaces with swatches, hex entry, token labels, and theme integration in Goshtoso.",
		"components/radio":            "Build radio groups and segmented controls with accessible labels, helper copy, and validation states in Go.",
		"components/range":            "Create range sliders with labels, helper text, value output, and accessible input behavior for Go forms.",
		"components/rating":           "Render rating inputs and display patterns with stars, labels, accessible values, and form integration in Goshtoso.",
		"components/search":           "Add command-palette search with a sidebar trigger, global modal, filtering, highlighting, and docs navigation.",
		"components/select":           "Build native-feeling select menus with trigger content, grouped options, and keyboard support in Go.",
		"components/schema-field":     "Generate form controls from JSON Schema defaults, values, and allow-list rules in server-rendered Go interfaces.",
		"components/sidebar":          "Create persistent or overlay side navigation with sections, search slots, active states, and responsive behavior.",
		"components/spinner":          "Show loading indicators with size, color, label, and accessible busy-state patterns for Go and HTMX UIs.",
		"components/steps":            "Render progress indicators for wizards, onboarding, and multi-step workflows in server-rendered Go applications.",
		"components/structured-input": "Compose inputs for prefixes, suffixes, segmented values, metadata rows, and constrained data entry in Go.",
		"components/table":            "Build sortable, paginated, filterable Go data tables with HTMX loading, rich cells, and server-rendered rows.",
		"components/tabs":             "Create segmented content navigation with keyboard support, panels, active states, and HTMX-loaded tab content.",
		"components/tags-list":        "Manage editable tag collections with add, remove, duplicate handling, and keyboard flows in Goshtoso.",
		"components/text-input":       "Render text fields with icons, search affordances, masks, validation, and password controls for Go forms.",
		"components/textarea":         "Build multi-line text entry with resize behavior, helper text, counters, and validation in server-rendered Go UIs.",
		"components/toast":            "Create transient notifications with positions, timing, actions, and status variants, including HTMX OOB toasts.",
		"components/toggle":           "Render binary switches for settings, feature flags, and compact on-off choices in Go applications.",
		"components/tooltip":          "Add contextual hints for icons, controls, and abbreviated UI labels with hover, focus, or click triggers.",
		"docs/theme":                  "Customize Goshtoso themes with Tailwind CSS tokens, dark mode, live previews, and server-rendered component examples.",
		"getting-started":             "Start a Go HTMX app with Goshtoso, templ, Tailwind CSS, local runtime assets, and copy-pasteable setup code.",
		"attributions":                "Review third-party licenses and asset attributions for the Goshtoso documentation site and component library.",
		"license":                     "Read the Goshtoso license terms for using the Go UI component library in personal and commercial projects.",
		"privacy":                     "Understand how the Goshtoso demo site uses browser storage for preferences and local example state without analytics.",
		"examples":                    "Explore full runnable Go apps built with Goshtoso components, HTMX fragments, Alpine.js state, and server-rendered templ pages.",
		"examples/todo":               "Run a cookie-backed HTMX todo list example built with Goshtoso Go components and server-rendered fragments.",
		"examples/expense":            "Track expenses in a cookie-backed HTMX example with search, category filters, pagination, a running total, and Goshtoso combobox, modal, and toast components.",
		"examples/logs":               "Explore a live log feed example using Go, SSE, htmx-ext-sse, filters, and Goshtoso interface components.",
		"examples/profile":            "Edit a profile settings screen composed from Goshtoso form, avatar, modal, toggle, and validation components.",
		"examples/ticker":             "Watch a live ticker table update with Go server-sent events, HTMX row swaps, and Goshtoso table components.",
		"examples/chat":               "Try a real-time Go chat interface with WebSockets, server-rendered message bubbles, and Goshtoso components.",
		"examples/wizard":             "Step through a cookie-backed onboarding wizard with per-step server-side validation, HTMX fragment swaps, and Goshtoso steps, form, and toast components.",
	}
	if description, ok := descriptions[key]; ok {
		return description
	}
	return entry.Title + " documentation for building interactive server-rendered Go interfaces with Goshtoso, templ, HTMX, Alpine.js, and Tailwind CSS."
}
