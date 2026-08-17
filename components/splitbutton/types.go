package splitbutton

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/dropdown"
)

// Action describes the dominant action in a SplitButton.
//
// An action renders as a native link when only Href is set. OnClick, HTMX, or
// Disabled switches it to a button so the consumer controls the behavior.
type Action struct {
	// Label is the visible action label.
	Label string
	// Href is the native navigation URL when no button action is configured.
	Href string
	// Icon is an optional leading icon.
	Icon templ.Component
	// OnClick is an Alpine.js expression for button actions.
	OnClick string
	// HTMX configures a declarative server action.
	HTMX *dropdown.HTMXConfig
	// Disabled renders an inert button.
	Disabled bool
	// Danger applies the destructive button tone.
	Danger bool
	// Tooltip sets a native title attribute.
	Tooltip string
	// ID sets the primary action's native ID.
	ID string
}

func (action Action) isButton() bool {
	return action.OnClick != "" || action.HTMX != nil || action.Disabled
}

// Config holds the dominant action and consumer-owned dropdown sections.
type Config struct {
	// ID is the unique SplitButton root ID. The menu root receives ID + "-menu".
	ID string
	// Primary is the always-visible dominant action.
	Primary Action
	// MenuLabel is the accessible name for the adjacent menu trigger.
	MenuLabel string
	// MenuTriggerIcon replaces the default chevron icon in the menu trigger.
	MenuTriggerIcon templ.Component
	// Sections supplies the consumer-owned menu items.
	Sections []dropdown.Section
	// MenuAlign controls which edge of the menu panel anchors to the trigger.
	MenuAlign dropdown.MenuAlign
	// Tone controls the primary and menu-trigger color treatment.
	Tone button.Tone
	// Size controls the primary and menu-trigger size.
	Size button.Size
	// RootClass appends classes to the connected group root.
	RootClass string
}

func (cfg Config) menuLabel() string {
	if label := strings.TrimSpace(cfg.MenuLabel); label != "" {
		return label
	}
	return "More actions"
}

func (cfg Config) tone() button.Tone {
	if cfg.Tone == "" {
		return button.TonePrimary
	}
	return cfg.Tone
}

func (cfg Config) size() button.Size {
	if cfg.Size == "" {
		return button.SizeMedium
	}
	return cfg.Size
}

func (cfg Config) rootClasses() string {
	base := "inline-flex items-stretch rounded-radius"
	if extra := strings.TrimSpace(cfg.RootClass); extra != "" {
		return base + " " + extra
	}
	return base
}

func (cfg Config) menuID() string {
	if cfg.ID == "" {
		return ""
	}
	return cfg.ID + "-menu"
}

func (action Action) tone(cfg Config) button.Tone {
	if action.Danger {
		return button.ToneDanger
	}
	return cfg.tone()
}
