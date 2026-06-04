package steps

import (
	"fmt"

	"github.com/a-h/templ"
)

// Status controls visual state for a step item.
type Status string

const (
	// StatusCompleted renders a filled step with a check icon.
	StatusCompleted Status = "completed"
	// StatusCurrent renders active step with highlighted number and outline.
	StatusCurrent Status = "current"
	// StatusUpcoming renders pending step with muted number.
	StatusUpcoming Status = "upcoming"
)

// Orientation controls overall layout direction.
type Orientation string

const (
	// OrientationHorizontal renders steps left-to-right.
	OrientationHorizontal Orientation = "horizontal"
	// OrientationVertical renders steps top-to-bottom.
	OrientationVertical Orientation = "vertical"
)

// Step represents a single item in progress flow.
type Step struct {
	// ID sets stable element id for the list item.
	ID string
	// Label is visible text next to step when labels are enabled.
	Label string
	// AriaLabel overrides accessible label on the list item.
	AriaLabel string
	// Number overrides displayed step number for non-completed states.
	// Defaults to 1-based index when zero.
	Number int
	// Status controls visual styling.
	Status Status
	// StepAttrs appends arbitrary attributes to the list item.
	StepAttrs templ.Attributes
}

// Config holds configuration for Steps component.
type Config struct {
	// ID sets stable element id on root ordered list.
	ID string
	// Steps are rendered in given order.
	Steps []Step
	// Orientation controls horizontal vs vertical layout.
	Orientation Orientation
	// ShowLabels toggles visible labels beside steps.
	ShowLabels bool
	// AriaLabel labels the ordered list.
	AriaLabel string
	// LiveRegion announces swapped state changes to assistive tech.
	LiveRegion bool
	// RootClass appends custom classes to root list element.
	RootClass string
	// RootAttrs appends arbitrary attributes to root ordered list.
	RootAttrs templ.Attributes
}

// normalizedOrientation returns default orientation.
func (cfg Config) normalizedOrientation() Orientation {
	if cfg.Orientation == OrientationVertical {
		return OrientationVertical
	}
	return OrientationHorizontal
}

// listClasses returns root ordered-list classes.
func (cfg Config) listClasses() string {
	base := "text-on-surface dark:text-on-surface-dark"
	if cfg.normalizedOrientation() == OrientationVertical {
		base += " flex w-min flex-col gap-14"
	} else {
		base += " flex w-full items-center gap-2"
	}
	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}
	return base
}

// resolvedAriaLabel returns default aria label.
func (cfg Config) resolvedAriaLabel() string {
	if cfg.AriaLabel != "" {
		return cfg.AriaLabel
	}
	return "progress"
}

// resolvedStep applies defaults for a single step.
func resolvedStep(step Step, index int) Step {
	if step.ID == "" {
		step.ID = fmt.Sprintf("step-%d", index+1)
	}
	if step.Number == 0 {
		step.Number = index + 1
	}
	if step.Status == "" {
		step.Status = StatusUpcoming
	}
	return step
}

// connectorClasses returns classes for connector before current step.
func connectorClasses(status Status, orientation Orientation) string {
	color := "bg-outline dark:bg-outline-dark"
	if status == StatusCompleted || status == StatusCurrent {
		color = "bg-primary dark:bg-primary-dark"
	}
	if orientation == OrientationVertical {
		return "absolute bottom-8 left-3 h-10 w-0.5 " + color
	}
	return "h-0.5 w-full " + color
}

// indicatorClasses returns classes for step circle.
func indicatorClasses(status Status) string {
	base := "flex size-6 shrink-0 items-center justify-center rounded-full border"
	switch status {
	case StatusCompleted:
		return base + " border-primary bg-primary text-on-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark"
	case StatusCurrent:
		return base + " border-primary bg-primary font-bold text-on-primary outline outline-2 outline-offset-2 outline-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark dark:outline-primary-dark"
	default:
		return base + " border-outline bg-surface-alt font-medium text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"
	}
}

// labelClasses returns classes for optional visible label.
func labelClasses(status Status) string {
	base := "inline w-max"
	switch status {
	case StatusCompleted:
		return base + " text-primary dark:text-primary-dark"
	case StatusCurrent:
		return base + " font-bold text-primary dark:text-primary-dark"
	default:
		return base + " text-on-surface dark:text-on-surface-dark"
	}
}
