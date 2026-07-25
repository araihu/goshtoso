package checkbox

// Tone represents checkbox color variants
type Tone string

const (
	TonePrimary   Tone = "primary"
	ToneSecondary Tone = "secondary"
	ToneInfo      Tone = "info"
	ToneSuccess   Tone = "success"
	ToneWarning   Tone = "warning"
	ToneDanger    Tone = "danger"
)

// Icon represents custom checkbox icon types
type Icon string

const (
	IconCheck Icon = "check" // Default checkmark
	IconXmark Icon = "xmark" // X mark
	IconMinus Icon = "minus" // Minus/dash
	IconPlus  Icon = "plus"  // Plus sign
)

// Animation represents checkbox animation styles
type Animation string

const (
	AnimationNone      Animation = ""           // No animation (default)
	AnimationSlideUp   Animation = "slide-up"   // Checkmark slides up
	AnimationScaleUp   Animation = "scale-up"   // Background scales up
	AnimationSlideDown Animation = "slide-down" // Background slides down
)

// Config holds configuration for a single checkbox
type Config struct {
	// ID is the unique identifier for the checkbox input
	ID string
	// Name is the form field name
	Name string
	// Value is the form field value
	Value string
	// Label is the text displayed next to the checkbox
	Label string
	// Checked sets the initial checked state
	Checked bool
	// Disabled disables the checkbox
	Disabled bool
	// Tone determines the color scheme (default: TonePrimary)
	Tone Tone
	// Icon determines the check icon (default: IconCheck)
	Icon Icon
	// Animation sets the animation style
	Animation Animation
	// HelperText adds helper text below the label
	HelperText string
	// HelperTextID is the ID for the helper text element (for aria-describedby)
	HelperTextID string
	// Container wraps the checkbox in a bordered container with label on the left
	Container bool
}

// GroupConfig holds configuration for a group of checkboxes
type GroupConfig struct {
	// Title is the heading above the group
	Title string
	// Items are the checkboxes in the group
	Items []Config
}

// iconPath returns the SVG path data for the configured icon
func (cfg Config) iconPath() string {
	switch cfg.Icon {
	case IconXmark:
		return "M6 18L18 6M6 6l12 12"
	case IconMinus:
		return "M18 12H6"
	case IconPlus:
		return "M12 4.5v15m7.5-7.5h-15"
	default:
		return "M4.5 12.75l6 6 9-13.5"
	}
}

// checkedBorderClass returns the checked border color class
func (cfg Config) checkedBorderClass() string {
	switch cfg.Tone {
	case ToneSecondary:
		return "checked:border-secondary dark:checked:border-secondary-dark"
	case ToneInfo:
		return "checked:border-info dark:checked:border-info"
	case ToneSuccess:
		return "checked:border-success dark:checked:border-success"
	case ToneWarning:
		return "checked:border-warning dark:checked:border-warning"
	case ToneDanger:
		return "checked:border-danger dark:checked:border-danger"
	default:
		return "checked:border-primary dark:checked:border-primary-dark"
	}
}

// checkedBgClass returns the checked background color class
func (cfg Config) checkedBgClass() string {
	switch cfg.Tone {
	case ToneSecondary:
		return "checked:before:bg-secondary dark:checked:before:bg-secondary-dark"
	case ToneInfo:
		return "checked:before:bg-info dark:checked:before:bg-info"
	case ToneSuccess:
		return "checked:before:bg-success dark:checked:before:bg-success"
	case ToneWarning:
		return "checked:before:bg-warning dark:checked:before:bg-warning"
	case ToneDanger:
		return "checked:before:bg-danger dark:checked:before:bg-danger"
	default:
		return "checked:before:bg-primary dark:checked:before:bg-primary-dark"
	}
}

