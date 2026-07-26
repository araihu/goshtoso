package pageheader

import (
	"strings"

	"github.com/a-h/templ"
)

// Config holds page identity, navigation context, actions, and target-specific hooks.
type Config struct {
	// Title is the page's primary heading.
	Title string
	// Description is optional supporting copy below the title.
	Description string
	// Breadcrumbs renders navigation context above the title.
	Breadcrumbs templ.Component
	// Actions renders page-level controls beside the title group.
	Actions templ.Component
	// RootClass appends CSS classes to the header root.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the header root.
	RootAttrs templ.Attributes
	// BreadcrumbsClass appends CSS classes to the breadcrumbs wrapper.
	BreadcrumbsClass string
	// BreadcrumbsAttrs appends arbitrary HTML attributes to the breadcrumbs wrapper.
	BreadcrumbsAttrs templ.Attributes
	// ActionsClass appends CSS classes to the actions wrapper.
	ActionsClass string
	// ActionsAttrs appends arbitrary HTML attributes to the actions wrapper.
	ActionsAttrs templ.Attributes
}

func (cfg Config) rootClasses() string {
	return appendClass(
		"flex flex-col gap-4 border-b border-outline pb-6 dark:border-outline-dark",
		cfg.RootClass,
	)
}

func (cfg Config) breadcrumbsClasses() string {
	return appendClass("min-w-0", cfg.BreadcrumbsClass)
}

func (cfg Config) actionsClasses() string {
	return appendClass(
		"flex shrink-0 flex-wrap items-center gap-2 sm:justify-end",
		cfg.ActionsClass,
	)
}

func appendClass(base, extra string) string {
	if extra = strings.TrimSpace(extra); extra != "" {
		return base + " " + extra
	}
	return base
}
