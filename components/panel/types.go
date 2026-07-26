package panel

import (
	"strings"

	"github.com/a-h/templ"
)

// Appearance controls the panel's background and border treatment.
type Appearance string

const (
	// AppearanceOutlined is the default surface with a border.
	AppearanceOutlined Appearance = ""
	// AppearanceSubtle uses the alternate surface without a border.
	AppearanceSubtle Appearance = "subtle"
	// AppearancePlain leaves background, border, and radius to the surrounding layout.
	AppearancePlain Appearance = "plain"
)

// Density controls padding inside panel regions.
type Density string

const (
	// DensityDefault uses the standard application spacing.
	DensityDefault Density = ""
	// DensityCompact reduces padding for dense operational interfaces.
	DensityCompact Density = "compact"
	// DensityRelaxed increases padding for focused detail and settings surfaces.
	DensityRelaxed Density = "relaxed"
)

// Config defines a neutral application surface with optional named regions.
// Header, Actions, Body, and Footer own no heading level or document semantics;
// consumers supply those through the rendered components and RootAttrs.
type Config struct {
	// Appearance controls background and border treatment.
	Appearance Appearance
	// Density controls padding in header, body, and footer regions.
	Density Density
	// Header renders arbitrary heading or context content.
	Header templ.Component
	// Actions renders controls aligned opposite Header.
	Actions templ.Component
	// Body renders the primary content. When nil, Panel renders templ children.
	Body templ.Component
	// Footer renders supporting content or secondary actions below the body.
	Footer templ.Component
	// RootClass appends CSS classes to the neutral root div.
	RootClass string
	// RootAttrs appends arbitrary HTML attributes to the neutral root div.
	RootAttrs templ.Attributes
	// HeaderClass appends CSS classes to the header region.
	HeaderClass string
	// HeaderAttrs appends arbitrary HTML attributes to the header region.
	HeaderAttrs templ.Attributes
	// ActionsClass appends CSS classes to the actions region.
	ActionsClass string
	// ActionsAttrs appends arbitrary HTML attributes to the actions region.
	ActionsAttrs templ.Attributes
	// BodyClass appends CSS classes to the body region.
	BodyClass string
	// BodyAttrs appends arbitrary HTML attributes to the body region.
	BodyAttrs templ.Attributes
	// FooterClass appends CSS classes to the footer region.
	FooterClass string
	// FooterAttrs appends arbitrary HTML attributes to the footer region.
	FooterAttrs templ.Attributes
}

func (cfg Config) rootClasses() string {
	base := "w-full text-on-surface dark:text-on-surface-dark"
	switch cfg.Appearance {
	case AppearanceSubtle:
		base += " rounded-radius bg-surface-alt dark:bg-surface-dark-alt"
	case AppearancePlain:
		base += " bg-transparent"
	default:
		base += " rounded-radius border border-outline bg-surface dark:border-outline-dark dark:bg-surface-dark"
	}
	return appendClass(base, cfg.RootClass)
}

func (cfg Config) regionPadding() string {
	switch cfg.Density {
	case DensityCompact:
		return "px-4 py-3"
	case DensityRelaxed:
		return "px-6 py-5"
	default:
		return "px-5 py-4"
	}
}

func (cfg Config) headerClasses() string {
	base := "flex min-w-0 flex-wrap items-start justify-between gap-3 border-b border-outline dark:border-outline-dark " + cfg.regionPadding()
	return appendClass(base, cfg.HeaderClass)
}

func (cfg Config) actionsClasses() string {
	return appendClass("ml-auto flex flex-wrap items-center justify-end gap-2", cfg.ActionsClass)
}

func (cfg Config) bodyClasses() string {
	return appendClass(cfg.regionPadding(), cfg.BodyClass)
}

func (cfg Config) footerClasses() string {
	base := "border-t border-outline dark:border-outline-dark " + cfg.regionPadding()
	return appendClass(base, cfg.FooterClass)
}

func appendClass(base, extra string) string {
	if extra = strings.TrimSpace(extra); extra != "" {
		return base + " " + extra
	}
	return base
}
