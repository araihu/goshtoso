// Package radio renders a PenguinUI-styled radio button.
//
// # Interaction primitives
//
// Radio exposes both HTMX and Alpine.js primitives as first-class Config
// fields so callers can wire either, or both, without escaping pitfalls:
//
//   - HTMX → server roundtrip on radio change. Set fields on cfg.HTMX
//     (Get/Post/Target/Swap/...). When any HTMX verb is set and
//     cfg.HTMX.Trigger is empty, the rendered hx-trigger defaults to
//     "change" — the native event radios fire on selection.
//   - Alpine → client-side state without a server. Set fields on
//     cfg.Alpine (Model/OnChange/BindChecked/BindDisabled/Data).
//   - Both → set both. HTMX hits the server while Alpine updates local
//     state. Useful for optimistic UI / hybrid flows.
//   - Attrs → generic templ.Attributes escape hatch for any
//     hx-*/x-*/data-*/aria-* not modelled above. Rendered LAST so the
//     caller can always override.
package radio

import "github.com/a-h/templ"

// Variant represents radio button color variants.
type Variant string

const (
	Primary   Variant = "primary"
	Secondary Variant = "secondary"
	Info      Variant = "info"
	Success   Variant = "success"
	Warning   Variant = "warning"
	Danger    Variant = "danger"
)

// Size represents radio input sizes.
type Size string

const (
	SizeSM Size = "sm"
	SizeMD Size = "md" // default
	SizeLG Size = "lg"
	SizeXL Size = "xl"
)

// HTMXConfig holds HTMX attributes for server-side interactions on change.
type HTMXConfig struct {
	Get       string
	Post      string
	Put       string
	Delete    string
	Patch     string
	Target    string
	Swap      string
	Trigger   string // default "change" when any verb is set
	Indicator string
	PushURL   bool
	Confirm   string
	Vals      string // hx-vals JSON string (caller-escaped)
	Include   string // hx-include selector
}

// HasHxVerb reports whether any HTMX verb field is set, used to default
// hx-trigger to "change" when the caller omits it.
func (h *HTMXConfig) HasHxVerb() bool {
	if h == nil {
		return false
	}
	return h.Get != "" || h.Post != "" || h.Put != "" || h.Delete != "" || h.Patch != ""
}

// AlpineConfig holds common Alpine.js directives for client-side behavior.
type AlpineConfig struct {
	Data         string // x-data
	Model        string // x-model
	OnChange     string // x-on:change
	BindChecked  string // x-bind:checked (JS expression)
	BindDisabled string // x-bind:disabled
}

// Config holds configuration for a single radio button.
type Config struct {
	// ID is the unique identifier for the radio input
	ID string
	// Name is the form field name (shared by all radios in a group)
	Name string
	// Value is the form field value
	Value string
	// Label is the text displayed next to the radio
	Label string
	// Checked sets the initial checked state
	Checked bool
	// Disabled disables the radio
	Disabled bool
	// Variant determines the color scheme (default: Primary)
	Variant Variant
	// Size sets the input box size (default: SizeMD)
	Size Size
	// Description adds helper text below the label
	Description string
	// DescriptionID is the id for the description element (for aria-describedby)
	DescriptionID string
	// BadgeColor wraps the label in a semi-solid badge.
	// Accepts: "success", "danger", "warning", "info", "neutral", "primary", "secondary"
	BadgeColor string
	// Container wraps the radio in a bordered container, showing the radio
	// bullet next to a label (still a discrete option visually).
	Container bool
	// Segmented renders a true segmented-control pill: the input is sr-only
	// and the label takes the full visual, switching background/text via
	// `has-checked:`. Designed to be grouped inside a `RadioBar` wrapper
	// (or any flex container with `divide-x` for connected segments).
	// When set, Container is ignored.
	Segmented bool
	// Class is appended to the label root element
	RootClass string

	// HTMX wires server interactions on change.
	HTMX *HTMXConfig
	// Alpine wires client-side state.
	Alpine *AlpineConfig
	// Attrs is an escape hatch applied LAST to the input — wins on conflict.
	InputAttrs templ.Attributes
}

// GroupConfig holds configuration for a group of radio buttons
type GroupConfig struct {
	// Title is the heading above the group
	Title string
	// Items are the radio buttons in the group
	Items []Config
}

// checkedBorderClass returns the checked border color class
func (cfg Config) checkedBorderClass() string {
	switch cfg.Variant {
	case Secondary:
		return "checked:border-secondary dark:checked:border-secondary-dark"
	case Info:
		return "checked:border-info dark:checked:border-info"
	case Success:
		return "checked:border-success dark:checked:border-success"
	case Warning:
		return "checked:border-warning dark:checked:border-warning"
	case Danger:
		return "checked:border-danger dark:checked:border-danger"
	default:
		return "checked:border-primary dark:checked:border-primary-dark"
	}
}

