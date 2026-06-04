package tooltip

import "github.com/a-h/templ"

// Position represents tooltip placement relative to the trigger
type Position string

const (
	Top    Position = "top"
	Bottom Position = "bottom"
	Left   Position = "left"
	Right  Position = "right"
)

// Trigger represents how the tooltip is activated
type Trigger string

const (
	Hover Trigger = "hover" // Default: peer-hover/peer-focus
	Click Trigger = "click" // Alpine.js click toggle
)

// Config holds configuration for the tooltip component
type Config struct {
	// ID is the tooltip element ID (used for aria-describedby)
	ID string
	// Label is the main tooltip text
	Label string
	// Description is optional secondary text (rich tooltip)
	Description string
	// Position determines where the tooltip appears
	Position Position
	// TriggerMode determines how the tooltip is activated
	TriggerMode Trigger
	// TriggerLabel is the text shown on the trigger button
	TriggerLabel string
	// Trigger is an optional custom trigger element (overrides TriggerLabel)
	Trigger templ.Component
}

// positionClasses returns CSS classes for tooltip positioning
func (cfg Config) positionClasses() string {
	switch cfg.Position {
	case Bottom:
		return "absolute top-full mt-2 left-1/2 -translate-x-1/2"
	case Left:
		return "absolute right-full mr-2 top-1/2 -translate-y-1/2"
	case Right:
		return "absolute left-full ml-2 top-1/2 -translate-y-1/2"
	default: // Top
		return "absolute bottom-full mb-2 left-1/2 -translate-x-1/2"
	}
}

// isRich returns true if tooltip has a description (rich tooltip)
func (cfg Config) isRich() bool {
	return cfg.Description != ""
}

// tooltipID returns the ID to use, defaulting to "tooltipExample"
func (cfg Config) tooltipID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	return "tooltipExample"
}

// triggerLabel returns the trigger button text
func (cfg Config) triggerLabel() string {
	if cfg.TriggerLabel != "" {
		return cfg.TriggerLabel
	}
	return "Hover Me"
}
