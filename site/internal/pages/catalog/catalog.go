// Package catalog owns the ordered component-page inventory used by the demo
// site's navigation, search, metadata, and documentation contract tests.
package catalog

import (
	"slices"
	"strings"

	"github.com/araihu/goshtoso/components"
)

const componentModulePath = "github.com/araihu/goshtoso/components/"
const pkgGoDevModuleURL = "https://pkg.go.dev/github.com/araihu/goshtoso@"

// Entry describes one public component documentation page.
type Entry struct {
	Key         string
	Path        string
	Title       string
	Active      string
	Description string
	Section     string
	Order       int
	Kinds       []components.Kind
	// Package overrides the package directory when the documentation route and
	// public Go package intentionally use different names.
	Package string
}

// GoPackagePath returns the public Go import path documented by this page.
// Route names use kebab case for readability; Go package directories do not.
func (e Entry) GoPackagePath() string {
	name := strings.TrimPrefix(e.Key, "components/")
	if e.Package != "" {
		name = e.Package
	}
	return componentModulePath + strings.ReplaceAll(name, "-", "")
}

// GoDocsURL returns the pkg.go.dev URL for this package at an exact module version.
func (e Entry) GoDocsURL(version string) string {
	name := strings.TrimPrefix(e.GoPackagePath(), componentModulePath)
	return pkgGoDevModuleURL + version + "/components/" + name
}

