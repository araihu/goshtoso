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

type config struct {
	text      string
	label     string
	size      Size
	icon      templ.Component
	rootClass string
	attrs     templ.Attributes
}

// Option configures a Kbd.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

// WithLabel sets accessible text for the key.
func WithLabel(label string) Option {
	return optionFunc(func(cfg *config) {
		cfg.label = label
	})
}

// WithSize sets the rendered key size.
func WithSize(size Size) Option {
	return optionFunc(func(cfg *config) {
		cfg.size = size
	})
}

// WithIcon adds an icon before the key text.
func WithIcon(icon templ.Component) Option {
	return optionFunc(func(cfg *config) {
		cfg.icon = icon
	})
}

// WithRootClass appends CSS classes to the kbd element.
func WithRootClass(class string) Option {
	return optionFunc(func(cfg *config) {
		cfg.rootClass = class
	})
}

// WithAttrs adds arbitrary attributes to the kbd element.
func WithAttrs(attrs templ.Attributes) Option {
	return optionFunc(func(cfg *config) {
		cfg.attrs = attrs
	})
}

func newConfig(text string, options []Option) config {
	cfg := config{
		text: text,
		size: SizeMD,
	}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

func (cfg config) rootClasses() string {
	classes := "inline-flex items-center justify-center gap-1 rounded-radius border border-outline bg-surface-alt font-mono font-medium text-on-surface-strong shadow-sm shadow-outline/30 dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark-strong dark:shadow-black/20 " + cfg.sizeClasses()
	if cfg.rootClass != "" {
		classes += " " + cfg.rootClass
	}
	return classes
}

func (cfg config) sizeClasses() string {
	switch cfg.size {
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

func (cfg config) iconClasses() string {
	switch cfg.size {
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

func (cfg config) accessibleLabel() string {
	if cfg.text != "" {
		return ""
	}
	return strings.TrimSpace(cfg.label)
}
