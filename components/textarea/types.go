package textarea

import "github.com/a-h/templ"

// State represents textarea validation state
type State string

const (
	StateDefault State = ""
	StateError   State = "error"
	StateSuccess State = "success"
)

// Config holds configuration for the textarea component
type Config struct {
	// ID is a unique identifier for the textarea
	ID string
	// Name is the form field name
	Name string
	// Label is the label text above the textarea
	Label string
	// Placeholder text when empty
	Placeholder string
	// Value is the current textarea content
	Value string
	// Rows is the number of visible text rows (default: 3)
	Rows int
	// Disabled disables the textarea
	Disabled bool
	// ReadOnly makes the textarea read-only
	ReadOnly bool
	// State is the validation state (default/error/success)
	State State
	// HelperText is the helper or error text below the textarea
	HelperText string
	// Class allows additional CSS classes on the container
	Class string
	// Attrs allows arbitrary HTML attributes on the <textarea> element (e.g. onkeydown, hx-*).
	Attrs templ.Attributes
}

// ContainerClasses returns CSS classes for the outer container.
// Width is determined by the parent layout — no max-width is imposed.
func (cfg Config) ContainerClasses() string {
	base := "flex w-full flex-col gap-1 text-on-surface dark:text-on-surface-dark"
	if cfg.Class != "" {
		return base + " " + cfg.Class
	}
	return base
}

// LabelClasses returns CSS classes for the label element
func (cfg Config) LabelClasses() string {
	switch cfg.State {
	case StateError:
		return "flex w-fit items-center gap-1 pl-0.5 text-sm text-danger"
	case StateSuccess:
		return "flex w-fit items-center gap-1 pl-0.5 text-sm text-success"
	default:
		return "w-fit pl-0.5 text-sm"
	}
}

// TextareaClasses returns CSS classes for the textarea element
func (cfg Config) TextareaClasses() string {
	base := "w-full rounded-radius border bg-surface px-2.5 py-2 text-sm text-on-surface-strong placeholder:text-on-surface-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary disabled:cursor-not-allowed disabled:bg-surface-alt disabled:text-on-surface-muted disabled:placeholder:text-on-surface-muted disabled:opacity-50 dark:bg-surface-dark dark:text-on-surface-dark-strong dark:placeholder:text-on-surface-dark-muted dark:focus-visible:outline-primary-dark dark:disabled:bg-surface-dark-alt dark:disabled:text-on-surface-dark-muted dark:disabled:placeholder:text-on-surface-dark-muted"

	switch cfg.State {
	case StateError:
		return base + " border-danger dark:border-danger"
	case StateSuccess:
		return base + " border-success dark:border-success"
	default:
		return base + " border-outline dark:border-outline-dark"
	}
}

// HelperTextClasses returns CSS classes for helper text
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

// GetRows returns the number of rows, defaulting to 3
func (cfg Config) GetRows() string {
	switch cfg.Rows {
	case 0:
		return "3"
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	case 5:
		return "5"
	case 6:
		return "6"
	case 7:
		return "7"
	case 8:
		return "8"
	case 9:
		return "9"
	case 10:
		return "10"
	default:
		return "3"
	}
}
