package search

import (
	"net/url"
	"strings"

	"github.com/a-h/templ"
)

// Item describes one search result rendered by Search.
type Item struct {
	// ID is an optional stable identifier for the result.
	ID string
	// Title is the primary result text.
	Title string
	// Description is optional supporting text.
	Description string
	// Href turns the result into a link.
	Href string
	// Section is optional grouping or eyebrow text.
	Section string
	// Keywords are extra terms included in client-side filtering.
	Keywords []string
	// Attrs are extra attributes spread onto the result button.
	Attrs templ.Attributes
}

// JSString escapes a value for use inside a single-quoted Alpine expression.
func JSString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

// Config holds configuration for the search component.
type Config struct {
	// ID is a unique identifier used for ARIA relationships.
	ID string
	// Label is the accessible label and optional visible trigger text.
	Label string
	// Placeholder is shown in the dialog input.
	Placeholder string
	// ShortcutText is rendered in the trigger KBD hint. Defaults to "⌘ K".
	ShortcutText string
	// GlobalShortcut enables Cmd/Ctrl+K window handling for this instance.
	GlobalShortcut bool
	// EscapeText is rendered in the dialog close KBD hint. Defaults to "Esc".
	EscapeText string
	// Items are the caller-provided result records.
	Items []Item
	// MaxResults limits the visible matches. Defaults to 4.
	MaxResults int
	// DescriptionMaxLength truncates result descriptions. Defaults to 120.
	DescriptionMaxLength int
	// EmptyText appears when no result matches the query.
	EmptyText string
	// RootClass allows additional classes on the root.
	RootClass string
	// TriggerClass allows additional classes on the trigger button.
	TriggerClass string
	// DialogClass allows additional classes on the dialog panel.
	DialogClass string
	// InputAttrs allows extra attributes on the search input.
	InputAttrs templ.Attributes
}

// GetID returns a stable ID.
func (cfg Config) GetID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	return "search"
}

// GetLabel returns the accessible label.
func (cfg Config) GetLabel() string {
	if cfg.Label != "" {
		return cfg.Label
	}
	return "Search"
}

// GetPlaceholder returns the input placeholder.
func (cfg Config) GetPlaceholder() string {
	if cfg.Placeholder != "" {
		return cfg.Placeholder
	}
	return "Search"
}

// GetShortcutText returns the trigger shortcut hint.
func (cfg Config) GetShortcutText() string {
	if cfg.ShortcutText != "" {
		return cfg.ShortcutText
	}
	return "⌘ K"
}

// GetEscapeText returns the dialog close shortcut hint.
func (cfg Config) GetEscapeText() string {
	if cfg.EscapeText != "" {
		return cfg.EscapeText
	}
	return "Esc"
}

// GetEmptyText returns the empty-state copy.
func (cfg Config) GetEmptyText() string {
	if cfg.EmptyText != "" {
		return cfg.EmptyText
	}
	return "No results found."
}

// GetMaxResults returns the maximum number of visible results.
func (cfg Config) GetMaxResults() int {
	if cfg.MaxResults > 0 {
		return cfg.MaxResults
	}
	return 4
}

// GetDescriptionMaxLength returns the maximum visible description length.
func (cfg Config) GetDescriptionMaxLength() int {
	if cfg.DescriptionMaxLength > 0 {
		return cfg.DescriptionMaxLength
	}
	return 120
}

// RootClasses returns classes for the component root.
func (cfg Config) RootClasses() string {
	classes := "w-full"
	if cfg.RootClass != "" {
		classes += " " + cfg.RootClass
	}
	return classes
}

// TriggerClasses returns classes for the search trigger.
func (cfg Config) TriggerClasses() string {
	classes := "flex min-h-10 w-full items-center gap-2 rounded-radius border border-outline bg-surface px-3 text-left text-sm text-on-surface-muted transition hover:border-outline-strong hover:text-on-surface-strong focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:border-outline-dark dark:bg-surface-dark/50 dark:text-on-surface-dark-muted dark:hover:border-outline-dark-strong dark:hover:text-on-surface-dark-strong dark:focus-visible:outline-primary-dark"
	if cfg.TriggerClass != "" {
		classes += " " + cfg.TriggerClass
	}
	return classes
}

// DialogClasses returns classes for the search panel.
func (cfg Config) DialogClasses() string {
	classes := "relative mt-16 flex w-full max-w-2xl flex-col text-on-surface shadow-2xl shadow-black/20 dark:text-on-surface-dark"
	if cfg.DialogClass != "" {
		classes += " " + cfg.DialogClass
	}
	return classes
}

// SearchText returns the normalized client-side search corpus for an item.
func (item Item) SearchText() string {
	parts := []string{item.Title, item.Description, item.Section}
	parts = append(parts, item.Keywords...)
	return strings.TrimSpace(strings.Join(parts, " "))
}

// SafeHref returns a navigation target only for schemes that cannot execute
// script when assigned to window.location.href.
func (item Item) SafeHref() string {
	href := strings.TrimSpace(item.Href)
	if href == "" {
		return ""
	}
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	switch strings.ToLower(u.Scheme) {
	case "", "http", "https", "mailto", "tel":
		return href
	default:
		return ""
	}
}
