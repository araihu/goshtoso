package selectfield

import "github.com/a-h/templ"

// State represents the validation state of the select
type State string

const (
	StateDefault State = ""
	StateError   State = "error"
	StateSuccess State = "success"
)

// Option represents a single selectable item
type Option struct {
	Value    string
	Label    string
	Selected bool
}

// ToOptions converts a slice of any type into []Option using the provided accessor functions.
// The selected parameter is compared against each value to set the Selected flag.
//
// Example:
//
//	type Region struct { Code, Name string }
//	regions := []Region{{Code: "us-east-1", Name: "US East"}, {Code: "eu-west-1", Name: "EU West"}}
//	opts := selectfield.ToOptions(regions, func(r Region) string { return r.Code }, func(r Region) string { return r.Name }, "eu-west-1")
func ToOptions[T any](items []T, valueFn func(T) string, labelFn func(T) string, selected string) []Option {
	opts := make([]Option, len(items))
	for i, item := range items {
		v := valueFn(item)
		opts[i] = Option{
			Value:    v,
			Label:    labelFn(item),
			Selected: v == selected,
		}
	}
	return opts
}

// Config holds configuration for the select component
type Config struct {
	// ID is a unique identifier for the select element
	ID string
	// Name is the form field name
	Name string
	// Label is the label text shown above the select
	Label string
	// Placeholder text when no selection (default: "Please Select")
	Placeholder string
	// Options is the list of available options
	Options []Option
	// State is the validation state (error, success, or default)
	State State
	// HelperText is shown below the select (e.g., error or success message)
	HelperText string
	// Disabled disables the select
	Disabled bool
	// Autocomplete sets the autocomplete attribute
	Autocomplete string
	// Class allows additional CSS classes on the wrapper
	Class string
	// AlpineModel sets x-model on the select for Alpine.js binding
	AlpineModel string
	// AlpineBindDisabled sets x-bind:disabled on the select
	AlpineBindDisabled string
	// Readonly renders the select as disabled (grayed out) + hidden input with value so it still submits
	Readonly bool
	// Attrs allows arbitrary HTML attributes on the <select> element (e.g., hx-post, hx-indicator)
	Attrs templ.Attributes
	// Shell enables "shell mode": the Select renders its trigger + dropdown
	// chrome but hosts arbitrary templ children as the dropdown body instead
	// of an option list. Used to wrap custom pickers (e.g. a color palette).
	Shell bool
	// TriggerLeading is optional content rendered at the start of the trigger.
	TriggerLeading templ.Component
	// ValueExpr is an Alpine expression (x-text) for the trigger's value text
	// in shell mode. Resolves against the host page's x-data scope.
	ValueExpr string
	// TriggerLabel is optional static text shown left-aligned in the trigger
	// in shell mode (e.g. a token/role name). When set, the ValueExpr value is
	// pushed to the right and rendered muted, matching a "Name … Value" row.
	TriggerLabel string
}

// ContainerClasses returns CSS classes for the outer wrapper.
// Width is determined by the parent layout — no max-width is imposed.
func (cfg Config) ContainerClasses() string {
	base := "relative flex w-full flex-col gap-1 text-on-surface dark:text-on-surface-dark"
	if cfg.Class != "" {
		base += " " + cfg.Class
	}
	return base
}

// SelectClasses returns CSS classes for the select element (legacy, kept for compatibility)
func (cfg Config) SelectClasses() string {
	base := "w-full appearance-none rounded-radius border bg-surface px-4 py-2 text-sm text-on-surface-strong focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary disabled:cursor-not-allowed disabled:bg-surface-alt disabled:text-on-surface-muted disabled:opacity-50 dark:bg-surface-dark dark:text-on-surface-dark-strong dark:focus-visible:outline-primary-dark dark:disabled:bg-surface-dark-alt dark:disabled:text-on-surface-dark-muted"

	switch cfg.State {
	case StateError:
		return base + " border-danger"
	case StateSuccess:
		return base + " border-success"
	default:
		return base + " border-outline dark:border-outline-dark"
	}
}

// TriggerClasses returns CSS classes for the custom dropdown trigger button
func (cfg Config) TriggerClasses() string {
	base := "inline-flex w-full items-center justify-between gap-2 rounded-radius border px-4 py-2 text-sm transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:opacity-50"

	if cfg.IsEffectivelyDisabled() {
		return base + " border-outline bg-surface-alt text-on-surface-muted opacity-50 cursor-not-allowed dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark-muted"
	}

	switch cfg.State {
	case StateError:
		return base + " border-danger bg-surface text-on-surface-strong dark:bg-surface-dark dark:text-on-surface-dark-strong"
	case StateSuccess:
		return base + " border-success bg-surface text-on-surface-strong dark:bg-surface-dark dark:text-on-surface-dark-strong"
	default:
		return base + " border-outline bg-surface text-on-surface-strong dark:border-outline-dark dark:bg-surface-dark dark:text-on-surface-dark-strong"
	}
}

// LabelClasses returns CSS classes for the label
func (cfg Config) LabelClasses() string {
	base := "w-fit pl-0.5 text-sm"

	switch cfg.State {
	case StateError:
		return "flex w-fit gap-1 pl-0.5 text-sm text-danger"
	case StateSuccess:
		return "flex w-fit gap-1 pl-0.5 text-sm text-success"
	default:
		return base
	}
}

// GetPlaceholder returns the placeholder text
func (cfg Config) GetPlaceholder() string {
	if cfg.Placeholder != "" {
		return cfg.Placeholder
	}
	return "Please Select"
}

// SelectedValue returns the value of the first selected option, or empty string
func (cfg Config) SelectedValue() string {
	for _, opt := range cfg.Options {
		if opt.Selected {
			return opt.Value
		}
	}
	return ""
}

// IsEffectivelyDisabled returns true if the select should render as disabled (Disabled or Readonly)
func (cfg Config) IsEffectivelyDisabled() bool {
	return cfg.Disabled || cfg.Readonly
}

// shellData returns the slim Alpine x-data for shell mode: only open state,
// since the hosted content owns the value.
func shellData() string {
	return `{ isOpen: false, openedWithKeyboard: false }`
}
