package banner

import "github.com/a-h/templ"

// Variant represents banner style variants
type Variant string

const (
	Default Variant = "default"
	Primary Variant = "primary"
	Info    Variant = "info"
	Success Variant = "success"
	Warning Variant = "warning"
	Danger  Variant = "danger"
)

// Position represents banner position
type Position string

const (
	PositionRelative Position = "relative" // Default inline
	PositionFixed    Position = "fixed"    // Fixed to top
)

// Config holds configuration for the banner
type Config struct {
	// Description is the main banner content
	Description string
	// Variant determines the color scheme
	Variant Variant
	// Position determines if banner is fixed or relative
	Position Position
	// Persistent disables the dismiss button (default: banners are dismissible)
	Persistent bool
	// DismissAction is the Alpine.js action when dismissed
	DismissAction string
	// CTA is the call-to-action button config (optional)
	CTA *CTAConfig
	// CookieBanner enables cookie consent mode
	CookieBanner bool
	// CookieConfig holds cookie banner specific settings
	CookieConfig *CookieBannerConfig
	// RootClass allows additional CSS classes on the banner root.
	RootClass string
}

// CTAConfig holds call-to-action button configuration
type CTAConfig struct {
	// ActionLabel is the button label
	ActionLabel string
	// Href is the link URL (if set, renders as anchor)
	Href string
	// OnClick is the Alpine.js click action
	OnClick string
}

// CookieBannerConfig holds cookie banner specific configuration
type CookieBannerConfig struct {
	// Title of the cookie banner
	Title string
	// Icon is an optional icon (emoji or component)
	Icon templ.Component
	// AcceptLabel is the accept button label
	AcceptLabel string
	// RejectLabel is the reject button label
	RejectLabel string
	// AcceptAction is the Alpine.js action for accept
	AcceptAction string
	// RejectAction is the Alpine.js action for reject
	RejectAction string
}

// ContainerClasses returns the container CSS classes
func (cfg Config) ContainerClasses() string {
	base := "flex w-full p-4 text-on-surface dark:text-on-surface-dark"

	// Position
	if cfg.Position == PositionFixed {
		base = "fixed inset-x-0 top-0 z-50 " + base
	}

	// Variant styles
	switch cfg.Variant {
	case Primary:
		base += " border-b border-primary bg-primary/10 dark:border-primary-dark dark:bg-primary-dark/10"
	case Info:
		base += " border-b border-info bg-info/10 dark:border-info dark:bg-info/10"
	case Success:
		base += " border-b border-success bg-success/10 dark:border-success dark:bg-success/10"
	case Warning:
		base += " border-b border-warning bg-warning/10 dark:border-warning dark:bg-warning/10"
	case Danger:
		base += " border-b border-danger bg-danger/10 dark:border-danger dark:bg-danger/10"
	default:
		base += " border-b border-outline bg-surface-alt dark:border-outline-dark dark:bg-surface-dark-alt"
	}

	return base + " " + cfg.RootClass
}

// TextClasses returns the text content classes
func (cfg Config) TextClasses() string {
	return "px-6 text-xs sm:text-sm text-pretty mx-auto"
}

// LinkClasses returns the link classes within the banner
func (cfg Config) LinkClasses() string {
	switch cfg.Variant {
	case Primary:
		return "font-medium text-primary underline-offset-2 hover:underline focus:underline focus:outline-hidden dark:text-primary-dark"
	case Info:
		return "font-medium text-info underline-offset-2 hover:underline focus:underline focus:outline-hidden"
	case Success:
		return "font-medium text-success underline-offset-2 hover:underline focus:underline focus:outline-hidden"
	case Warning:
		return "font-medium text-warning underline-offset-2 hover:underline focus:underline focus:outline-hidden"
	case Danger:
		return "font-medium text-danger underline-offset-2 hover:underline focus:underline focus:outline-hidden"
	default:
		return "font-medium text-primary underline-offset-2 hover:underline focus:underline focus:outline-hidden dark:text-primary-dark"
	}
}

// CTAClasses returns the CTA button classes
func (cfg Config) CTAClasses() string {
	return "whitespace-nowrap bg-primary px-4 py-1 text-center text-xs font-medium tracking-wide text-on-primary transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:opacity-100 active:outline-offset-0 disabled:cursor-not-allowed disabled:opacity-75 dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark rounded-radius"
}

// CookieContainerClasses returns the container CSS classes for cookie banners.
func (cfg Config) CookieContainerClasses() string {
	position := "fixed bottom-4 left-4 right-4 sm:left-auto sm:right-4 sm:max-w-sm z-50"
	if cfg.Position == PositionRelative {
		position = "absolute bottom-4 left-4 right-4 sm:left-auto sm:right-4 sm:max-w-sm"
	}

	return position + " flex flex-col gap-4 border border-outline bg-surface-alt/50 text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt/50 dark:text-on-surface-dark rounded-radius " + cfg.RootClass
}
