package combobox

import "slices"

// SelectionMode determines if combobox supports single or multi-select
type SelectionMode string

const (
	ModeSingle   SelectionMode = "single"
	ModeMultiple SelectionMode = "multiple"
)

// Size represents combobox size variants
type Size string

const (
	SizeSM Size = "sm"
	SizeMD Size = "md" // default
	SizeLG Size = "lg"
)

// Option represents a selectable item in the combobox dropdown.
type Option struct {
	// Value is the form submission value.
	Value string
	// Label is the display text shown to the user.
	Label string
	// Img is an optional image URL (e.g. avatar, flag).
	Img string
	// Initials are displayed as avatar fallback when Img is empty (e.g. "JD").
	Initials string
	// Meta is optional secondary text (e.g. email, description).
	Meta string
	// Badge is an optional small label rendered next to the option (e.g. "Team", "User").
	Badge string
	// BadgeColor determines the badge color variant (info, success, warning, danger, neutral).
	BadgeColor string
}

// ToOptions converts a slice of any type into []Option using the provided accessor functions.
func ToOptions[T any](items []T, valueFn func(T) string, labelFn func(T) string) []Option {
	opts := make([]Option, len(items))
	for i, item := range items {
		opts[i] = Option{
			Value: valueFn(item),
			Label: labelFn(item),
		}
	}
	return opts
}

// SelectedValues returns the values of options that match the selected value.
// Useful for converting a single selected string into the []string expected by Config.Selected.
func SelectedValues(selected string) []string {
	if selected == "" {
		return nil
	}
	return []string{selected}
}

// State represents the validation state of the combobox
type State string

const (
	StateDefault State = ""
	StateError   State = "error"
	StateSuccess State = "success"
)

// Config holds configuration for the combobox
type Config struct {
	// ID is a unique identifier for the combobox
	ID string
	// Name is the form field name
	Name string
	// Label is the label text shown above the combobox
	Label string
	// Placeholder text when no selection
	Placeholder string
	// State is the validation state (error, success, or default)
	State State
	// Options is the list of available options (for static data)
	Options []Option
	// Selected contains the currently selected values
	Selected []string
	// Mode determines single or multi-select
	Mode SelectionMode
	// EnableSearch adds a search field to filter options
	EnableSearch bool
	// SearchPlaceholder text for the search field
	SearchPlaceholder string
	// EnableClearAll shows a "Clear all" button when items are selected
	EnableClearAll bool
	// ClearAllText is the text for the clear all button (default: "Clear all")
	ClearAllText string
	// Size of the combobox
	Size Size
	// Disabled disables the combobox
	Disabled bool
	// HTMX endpoint for lazy loading options
	// When set, options are loaded from the server via HTMX
	HXMTEndpoint string
	// HTMXTrigger for loading options (default: "click")
	HTMXTrigger string
	// HTMXSearchParam is the query param name for search (default: "search")
	HTMXSearchParam string
	// Class allows additional CSS classes
	Class string
	// NoResultsText shown when search returns no matches (default: "No matches found")
	NoResultsText string
}

// SizeClasses returns CSS classes for the combobox size
func (cfg Config) SizeClasses() string {
	switch cfg.Size {
	case SizeSM:
		return "min-w-32"
	case SizeLG:
		return "min-w-48"
	default:
		return "min-w-40"
	}
}

// TriggerClasses returns CSS classes for the trigger button
func (cfg Config) TriggerClasses() string {
	base := "inline-flex w-full items-center justify-between gap-2 rounded-radius border px-3 py-2 text-sm font-medium transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark"

	if cfg.Disabled {
		return base + " border-outline bg-surface-alt/50 text-on-surface/50 cursor-not-allowed dark:border-outline-dark dark:bg-surface-dark-alt/30 dark:text-on-surface-dark/50"
	}

	switch cfg.State {
	case StateError:
		return base + " border-danger"
	case StateSuccess:
		return base + " border-success"
	default:
		return base
	}
}

// LabelClasses returns CSS classes for the label based on validation state
func (cfg Config) LabelClasses() string {
	base := "block pb-1 text-sm text-on-surface dark:text-on-surface-dark"
	switch cfg.State {
	case StateError:
		return "flex w-fit gap-1 pb-1 text-sm text-danger"
	case StateSuccess:
		return "flex w-fit gap-1 pb-1 text-sm text-success"
	default:
		return base
	}
}

