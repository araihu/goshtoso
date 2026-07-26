package toolbar

import (
	"strings"

	"github.com/a-h/templ"
)

// Config holds toolbar regions, accessible labeling, and target-specific hooks.
type Config struct {
	// Label is the toolbar's accessible name; default is "Page tools".
	Label string
	// Search renders the primary search control.
	Search templ.Component
	// Filters renders filter and view controls.
	Filters templ.Component
	// Actions renders page or collection actions.
	Actions templ.Component
	// Sticky keeps the toolbar visible at the top of its scroll container.
	Sticky bool
	// RootClass appends CSS classes to the toolbar root.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the toolbar root.
	RootAttrs templ.Attributes
	// SearchClass appends CSS classes to the search wrapper.
	SearchClass string
	// SearchAttrs appends arbitrary HTML attributes to the search wrapper.
	SearchAttrs templ.Attributes
	// FiltersClass appends CSS classes to the filters wrapper.
	FiltersClass string
	// FiltersAttrs appends arbitrary HTML attributes to the filters wrapper.
	FiltersAttrs templ.Attributes
	// ActionsClass appends CSS classes to the actions wrapper.
	ActionsClass string
	// ActionsAttrs appends arbitrary HTML attributes to the actions wrapper.
	ActionsAttrs templ.Attributes
}

func (cfg Config) label() string {
	if label := strings.TrimSpace(cfg.Label); label != "" {
		return label
	}
	return "Page tools"
}

func (cfg Config) rootClasses() string {
	base := "flex flex-wrap items-center gap-3 rounded-radius border border-outline bg-surface-alt p-3 dark:border-outline-dark dark:bg-surface-dark-alt"
	if cfg.Sticky {
		base += " sticky top-0 z-20"
	}
	return appendClass(base, cfg.RootClass)
}

func (cfg Config) searchClasses() string {
	return appendClass("min-w-0 flex-1 basis-64", cfg.SearchClass)
}

func (cfg Config) filtersClasses() string {
	return appendClass("flex flex-wrap items-center gap-2", cfg.FiltersClass)
}

func (cfg Config) actionsClasses() string {
	return appendClass("ml-auto flex flex-wrap items-center gap-2", cfg.ActionsClass)
}

func appendClass(base, extra string) string {
	if extra = strings.TrimSpace(extra); extra != "" {
		return base + " " + extra
	}
	return base
}
