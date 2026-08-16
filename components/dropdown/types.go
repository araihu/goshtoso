package dropdown

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/popover"
)

// HTMXConfig holds declarative HTMX attributes for an item action.
//
// Set one request method. When both Get and Post are set, Post takes
// precedence so rendering emits one unambiguous request method. When HTMX is
// set, Item renders as a button and Href is ignored.
type HTMXConfig struct {
	// Get is the URL for an HTMX GET request.
	Get string
	// Post is the URL for an HTMX POST request.
	Post string
	// Target is the CSS selector receiving the response.
	Target string
	// Swap is the HTMX swap strategy.
	Swap string
	// Trigger overrides HTMX's default click trigger.
	Trigger string
	// Vals is additional request values as an HTMX JSON expression.
	Vals string
	// Confirm is the confirmation message shown before the request.
	Confirm string
}

func (cfg *HTMXConfig) attrs() templ.Attributes {
	if cfg == nil {
		return nil
	}
	attrs := templ.Attributes{}
	if cfg.Post != "" {
		attrs["hx-post"] = cfg.Post
	} else if cfg.Get != "" {
		attrs["hx-get"] = cfg.Get
	}
	if cfg.Target != "" {
		attrs["hx-target"] = cfg.Target
	}
	if cfg.Swap != "" {
		attrs["hx-swap"] = cfg.Swap
	}
	if cfg.Trigger != "" {
		attrs["hx-trigger"] = cfg.Trigger
	}
	if cfg.Vals != "" {
		attrs["hx-vals"] = cfg.Vals
	}
	if cfg.Confirm != "" {
		attrs["hx-confirm"] = cfg.Confirm
	}
	return attrs
}

// TriggerMode determines how the dropdown is activated
type TriggerMode string

const (
	TriggerClick   TriggerMode = "click"
	TriggerHover   TriggerMode = "hover"
	TriggerContext TriggerMode = "context"
)

// MenuAlign controls which edge of the trigger the menu panel anchors to.
// AlignStart (default) pins the panel's left edge to the trigger's left edge,
// so the panel opens rightward. AlignEnd pins the panel's right edge to the
// trigger's right edge, so it opens leftward — use this for triggers near the
// right edge of the viewport to avoid horizontal overflow.
type MenuAlign string

const (
	AlignStart MenuAlign = "start"
	AlignEnd   MenuAlign = "end"
)

// Item represents a single menu item in the dropdown.
//
// An Item renders as either an anchor (default) or a button. It renders as a
// button when OnClick, HTMX, or Disabled is set. Href is then ignored so links
// remain a native navigation fallback only when no action is configured.
type Item struct {
	// Label is the display text for the menu item
	Label string
	// Href is the link URL (use "#" for non-navigating items).
	// Ignored when OnClick, HTMX, or Disabled is set.
	Href string
	// Icon is an optional icon component rendered before the label
	Icon templ.Component
	// Caption is optional secondary text rendered below the label.
	Caption string
	// TrailingIcon is an optional consumer-owned icon rendered after the item content.
	TrailingIcon templ.Component
	// Target is passed through to native anchor items. It is ignored for buttons.
	Target string
	// Rel is passed through to native anchor items. It is ignored for buttons.
	Rel string
	// Shortcut is an optional keyboard shortcut label (e.g., "Z", "X")
	Shortcut string
	// ShortcutIcon is an optional icon for the shortcut modifier key
	ShortcutIcon templ.Component

	// OnClick is an Alpine.js expression invoked on click (e.g., "open = true").
	// Setting this renders the item as a <button> instead of an anchor.
	OnClick string
	// HTMX configures a declarative server action. It may be combined with
	// OnClick; both attributes render on the same button.
	HTMX *HTMXConfig
	// Disabled renders the item as a disabled <button> with muted styling.
	// Clicks are suppressed.
	Disabled bool
	// Danger applies destructive styling (red text, red hover) — for actions
	// like "Delete" or "Remove".
	Danger bool
	// Tooltip sets a native title attribute on the item. Useful when Disabled
	// to explain why the action isn't available.
	Tooltip string
	// ID sets the element id — optional, for htmx/Alpine targeting.
	ID string
}

// IsButton reports whether the item should render as a <button> rather than
// an anchor. Buttons are required for click handlers and disabled state.
func (i Item) isButton() bool {
	return i.OnClick != "" || i.HTMX != nil || i.Disabled
}

// Section groups items with an optional heading.
type Section struct {
	// Items is the list of items in this section.
	Items []Item
}

