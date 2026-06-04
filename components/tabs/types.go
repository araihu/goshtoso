package tabs

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

// Tab represents a single tab with its label and panel content
type Tab struct {
	// ID is the unique identifier for the tab (used in Alpine.js state)
	ID string
	// Label is the display text for the tab button
	Label string
	// Icon is an optional icon component rendered before the label
	Icon templ.Component
	// Badge is an optional badge text (e.g., count) shown after the label
	Badge string
	// Content is the tab panel content (used for static/inline content)
	Content templ.Component
	// HTMX enables lazy loading of tab content via an HTMX request.
	// When set, the panel issues an hx-get on first activation instead of
	// rendering Content inline.
	HTMX *TabHTMX
}

// TabHTMX configures HTMX lazy loading for a single tab panel
type TabHTMX struct {
	// Get is the URL to fetch content from (hx-get)
	Get string
	// Swap controls how the response is inserted (hx-swap, default "innerHTML")
	Swap string
	// Indicator is a CSS selector for a loading indicator element (hx-indicator)
	Indicator string
}

// Config holds configuration for the Tabs component
type Config struct {
	// ID is a unique identifier for this tabs instance (used for ARIA attributes)
	ID string
	// Tabs is the list of tabs to render
	Tabs []Tab
	// DefaultTab is the ID of the initially selected tab (defaults to first tab)
	DefaultTab string
	// RootClass allows additional CSS classes on the container.
	RootClass string
	// SyncHash syncs the active tab with the URL fragment (hash).
	// When true: reads hash on init to select tab, updates hash on tab change.
	// Invalid hash values fall back to DefaultTab.
	SyncHash bool
}

// tabsData generates the Alpine.js x-data object literal for the tab component.
// Uses inline x-data matching the combobox v1 pattern — Go generates the JS
// object literal, templ's HTML escaping (&#39; etc.) is transparent to Alpine.
func tabsData(cfg Config) string {
	defaultTab := cfg.DefaultTab
	if defaultTab == "" && len(cfg.Tabs) > 0 {
		defaultTab = cfg.Tabs[0].ID
	}

	if !cfg.SyncHash {
		return fmt.Sprintf(`{selectedTab:'%s'}`, jsEscapeSingle(defaultTab))
	}

	// Build JS array of valid tab IDs for hash validation
	jsIDs := make([]string, len(cfg.Tabs))
	for i, t := range cfg.Tabs {
		jsIDs[i] = "'" + jsEscapeSingle(t.ID) + "'"
	}
	var validArr strings.Builder
	for i, id := range jsIDs {
		if i > 0 {
			validArr.WriteString(",")
		}
		validArr.WriteString(id)
	}

	return fmt.Sprintf(
		`{selectedTab:'%s',init(){`+
			`var h=window.location.hash.slice(1);`+
			`var v=[%s];`+
			`if(h&&v.includes(h))this.selectedTab=h;`+
			`this.$watch('selectedTab',function(t){history.replaceState(null,'','#'+t)});`+
			`}}`,
		jsEscapeSingle(defaultTab), validArr.String())
}

// ActiveClasses returns the CSS classes for the active tab button
func ActiveClasses() string {
	return "font-bold text-primary border-b-2 border-primary dark:border-primary-dark dark:text-primary-dark"
}

// InactiveClasses returns the CSS classes for inactive tab buttons
func InactiveClasses() string {
	return "text-on-surface font-medium dark:text-on-surface-dark dark:hover:border-b-outline-dark-strong dark:hover:text-on-surface-dark-strong hover:border-b-2 hover:border-b-outline-strong hover:text-on-surface-strong"
}

// BadgeActiveClasses returns CSS for badge when tab is active
func BadgeActiveClasses() string {
	return "border-primary bg-primary/10 dark:bg-primary-dark dark:border-primary-dark dark:text-on-primary-dark"
}

// BadgeInactiveClasses returns CSS for badge when tab is inactive
func BadgeInactiveClasses() string {
	return "border-outline dark:border-outline-dark bg-surface-alt dark:bg-surface-dark-alt"
}
