package alert

// Tone represents alert color variants
type Tone string

const (
	ToneInfo    Tone = "info"
	ToneSuccess Tone = "success"
	ToneWarning Tone = "warning"
	ToneDanger  Tone = "danger"
)

// LinkConfig holds configuration for an alert link
type LinkConfig struct {
	// Label is the link label
	Label string
	// Href is the link URL
	Href string
}

// HTMXConfig holds HTMX attributes for an alert action button.
type HTMXConfig struct {
	// Get triggers an HTMX GET request.
	Get string
	// Post triggers an HTMX POST request.
	Post string
	// Target is the HTMX target selector.
	Target string
	// Swap is the HTMX swap strategy.
	Swap string
}

// ActionConfig holds configuration for alert action buttons
type ActionConfig struct {
	// PrimaryLabel is the primary action button label
	PrimaryLabel string
	// PrimaryOnClick is the Alpine.js action for the primary button
	PrimaryOnClick string
	// PrimaryHTMX configures HTMX behavior for the primary button.
	PrimaryHTMX *HTMXConfig
	// DismissLabel is the secondary dismiss button label (defaults to "Dismiss")
	DismissLabel string
}

// Config holds configuration for the alert component
type Config struct {
	// Title is the alert heading
	Title string
	// Description is the alert body text
	Description string
	// Tone determines the color scheme (info, success, warning, danger)
	Tone Tone
	// Dismissible enables the dismiss button with Alpine.js transition
	Dismissible bool
	// Link adds a link action to the alert
	Link *LinkConfig
	// Action adds primary + dismiss action buttons
	Action *ActionConfig
	// ListItems adds a bullet list below the description
	ListItems []string
	// RootClass allows additional CSS classes on the alert root.
	RootClass string
}

// ContainerClasses returns the outer container CSS classes
func (cfg Config) containerClasses() string {
	base := "relative w-full overflow-hidden rounded-radius border bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark"

	switch cfg.Tone {
	case ToneInfo:
		base += " border-info"
	case ToneSuccess:
		base += " border-success"
	case ToneWarning:
		base += " border-warning"
	case ToneDanger:
		base += " border-danger"
	default:
		base += " border-info"
	}

	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}

	return base
}

// InnerClasses returns the inner wrapper CSS classes
func (cfg Config) innerClasses() string {
	base := "flex w-full items-center gap-2 p-4"

	switch cfg.Tone {
	case ToneInfo:
		base += " bg-info/10"
	case ToneSuccess:
		base += " bg-success/10"
	case ToneWarning:
		base += " bg-warning/10"
	case ToneDanger:
		base += " bg-danger/10"
	default:
		base += " bg-info/10"
	}

	return base
}

// IconBadgeClasses returns the icon badge CSS classes
func (cfg Config) iconBadgeClasses() string {
	base := "rounded-full p-1"

	switch cfg.Tone {
	case ToneInfo:
		base += " bg-info/15 text-info"
	case ToneSuccess:
		base += " bg-success/15 text-success"
	case ToneWarning:
		base += " bg-warning/15 text-warning"
	case ToneDanger:
		base += " bg-danger/15 text-danger"
	default:
		base += " bg-info/15 text-info"
	}

	return base
}

// TitleClasses returns the title CSS classes
func (cfg Config) titleClasses() string {
	base := "text-sm font-semibold"

	switch cfg.Tone {
	case ToneInfo:
		base += " text-info"
	case ToneSuccess:
		base += " text-success"
	case ToneWarning:
		base += " text-warning"
	case ToneDanger:
		base += " text-danger"
	default:
		base += " text-info"
	}

	return base
}

// LinkClasses returns the link CSS classes
func (cfg Config) linkClasses() string {
	base := "whitespace-nowrap ml-auto text-sm font-medium tracking-wide transition hover:opacity-75 text-center active:opacity-100"

	switch cfg.Tone {
	case ToneInfo:
		base += " text-info"
	case ToneSuccess:
		base += " text-success"
	case ToneWarning:
		base += " text-warning"
	case ToneDanger:
		base += " text-danger"
	default:
		base += " text-info"
	}

	return base
}

// PrimaryActionClasses returns the primary action button CSS classes
func (cfg Config) primaryActionClasses() string {
	base := "whitespace-nowrap text-center text-sm font-semibold tracking-wide transition hover:opacity-75 active:opacity-100"

	switch cfg.Tone {
	case ToneInfo:
		base += " text-info"
	case ToneSuccess:
		base += " text-success"
	case ToneWarning:
		base += " text-warning"
	case ToneDanger:
		base += " text-danger"
	default:
		base += " text-info"
	}

	return base
}

// ListClasses returns the list CSS classes
func (cfg Config) listClasses() string {
	base := "mt-2 list-inside list-disc pl-2 text-xs font-medium sm:text-sm"

	if cfg.Tone == ToneDanger {
		base += " text-danger"
	}

	return base
}
