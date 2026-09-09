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
	// HideSeparator removes the bottom border while preserving header spacing.
	// The separator is visible by default.
	HideSeparator bool
	// RootClass appends CSS classes to the header root.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the header root.
	RootAttrs templ.Attributes
	// BreadcrumbsClass appends CSS classes to the breadcrumbs wrapper.
	BreadcrumbsClass string
	// BreadcrumbsAttrs appends arbitrary HTML attributes to the breadcrumbs wrapper.
	BreadcrumbsAttrs templ.Attributes
	// TitleClass appends CSS classes to the h1 heading.
	TitleClass string
	// TitleAttrs appends arbitrary HTML attributes to the h1 heading.
	TitleAttrs templ.Attributes
	// ActionsClass appends CSS classes to the actions wrapper.
	ActionsClass string
	// ActionsAttrs appends arbitrary HTML attributes to the actions wrapper.
	ActionsAttrs templ.Attributes
}

func (cfg Config) rootClasses() string {
	base := "flex flex-col gap-4 pb-6"
	if !cfg.HideSeparator {
		base += " border-b border-outline dark:border-outline-dark"
	}
	return appendClass(base, cfg.RootClass)
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

func (cfg Config) titleClasses() string {
	return appendClass(
		"text-2xl font-bold leading-tight text-on-surface-strong sm:text-3xl dark:text-on-surface-dark-strong",
		cfg.TitleClass,
	)
}

func appendClass(base, extra string) string {
	if extra = strings.TrimSpace(extra); extra != "" {
		return base + " " + extra
	}
	return base
}
