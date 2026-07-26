package appshell

import (
	"maps"
	"strings"

	"github.com/a-h/templ"
)

// Config holds the application-shell regions and target-specific HTML hooks.
type Config struct {
	// Header renders the persistent top region; nil omits the header wrapper.
	Header templ.Component
	// Sidebar renders desktop navigation; nil omits the sidebar wrapper.
	Sidebar templ.Component
	// Content renders inside the shell's single scrollable main region. When
	// nil, AppShell renders its templ children as the content slot.
	Content templ.Component
	// MainID identifies the main region and skip-link target; default is "main-content".
	MainID string
	// SkipLinkLabel is the visible keyboard-focus label; default is "Skip to main content".
	SkipLinkLabel string
	// RootClass appends CSS classes to the outer shell element.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the outer shell element.
	RootAttrs templ.Attributes
	// HeaderClass appends CSS classes to the header wrapper.
	HeaderClass string
	// HeaderAttrs appends arbitrary HTML attributes to the header wrapper.
	HeaderAttrs templ.Attributes
	// SidebarClass appends CSS classes to the sidebar wrapper.
	SidebarClass string
	// SidebarAttrs appends arbitrary HTML attributes to the sidebar wrapper.
	SidebarAttrs templ.Attributes
	// MainClass appends CSS classes to the main region.
	MainClass string
	// MainAttrs appends arbitrary HTML attributes to the main region. AppShell
	// supplies tabindex="-1" by default so the skip-link target can receive
	// programmatic focus; set tabindex here to override that default.
	MainAttrs templ.Attributes
}

func (cfg Config) mainID() string {
	if id := strings.TrimSpace(cfg.MainID); id != "" {
		return id
	}
	return "main-content"
}

func (cfg Config) skipLinkLabel() string {
	if label := strings.TrimSpace(cfg.SkipLinkLabel); label != "" {
		return label
	}
	return "Skip to main content"
}

func (cfg Config) rootClasses() string {
	return appendClass(
		"relative flex min-h-screen flex-col overflow-hidden bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark",
		cfg.RootClass,
	)
}

func (cfg Config) headerClasses() string {
	return appendClass(
		"shrink-0 border-b border-outline bg-surface dark:border-outline-dark dark:bg-surface-dark",
		cfg.HeaderClass,
	)
}

func (cfg Config) sidebarClasses() string {
	return appendClass(
		"hidden w-72 shrink-0 overflow-y-auto border-r border-outline bg-surface dark:border-outline-dark dark:bg-surface-dark lg:block",
		cfg.SidebarClass,
	)
}

func (cfg Config) mainClasses() string {
	return appendClass(
		"min-w-0 flex-1 overflow-y-auto bg-surface p-4 sm:p-6 lg:p-8 dark:bg-surface-dark",
		cfg.MainClass,
	)
}

func (cfg Config) mainAttrs() templ.Attributes {
	attrs := make(templ.Attributes, len(cfg.MainAttrs)+1)
	maps.Copy(attrs, cfg.MainAttrs)
	if _, overridden := attrs["tabindex"]; !overridden {
		attrs["tabindex"] = "-1"
	}
	return attrs
}

func appendClass(base, extra string) string {
	if extra = strings.TrimSpace(extra); extra != "" {
		return base + " " + extra
	}
	return base
}
