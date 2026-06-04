package rangeinput

import (
	"math"
	"strconv"

	"github.com/a-h/templ"
)

// Tick labels one position in the optional tick row below the range input.
type Tick struct {
	// Label is the visible tick text. Use "|" for an unlabeled mark.
	Label string
	// HideOnMobile hides the tick below the sm breakpoint.
	HideOnMobile bool
}

// Config holds configuration for the range component.
type Config struct {
	// ID is the unique identifier for the range input.
	ID string
	// Name is the form field name.
	Name string
	// Label is the visible label shown above or beside the input.
	Label string
	// Value is the current value (default: "0").
	Value string
	// Min is the minimum value (default: "0").
	Min string
	// Max is the maximum value (default: "100").
	Max string
	// Step is the increment between valid values (default: "1").
	Step string
	// Disabled disables the range input.
	Disabled bool
	// Required marks the input as required.
	Required bool
	// ShowTicks renders a tick row below the input.
	ShowTicks bool
	// Ticks overrides the default tick labels used when ShowTicks is true.
	Ticks []Tick
	// ShowValue renders a live value badge bound with Alpine.js.
	ShowValue bool
	// LeadingIcon renders before the label/input row.
	LeadingIcon templ.Component
	// TrailingIcon renders after the input row.
	TrailingIcon templ.Component
	// RootClass allows additional CSS classes on the outer container.
	RootClass string
	// InputClass allows additional CSS classes on the input.
	InputClass string
	// InputAttrs allows arbitrary attributes on the input, including hx-* or x-* hooks.
	InputAttrs templ.Attributes
}

// ValueOrDefault returns the input value, defaulting to "0".
func (cfg Config) ValueOrDefault() string {
	if cfg.Value != "" {
		return cfg.Value
	}
	return "0"
}

// MinOrDefault returns the minimum range value, defaulting to "0".
func (cfg Config) MinOrDefault() string {
	if cfg.Min != "" {
		return cfg.Min
	}
	return "0"
}

// MaxOrDefault returns the maximum range value, defaulting to "100".
func (cfg Config) MaxOrDefault() string {
	if cfg.Max != "" {
		return cfg.Max
	}
	return "100"
}

// StepOrDefault returns the step value, defaulting to "1".
func (cfg Config) StepOrDefault() string {
	if cfg.Step != "" {
		return cfg.Step
	}
	return "1"
}

// RootClasses returns the CSS classes for the outer container.
func (cfg Config) RootClasses() string {
	base := "flex w-full flex-col gap-2 text-on-surface dark:text-on-surface-dark"
	if cfg.RootClass != "" {
		return base + " " + cfg.RootClass
	}
	return base
}

// ControlClasses returns the CSS classes for the input row.
func (cfg Config) ControlClasses() string {
	if cfg.LeadingIcon != nil || cfg.TrailingIcon != nil || cfg.ShowValue {
		return "flex w-full items-center gap-3"
	}
	return "w-full"
}

// InputClasses returns the CSS classes for the native range input.
func (cfg Config) InputClasses() string {
	base := "w-full cursor-pointer appearance-none bg-transparent disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none [&::-moz-range-thumb]:size-4 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:bg-primary dark:[&::-moz-range-thumb]:bg-primary-dark [&::-moz-range-track]:h-2 [&::-moz-range-track]:rounded-radius [&::-moz-range-track]:bg-surface-alt dark:[&::-moz-range-track]:bg-surface-dark-alt [&::-webkit-slider-runnable-track]:h-2 [&::-webkit-slider-runnable-track]:rounded-radius [&::-webkit-slider-runnable-track]:bg-surface-alt dark:[&::-webkit-slider-runnable-track]:bg-surface-dark-alt [&::-webkit-slider-thumb]:-mt-1 [&::-webkit-slider-thumb]:size-4 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-primary dark:[&::-webkit-slider-thumb]:bg-primary-dark focus-visible:[&::-moz-range-thumb]:outline-2 focus-visible:[&::-moz-range-thumb]:outline-offset-2 focus-visible:[&::-moz-range-thumb]:outline-primary dark:focus-visible:[&::-moz-range-thumb]:outline-primary-dark focus-visible:[&::-webkit-slider-thumb]:outline-2 focus-visible:[&::-webkit-slider-thumb]:outline-offset-2 focus-visible:[&::-webkit-slider-thumb]:outline-primary dark:focus-visible:[&::-webkit-slider-thumb]:outline-primary-dark"
	if cfg.InputClass != "" {
		return base + " " + cfg.InputClass
	}
	return base
}

// LabelClasses returns the CSS classes for the label.
func (cfg Config) LabelClasses() string {
	return "w-fit pl-0.5 text-sm font-medium text-on-surface-strong dark:text-on-surface-dark-strong"
}

// IconClasses returns the CSS classes for icon wrappers.
func (cfg Config) IconClasses() string {
	return "shrink-0 text-on-surface-muted dark:text-on-surface-dark-muted"
}

// ValueClasses returns the CSS classes for the live value badge.
func (cfg Config) ValueClasses() string {
	return "min-w-10 rounded-radius bg-surface-alt px-2 py-1 text-center text-sm font-medium tabular-nums text-on-surface-strong dark:bg-surface-dark-alt dark:text-on-surface-dark-strong"
}

// AlpineData returns simple Alpine state for the live value display.
func (cfg Config) AlpineData() string {
	value, err := strconv.ParseFloat(cfg.ValueOrDefault(), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		value = 0
	}
	return "{ currentVal: " + strconv.FormatFloat(value, 'f', -1, 64) + " }"
}

// TickLabels returns configured ticks or a default 0-to-100 scale.
func (cfg Config) TickLabels() []Tick {
	if len(cfg.Ticks) > 0 {
		return cfg.Ticks
	}

	ticks := make([]Tick, 0, 21)
	ticks = append(ticks, Tick{Label: cfg.MinOrDefault()})
	for i := 1; i < 20; i++ {
		ticks = append(ticks, Tick{Label: "|", HideOnMobile: i%2 == 0})
	}
	ticks = append(ticks, Tick{Label: cfg.MaxOrDefault()})
	return ticks
}

// TickClasses returns the CSS classes for a tick.
func (cfg Config) TickClasses(tick Tick) string {
	base := "text-xs text-on-surface-muted dark:text-on-surface-dark-muted"
	if tick.HideOnMobile {
		return "hidden sm:inline " + base
	}
	return base
}