// Config holds configuration for the dropdown component
type Config struct {
	// ID is the unique identifier for the dropdown
	ID string
	// Label is the text shown on the trigger button
	Label string
	// TriggerMode determines how the dropdown opens (click, hover, context)
	TriggerMode TriggerMode
	// Sections groups items with dividers between sections
	Sections []Section
	// TriggerIcon is an optional custom trigger icon.
	// Context mode always shows an icon (defaults to horizontal dots). Labeled
	// click and hover modes render it as a leading icon.
	TriggerIcon templ.Component
	// TriggerIconOnly, in click or hover mode, renders TriggerIcon alone
	// inside a square button — no label, no chevron. Use this for icon-only
	// overflow triggers (e.g., a vertical-dots "…" affordance) without
	// inheriting TriggerContext's <li> item semantics.
	TriggerIconOnly bool
	// MenuAlign controls which edge of the trigger the menu anchors to.
	// Defaults to AlignStart (panel opens rightward). Use AlignEnd for
	// triggers at the right edge of the viewport.
	MenuAlign MenuAlign
	// Trigger replaces the default trigger with a consumer-owned component.
	// It must render one native interactive element.
	Trigger templ.Component
	// RootClass appends classes to the popover root.
	RootClass string
}

// GetTriggerMode returns the trigger mode with a default of click
func (cfg Config) getTriggerMode() TriggerMode {
	if cfg.TriggerMode == "" {
		return TriggerClick
	}
	return cfg.TriggerMode
}

// HasDividers returns true if there are multiple sections
func (cfg Config) hasDividers() bool {
	return len(cfg.Sections) > 1
}

// HasIcons returns true if any item has an icon
func (cfg Config) hasIcons() bool {
	for _, section := range cfg.Sections {
		for _, item := range section.Items {
			if item.Icon != nil || item.TrailingIcon != nil {
				return true
			}
		}
	}
	return false
}

// HasShortcuts returns true if any item has a shortcut
func (cfg Config) hasShortcuts() bool {
	for _, section := range cfg.Sections {
		for _, item := range section.Items {
			if item.Shortcut != "" {
				return true
			}
		}
	}
	return false
}

// IsContextMenu returns true if this is a context menu trigger
func (cfg Config) isContextMenu() bool {
	return cfg.getTriggerMode() == TriggerContext
}

// UseIconOnlyTrigger reports whether the click/hover trigger should render
// the icon alone (no label + chevron).
func (cfg Config) useIconOnlyTrigger() bool {
	return cfg.TriggerIconOnly && cfg.TriggerIcon != nil && !cfg.isContextMenu()
}

func (cfg Config) popoverPlacement() popover.Placement {
	if cfg.MenuAlign == AlignEnd {
		return popover.PlacementBottomEnd
	}
	return popover.PlacementBottomStart
}

func (cfg Config) panelClass() string {
	if cfg.hasDividers() {
		return "divide-y divide-outline dark:divide-outline-dark"
	}
	return ""
}

// ItemClasses returns the CSS classes for a dropdown menu item
func (cfg Config) itemClasses(hasIcon bool) string {
	base := "bg-surface-alt px-4 py-2 text-left text-sm text-on-surface hover:bg-surface-dark-alt/5 hover:text-on-surface-strong focus-visible:bg-surface-dark-alt/10 focus-visible:text-on-surface-strong focus-visible:outline-hidden dark:bg-surface-dark-alt dark:text-on-surface-dark dark:hover:bg-surface-alt/5 dark:hover:text-on-surface-dark-strong dark:focus-visible:bg-surface-alt/10 dark:focus-visible:text-on-surface-dark-strong"
	if hasIcon {
		return "flex items-center gap-2 " + base
	}
	return base
}

// DangerClasses returns the destructive-variant classes applied in addition
// to ItemClasses when Item.Danger is true. Palette matches the navbar
// UserMenuItem danger styling for parity.
func (cfg Config) dangerClasses() string {
	return "text-danger hover:bg-danger/5 hover:text-danger focus-visible:bg-danger/10 focus-visible:text-danger dark:text-danger dark:hover:bg-danger/10 dark:hover:text-danger dark:focus-visible:bg-danger/10 dark:focus-visible:text-danger"
}

// DisabledClasses returns the classes applied when Item.Disabled is true.
// opacity-50 + cursor-not-allowed communicates the state; pointer-events-none
// backs up the native disabled attribute against Alpine event listeners.
func (cfg Config) disabledClasses() string {
	return "opacity-50 cursor-not-allowed pointer-events-none"
}

// ButtonClasses returns the CSS classes for the trigger button
func (cfg Config) buttonClasses() string {
	if cfg.isContextMenu() {
		return "inline-flex items-center bg-transparent transition motion-reduce:transition-none hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-outline-strong active:opacity-100 dark:focus-visible:outline-outline-dark-strong"
	}
	if cfg.useIconOnlyTrigger() {
		return "inline-flex items-center justify-center rounded-radius border border-outline bg-surface-alt p-2 transition motion-reduce:transition-none hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-outline-strong dark:border-outline-dark dark:bg-surface-dark-alt dark:focus-visible:outline-outline-dark-strong"
	}
	return "inline-flex items-center gap-2 whitespace-nowrap rounded-radius border border-outline bg-surface-alt px-4 py-2 text-sm font-medium tracking-wide transition motion-reduce:transition-none hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-outline-strong dark:border-outline-dark dark:bg-surface-dark-alt dark:focus-visible:outline-outline-dark-strong"
}
