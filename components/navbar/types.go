package navbar

import "github.com/a-h/templ"

// NavLink represents a navigation link in the navbar
type NavLink struct {
	// Label is the display text
	Label string
	// Href is the link URL
	Href string
	// Active marks this link as the current page
	Active bool
	// LinkAttrs are extra HTML attributes on the <a> tag
	LinkAttrs templ.Attributes
}

// UserProfile holds user information for the avatar dropdown
type UserProfile struct {
	// Name is the user's display name
	Name string
	// Email is the user's email address
	Email string
	// Avatar is an optional component rendered as the avatar trigger button content.
	// When nil, a default user icon is rendered.
	Avatar templ.Component
}

// UserMenuItem represents a single item in the user dropdown menu
type UserMenuItem struct {
	// Label is the display text
	Label string
	// Href is the link URL
	Href string
	// Icon is an optional icon rendered before the label
	Icon templ.Component
	// LinkAttrs are extra HTML attributes on the <a> tag
	LinkAttrs templ.Attributes
	// Danger renders the item in danger color (e.g., sign out)
	Danger bool
}

// ActionPosition controls where an action item is rendered in the navbar
type ActionPosition string

const (
	// ActionLeft renders the action after the brand (default)
	ActionLeft ActionPosition = "left"
	// ActionRight renders the action before the user avatar
	ActionRight ActionPosition = "right"
)

// ActionItem is a custom component rendered in the navbar at a configurable position.
type ActionItem struct {
	// Content is the component to render
	Content templ.Component
	// Position controls placement: "left" (after brand) or "right" (before avatar).
	// Default: "left"
	Position ActionPosition
}

// Config holds configuration for the navbar component
type Config struct {
	// Brand is the logo/brand component (left side)
	Brand templ.Component
	// BrandHref is the link for the brand (default: "/")
	BrandHref string
	// Links are the desktop navigation links
	Links []NavLink
	// Actions are custom components (e.g., dark mode toggle, theme selector)
	// rendered at configurable positions. Default position is left (after brand).
	Actions []ActionItem
	// User holds user profile data for the avatar dropdown (nil = no avatar)
	User *UserProfile
	// UserMenu contains dropdown items under the avatar
	UserMenu []UserMenuItem
	// NavClass allows additional CSS classes on the outer <nav>.
	NavClass string
	// NavAttrs are extra HTML attributes on the <nav> element
	NavAttrs templ.Attributes
}

// LeftActions returns action items positioned on the left
func (cfg Config) leftActions() []ActionItem {
	var items []ActionItem
	for _, a := range cfg.Actions {
		if a.Position == "" || a.Position == ActionLeft {
			items = append(items, a)
		}
	}
	return items
}

// RightActions returns action items positioned on the right
func (cfg Config) rightActions() []ActionItem {
	var items []ActionItem
	for _, a := range cfg.Actions {
		if a.Position == ActionRight {
			items = append(items, a)
		}
	}
	return items
}

// NavClasses returns the CSS classes for the outer nav element
func (cfg Config) navClasses() string {
	base := "flex items-center justify-between border-b border-outline px-6 py-4 dark:border-outline-dark"
	if cfg.NavClass != "" {
		return base + " " + cfg.NavClass
	}
	return base
}

// LinkClasses returns the CSS classes for a nav link
func linkClasses(active bool) string {
	if active {
		return "font-bold text-primary underline-offset-2 hover:text-primary focus:outline-hidden focus:underline dark:text-primary-dark dark:hover:text-primary-dark"
	}
	return "font-medium text-on-surface underline-offset-2 hover:text-primary focus:outline-hidden focus:underline dark:text-on-surface-dark dark:hover:text-primary-dark"
}

// MenuItemClasses returns the CSS classes for a user menu item
func menuItemClasses(danger bool) string {
	if danger {
		return "block bg-surface-alt px-4 py-2 text-sm text-danger hover:bg-danger/5 focus-visible:bg-danger/10 focus-visible:outline-hidden dark:bg-surface-dark-alt dark:hover:bg-danger/10"
	}
	return "block bg-surface-alt px-4 py-2 text-sm text-on-surface hover:bg-surface-dark-alt/5 hover:text-on-surface-strong focus-visible:bg-surface-dark-alt/10 focus-visible:text-on-surface-strong focus-visible:outline-hidden dark:bg-surface-dark-alt dark:text-on-surface-dark dark:hover:bg-surface-alt/5 dark:hover:text-on-surface-dark-strong dark:focus-visible:bg-surface-alt/10 dark:focus-visible:text-on-surface-dark-strong"
}
