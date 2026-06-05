package link

import (
	"github.com/a-h/templ"
)

// Style controls the visual treatment of a link.
type Style string

const (
	// StyleText renders a standard inline text link.
	StyleText Style = "text"
	// StyleButton renders a link with primary button styling.
	StyleButton Style = "button"
)

// Size controls the dimensions of button-styled links.
type Size string

const (
	SizeSmall  Size = "sm"
	SizeMedium Size = "md"
	SizeLarge  Size = "lg"
	SizeXLarge Size = "xl"
)

// IconPosition controls where Icon is rendered relative to the link content.
type IconPosition string

const (
	IconLeading  IconPosition = "leading"
	IconTrailing IconPosition = "trailing"
)

// Config holds configuration for the Link component.
type Config struct {
	// Href is the link destination. Defaults to "#".
	Href string
	// Target is the native anchor target attribute.
	Target string
	// Rel is the native anchor rel attribute. Defaults to "noopener noreferrer"
	// when Target is "_blank".
	Rel string
	// Role overrides the native role. StyleButton defaults to role="button".
	Role string
	// ID is the HTML id attribute for the anchor.
	ID string
	// Style controls the link treatment. Defaults to StyleText.
	Style Style
	// Size controls button-styled link dimensions. Defaults to SizeMedium.
	Size Size
	// Icon is an optional icon rendered inside the link.
	Icon templ.Component
	// IconPosition controls Icon placement. Defaults to IconTrailing.
	IconPosition IconPosition
	// Class is additional CSS classes appended to the anchor.
	Class string
	// Attrs are arbitrary attributes on the anchor element.
	Attrs templ.Attributes
}

func (cfg Config) href() templ.SafeURL {
	if cfg.Href == "" {
		return "#"
	}
	return templ.URL(cfg.Href)
}

func (cfg Config) rel() string {
	if cfg.Rel != "" {
		return cfg.Rel
	}
	if cfg.Target == "_blank" {
		return "noopener noreferrer"
	}
	return ""
}

func (cfg Config) role() string {
	if cfg.Role != "" {
		return cfg.Role
	}
	if cfg.Style == StyleButton {
		return "button"
	}
	return ""
}

func (cfg Config) classes() string {
	base := cfg.textClasses()
	if cfg.Style == StyleButton {
		base = cfg.buttonClasses()
	}
	if cfg.Icon != nil {
		base += " inline-flex items-center gap-1.5"
	}
	if cfg.Class != "" {
		base += " " + cfg.Class
	}
	return base
}

func (cfg Config) textClasses() string {
	return "font-medium text-primary underline-offset-2 transition-colors hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:text-primary-dark dark:focus-visible:outline-primary-dark"
}

func (cfg Config) buttonClasses() string {
	return "whitespace-nowrap rounded-2xl border border-primary bg-primary text-on-primary font-medium tracking-wide text-center transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:opacity-100 active:outline-offset-0 dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark " + cfg.buttonSizeClasses()
}

func (cfg Config) buttonSizeClasses() string {
	switch cfg.Size {
	case SizeSmall:
		return "px-4 py-2 text-xs"
	case SizeLarge:
		return "px-4 py-2 text-base"
	case SizeXLarge:
		return "px-4 py-2 text-lg"
	default:
		return "px-4 py-2 text-sm"
	}
}
