package badge

import "github.com/a-h/templ"

// Tone represents badge style variants
type Tone string

const (
	ToneDefault   Tone = "default"
	ToneInverse   Tone = "inverse"
	TonePrimary   Tone = "primary"
	ToneSecondary Tone = "secondary"
	ToneInfo      Tone = "info"
	ToneSuccess   Tone = "success"
	ToneWarning   Tone = "warning"
	ToneDanger    Tone = "danger"
)

// Appearance represents the badge fill treatment.
type Appearance string

const (
	AppearanceSolid Appearance = ""     // Default solid background
	AppearanceSoft  Appearance = "soft" // Subtle background with border
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
	// Tone determines the color scheme
	Tone Tone
	// Appearance determines solid or soft rendering.
	Appearance Appearance
	// Size of the badge
	Size Size
	// Icon is an optional icon component
	Icon templ.Component
	// Indicator adds a colored dot indicator
	Indicator bool
	// IndicatorColor overrides the default indicator color
	IndicatorColor string
	// RootClass allows additional CSS classes on the badge root.
	RootClass string
}

// SizeClasses returns the CSS classes for the size (text + padding)
func (cfg Config) sizeClasses() string {
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
func (cfg Config) sizeTextClass() string {
	switch cfg.Size {
	case SizeSM:
		return "text-[10px]"
	case SizeLG:
		return "text-sm"
	default:
		return "text-xs"
	}
}

// toneClasses returns the CSS classes for solid variant
func (cfg Config) toneClasses() string {
	switch cfg.Tone {
	case ToneInverse:
		return "border border-outline-dark bg-surface-dark-alt text-on-surface-dark dark:border-outline dark:bg-surface-alt dark:text-on-surface"
	case TonePrimary:
		return "border border-primary bg-primary text-on-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark"
	case ToneSecondary:
		return "border border-secondary bg-secondary text-on-secondary dark:border-secondary-dark dark:bg-secondary-dark dark:text-on-secondary-dark"
	case ToneInfo:
		return "border border-info bg-info text-on-info"
	case ToneSuccess:
		return "border border-success bg-success text-on-success"
	case ToneWarning:
		return "border border-warning bg-warning text-on-warning"
	case ToneDanger:
		return "border border-danger bg-danger text-on-danger"
	default:
		return "border border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"
	}
}

// softToneClasses returns the CSS classes for soft variant
func (cfg Config) softToneClasses() string {
	switch cfg.Tone {
	case ToneInverse:
		return "border border-outline-dark bg-surface text-on-surface dark:border-outline dark:bg-surface-dark dark:text-on-surface-dark"
	case TonePrimary:
		return "border border-primary bg-surface text-on-surface-strong dark:border-primary-dark dark:bg-surface-dark dark:text-on-surface-dark-strong"
	case ToneSecondary:
		return "border border-secondary bg-surface text-on-surface-strong dark:border-secondary-dark dark:bg-surface-dark dark:text-on-surface-dark-strong"
	case ToneInfo:
		return "border border-info bg-surface text-on-surface-strong dark:border-info dark:bg-surface-dark dark:text-on-surface-dark-strong"
	case ToneSuccess:
		return "border border-success bg-surface text-on-surface-strong dark:border-success dark:bg-surface-dark dark:text-on-surface-dark-strong"
	case ToneWarning:
		return "border border-warning bg-surface text-on-surface-strong dark:border-warning dark:bg-surface-dark dark:text-on-surface-dark-strong"
	case ToneDanger:
		return "border border-danger bg-surface text-on-surface-strong dark:border-danger dark:bg-surface-dark dark:text-on-surface-dark-strong"
	default:
		return "border border-outline bg-surface text-on-surface dark:border-outline-dark dark:bg-surface-dark dark:text-on-surface-dark"
	}
}

// SoftInnerClasses returns the inner span classes for soft variant
func (cfg Config) softInnerClasses() string {
	switch cfg.Tone {
	case ToneInverse:
		return "bg-surface-dark-alt/10 dark:bg-surface-alt/10"
	case TonePrimary:
		return "bg-primary/10 dark:bg-primary-dark/10"
	case ToneSecondary:
		return "bg-secondary/10 dark:bg-secondary-dark/10"
	case ToneInfo:
		return "bg-info/10 dark:bg-info/10"
	case ToneSuccess:
		return "bg-success/10 dark:bg-success/10"
	case ToneWarning:
		return "bg-warning/10 dark:bg-warning/10"
	case ToneDanger:
		return "bg-danger/10 dark:bg-danger/10"
	default:
		return "bg-surface-alt/10 dark:bg-surface-dark-alt/10"
	}
}

// IndicatorClasses returns the indicator dot classes
func (cfg Config) indicatorClasses() string {
	if cfg.IndicatorColor != "" {
		return "size-1.5 rounded-full " + cfg.IndicatorColor
	}

	switch cfg.Tone {
	case ToneInverse:
		return "size-1.5 rounded-full bg-on-surface dark:bg-on-surface-dark"
	case TonePrimary:
		return "size-1.5 rounded-full bg-primary dark:bg-primary-dark"
	case ToneSecondary:
		return "size-1.5 rounded-full bg-secondary dark:bg-secondary-dark"
	case ToneInfo:
		return "size-1.5 rounded-full bg-info"
	case ToneSuccess:
		return "size-1.5 rounded-full bg-success"
	case ToneWarning:
		return "size-1.5 rounded-full bg-warning"
	case ToneDanger:
		return "size-1.5 rounded-full bg-danger"
	default:
		return "size-1.5 rounded-full bg-on-surface dark:bg-on-surface-dark"
	}
}

// IsSoft returns true if badge uses soft style
func (cfg Config) isSoft() bool {
	return cfg.Appearance == AppearanceSoft
}
