package selectfield

import (
	"encoding/base64"
	"encoding/json"
	"maps"
	"strings"

	"github.com/a-h/templ"
)

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
func toOptions[T any](items []T, valueFn func(T) string, labelFn func(T) string, selected string) []Option {
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

// AlpineConfig wires client-side Alpine bindings.
type AlpineConfig struct {
	Model        string
	BindDisabled string
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
	// Required exposes accessible required state on the composite trigger.
	// Enforce selection in server validation because the submitted value is
	// represented by a hidden text control.
	Required bool
	// Disabled disables the select
	Disabled bool
	// Autocomplete sets the autocomplete attribute
	Autocomplete string
	// RootClass allows additional CSS classes on the wrapper.
	RootClass string
	// Alpine wires client-side state.
	Alpine *AlpineConfig
	// Readonly renders the select as disabled (grayed out) + hidden input with value so it still submits
	Readonly bool
	// InputAttrs allows arbitrary HTML attributes on the hidden submission input.
	// To restore a draft from external JavaScript, set this input's value and
	// dispatch a bubbling input or change event; Select synchronizes its visible
	// value and live option state from either standard event.
	InputAttrs templ.Attributes
	// TriggerAttrs appends non-conflicting HTML attributes to the focusable
	// combobox trigger. Use it for ARIA relationships and event hooks.
	TriggerAttrs templ.Attributes
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
func (cfg Config) containerClasses() string {
	base := "relative flex w-full flex-col gap-1 text-on-surface dark:text-on-surface-dark"
	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}
	return base
}

// SelectClasses returns CSS classes for the select element (legacy, kept for compatibility)
func (cfg Config) selectClasses() string {
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
func (cfg Config) triggerClasses() string {
	base := "inline-flex w-full items-center justify-between gap-2 rounded-radius border px-4 py-2 text-sm transition hover:contrast-125 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:contrast-100"

	if cfg.isEffectivelyDisabled() {
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
func (cfg Config) labelClasses() string {
	base := "w-fit pl-0.5 text-sm"

	switch cfg.State {
	case StateError:
		return "flex w-fit gap-1 pl-0.5 text-sm text-danger-text dark:text-danger-text-dark"
	case StateSuccess:
		return "flex w-fit gap-1 pl-0.5 text-sm text-success-text dark:text-success-text-dark"
	default:
		return base
	}
}

// GetPlaceholder returns the placeholder text
func (cfg Config) getPlaceholder() string {
	if cfg.Placeholder != "" {
		return cfg.Placeholder
	}
	return "Please Select"
}

// SelectedValue returns the value of the first selected option, or empty string
func (cfg Config) selectedValue() string {
	for _, opt := range cfg.Options {
		if opt.Selected {
			return opt.Value
		}
	}
	return ""
}

// IsEffectivelyDisabled returns true if the select should render as disabled (Disabled or Readonly)
func (cfg Config) isEffectivelyDisabled() bool {
	return cfg.Disabled || cfg.Readonly
}

func (cfg Config) triggerAttributes() templ.Attributes {
	attrs := make(templ.Attributes, len(cfg.TriggerAttrs)+2)
	maps.Copy(attrs, cfg.TriggerAttrs)
	if cfg.HelperText != "" && cfg.ID != "" {
		existing, _ := attrs["aria-describedby"].(string)
		attrs["aria-describedby"] = mergeAttributeTokens(existing, cfg.ID+"-helper")
	}
	if cfg.State == StateError {
		attrs["aria-invalid"] = "true"
	}
	if cfg.Required {
		if _, exists := attrs["aria-required"]; !exists {
			attrs["aria-required"] = "true"
		}
	}
	return attrs
}

func mergeAttributeTokens(values ...string) string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		for token := range strings.FieldsSeq(value) {
			if !seen[token] {
				seen[token] = true
				result = append(result, token)
			}
		}
	}
	return strings.Join(result, " ")
}

type factoryData struct {
	Placeholder    string          `json:"placeholder"`
	Options        []factoryOption `json:"options"`
	SelectedValues []string        `json:"selectedValues"`
	ActiveIndex    int             `json:"activeIndex"`
	AlpineModel    string          `json:"alpineModel,omitempty"`
}

type factoryOption struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Selected bool   `json:"selected"`
}

// factoryDataJSON serializes Select's non-executable state for its factory.
func (cfg Config) factoryDataJSON() string {
	data := factoryData{
		Placeholder:    cfg.getPlaceholder(),
		Options:        []factoryOption{},
		SelectedValues: []string{},
	}
	for _, option := range cfg.Options {
		data.Options = append(data.Options, factoryOption(option))
	}
	for index, option := range cfg.Options {
		if option.Selected {
			data.SelectedValues = []string{option.Value}
			data.ActiveIndex = index
			break
		}
	}
	if cfg.Alpine != nil {
		data.AlpineModel = cfg.Alpine.Model
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return `{"placeholder":"Please Select","options":[],"selectedValues":[],"activeIndex":0}`
	}
	return string(encoded)
}

func (cfg Config) factoryData() string {
	return base64.StdEncoding.EncodeToString([]byte(cfg.factoryDataJSON()))
}

func jsEscapeSingle(s string) string {
	var result strings.Builder
	for _, c := range s {
		switch c {
		case '\'':
			result.WriteString(`\'`)
		case '\\':
			result.WriteString(`\\`)
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\t':
			result.WriteString(`\t`)
		case '\u2028':
			result.WriteString(`\u2028`)
		case '\u2029':
			result.WriteString(`\u2029`)
		default:
			result.WriteString(string(c))
		}
	}
	return result.String()
}