// checkedBgClass returns the checked background color class on the before pseudo
func (cfg Config) checkedBgClass() string {
	switch cfg.Variant {
	case Secondary:
		return "checked:before:bg-secondary dark:checked:before:bg-secondary-dark"
	case Info:
		return "checked:before:bg-info dark:checked:before:bg-info"
	case Success:
		return "checked:before:bg-success dark:checked:before:bg-success"
	case Warning:
		return "checked:before:bg-warning dark:checked:before:bg-warning"
	case Danger:
		return "checked:before:bg-danger dark:checked:before:bg-danger"
	default:
		return "checked:before:bg-primary dark:checked:before:bg-primary-dark"
	}
}

// focusCheckedClass returns the focus outline color when checked
func (cfg Config) focusCheckedClass() string {
	switch cfg.Variant {
	case Secondary:
		return "checked:focus:outline-secondary dark:checked:focus:outline-secondary-dark"
	case Info:
		return "checked:focus:outline-info dark:checked:focus:outline-info"
	case Success:
		return "checked:focus:outline-success dark:checked:focus:outline-success"
	case Warning:
		return "checked:focus:outline-warning dark:checked:focus:outline-warning"
	case Danger:
		return "checked:focus:outline-danger dark:checked:focus:outline-danger"
	default:
		return "checked:focus:outline-primary dark:checked:focus:outline-primary-dark"
	}
}

// sizeBoxClass returns the radio input box size class
func (cfg Config) sizeBoxClass() string {
	switch cfg.Size {
	case SizeSM:
		return "size-3"
	case SizeLG:
		return "size-5"
	case SizeXL:
		return "size-6"
	default:
		return "size-4"
	}
}

// BadgeClasses returns CSS classes for a badge-styled label
func BadgeClasses(color string) string {
	base := "w-fit rounded-radius px-2 py-0.5 text-xs font-medium"
	switch color {
	case "success":
		return base + " bg-success/10 text-success"
	case "danger":
		return base + " bg-danger/10 text-danger"
	case "warning":
		return base + " bg-warning/10 text-warning"
	case "info":
		return base + " bg-info/10 text-info"
	case "primary":
		return base + " bg-primary/10 text-primary dark:text-primary-dark"
	case "secondary":
		return base + " bg-secondary/10 text-secondary"
	case "neutral":
		return base + " bg-on-surface/10 text-on-surface dark:text-on-surface-dark"
	default:
		return ""
	}
}

// InputClasses returns the full CSS class string for the radio input
func (cfg Config) InputClasses() string {
	bg := "bg-surface-alt dark:bg-surface-dark-alt"
	if cfg.Container {
		bg = "bg-surface dark:bg-surface-dark"
	}

	base := "before:content[''] peer relative shrink-0 appearance-none overflow-hidden rounded-full border border-outline " + bg +
		" before:absolute before:inset-0 before:scale-0 before:rounded-full before:transition before:duration-200 checked:before:scale-[0.55]" +
		" focus:outline-2 focus:outline-offset-2 focus:outline-outline-strong active:outline-offset-0 disabled:cursor-not-allowed" +
		" dark:border-outline-dark dark:focus:outline-outline-dark-strong"

	return base + " " + cfg.sizeBoxClass() + " " + cfg.checkedBorderClass() + " " + cfg.checkedBgClass() + " " + cfg.focusCheckedClass()
}

// SegmentedLabelClasses returns the label classes for the Segmented variant.
// The label fully replaces the visual; the input is sr-only and acts as the
// state holder. Active styling rides on Tailwind v4's `has-checked:` selector
// (label has a checked input child → primary fill).
func (cfg Config) SegmentedLabelClasses() string {
	checkedVariant := segmentedCheckedClasses(cfg.Variant)
	return "cursor-pointer select-none inline-flex items-center justify-center px-3 py-1.5 text-sm font-medium transition-colors " +
		"text-on-surface hover:bg-surface dark:text-on-surface-dark dark:hover:bg-surface-dark " +
		"has-disabled:cursor-not-allowed has-disabled:opacity-75 " + checkedVariant
}

func segmentedCheckedClasses(v Variant) string {
	switch v {
	case Secondary:
		return "has-checked:bg-secondary has-checked:text-on-secondary dark:has-checked:bg-secondary-dark dark:has-checked:text-on-secondary-dark"
	case Info:
		return "has-checked:bg-info has-checked:text-on-info"
	case Success:
		return "has-checked:bg-success has-checked:text-on-success"
	case Warning:
		return "has-checked:bg-warning has-checked:text-on-warning"
	case Danger:
		return "has-checked:bg-danger has-checked:text-on-danger"
	default:
		return "has-checked:bg-primary has-checked:text-on-primary dark:has-checked:bg-primary-dark dark:has-checked:text-on-primary-dark"
	}
}

// HasAlpine reports whether any Alpine field is set.
func (cfg Config) HasAlpine() bool {
	return cfg.Alpine != nil
}

// HasHTMX reports whether the HTMX block is non-nil.
func (cfg Config) HasHTMX() bool {
	return cfg.HTMX != nil
}