// focusCheckedClass returns the focus outline color when checked
func (cfg Config) focusCheckedClass() string {
	switch cfg.Tone {
	case ToneSecondary:
		return "checked:focus:outline-secondary dark:checked:focus:outline-secondary-dark"
	case ToneInfo:
		return "checked:focus:outline-info dark:checked:focus:outline-info"
	case ToneSuccess:
		return "checked:focus:outline-success dark:checked:focus:outline-success"
	case ToneWarning:
		return "checked:focus:outline-warning dark:checked:focus:outline-warning"
	case ToneDanger:
		return "checked:focus:outline-danger dark:checked:focus:outline-danger"
	default:
		return "checked:focus:outline-primary dark:checked:focus:outline-primary-dark"
	}
}

// svgTextClass returns the text color class for the SVG icon
func (cfg Config) svgTextClass() string {
	switch cfg.Tone {
	case ToneSecondary:
		return "text-on-secondary dark:text-on-secondary-dark"
	case ToneInfo:
		return "text-on-info dark:text-on-info-dark"
	case ToneSuccess:
		return "text-on-success dark:text-on-success-dark"
	case ToneWarning:
		return "text-on-warning dark:text-on-warning-dark"
	case ToneDanger:
		return "text-on-danger dark:text-on-danger-dark"
	default:
		return "text-on-primary dark:text-on-primary-dark"
	}
}

// InputClasses returns the full CSS class string for the checkbox input
func (cfg Config) inputClasses() string {
	base := "before:content[''] peer relative size-4 appearance-none overflow-hidden rounded-sm border border-outline bg-surface-alt before:absolute before:inset-0 focus:outline-2 focus:outline-offset-2 focus:outline-outline-strong active:outline-offset-0 disabled:cursor-not-allowed dark:border-outline-dark dark:bg-surface-dark-alt dark:focus:outline-outline-dark-strong"

	if cfg.Container {
		// Container variant uses bg-surface instead of bg-surface-alt
		base = "before:content[''] peer relative size-4 appearance-none overflow-hidden rounded-sm border border-outline bg-surface before:absolute before:inset-0 focus:outline-2 focus:outline-offset-2 focus:outline-outline-strong active:outline-offset-0 disabled:cursor-not-allowed dark:border-outline-dark dark:bg-surface-dark dark:focus:outline-outline-dark-strong"
	}

	classes := base + " " + cfg.checkedBorderClass() + " " + cfg.checkedBgClass() + " " + cfg.focusCheckedClass()

	// Add animation-specific classes
	switch cfg.Animation {
	case AnimationScaleUp:
		classes += " before:scale-0 before:rounded-full before:transition before:duration-200 checked:before:scale-125"
	case AnimationSlideDown:
		classes += " before:-translate-y-4 before:transition before:duration-200 checked:before:translate-y-0"
	}

	return classes
}

// SvgClasses returns the CSS class string for the SVG icon
func (cfg Config) svgClasses() string {
	base := "pointer-events-none invisible absolute left-1/2 top-1/2 size-3 -translate-x-1/2 -translate-y-1/2 peer-checked:visible"

	classes := base + " " + cfg.svgTextClass()

	// Add animation-specific classes
	switch cfg.Animation {
	case AnimationSlideUp:
		classes = "pointer-events-none invisible absolute left-1/2 top-1/2 size-3 -translate-x-1/2 -translate-y-1/4 peer-checked:-translate-y-1/2 transition duration-200 peer-checked:visible " + cfg.svgTextClass()
	case AnimationScaleUp:
		classes = "pointer-events-none invisible absolute left-1/2 top-1/2 size-3 -translate-x-1/2 -translate-y-1/2 scale-0 transition duration-200 delay-200 peer-checked:scale-100 peer-checked:visible " + cfg.svgTextClass()
	case AnimationSlideDown:
		classes = "pointer-events-none invisible absolute left-1/2 top-1/2 size-3 -translate-y-1/2 -translate-x-1/2 opacity-0 transition delay-200 duration-200 peer-checked:visible peer-checked:opacity-100 " + cfg.svgTextClass()
	}

	return classes
}
