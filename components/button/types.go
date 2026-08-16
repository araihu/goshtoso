package button

import "github.com/a-h/templ"

// Tone controls the semantic color treatment of a button.
type Tone string

const (
	TonePrimary   Tone = "primary"
	ToneSecondary Tone = "secondary"
	ToneAlternate Tone = "alternate"
	ToneInverse   Tone = "inverse"
	ToneInfo      Tone = "info"
	ToneDanger    Tone = "danger"
	ToneWarning   Tone = "warning"
	ToneSuccess   Tone = "success"
)

// Size represents button sizes.
type Size string

const (
	SizeSmall  Size = "sm"
	SizeMedium Size = "md"
	SizeLarge  Size = "lg"
	SizeXLarge Size = "xl"
)

// HTMXConfig holds HTMX attributes for server-side interactions.
type HTMXConfig struct {
	// Get is the URL for an HTMX GET request.
	Get string
	// Post is the URL for an HTMX POST request.
	Post string
	// Put is the URL for an HTMX PUT request.
	Put string
	// Delete is the URL for an HTMX DELETE request.
	Delete string
	// Patch is the URL for an HTMX PATCH request.
	Patch string
	// Target is the CSS selector for the element to swap the response into.
	Target string
	// Swap is the htmx swap strategy (e.g. "innerHTML", "outerHTML").
	Swap string
	// Trigger is the htmx trigger that initiates the request.
	Trigger string
	// Indicator is the CSS selector for the loading indicator element.
	Indicator string
	// PushURL pushes the request URL to the browser history.
	PushURL bool
	// Confirm is a confirmation message shown before the request is sent.
	Confirm string
	// Vals is additional values to submit with the request as JSON.
	Vals string
}

// AlpineConfig holds Alpine.js directives for client-side interactions.
type AlpineConfig struct {
	// OnClick is the Alpine x-on:click expression.
	OnClick string
	// BindDisabled is the x-bind:disabled expression.
	BindDisabled string
	// Show is the x-show expression controlling visibility.
	Show string
	// Transition enables x-transition on the element.
	Transition bool
	// Data is the x-data expression for component state.
	Data string
}

type config struct {
	tone        Tone
	size        Size
	buttonType  string
	disabled    bool
	id          string
	rootClass   string
	htmx        *HTMXConfig
	alpine      *AlpineConfig
	loadingText string
	attrs       templ.Attributes
}

// Option configures a Button.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

// WithTone sets the button's semantic color treatment.
func WithTone(tone Tone) Option {
	return optionFunc(func(cfg *config) {
		cfg.tone = tone
	})
}

// WithSize sets the button size.
func WithSize(size Size) Option {
	return optionFunc(func(cfg *config) {
		cfg.size = size
	})
}

// WithType sets the native button type attribute.
func WithType(buttonType string) Option {
	return optionFunc(func(cfg *config) {
		cfg.buttonType = buttonType
	})
}

// Disabled disables the button.
func Disabled() Option {
	return optionFunc(func(cfg *config) {
		cfg.disabled = true
	})
}

// WithID sets the button's HTML id attribute.
func WithID(id string) Option {
	return optionFunc(func(cfg *config) {
		cfg.id = id
	})
}

// WithRootClass appends CSS classes to the button.
func WithRootClass(class string) Option {
	return optionFunc(func(cfg *config) {
		cfg.rootClass = class
	})
}

// WithAttrs adds arbitrary attributes to the native button element.
//
// Use this escape hatch for standard form attributes such as name, value, and
// formaction, or for data-* and aria-* attributes not covered by other options.
func WithAttrs(attrs templ.Attributes) Option {
	return optionFunc(func(cfg *config) {
		cfg.attrs = attrs
	})
}

// WithHTMX configures HTMX attributes for the button.
func WithHTMX(htmx *HTMXConfig) Option {
	return optionFunc(func(cfg *config) {
		cfg.htmx = htmx
	})
}

// WithAlpine configures Alpine.js directives for the button.
func WithAlpine(alpine *AlpineConfig) Option {
	return optionFunc(func(cfg *config) {
		cfg.alpine = alpine
	})
}

