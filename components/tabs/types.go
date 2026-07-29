package tabs

import (
	"encoding/base64"
	"encoding/json"

	"github.com/a-h/templ"
)

// Tab represents a single tab with its label and panel content
type Tab struct {
	// ID is the unique identifier for the tab (used in Alpine.js state)
	ID string
	// Label is the display text for the tab button
	Label string
	// LabelSlot optionally replaces the visible label while Label remains accessible text.
	LabelSlot templ.Component
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

func defaultTab(cfg Config) string {
	defaultTab := cfg.DefaultTab
	if defaultTab == "" && len(cfg.Tabs) > 0 {
		defaultTab = cfg.Tabs[0].ID
	}
	return defaultTab
}

func tabIDsJSON(cfg Config) string {
	ids := make([]string, len(cfg.Tabs))
	for i, tab := range cfg.Tabs {
		ids[i] = tab.ID
	}
	encoded, err := json.Marshal(ids)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func tabsData(cfg Config) string {
	data := struct {
		Default  string   `json:"default"`
		IDs      []string `json:"ids"`
		SyncHash bool     `json:"syncHash"`
	}{defaultTab(cfg), make([]string, len(cfg.Tabs)), cfg.SyncHash}
	for i, tab := range cfg.Tabs {
		data.IDs[i] = tab.ID
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return base64.StdEncoding.EncodeToString([]byte(`{"default":"","ids":[],"syncHash":false}`))
	}
	return base64.StdEncoding.EncodeToString(encoded)
}

// ActiveClasses returns the CSS classes for the active tab button
func activeClasses() string {
	return "font-bold text-primary border-b-2 border-primary dark:border-primary-dark dark:text-primary-dark"
}

// InactiveClasses returns the CSS classes for inactive tab buttons
func inactiveClasses() string {
	return "text-on-surface font-medium dark:text-on-surface-dark dark:hover:border-b-outline-dark-strong dark:hover:text-on-surface-dark-strong hover:border-b-2 hover:border-b-outline-strong hover:text-on-surface-strong"
}

// BadgeActiveClasses returns CSS for badge when tab is active
func badgeActiveClasses() string {
	return "border-primary bg-primary/10 dark:bg-primary-dark dark:border-primary-dark dark:text-on-primary-dark"
}

// BadgeInactiveClasses returns CSS for badge when tab is inactive
func badgeInactiveClasses() string {
	return "border-outline dark:border-outline-dark bg-surface-alt dark:bg-surface-dark-alt"
}
