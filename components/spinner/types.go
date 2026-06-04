package spinner

// Variant represents spinner color variants
type Variant string

const (
	Default   Variant = "default"
	Primary   Variant = "primary"
	Secondary Variant = "secondary"
	Info      Variant = "info"
	Success   Variant = "success"
	Warning   Variant = "warning"
	Danger    Variant = "danger"
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
	// Variant determines the color scheme
	Variant Variant
	// Size of the spinner
	Size Size
	// Class allows additional CSS classes
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
	switch cfg.Variant {
	case Primary:
		return "fill-primary dark:fill-primary-dark"
	case Secondary:
		return "fill-secondary dark:fill-secondary-dark"
	case Info:
		return "fill-info dark:fill-info"
	case Success:
		return "fill-success dark:fill-success"
	case Warning:
		return "fill-warning dark:fill-warning"
	case Danger:
		return "fill-danger dark:fill-danger"
	default:
		return "fill-on-surface dark:fill-on-surface-dark"
	}
}
