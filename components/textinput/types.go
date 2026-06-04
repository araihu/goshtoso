package textinput

import (
	"strconv"

	"github.com/a-h/templ"
)

// State represents the validation state of the input
type State string

const (
	StateDefault State = ""
	StateError   State = "error"
	StateSuccess State = "success"
)

// InputType represents the HTML input type
type InputType string

const (
	TypeText          InputType = "text"
	TypePassword      InputType = "password"
	TypeSearch        InputType = "search"
	TypeEmail         InputType = "email"
	TypeTel           InputType = "tel"
	TypeURL           InputType = "url"
	TypeNumber        InputType = "number"
	TypeDate          InputType = "date"
	TypeDateTimeLocal InputType = "datetime-local"
)

// Config holds configuration for the text input component
type Config struct {
	// ID is a unique identifier for the input
	ID string
	// Name is the form field name
	Name string
	// Label is the label text shown above the input
	Label string
	// Placeholder text when input is empty
	Placeholder string
	// Value is the current input value
	Value string
	// Type is the HTML input type (default: "text")
	Type InputType
	// State represents validation state (error, success, or default)
	State State
	// HelperText is optional text shown below the input (used for error/success messages)
	HelperText string
	// Disabled disables the input
	Disabled bool
	// Required marks the input as required
	Required bool
	// Autocomplete is the HTML autocomplete attribute value
	Autocomplete string
	// Mask is an Alpine.js x-mask pattern (e.g. "(999) 999-9999")
	Mask string
	// Pattern is an HTML pattern attribute for regex validation (e.g. "[a-z0-9]{6}")
	Pattern string
	// MaxLength limits the number of characters (0 = no limit)
	MaxLength int
	// Readonly makes the input non-editable but still submittable
	Readonly bool
	// Class allows additional CSS classes on the container
	RootClass string
	// Attrs allows arbitrary HTML attributes on the <input> element (e.g., hx-post, hx-indicator)
	InputAttrs templ.Attributes
}

// GetType returns the input type, defaulting to "text"
func (cfg Config) GetType() InputType {
	if cfg.Type != "" {
		return cfg.Type
	}
	return TypeText
}

// InputClasses returns the CSS classes for the input element
func (cfg Config) InputClasses() string {
	base := "w-full rounded-radius border bg-surface px-2 py-2 text-sm text-on-surface-strong placeholder:text-on-surface-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary disabled:cursor-not-allowed disabled:bg-surface-alt disabled:text-on-surface-muted disabled:placeholder:text-on-surface-muted disabled:opacity-50 dark:bg-surface-dark dark:text-on-surface-dark-strong dark:placeholder:text-on-surface-dark-muted dark:focus-visible:outline-primary-dark dark:disabled:bg-surface-dark-alt dark:disabled:text-on-surface-dark-muted dark:disabled:placeholder:text-on-surface-dark-muted"

	switch cfg.State {
	case StateError:
		return base + " border-danger"
	case StateSuccess:
		return base + " border-success"
	default:
		return base + " border-outline dark:border-outline-dark"
	}
}

// LabelClasses returns the CSS classes for the label element
func (cfg Config) LabelClasses() string {
	base := "w-fit pl-0.5 text-sm"

	switch cfg.State {
	case StateError:
		return "flex " + base + " items-center gap-1 text-danger"
	case StateSuccess:
		return "flex " + base + " items-center gap-1 text-success"
	default:
		return base
	}
}

// HelperTextClasses returns the CSS classes for the helper text
func (cfg Config) HelperTextClasses() string {
	switch cfg.State {
	case StateError:
		return "pl-0.5 text-xs text-danger"
	case StateSuccess:
		return "pl-0.5 text-xs text-success"
	default:
		return "pl-0.5 text-xs text-on-surface/60 dark:text-on-surface-dark/60"
	}
}

// ContainerClasses returns the CSS classes for the outer container.
// Width is determined by the parent layout — no max-width is imposed.
func (cfg Config) ContainerClasses() string {
	base := "flex w-full flex-col gap-1 text-on-surface dark:text-on-surface-dark"
	if cfg.RootClass != "" {
		return base + " " + cfg.RootClass
	}
	return base
}

// IsPassword returns true if the input type is password
func (cfg Config) IsPassword() bool {
	return cfg.Type == TypePassword
}

// IsSearch returns true if the input type is search
func (cfg Config) IsSearch() bool {
	return cfg.Type == TypeSearch
}

// HasMask returns true if a mask pattern is set
func (cfg Config) HasMask() bool {
	return cfg.Mask != ""
}

// HasPattern returns true if a pattern is set
func (cfg Config) HasPattern() bool {
	return cfg.Pattern != ""
}

// HasMaxLength returns true if a max length is set
func (cfg Config) HasMaxLength() bool {
	return cfg.MaxLength > 0
}

// MaxLengthStr returns MaxLength as a string for the HTML attribute
func (cfg Config) MaxLengthStr() string {
	return strconv.Itoa(cfg.MaxLength)
}
