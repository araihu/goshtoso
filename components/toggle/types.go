package toggle

import "github.com/a-h/templ"

// Tone represents toggle color variants
type Tone string

const (
	TonePrimary   Tone = "primary"
	ToneSecondary Tone = "secondary"
	ToneInfo      Tone = "info"
	ToneSuccess   Tone = "success"
	ToneWarning   Tone = "warning"
	ToneDanger    Tone = "danger"
)

// Appearance represents the toggle layout treatment.
type Appearance string

const (
	AppearanceDefault   Appearance = ""          // Inline toggle with label
	AppearanceContainer Appearance = "container" // Toggle wrapped in bordered container
)

// Config holds configuration for the toggle component
type Config struct {
	// ID is the unique identifier for the toggle input
	ID string
	// Label is the text label displayed next to the toggle
	Label string
	// Tone determines the checked color scheme (default: TonePrimary)
	Tone Tone
	// Appearance determines the layout treatment (default or container).
	Appearance Appearance
	// Checked sets the initial checked state
	Checked bool
	// Disabled disables the toggle
	Disabled bool
	// Name is the form field name
	Name string
	// Value makes the checkbox submit this value when checked, turning the toggle into a real form control.
	// Requires Name; when set, the always-off hidden input is omitted.
	Value string
	// RootClass allows additional CSS classes on the label.
	RootClass string
	// InputAttrs are extra attributes applied to the <input> element.
	// (e.g. x-on:change, x-bind:checked for Alpine binding).
	// Note: "checked" and "disabled" are already set from Config — use x-bind:checked / x-bind:disabled in InputAttrs for dynamic control rather than passing raw checked/disabled keys.
	InputAttrs templ.Attributes
}

// ToggleClasses returns the CSS classes for the toggle track div
func (cfg Config) toggleClasses() string {
	base := "relative h-6 w-11 after:h-5 after:w-5 peer-checked:after:translate-x-5 rounded-full border border-control-outline after:absolute after:bottom-0 after:left-[0.0625rem] after:top-0 after:my-auto after:rounded-full after:bg-on-surface after:transition-all after:content-[''] peer-focus:outline-2 peer-focus:outline-offset-2 peer-focus:outline-outline-strong peer-active:outline-offset-0 peer-disabled:cursor-not-allowed dark:border-control-outline-dark dark:after:bg-on-surface-dark dark:peer-focus:outline-outline-dark-strong"

	switch cfg.Appearance {
	case AppearanceContainer:
		base += " bg-surface dark:bg-surface-dark"
	default:
		base += " bg-surface-alt dark:bg-surface-dark-alt"
	}

	base += " " + cfg.checkedClasses()

	return base
}

// checkedClasses returns the peer-checked classes for the variant
func (cfg Config) checkedClasses() string {
	switch cfg.Tone {
	case ToneSecondary:
		return "peer-checked:bg-secondary peer-checked:after:bg-on-secondary peer-focus:peer-checked:outline-secondary dark:peer-checked:bg-secondary-dark dark:peer-checked:after:bg-on-secondary-dark dark:peer-focus:peer-checked:outline-secondary-dark"
	case ToneInfo:
		return "peer-checked:bg-info peer-checked:after:bg-on-info peer-focus:peer-checked:outline-info dark:peer-checked:bg-info dark:peer-checked:after:bg-on-info dark:peer-focus:peer-checked:outline-info"
	case ToneSuccess:
		return "peer-checked:bg-success peer-checked:after:bg-on-success peer-focus:peer-checked:outline-success dark:peer-checked:bg-success dark:peer-checked:after:bg-on-success dark:peer-focus:peer-checked:outline-success"
	case ToneWarning:
		return "peer-checked:bg-warning peer-checked:after:bg-on-warning peer-focus:peer-checked:outline-warning dark:peer-checked:bg-warning dark:peer-checked:after:bg-on-warning dark:peer-focus:peer-checked:outline-warning"
	case ToneDanger:
		return "peer-checked:bg-danger peer-checked:after:bg-on-danger peer-focus:peer-checked:outline-danger dark:peer-checked:bg-danger dark:peer-checked:after:bg-on-danger dark:peer-focus:peer-checked:outline-danger"
	default: // TonePrimary
		return "peer-checked:bg-primary peer-checked:after:bg-on-primary peer-focus:peer-checked:outline-primary dark:peer-checked:bg-primary-dark dark:peer-checked:after:bg-on-primary-dark dark:peer-focus:peer-checked:outline-primary-dark"
	}
}

// LabelClasses returns the CSS classes for the label container
func (cfg Config) labelClasses() string {
	base := "inline-flex items-center gap-3"

	if cfg.Appearance == AppearanceContainer {
		base = "inline-flex min-w-52 items-center justify-between gap-3 rounded-radius border border-outline bg-surface-alt px-4 py-1.5 dark:border-outline-dark dark:bg-surface-dark-alt"
	}

	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}

	return base
}
