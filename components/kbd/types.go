package kbd

import (
	"strings"

	"github.com/a-h/templ"
)

// Size represents the rendered key size.
type Size string

const (
	SizeXS Size = "xs"
	SizeSM Size = "sm"
	SizeMD Size = "md"
	SizeLG Size = "lg"
)

// Config holds configuration for the kbd component.
type Config struct {
	// Text is the visible keyboard key text.
	Text string
	// Label is optional accessible text, useful for icon-only keys.
	Label string
	// Size controls text, padding, and icon dimensions.
	Size Size
	// Icon is an optional icon component rendered before Text.
	Icon templ.Component
	// Class allows additional CSS classes on the kbd element.
	Class string
	// Attrs allows caller-supplied attributes on the kbd element.
	Attrs templ.Attributes
}

// RootClasses returns the classes for the kbd element.
func (cfg Config) RootClasses() string {
	classes := "inline-flex items-center justify-center gap-1 rounded-radius border border-outline bg-surface-alt font-mono font-medium text-on-surface-strong shadow-sm shadow-outline/30 dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark-strong dark:shadow-black/20 " + cfg.SizeClasses()
	if cfg.Class != "" {
		classes += " " + cfg.Class
	}
	return classes
}

// SizeClasses returns the CSS classes for the configured size.
func (cfg Config) SizeClasses() string {
	switch cfg.Size {
	case SizeXS:
		return "min-h-5 min-w-5 px-1 py-0.5 text-[10px] leading-none"
	case SizeSM:
		return "min-h-6 min-w-6 px-1.5 py-0.5 text-xs leading-none"
	case SizeLG:
		return "min-h-9 min-w-9 px-2.5 py-1 text-base leading-none"
	default:
		return "min-h-7 min-w-7 px-2 py-1 text-sm leading-none"
	}
}

// IconClasses returns the icon size classes for the configured size.
func (cfg Config) IconClasses() string {
	switch cfg.Size {
	case SizeXS:
		return "size-3"
	case SizeSM:
		return "size-3.5"
	case SizeLG:
		return "size-5"
	default:
		return "size-4"
	}
}

// AccessibleLabel returns the non-visual label for icon-only keys.
func (cfg Config) AccessibleLabel() string {
	if cfg.Text != "" {
		return ""
	}
	return strings.TrimSpace(cfg.Label)
}
