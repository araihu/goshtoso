package toggle

import "github.com/a-h/templ"

// Variant represents toggle color variants
type Variant string

const (
	Primary   Variant = "primary"
	Secondary Variant = "secondary"
	Info      Variant = "info"
	Success   Variant = "success"
	Warning   Variant = "warning"
	Danger    Variant = "danger"
)

// Style represents toggle layout style
type Style string

const (
	StyleDefault   Style = "default"   // Inline toggle with label
	StyleContainer Style = "container" // Toggle wrapped in bordered container
)

// Config holds configuration for the toggle component
type Config struct {
	// ID is the unique identifier for the toggle input
	ID string
	// Label is the text label displayed next to the toggle
	Label string
	// Variant determines the checked color scheme (default: Primary)
	Variant Variant
	// Style determines the layout style (default or container)
	Style Style
	// Checked sets the initial checked state
	Checked bool
	// Disabled disables the toggle
	Disabled bool
	// Name is the form field name
	Name string
	// Value makes the checkbox submit this value when checked, turning the toggle into a real form control.
	// Requires Name; when set, the always-off hidden input is omitted.
	Value string
	// Class allows additional CSS classes on the label
	RootClass string
	// Attrs are extra attributes applied to the <input> element
	// (e.g. x-on:change, x-bind:checked for Alpine binding).
	// Note: "checked" and "disabled" are already set from Config — use x-bind:checked / x-bind:disabled in Attrs for dynamic control rather than passing raw checked/disabled keys.
	InputAttrs templ.Attributes
}

// ToggleClasses returns the CSS classes for the toggle track div
func (cfg Config) ToggleClasses() string {
	base := "relative h-6 w-11 after:h-5 after:w-5 peer-checked:after:translate-x-5 rounded-full border border-outline after:absolute after:bottom-0 after:left-[0.0625rem] after:top-0 after:my-auto after:rounded-full after:bg-on-surface after:transition-all after:content-[''] peer-focus:outline-2 peer-focus:outline-offset-2 peer-focus:outline-outline-strong peer-active:outline-offset-0 peer-disabled:cursor-not-allowed peer-disabled:opacity-70 dark:border-outline-dark dark:after:bg-on-surface-dark dark:peer-focus:outline-outline-dark-strong"

	switch cfg.Style {
	case StyleContainer:
		base += " bg-surface dark:bg-surface-dark"
	default:
		base += " bg-surface-alt dark:bg-surface-dark-alt"
	}

	base += " " + cfg.checkedClasses()

	return base
}

// checkedClasses returns the peer-checked classes for the variant
func (cfg Config) checkedClasses() string {
	switch cfg.Variant {
	case Secondary:
		return "peer-checked:bg-secondary peer-checked:after:bg-on-secondary peer-focus:peer-checked:outline-secondary dark:peer-checked:bg-secondary-dark dark:peer-checked:after:bg-on-secondary-dark dark:peer-focus:peer-checked:outline-secondary-dark"
	case Info:
		return "peer-checked:bg-info peer-checked:after:bg-on-info peer-focus:peer-checked:outline-info dark:peer-checked:bg-info dark:peer-checked:after:bg-on-info dark:peer-focus:peer-checked:outline-info"
	case Success:
		return "peer-checked:bg-success peer-checked:after:bg-on-success peer-focus:peer-checked:outline-success dark:peer-checked:bg-success dark:peer-checked:after:bg-on-success dark:peer-focus:peer-checked:outline-success"
	case Warning:
		return "peer-checked:bg-warning peer-checked:after:bg-on-warning peer-focus:peer-checked:outline-warning dark:peer-checked:bg-warning dark:peer-checked:after:bg-on-warning dark:peer-focus:peer-checked:outline-warning"
	case Danger:
		return "peer-checked:bg-danger peer-checked:after:bg-on-danger peer-focus:peer-checked:outline-danger dark:peer-checked:bg-danger dark:peer-checked:after:bg-on-danger dark:peer-focus:peer-checked:outline-danger"
	default: // Primary
		return "peer-checked:bg-primary peer-checked:after:bg-on-primary peer-focus:peer-checked:outline-primary dark:peer-checked:bg-primary-dark dark:peer-checked:after:bg-on-primary-dark dark:peer-focus:peer-checked:outline-primary-dark"
	}
}

// LabelClasses returns the CSS classes for the label container
func (cfg Config) LabelClasses() string {
	base := "inline-flex items-center gap-3"

	if cfg.Style == StyleContainer {
		base = "inline-flex min-w-52 items-center justify-between gap-3 rounded-radius border border-outline bg-surface-alt px-4 py-1.5 dark:border-outline-dark dark:bg-surface-dark-alt"
	}

	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}

	return base
}
