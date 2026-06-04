package accordion

import "github.com/a-h/templ"

// Variant represents accordion style variants
type Variant string

const (
	// Default variant with background and borders
	Default Variant = "default"
	// NoBackground variant without background styling
	NoBackground Variant = "no-background"
	// Split variant with different layout
	Split Variant = "split"
	// SingleOpen ensures only one item can be open at a time
	SingleOpen Variant = "single-open"
)

// AccordionConfig holds configuration for the accordion
type AccordionConfig struct {
	// Items are the accordion sections
	Items []AccordionItem
	// AllowMultiple allows multiple items to be open simultaneously
	// If false (default), only one item can be open at a time
	AllowMultiple bool
	// Variant determines the visual style
	Variant Variant
	// ID is the container ID for accessibility
	ID string
	// Class allows additional CSS classes
	RootClass string
}

// AccordionItem represents a single accordion section
type AccordionItem struct {
	// ID unique identifier for the item (used for accessibility)
	ID string
	// Title is the header text
	Title string
	// Content is the body content (can be templ.Component or string)
	Content templ.Component
	// Icon is an optional leading icon (templ.Component)
	Icon templ.Component
	// Disabled prevents interaction with this item
	Disabled bool
	// InitiallyExpanded sets the initial state
	InitiallyExpanded bool
}

// AccordionItemData is used internally for rendering
type AccordionItemData struct {
	// Item is the accordion section being rendered.
	Item AccordionItem
	// Index is the zero-based position of the item in the accordion.
	Index int
	// AllowMultiple indicates whether multiple sections can be open simultaneously.
	AllowMultiple bool
	// Variant is the visual style variant inherited from the parent accordion.
	Variant Variant
	// ContainerID is the parent accordion's element ID for accessibility.
	ContainerID string
}

// ContainerClasses returns the container CSS classes based on variant
func (cfg AccordionConfig) ContainerClasses() string {
	// Split renders each item as its own gapped card, so the container drops
	// the shared divider/border/background and just lays the cards out vertically.
	if cfg.Variant == Split {
		return "flex w-full flex-col gap-4 text-on-surface dark:text-on-surface-dark"
	}

	base := "w-full divide-y divide-outline overflow-hidden rounded-radius border border-outline text-on-surface dark:divide-outline-dark dark:border-outline-dark dark:text-on-surface-dark"

	switch cfg.Variant {
	case NoBackground:
		return base + " bg-surface dark:bg-surface-dark"
	default:
		return base + " bg-surface-alt/40 dark:bg-surface-dark-alt/50"
	}
}

// ItemContainerClasses returns per-item wrapper classes. Only the Split variant
// uses these: each item becomes a self-contained bordered card. Other variants
// return "" since the shared container already provides borders and dividers.
func (data AccordionItemData) ItemContainerClasses() string {
	if data.Variant == Split {
		return "overflow-hidden rounded-radius border border-outline bg-surface-alt/40 dark:border-outline-dark dark:bg-surface-dark-alt/50"
	}
	return ""
}

// ItemButtonClasses returns button classes based on variant and state
func (data AccordionItemData) ItemButtonClasses() string {
	base := "flex w-full items-center justify-between gap-4 p-4 text-left underline-offset-2 focus-visible:underline focus-visible:outline-hidden"

	switch data.Variant {
	case NoBackground:
		return base + " bg-surface hover:bg-surface-alt focus-visible:bg-surface-alt dark:bg-surface-dark dark:hover:bg-surface-dark-alt dark:focus-visible:bg-surface-dark-alt"
	default:
		return base + " bg-surface-alt hover:bg-surface-alt/75 focus-visible:bg-surface-alt/75 dark:bg-surface-dark-alt dark:hover:bg-surface-dark-alt/75 dark:focus-visible:bg-surface-dark-alt/75"
	}
}

// ExpandedClasses returns classes when item is expanded
func (data AccordionItemData) ExpandedClasses() string {
	return "text-on-surface-strong dark:text-on-surface-dark-strong font-bold"
}

// CollapsedClasses returns classes when item is collapsed
func (data AccordionItemData) CollapsedClasses() string {
	return "text-on-surface dark:text-on-surface-dark font-medium"
}

// ContentClasses returns content container classes
func (data AccordionItemData) ContentClasses() string {
	return "p-4 text-sm sm:text-base text-pretty"
}