// TriggerStateClasses returns dynamic classes based on selection state
func (cfg Config) TriggerStateClasses(selected bool) string {
	if selected {
		return "border-secondary bg-secondary/10 text-on-surface-strong dark:border-secondary-dark dark:bg-secondary-dark/15 dark:text-on-surface-dark-strong"
	}
	return "border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt/50 dark:text-on-surface-dark"
}

// DropdownClasses returns CSS classes for the dropdown
func (cfg Config) DropdownClasses() string {
	return "absolute left-0 top-full z-30 mt-1 min-w-full overflow-hidden rounded-radius border border-outline bg-surface-alt shadow-lg dark:border-outline-dark dark:bg-surface-dark-alt"
}

// OptionClasses returns CSS classes for option items
func (cfg Config) OptionClasses() string {
	return "combobox-option inline-flex justify-between items-center gap-4 bg-surface-alt px-4 py-2 text-sm text-on-surface hover:bg-surface-dark-alt/5 hover:text-on-surface-strong focus-visible:bg-surface-dark-alt/5 focus-visible:text-on-surface-strong focus-visible:outline-hidden dark:bg-surface-dark-alt dark:text-on-surface-dark dark:hover:bg-surface-alt/5 dark:hover:text-on-surface-dark-strong dark:focus-visible:bg-surface-alt/10 dark:focus-visible:text-on-surface-dark-strong cursor-pointer"
}

// CheckboxClasses returns CSS classes for checkbox input
func (cfg Config) CheckboxClasses() string {
	return "peer relative size-4 appearance-none overflow-hidden rounded-sm border border-outline bg-surface-alt before:absolute before:inset-0 before:content-[''] checked:border-primary checked:before:bg-primary focus:outline-2 focus:outline-offset-2 focus:outline-outline-strong checked:focus:outline-primary active:outline-offset-0 dark:border-outline-dark dark:bg-surface-dark-alt dark:checked:border-primary-dark dark:checked:before:bg-primary-dark"
}

// SearchFieldClasses returns CSS classes for the search input
func (cfg Config) SearchFieldClasses() string {
	return "w-full bg-transparent py-2.5 pl-9 pr-3 text-sm text-on-surface placeholder:text-on-surface/40 focus:outline-none dark:text-on-surface-dark dark:placeholder:text-on-surface-dark/40"
}

// IsMultiple returns true if the combobox supports multi-select
func (cfg Config) IsMultiple() bool {
	return cfg.Mode == ModeMultiple
}

// HasSelection returns true if there are selected values
func (cfg Config) HasSelection() bool {
	return len(cfg.Selected) > 0
}

// IsSelected returns true if the given value is selected
func (cfg Config) IsSelected(value string) bool {
	return slices.Contains(cfg.Selected, value)
}

// GetSearchPlaceholder returns the search placeholder text
func (cfg Config) GetSearchPlaceholder() string {
	if cfg.SearchPlaceholder != "" {
		return cfg.SearchPlaceholder
	}
	return "Search..."
}

// GetNoResultsText returns the no results text
func (cfg Config) GetNoResultsText() string {
	if cfg.NoResultsText != "" {
		return cfg.NoResultsText
	}
	return "No matches found"
}

// GetClearAllText returns the clear all button text
func (cfg Config) GetClearAllText() string {
	if cfg.ClearAllText != "" {
		return cfg.ClearAllText
	}
	return "Clear all"
}

// GetHTMXTrigger returns the HTMX trigger
func (cfg Config) GetHTMXTrigger() string {
	if cfg.HTMXTrigger != "" {
		return cfg.HTMXTrigger
	}
	return "click"
}

// GetHTMXSearchParam returns the search param name
func (cfg Config) GetHTMXSearchParam() string {
	if cfg.HTMXSearchParam != "" {
		return cfg.HTMXSearchParam
	}
	return "search"
}

// GetPlaceholder returns the placeholder text
func (cfg Config) GetPlaceholder() string {
	if cfg.Placeholder != "" {
		return cfg.Placeholder
	}
	return "Please Select"
}
