package sidebar

import "github.com/a-h/templ"

// Item represents a single navigation item in the sidebar
type Item struct {
	// ID is the unique identifier for the item
	ID string
	// Label is the display text
	Label string
	// Icon is an optional icon (templ.Component)
	Icon templ.Component
	// Href is the link URL
	Href string
	// Active indicates if this item is currently active
	Active bool
	// Disabled prevents interaction
	Disabled bool
	// Badge is an optional badge text (e.g., "New", "Soon")
	Badge string
	// Items contains child items for nested navigation
	Items []Item
	// LinkAttrs are extra HTML attributes spread onto the <a> element.
	// Use for HTMX, Alpine.js, or other consumer-specific attributes.
	LinkAttrs templ.Attributes
}

// Config holds configuration for the sidebar
type Config struct {
	// Items are the top-level navigation items (icon links above sections)
	Items []Item
	// SectionsTitle is an optional header rendered above the sections (e.g. "Components")
	SectionsTitle string
	// Sections are grouped navigation sections with titles and left-border items
	Sections []Section
	// Logo is the logo component (optional)
	Logo templ.Component
	// LogoText is the text-only logo if no component provided
	LogoText string
	// LogoHref is the link for the logo (default: "/")
	LogoHref string
	// ShowSearch enables the search input
	ShowSearch bool
	// SearchPlaceholder is the placeholder text for search
	SearchPlaceholder string
	// RootClass allows additional CSS classes on the sidebar root.
	RootClass string
	// SearchSlot replaces the default search input with a custom component.
	// When set, ShowSearch is ignored.
	SearchSlot templ.Component
	// FooterSlot renders content at the bottom of the sidebar (e.g., profile menu).
	FooterSlot templ.Component
}

// Section represents a group of navigation items
type Section struct {
	// Title is the section header
	Title string
	// Items are the navigation items in this section
	Items []Item
	// Collapsible enables collapse/expand with Alpine.js
	Collapsible bool
}

// ContainerClasses returns the container CSS classes
func (cfg Config) ContainerClasses() string {
	return "h-full w-full border-r border-outline bg-surface dark:border-outline-dark dark:bg-surface-dark flex flex-col"
}

// NavClasses returns the navigation container classes
func (cfg Config) NavClasses() string {
	return "flex-1 overflow-y-auto sidebar-scroll scrollbar-custom p-4"
}
