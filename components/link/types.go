package link

import "github.com/a-h/templ"

// Appearance controls the visual treatment of a link.
type Appearance string

const (
	AppearanceText   Appearance = "text"
	AppearanceButton Appearance = "button"
)

// Size controls the dimensions of button-appearance links.
type Size string

const (
	SizeSmall  Size = "sm"
	SizeMedium Size = "md"
	SizeLarge  Size = "lg"
	SizeXLarge Size = "xl"
)

// IconPosition controls where an icon is rendered relative to link content.
type IconPosition string

const (
	IconLeading  IconPosition = "leading"
	IconTrailing IconPosition = "trailing"
)

type config struct {
	href         string
	target       string
	rel          string
	role         string
	id           string
	appearance   Appearance
	size         Size
	icon         templ.Component
	iconPosition IconPosition
	rootClass    string
	attrs        templ.Attributes
}

// Option configures a Link.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

// WithTarget sets the native anchor target attribute.
func WithTarget(target string) Option {
	return optionFunc(func(cfg *config) {
		cfg.target = target
	})
}

// WithRel sets the native anchor rel attribute.
func WithRel(rel string) Option {
	return optionFunc(func(cfg *config) {
		cfg.rel = rel
	})
}

// WithRole sets the anchor role.
func WithRole(role string) Option {
	return optionFunc(func(cfg *config) {
		cfg.role = role
	})
}

// WithID sets the anchor's HTML id attribute.
func WithID(id string) Option {
	return optionFunc(func(cfg *config) {
		cfg.id = id
	})
}

// WithAppearance sets the link's visual treatment.
func WithAppearance(appearance Appearance) Option {
	return optionFunc(func(cfg *config) {
		cfg.appearance = appearance
	})
}

// WithSize sets the dimensions of a button-appearance link.
func WithSize(size Size) Option {
	return optionFunc(func(cfg *config) {
		cfg.size = size
	})
}

// WithIcon adds an icon to the link.
func WithIcon(icon templ.Component) Option {
	return optionFunc(func(cfg *config) {
		cfg.icon = icon
	})
}

// WithIconPosition sets icon placement relative to link content.
func WithIconPosition(position IconPosition) Option {
	return optionFunc(func(cfg *config) {
		cfg.iconPosition = position
	})
}

// WithRootClass appends CSS classes to the anchor.
func WithRootClass(class string) Option {
	return optionFunc(func(cfg *config) {
		cfg.rootClass = class
	})
}

// WithAttrs adds arbitrary attributes to the anchor.
func WithAttrs(attrs templ.Attributes) Option {
	return optionFunc(func(cfg *config) {
		cfg.attrs = attrs
	})
}

func newConfig(href string, options []Option) config {
	cfg := config{
		href:         href,
		appearance:   AppearanceText,
		size:         SizeMedium,
		iconPosition: IconTrailing,
	}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

func (cfg config) safeHref() templ.SafeURL {
	return templ.URL(cfg.href)
}

func (cfg config) effectiveRel() string {
	if cfg.rel != "" {
		return cfg.rel
	}
	if cfg.target == "_blank" {
		return "noopener noreferrer"
	}
	return ""
}

func (cfg config) effectiveRole() string {
	if cfg.role != "" {
		return cfg.role
	}
	if cfg.appearance == AppearanceButton {
		return "button"
	}
	return ""
}

func (cfg config) classes() string {
	base := cfg.textClasses()
	if cfg.appearance == AppearanceButton {
		base = cfg.buttonClasses()
	}
	if cfg.icon != nil {
		base += " inline-flex items-center gap-1.5"
	}
	if cfg.rootClass != "" {
		base += " " + cfg.rootClass
	}
	return base
}

func (cfg config) textClasses() string {
	return "font-medium text-primary underline-offset-2 transition-colors hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:text-primary-dark dark:focus-visible:outline-primary-dark"
}

func (cfg config) buttonClasses() string {
	return "whitespace-nowrap rounded-2xl border border-primary bg-primary text-on-primary font-medium tracking-wide text-center transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:opacity-100 active:outline-offset-0 dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark " + cfg.buttonSizeClasses()
}

func (cfg config) buttonSizeClasses() string {
	switch cfg.size {
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
