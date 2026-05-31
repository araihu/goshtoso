package components

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
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
	"components/accordion":       {"Accordion", "accordion", accordionDemoContent},
	"components/alert":           {"Alert", "alert", alertDemoContent},
	"components/avatar":          {"Avatar", "avatar", avatarDemoContent},
	"components/badge":           {"Badge", "badge", badgeDemoContent},
	"components/banner":          {"Banner", "banner", bannerDemoContent},
	"components/breadcrumbs":     {"Breadcrumbs", "breadcrumbs", breadcrumbsDemoContent},
	"components/button":          {"Buttons", "button", buttonDemoContent},
	"components/card":            {"Card", "card", cardDemoContent},
	"components/carousel":        {"Carousel", "carousel", carouselDemoContent},
	"components/checkbox":        {"Checkbox", "checkbox", checkboxDemoContent},
	"components/codeblock":       {"Code Block", "codeblock", codeBlockDemoContent},
	"components/combobox":        {"Combobox", "combobox", comboboxDemoContent},
	"components/combobox-new":    {"Combobox (HTMX SSR rewrite)", "combobox-new", comboboxNewDemoContent},
	"components/dropdown":        {"Dropdown", "dropdown", dropdownDemoContent},
	"components/fileinput":       {"File Input", "fileinput", fileInputDemoContent},
	"components/form":            {"Form", "form", formDemoContent},
	"components/form-validation": {"Form Validation", "form-validation", formValidationDemoContent},
	"components/key-value":       {"Key Value", "key-value", keyValueDemoContent},
	"components/modal":           {"Modal", "modal", modalDemoContent},
	"components/navbar":          {"Navbar", "navbar", navbarDemoContent},
	"components/pagination":      {"Pagination", "pagination", paginationDemoContent},
	"components/palette":         {"Palette", "palette", paletteDemoContent},
	"components/radio":           {"Radio", "radio", radioDemoContent},
	"components/select":          {"Select", "select", selectDemoContent},
	"components/sidebar":         {"Sidebar", "sidebar", sidebarDemoContent},
	"components/spinner":         {"Spinner", "spinner", spinnerDemoContent},
	"components/steps":           {"Steps", "steps", stepsDemoContent},
	"components/table":           {"Table", "table", tableDemoContent},
	"components/tabs":            {"Tabs", "tabs", tabsDemoContent},
	"components/tags-list":       {"Tags List", "tags-list", tagsListDemoContent},
	"components/text-input":      {"Text Input", "text-input", textInputDemoContent},
	"components/textarea":        {"Textarea", "textarea", textareaDemoContent},
	"components/toast":           {"Toast", "toast", toastDemoContent},
	"components/toggle":          {"Toggle", "toggle", toggleDemoContent},
	"components/tooltip":         {"Tooltip", "tooltip", tooltipDemoContent},
	"components/triplet":         {"Triplet", "triplet", tripletDemoContent},
	"docs/theme":                 {"Theme", "theme", themeDemoContent},
	"getting-started":            {"Getting Started", "", gettingStartedContent},
	"examples":                   {"Examples", "examples", examples.IndexContent},
	"examples/todo":              {"Todo List", "todo", examples.TodoContent},
	"examples/logs":              {"Live Log Feed", "logs", examples.LogsContent},
}

// LookupDemo returns the entry for a given canonical key (no leading slash).
func LookupDemo(key string) (DemoEntry, bool) {
	e, ok := Demos[key]
	return e, ok
}
