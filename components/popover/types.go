package popover

import (
	"strings"

	"github.com/a-h/templ"
)

// Activation controls how the popover is opened.
type Activation string

const (
	ActivationClick   Activation = "click"
	ActivationHover   Activation = "hover"
	ActivationContext Activation = "context"
)

// Placement controls the panel's CSS position relative to the trigger.
type Placement string

const (
	PlacementBottomStart Placement = "bottom-start"
	PlacementBottomEnd   Placement = "bottom-end"
	PlacementTopStart    Placement = "top-start"
	PlacementTopEnd      Placement = "top-end"
)

// Config holds the trigger, content, and lifecycle behavior for a Popover.
//
// Trigger should render one native interactive element. Content is arbitrary
// consumer-owned markup. The popover wraps Trigger in a contents marker so it
// can decorate an existing Button or Link without nesting interactive
// elements.
type Config struct {
	// ID is the optional unique root identifier. When set, the panel ID is
	// ID + "-panel" and the runtime connects aria-controls to it.
	ID string
	// Trigger is the consumer-owned interactive trigger component.
	Trigger templ.Component
	// Content is the consumer-owned panel content.
	Content templ.Component
	// Activation controls whether the popover opens on click, hover, or
	// context-menu activation. The default is ActivationClick.
	Activation Activation
	// Placement controls the panel's CSS placement. The default is
	// PlacementBottomStart.
	Placement Placement
	// Role sets the optional ARIA role on the panel, such as "menu" or
	// "dialog".
	Role string
	// Label sets the optional accessible name on the panel.
	Label string
	// RootClass appends classes to the relative root.
	RootClass string
	// PanelClass appends classes to the positioned panel.
	PanelClass string
	// TrapFocus enables Alpine's focus trap while keyboard-open.
	TrapFocus bool
}

func (cfg Config) activation() Activation {
	switch cfg.Activation {
	case ActivationHover, ActivationContext:
		return cfg.Activation
	default:
		return ActivationClick
	}
}

func (cfg Config) placement() Placement {
	switch cfg.Placement {
	case PlacementBottomEnd, PlacementTopStart, PlacementTopEnd:
		return cfg.Placement
	default:
		return PlacementBottomStart
	}
}

func (cfg Config) rootClasses() string {
	base := "relative w-fit"
	if extra := strings.TrimSpace(cfg.RootClass); extra != "" {
		return base + " " + extra
	}
	return base
}

func (cfg Config) panelID() string {
	if strings.TrimSpace(cfg.ID) == "" {
		return ""
	}
	return cfg.ID + "-panel"
}

func (cfg Config) panelClasses() string {
	position := "left-0 top-full mt-2"
	switch cfg.placement() {
	case PlacementBottomEnd:
		position = "right-0 top-full mt-2"
	case PlacementTopStart:
		position = "bottom-full left-0 mb-2"
	case PlacementTopEnd:
		position = "bottom-full right-0 mb-2"
	}
	base := "absolute z-30 flex w-fit min-w-48 flex-col overflow-hidden rounded-radius border border-outline bg-surface-alt shadow-md dark:border-outline-dark dark:bg-surface-dark-alt"
	if extra := strings.TrimSpace(cfg.PanelClass); extra != "" {
		return position + " " + base + " " + extra
	}
	return position + " " + base
}
