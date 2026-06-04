package badge

import "github.com/a-h/templ"

// Variant represents badge style variants
type Variant string

const (
	Default   Variant = "default"
	Inverse   Variant = "inverse"
	Primary   Variant = "primary"
	Secondary Variant = "secondary"
	Info      Variant = "info"
	Success   Variant = "success"
	Warning   Variant = "warning"
	Danger    Variant = "danger"
)

// Style represents badge style (solid or soft)
type Style string

const (
	StyleSolid Style = "solid" // Default solid background
	StyleSoft  Style = "soft"  // Subtle background with border
)

// Size represents badge size
type Size string

const (
	SizeSM Size = "sm" // Small
	SizeMD Size = "md" // Medium (default)
	SizeLG Size = "lg" // Large
)

// Config holds configuration for the badge
type Config struct {
	// Label is the badge content
	Label string
	// Variant determines the color scheme
	Variant Variant
	// Style determines solid or soft appearance
	Style Style
	// Size of the badge
	Size Size
	// Icon is an optional icon component
	Icon templ.Component
	// Indicator adds a colored dot indicator
	Indicator bool
	// IndicatorColor overrides the default indicator color
	IndicatorColor string
	// Class allows additional CSS classes
	RootClass string
}

// SizeClasses returns the CSS classes for the size (text + padding)
func (cfg Config) SizeClasses() string {
	switch cfg.Size {
	case SizeSM:
		return "text-[10px] px-1.5 py-0.5"
	case SizeLG:
		return "text-sm px-3 py-1.5"
	default:
		return "text-xs px-2 py-1"
	}
}

// SizeTextClass returns only the text size class (no padding),
// used on the outer container of badgeWithInner so the inner span controls padding.
func (cfg Config) SizeTextClass() string {
	switch cfg.Size {
	case SizeSM:
		return "text-[10px]"
	case SizeLG:
		return "text-sm"
	default:
		return "text-xs"
	}
}

// VariantClasses returns the CSS classes for solid variant
func (cfg Config) VariantClasses() string {
	switch cfg.Variant {
	case Inverse:
		return "border border-outline-dark bg-surface-dark-alt text-on-surface-dark dark:border-outline dark:bg-surface-alt dark:text-on-surface"
	case Primary:
		return "border border-primary bg-primary text-on-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark"
	case Secondary:
		return "border border-secondary bg-secondary text-on-secondary dark:border-secondary-dark dark:bg-secondary-dark dark:text-on-secondary-dark"
	case Info:
		return "border border-info bg-info text-on-info"
	case Success:
		return "border border-success bg-success text-on-success"
	case Warning:
		return "border border-warning bg-warning text-on-warning"
	case Danger:
		return "border border-danger bg-danger text-on-danger"
	default:
		return "border border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"
	}
}

// SoftVariantClasses returns the CSS classes for soft variant
func (cfg Config) SoftVariantClasses() string {
	switch cfg.Variant {
	case Inverse:
		return "border border-outline-dark bg-surface text-on-surface dark:border-outline dark:bg-surface-dark dark:text-on-surface-dark"
	case Primary:
		return "border border-primary bg-surface text-primary dark:border-primary-dark dark:bg-surface-dark dark:text-primary-dark"
	case Secondary:
		return "border border-secondary bg-surface text-secondary dark:border-secondary-dark dark:bg-surface-dark dark:text-secondary-dark"
	case Info:
		return "border border-info bg-surface text-info dark:border-info dark:bg-surface-dark dark:text-info"
	case Success:
		return "border border-success bg-surface text-success dark:border-success dark:bg-surface-dark dark:text-success"
	case Warning:
		return "border border-warning bg-surface text-warning dark:border-warning dark:bg-surface-dark dark:text-warning"
	case Danger:
		return "border border-danger bg-surface text-danger dark:border-danger dark:bg-surface-dark dark:text-danger"
	default:
		return "border border-outline bg-surface text-on-surface dark:border-outline-dark dark:bg-surface-dark dark:text-on-surface-dark"
	}
}

// SoftInnerClasses returns the inner span classes for soft variant
func (cfg Config) SoftInnerClasses() string {
	switch cfg.Variant {
	case Inverse:
		return "bg-surface-dark-alt/10 dark:bg-surface-alt/10"
	case Primary:
		return "bg-primary/10 dark:bg-primary-dark/10"
	case Secondary:
		return "bg-secondary/10 dark:bg-secondary-dark/10"
	case Info:
		return "bg-info/10 dark:bg-info/10"
	case Success:
		return "bg-success/10 dark:bg-success/10"
	case Warning:
		return "bg-warning/10 dark:bg-warning/10"
	case Danger:
		return "bg-danger/10 dark:bg-danger/10"
	default:
		return "bg-surface-alt/10 dark:bg-surface-dark-alt/10"
	}
}

// IndicatorClasses returns the indicator dot classes
func (cfg Config) IndicatorClasses() string {
	if cfg.IndicatorColor != "" {
		return "size-1.5 rounded-full " + cfg.IndicatorColor
	}

	switch cfg.Variant {
	case Inverse:
		return "size-1.5 rounded-full bg-on-surface dark:bg-on-surface-dark"
	case Primary:
		return "size-1.5 rounded-full bg-primary dark:bg-primary-dark"
	case Secondary:
		return "size-1.5 rounded-full bg-secondary dark:bg-secondary-dark"
	case Info:
		return "size-1.5 rounded-full bg-info"
	case Success:
		return "size-1.5 rounded-full bg-success"
	case Warning:
		return "size-1.5 rounded-full bg-warning"
	case Danger:
		return "size-1.5 rounded-full bg-danger"
	default:
		return "size-1.5 rounded-full bg-on-surface dark:bg-on-surface-dark"
	}
}

// IsSoft returns true if badge uses soft style
func (cfg Config) IsSoft() bool {
	return cfg.Style == StyleSoft
}