var componentPages = []Entry{
	{
		Key:         "components/app-shell",
		Path:        "/components/app-shell",
		Title:       "App Shell",
		Active:      "app-shell",
		Description: "Compose application frames with a keyboard skip link, persistent header, responsive sidebar, and one scrollable main region.",
		Section:     "Composition",
		Order:       0,
		Kinds:       []components.Kind{components.KindAppShell},
	},
	{
		Key:         "components/page-header",
		Path:        "/components/page-header",
		Title:       "Page Header",
		Active:      "page-header",
		Description: "Establish page hierarchy with breadcrumb context, a primary title, concise description, and page-level actions.",
		Section:     "Composition",
		Order:       1,
		Kinds:       []components.Kind{components.KindPageHeader},
	},
	{
		Key:         "components/toolbar",
		Path:        "/components/toolbar",
		Title:       "Toolbar",
		Active:      "toolbar",
		Description: "Group search, filters, view controls, and collection actions in a responsive, accessible application toolbar.",
		Section:     "Composition",
		Order:       2,
		Kinds:       []components.Kind{components.KindToolbar},
	},
	{
		Key:         "components/action-group",
		Path:        "/components/action-group",
		Title:       "Action Group",
		Active:      "action-group",
		Description: "Keep a primary action visible while lower-priority actions collapse into one accessible flat overflow menu at constrained widths.",
		Section:     "Composition",
		Order:       3,
		Kinds:       []components.Kind{components.KindActionGroup},
	},
	{
		Key:         "components/panel",
		Path:        "/components/panel",
		Title:       "Panel",
		Active:      "panel",
		Description: "Frame neutral full-width application regions with arbitrary header, actions, body, and footer content without imposing heading or article semantics.",
		Section:     "Composition",
		Order:       4,
		Kinds:       []components.Kind{components.KindPanel},
	},
	{
		Key:         "components/empty-state",
		Path:        "/components/empty-state",
		Title:       "Empty State",
		Active:      "empty-state",
		Description: "Explain absent content and offer a clear next action with an instructional empty-state component.",
		Section:     "Composition",
		Order:       5,
		Kinds:       []components.Kind{components.KindEmptyState},
	},
	{
		Key:         "components/skeleton",
		Path:        "/components/skeleton",
		Title:       "Skeleton",
		Active:      "skeleton",
		Description: "Render accessible text, rectangle, and circle loading placeholders with reduced-motion support.",
		Section:     "Composition",
		Order:       6,
		Kinds:       []components.Kind{components.KindSkeleton},
	},
	{
		Key:         "components/accordion",
		Path:        "/components/accordion",
		Title:       "Accordion",
		Active:      "accordion",
		Description: "Build accessible accordion interfaces in Go with server-rendered templ markup, HTMX-friendly content, and Alpine.js disclosure behavior.",
		Section:     "Display",
		Order:       7,
		Kinds:       []components.Kind{components.KindAccordion},
	},
	{
		Key:         "components/avatar",
		Path:        "/components/avatar",
		Title:       "Avatar",
		Active:      "avatar",
		Description: "Render Go avatar components with images, initials, placeholders, status indicators, borders, and stacked group layouts.",
		Section:     "Display",
		Order:       8,
		Kinds:       []components.Kind{components.KindAvatar, components.KindAvatarStack},
	},
	{
		Key:         "components/badge",
		Path:        "/components/badge",
		Title:       "Badge",
		Active:      "badge",
		Description: "Use Goshtoso badge components for statuses, counts, semantic colors, icon labels, and notification dots in Go applications.",
		Section:     "Display",
		Order:       9,
		Kinds: []components.Kind{
			components.KindBadge,
			components.KindNotificationBadge,
			components.KindNotificationDot,
			components.KindAnimatingDot,
		},
	},
	{
		Key:         "components/banner",
		Path:        "/components/banner",
		Title:       "Banner",
		Active:      "banner",
		Description: "Create inline announcement and alert banners with tones, icons, actions, and dismiss patterns for server-rendered Go UIs.",
		Section:     "Display",
		Order:       10,
		Kinds:       []components.Kind{components.KindBanner, components.KindCookieBanner},
	},
	{
		Key:         "components/card",
		Path:        "/components/card",
		Title:       "Card",
		Active:      "card",
		Description: "Compose structured card layouts in Go with media, headings, body content, actions, and responsive horizontal or vertical layouts.",
		Section:     "Display",
		Order:       11,
		Kinds:       []components.Kind{components.KindCard},
	},
	{
		Key:         "components/carousel",
		Path:        "/components/carousel",
		Title:       "Carousel",
		Active:      "carousel",
		Description: "Render carousel content rails for images, feature panels, and card-framed slides with Goshtoso, templ, and Alpine.js.",
		Section:     "Display",
		Order:       12,
		Kinds:       []components.Kind{components.KindCarousel, components.KindCardCarousel},
	},
	{
		Key:         "components/chatbubble",
		Path:        "/components/chatbubble",
		Title:       "Chat Bubble",
		Active:      "chatbubble",
		Description: "Create conversation message bubbles with alignment, avatars, metadata, and grouped thread layouts for Go chat interfaces.",
		Section:     "Display",
		Order:       13,
		Kinds:       []components.Kind{components.KindChatBubble, components.KindTypingIndicator},
	},
	{
		Key:         "components/codeblock",
		Path:        "/components/codeblock",
		Title:       "Code Block",
		Active:      "codeblock",
		Description: "Show copy-pasteable code snippets with semantic pre and code markup, labels, scroll bounds, copy actions, and themes.",
		Section:     "Display",
		Order:       14,
		Kinds:       []components.Kind{components.KindCodeBlock},
	},
	{
		Key:         "components/dependencies",
		Path:        "/components/dependencies",
		Title:       "Dependencies",
		Active:      "dependencies",
		Description: "Load Goshtoso's local CSS, Alpine.js, HTMX, and runtime helpers from embedded assets without CDN dependencies.",
		Section:     "Display",
		Order:       15,
		Kinds:       []components.Kind{components.KindDependencies, components.KindDependenciesMinimal},
		Package:     "head",
	},
	{
		Key:         "components/kbd",
		Path:        "/components/kbd",
		Title:       "KBD",
		Active:      "kbd",
		Description: "Display semantic keyboard shortcut hints for command palettes, forms, and toolbar controls in Go documentation.",
		Section:     "Display",
		Order:       16,
		Kinds:       []components.Kind{components.KindKbd},
	},
	{
		Key:         "components/table",
		Path:        "/components/table",
		Title:       "Table",
		Active:      "table",
		Description: "Build sortable, paginated, filterable Go data tables with HTMX loading, rich cells, and server-rendered rows.",
		Section:     "Display",
		Order:       17,
		Kinds: []components.Kind{
			components.KindTable,
			components.KindTableHeadContent,
			components.KindTableRows,
			components.KindTableRow,
			components.KindTablePaginationNav,
			components.KindTableImageCell,
		},
	},
	{
		Key:         "components/button",
		Path:        "/components/button",
		Title:       "Button",
		Active:      "button",
		Description: "Build Go button components with variants, sizes, icons, loading states, and HTMX request integrations.",
		Section:     "Input",
		Order:       18,
		Kinds:       []components.Kind{components.KindButton},
	},
	{
		Key:         "components/checkbox",
		Path:        "/components/checkbox",
		Title:       "Checkbox",
		Active:      "checkbox",
		Description: "Build checkbox fields and groups with accessible labels, helper text, disabled states, and validation in Go.",
		Section:     "Input",
		Order:       19,
		Kinds:       []components.Kind{components.KindCheckbox, components.KindCheckboxGroup},
	},
	{
		Key:         "components/combobox",
		Path:        "/components/combobox",
		Title:       "Combobox",
		Active:      "combobox",
		Description: "Add searchable single and multi-select combobox controls with keyboard navigation to server-rendered Go forms.",
		Section:     "Input",
		Order:       20,
		Kinds:       []components.Kind{components.KindCombobox},
	},
	{
		Key:         "components/fileinput",
		Path:        "/components/fileinput",
		Title:       "File Input",
		Active:      "fileinput",
		Description: "Create file upload controls with labels, helper text, selected-file display, native accept hints, and drop-zone or upload-button appearances in Goshtoso.",
		Section:     "Input",
		Order:       21,
		Kinds:       []components.Kind{components.KindFileInput},
	},
	{
		Key:         "components/form",
		Path:        "/components/form",
		Title:       "Form",
		Active:      "form",
		Description: "Build form layout patterns, field groups, server validation messages, and HTMX submit behavior with Go and templ.",
		Section:     "Input",
		Order:       22,
		Kinds: []components.Kind{
			components.KindForm,
			components.KindFormSection,
			components.KindFormCollapsibleSection,
			components.KindFormFlipSection,
			components.KindFormSubSection,
			components.KindFormFieldGroup,
			components.KindFormErrors,
		},
	},
	{
		Key:         "components/radio",
		Path:        "/components/radio",
		Title:       "Radio",
		Active:      "radio",
		Description: "Build radio groups and segmented controls with accessible labels, helper copy, and validation states in Go.",
		Section:     "Input",
		Order:       23,
		Kinds:       []components.Kind{components.KindRadio, components.KindRadioBar, components.KindRadioGroup},
	},
	{
		Key:         "components/range",
		Path:        "/components/range",
		Title:       "Range",
		Active:      "range",
		Description: "Create range sliders with labels, helper text, live value output, generated or custom ticks, and accessible input behavior for Go forms.",
		Section:     "Input",
		Order:       24,
		Kinds:       []components.Kind{components.KindRange},
	},
	{
		Key:         "components/rating",
		Path:        "/components/rating",
		Title:       "Rating",
		Active:      "rating",
		Description: "Render rating inputs and display patterns with stars, labels, accessible values, and form integration in Goshtoso.",
		Section:     "Input",
		Order:       25,
		Kinds:       []components.Kind{components.KindRating, components.KindRatingDisplay},
	},
	{
		Key:         "components/palette",
		Path:        "/components/palette",
		Title:       "Palette",
		Active:      "palette",
		Description: "Render color picking surfaces with swatches, hex entry, token labels, and theme integration in Goshtoso.",
		Section:     "Input",
		Order:       26,
		Kinds:       []components.Kind{components.KindPalette},
	},
	{
		Key:         "components/search",
		Path:        "/components/search",
		Title:       "Search",
		Active:      "search",
		Description: "Add command-palette search with a sidebar trigger, global modal, filtering, highlighting, and docs navigation.",
		Section:     "Input",
		Order:       27,
		Kinds:       []components.Kind{components.KindSearch, components.KindSearchField, components.KindSearchModal},
	},
	{
		Key:         "components/select",
		Path:        "/components/select",
		Title:       "Select",
		Active:      "select",
		Description: "Build native-feeling select menus with trigger content, grouped options, and keyboard support in Go.",
		Section:     "Input",
		Order:       28,
		Kinds:       []components.Kind{components.KindSelect},
	},
	{
		Key:         "components/schema-form",
		Path:        "/components/schema-form",
		Title:       "Schema Form",
		Active:      "schema-form",
		Description: "Generate complete form sections from JSON Schema defaults, submitted values, and allow-list rules in server-rendered Go interfaces.",
		Section:     "Input",
		Order:       29,
		Kinds:       []components.Kind{components.KindSchemaFormFields},
	},
	{
		Key:         "components/structured-input",
		Path:        "/components/structured-input",
		Title:       "Structured Input",
		Active:      "structured-input",
		Description: "Build repeatable structured form rows with typed columns for text and select controls, nested submitted names, defaults, and add/remove actions in Go.",
		Section:     "Input",
		Order:       30,
		Kinds:       []components.Kind{components.KindStructuredInput},
	},
	{
		Key:         "components/tags-list",
		Path:        "/components/tags-list",
		Title:       "Tags List",
		Active:      "tags-list",
		Description: "Manage editable tag collections with add, remove, duplicate-preserving values, and keyboard flows in Goshtoso.",
		Section:     "Input",
		Order:       31,
		Kinds:       []components.Kind{components.KindTagsList},
	},
	{
		Key:         "components/text-input",
		Path:        "/components/text-input",
		Title:       "Text Input",
		Active:      "text-input",
		Description: "Render text fields with icons, search affordances, masks, validation, and password controls for Go forms.",
		Section:     "Input",
		Order:       32,
		Kinds:       []components.Kind{components.KindTextInput},
	},
	{
		Key:         "components/textarea",
		Path:        "/components/textarea",
		Title:       "Textarea",
		Active:      "textarea",
		Description: "Build multi-line text entry with resize behavior, helper text, row defaults, and validation states in server-rendered Go UIs.",
		Section:     "Input",
		Order:       33,
		Kinds:       []components.Kind{components.KindTextarea, components.KindTextareaWithActions},
	},
	{
		Key:         "components/toggle",
		Path:        "/components/toggle",
		Title:       "Toggle",
		Active:      "toggle",
		Description: "Render binary switches for settings, feature flags, and compact on-off choices in Go applications.",
		Section:     "Input",
		Order:       34,
		Kinds:       []components.Kind{components.KindToggle},
	},
	{
		Key:         "components/alert",
		Path:        "/components/alert",
		Title:       "Alert",
		Active:      "alert",
		Description: "Alert documentation for building interactive server-rendered Go interfaces with Goshtoso, templ, HTMX, Alpine.js, and Tailwind CSS.",
		Section:     "Feedback",
		Order:       35,
		Kinds:       []components.Kind{components.KindAlert},
	},
	{
		Key:         "components/toast",
		Path:        "/components/toast",
		Title:       "Toast",
		Active:      "toast",
		Description: "Create transient notifications with timing, actions, semantic tones, sender messages, and HTMX out-of-band toasts.",
		Section:     "Feedback",
		Order:       36,
		Kinds: []components.Kind{
			components.KindToastContainer,
			components.KindToast,
			components.KindMessageToast,
			components.KindOOBToast,
			components.KindOOBMessageToast,
		},
	},
	{
		Key:         "components/modal",
		Path:        "/components/modal",
		Title:       "Modal",
		Active:      "modal",
		Description: "Create modal dialogs with focus management, actions, dismissal, and scroll handling using Goshtoso and Alpine.js.",
		Section:     "Feedback",
		Order:       37,
		Kinds:       []components.Kind{components.KindModal, components.KindAlertDialog},
	},
	{
		Key:         "components/drawer",
		Path:        "/components/drawer",
		Title:       "Drawer",
		Active:      "drawer",
		Description: "Create slide-over panels for details, filters, and HTMX workflows with Alpine.js state and focus trapping.",
		Section:     "Feedback",
		Order:       38,
		Kinds:       []components.Kind{components.KindDrawer},
	},
	{
		Key:         "components/spinner",
		Path:        "/components/spinner",
		Title:       "Spinner",
		Active:      "spinner",
		Description: "Render decorative animated loading glyphs with semantic tones and sizes for Go and HTMX interfaces.",
		Section:     "Feedback",
		Order:       39,
		Kinds:       []components.Kind{components.KindSpinner},
	},
	{
		Key:         "components/steps",
		Path:        "/components/steps",
		Title:       "Steps",
		Active:      "steps",
		Description: "Render progress indicators for wizards, onboarding, and multi-step workflows in server-rendered Go applications.",
		Section:     "Feedback",
		Order:       40,
		Kinds:       []components.Kind{components.KindSteps},
	},
	{
		Key:         "components/tooltip",
		Path:        "/components/tooltip",
		Title:       "Tooltip",
		Active:      "tooltip",
		Description: "Add contextual hints for icons, controls, and abbreviated UI labels with hover, focus, or click triggers.",
		Section:     "Feedback",
		Order:       41,
		Kinds:       []components.Kind{components.KindTooltip},
	},
	{
		Key:         "components/breadcrumbs",
		Path:        "/components/breadcrumbs",
		Title:       "Breadcrumbs",
		Active:      "breadcrumbs",
		Description: "Add hierarchical breadcrumb navigation to Go documentation routes, nested pages, and HTMX-enhanced server-rendered apps.",
		Section:     "Navigation",
		Order:       42,
		Kinds:       []components.Kind{components.KindBreadcrumbs},
	},
	{
		Key:         "components/dropdown",
		Path:        "/components/dropdown",
		Title:       "Dropdown",
		Active:      "dropdown",
		Description: "Render dropdown menus with grouped actions, icons, alignment, and keyboard-friendly behavior for Go UI workflows.",
		Section:     "Navigation",
		Order:       43,
		Kinds:       []components.Kind{components.KindDropdown},
	},
	{
		Key:         "components/link",
		Path:        "/components/link",
		Title:       "Link",
		Active:      "link",
		Description: "Render accessible link components for inline navigation, external links, icon links, and disabled states in Go UIs.",
		Section:     "Navigation",
		Order:       44,
		Kinds:       []components.Kind{components.KindLink},
	},
	{
		Key:         "components/navbar",
		Path:        "/components/navbar",
		Title:       "Navbar",
		Active:      "navbar",
		Description: "Build top navigation bars with links, actions, responsive menus, and brand areas for server-rendered Go sites.",
		Section:     "Navigation",
		Order:       45,
		Kinds:       []components.Kind{components.KindNavbar},
	},
	{
		Key:         "components/pagination",
		Path:        "/components/pagination",
		Title:       "Pagination",
		Active:      "pagination",
		Description: "Add pagination controls for tables, lists, and HTMX-powered result sets in Go applications.",
		Section:     "Navigation",
		Order:       46,
		Kinds:       []components.Kind{components.KindPagination},
	},
	{
		Key:         "components/sidebar",
		Path:        "/components/sidebar",
		Title:       "Sidebar",
		Active:      "sidebar",
		Description: "Create persistent or overlay side navigation with sections, search slots, active states, and responsive behavior.",
		Section:     "Navigation",
		Order:       47,
		Kinds:       []components.Kind{components.KindSidebar, components.KindSidebarOverlay},
	},
	{
		Key:         "components/tabs",
		Path:        "/components/tabs",
		Title:       "Tabs",
		Active:      "tabs",
		Description: "Create segmented content navigation with keyboard support, panels, active states, and HTMX-loaded tab content.",
		Section:     "Navigation",
		Order:       48,
		Kinds:       []components.Kind{components.KindTabs},
	},
}

// ComponentPages returns the component documentation entries in sidebar order.
// Both the entry slice and each entry's Kinds slice are defensive copies.
func ComponentPages() []Entry {
	pages := slices.Clone(componentPages)
	for i := range pages {
		pages[i].Kinds = slices.Clone(pages[i].Kinds)
	}
	return pages
}

// Lookup returns the component documentation entry with the canonical route key.
func Lookup(key string) (Entry, bool) {
	for _, entry := range componentPages {
		if entry.Key == key {
			entry.Kinds = slices.Clone(entry.Kinds)
			return entry, true
		}
	}
	return Entry{}, false
}

// LookupActive returns the component documentation entry for a navigation key.
func LookupActive(active string) (Entry, bool) {
	for _, entry := range componentPages {
		if entry.Active == active {
			entry.Kinds = slices.Clone(entry.Kinds)
			return entry, true
		}
	}
	return Entry{}, false
}
