package emptystate

import (
	"strings"

	"github.com/a-h/templ"
)

// Config holds empty-state guidance, optional visuals and actions, and target-specific hooks.
type Config struct {
	// Title names the empty state; default is "Nothing here yet".
	Title string
	// Description explains what appears here or how to proceed.
	Description string
	// Icon renders an optional decorative visual above the title.
	Icon templ.Component
	// Action renders an optional recovery or creation control.
	Action templ.Component
	// RootClass appends CSS classes to the empty-state root.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the empty-state root.
	RootAttrs templ.Attributes
	// IconClass appends CSS classes to the icon wrapper.
	IconClass string
	// IconAttrs appends arbitrary HTML attributes to the icon wrapper.
	IconAttrs templ.Attributes
	// ActionClass appends CSS classes to the action wrapper.
	ActionClass string
	// ActionAttrs appends arbitrary HTML attributes to the action wrapper.
	ActionAttrs templ.Attributes
}

func (cfg Config) title() string {
	if title := strings.TrimSpace(cfg.Title); title != "" {
		return title
	}
	return "Nothing here yet"
}

func (cfg Config) description() string {
	if description := strings.TrimSpace(cfg.Description); description != "" {
		return description
	}
	return "Items will appear here when they are available."
}

func (cfg Config) rootClasses() string {
	return appendClass(
		"flex min-h-48 flex-col items-center justify-center gap-3 rounded-radius border border-dashed border-outline bg-surface-alt/30 px-6 py-12 text-center text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt/30 dark:text-on-surface-dark",
		cfg.RootClass,
	)
}

func (cfg Config) iconClasses() string {
	return appendClass(
		"flex size-12 items-center justify-center rounded-full bg-surface-alt text-on-surface-muted dark:bg-surface-dark-alt dark:text-on-surface-dark-muted",
		cfg.IconClass,
	)
}

func (cfg Config) actionClasses() string {
	return appendClass("mt-2 flex flex-wrap items-center justify-center gap-2", cfg.ActionClass)
}

func appendClass(base, extra string) string {
	if extra = strings.TrimSpace(extra); extra != "" {
		return base + " " + extra
	}
	return base
}
