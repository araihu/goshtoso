package spinner

// Tone represents spinner color variants
type Tone string

const (
	ToneDefault   Tone = "default"
	TonePrimary   Tone = "primary"
	ToneSecondary Tone = "secondary"
	ToneInfo      Tone = "info"
	ToneSuccess   Tone = "success"
	ToneWarning   Tone = "warning"
	ToneDanger    Tone = "danger"
)

// Size represents spinner size
type Size string

const (
	SizeSM Size = "sm"
	SizeMD Size = "md" // Default
	SizeLG Size = "lg"
	SizeXL Size = "xl"
)

// Config holds configuration for the spinner component
type Config struct {
	// Tone determines the color scheme
	Tone Tone
	// Size of the spinner
	Size Size
	// RootClass allows additional CSS classes on the spinner root.
	RootClass string
}

// SizeClasses returns the CSS size class for the spinner
func (cfg Config) SizeClasses() string {
	switch cfg.Size {
	case SizeSM:
		return "size-4"
	case SizeLG:
		return "size-8"
	case SizeXL:
		return "size-12"
	default:
		return "size-5"
	}
}

// FillClasses returns the CSS fill classes for the spinner variant
func (cfg Config) FillClasses() string {
	switch cfg.Tone {
	case TonePrimary:
		return "fill-primary dark:fill-primary-dark"
	case ToneSecondary:
		return "fill-secondary dark:fill-secondary-dark"
	case ToneInfo:
		return "fill-info dark:fill-info"
	case ToneSuccess:
		return "fill-success dark:fill-success"
	case ToneWarning:
		return "fill-warning dark:fill-warning"
	case ToneDanger:
		return "fill-danger dark:fill-danger"
	default:
		return "fill-on-surface dark:fill-on-surface-dark"
	}
}
