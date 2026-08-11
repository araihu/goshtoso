package tooltip

import "github.com/a-h/templ"

// Position represents tooltip placement relative to the trigger.
type Position string

const (
	PositionTop    Position = "top"
	PositionBottom Position = "bottom"
	PositionLeft   Position = "left"
	PositionRight  Position = "right"
)

// Activation represents how the tooltip is activated.
type Activation string

const (
	ActivationHover Activation = "hover"
	// ActivationClick makes the Tooltip persistent. Its actual trigger keeps
	// aria-describedby, controls and reflects the Tooltip's expanded state, and
	// supports click, Enter, Space, Escape, and outside-click interaction. A
	// Tooltip is descriptive content, so this mode does not add aria-haspopup.
	ActivationClick Activation = "click"
)

type config struct {
	id           string
	label        string
	description  string
	position     Position
	activation   Activation
	triggerLabel string
	trigger      templ.Component
}

// Option configures a Tooltip.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

// WithDescription adds secondary text to create a rich tooltip.
func WithDescription(description string) Option {
	return optionFunc(func(cfg *config) {
		cfg.description = description
	})
}

// WithPosition sets tooltip placement relative to the trigger.
func WithPosition(position Position) Option {
	return optionFunc(func(cfg *config) {
		cfg.position = position
	})
}

// WithActivation sets how the tooltip is activated. ActivationClick opts into
// the persistent, reflected-state behavior documented by ActivationClick.
func WithActivation(activation Activation) Option {
	return optionFunc(func(cfg *config) {
		cfg.activation = activation
	})
}

// WithTriggerLabel sets the text shown on the default trigger button.
func WithTriggerLabel(label string) Option {
	return optionFunc(func(cfg *config) {
		cfg.triggerLabel = label
	})
}

// WithTrigger replaces the default trigger button with a custom component.
func WithTrigger(trigger templ.Component) Option {
	return optionFunc(func(cfg *config) {
		cfg.trigger = trigger
	})
}

func newConfig(id, label string, options []Option) config {
	cfg := config{
		id:           id,
		label:        label,
		position:     PositionTop,
		activation:   ActivationHover,
		triggerLabel: "Hover Me",
	}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

func (cfg config) positionClasses() string {
	switch cfg.position {
	case PositionBottom:
		return "absolute top-full mt-2 left-1/2 -translate-x-1/2"
	case PositionLeft:
		return "absolute right-full mr-2 top-1/2 -translate-y-1/2"
	case PositionRight:
		return "absolute left-full ml-2 top-1/2 -translate-y-1/2"
	default:
		return "absolute bottom-full mb-2 left-1/2 -translate-x-1/2"
	}
}

func (cfg config) isRich() bool {
	return cfg.description != ""
}