// WithLoadingText sets text shown while this button or an ancestor HTMX form is
// requesting. When the button owns the HTMX request it is disabled automatically;
// ancestor forms should set hx-disabled-elt="find button[type='submit']".
func WithLoadingText(text string) Option {
	return optionFunc(func(cfg *config) {
		cfg.loadingText = text
	})
}

func newConfig(options []Option) config {
	cfg := config{
		tone:       TonePrimary,
		size:       SizeMedium,
		buttonType: "button",
	}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

// toneClasses returns the Tailwind utility classes for a tone.
func toneClasses(tone Tone) string {
	switch tone {
	case TonePrimary:
		return "bg-primary text-on-primary border-primary dark:bg-primary-dark dark:text-on-primary-dark dark:border-primary-dark"
	case ToneSecondary:
		return "bg-secondary text-on-secondary border-secondary dark:bg-secondary-dark dark:text-on-secondary-dark dark:border-secondary-dark"
	case ToneAlternate:
		return "bg-surface-alt text-on-surface-strong border-surface-alt dark:bg-surface-dark-alt dark:text-on-surface-dark-strong dark:border-surface-dark-alt"
	case ToneInverse:
		return "bg-surface-dark text-on-surface-dark border-surface-dark dark:bg-surface dark:text-on-surface dark:border-surface"
	case ToneInfo:
		return "bg-info-action text-on-info-action border-info-action dark:bg-info-action-dark dark:text-on-info-action-dark dark:border-info-action-dark"
	case ToneDanger:
		return "bg-danger-action text-on-danger-action border-danger-action dark:bg-danger-action-dark dark:text-on-danger-action-dark dark:border-danger-action-dark"
	case ToneWarning:
		return "bg-warning-action text-on-warning-action border-warning-action dark:bg-warning-action-dark dark:text-on-warning-action-dark dark:border-warning-action-dark"
	case ToneSuccess:
		return "bg-success-action text-on-success-action border-success-action dark:bg-success-action-dark dark:text-on-success-action-dark dark:border-success-action-dark"
	default:
		return "bg-primary text-on-primary border-primary"
	}
}

func sizeClasses(size Size) string {
	switch size {
	case SizeSmall:
		return "px-4 py-2 text-xs"
	case SizeLarge:
		return "px-4 py-2 text-base"
	case SizeXLarge:
		return "px-4 py-2 text-lg"
	default:
		return "px-4 py-2 text-sm"
	}
}

func buttonClasses(cfg config) string {
	base := "inline-flex items-center justify-center gap-2 min-h-11 min-w-11 whitespace-nowrap rounded-2xl font-medium tracking-wide transition motion-reduce:transition-none hover:contrast-125 text-center focus-visible:outline-2 focus-visible:outline-offset-2 active:contrast-100 active:outline-offset-0 disabled:opacity-75 disabled:cursor-not-allowed border"
	outline := focusOutlineClasses(cfg.tone)
	return base + " " + toneClasses(cfg.tone) + " " + sizeClasses(cfg.size) + " " + outline + " " + cfg.rootClass
}

func focusOutlineClasses(tone Tone) string {
	switch tone {
	case ToneSecondary:
		return "focus-visible:outline-secondary dark:focus-visible:outline-secondary-dark"
	case ToneAlternate, ToneInverse:
		return "focus-visible:outline-on-surface-strong dark:focus-visible:outline-on-surface-dark-strong"
	case ToneInfo:
		return "focus-visible:outline-info-action dark:focus-visible:outline-info-action-dark"
	case ToneDanger:
		return "focus-visible:outline-danger-action dark:focus-visible:outline-danger-action-dark"
	case ToneWarning:
		return "focus-visible:outline-warning-action dark:focus-visible:outline-warning-action-dark"
	case ToneSuccess:
		return "focus-visible:outline-success-action dark:focus-visible:outline-success-action-dark"
	default:
		return "focus-visible:outline-primary dark:focus-visible:outline-primary-dark"
	}
}
