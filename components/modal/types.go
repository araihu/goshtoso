package modal

import (
	"strings"
	"unicode"
)

// Variant represents modal color variants (used in alert mode)
type Variant string

const (
	Default Variant = "default"
	Success Variant = "success"
	Info    Variant = "info"
	Warning Variant = "warning"
	Danger  Variant = "danger"
)

// ButtonAction holds optional HTMX and Alpine.js actions for a button
type ButtonAction struct {
	// OnClick is a custom Alpine.js expression (appended after modal close)
	OnClick string
	// HxGet triggers an HTMX GET request
	HxGet string
	// HxPost triggers an HTMX POST request
	HxPost string
	// HxTarget is the HTMX target selector
	HxTarget string
	// HxSwap is the HTMX swap strategy
	HxSwap string
}

// Config holds configuration for the modal component
type Config struct {
	// ID is a unique identifier used for aria-labelledby (required)
	ID string
	// Title is the modal heading
	Title string
	// Body is the modal body text
	Body string
	// TriggerText is the trigger button label
	TriggerText string
	// PrimaryText is the primary action button label
	PrimaryText string
	// PrimaryAction holds optional HTMX/JS actions for the primary button
	PrimaryAction *ButtonAction
	// SecondaryText is the secondary/dismiss button label (default mode only)
	SecondaryText string
	// SecondaryAction holds optional HTMX/JS actions for the secondary button
	SecondaryAction *ButtonAction
	// Variant determines the color scheme (used for alert mode and trigger button)
	Variant Variant
	// AlertMode renders the alert-style modal (icon header, centered body, single CTA)
	AlertMode bool
	// Class allows additional CSS classes on the dialog
	PanelClass string
}

func (cfg Config) StateVar() string {
	return safeJSIdentifier(cfg.ID, "modal") + "IsOpen"
}

func (cfg Config) TitleID() string {
	base := cfg.ID
	if base == "" {
		base = "modal"
	}
	return base + "Title"
}

// TriggerClasses returns the trigger button CSS classes
func (cfg Config) TriggerClasses() string {
	if cfg.AlertMode {
		return cfg.alertTriggerClasses()
	}
	return "whitespace-nowrap rounded-radius border border-primary dark:border-primary-dark bg-primary px-4 py-2 text-center text-sm font-medium tracking-wide text-on-primary transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:opacity-100 active:outline-offset-0 dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark"
}

func (cfg Config) alertTriggerClasses() string {
	base := "w-36 whitespace-nowrap rounded-radius border px-4 py-2 text-center text-sm font-medium tracking-wide transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 active:opacity-100 active:outline-offset-0"

	switch cfg.Variant {
	case Success:
		base += " border-success bg-success text-on-success focus-visible:outline-success"
	case Info:
		base += " border-info bg-info text-on-info focus-visible:outline-info"
	case Warning:
		base += " border-warning bg-warning text-on-warning focus-visible:outline-warning"
	case Danger:
		base += " border-danger bg-danger text-on-danger focus-visible:outline-danger"
	default:
		base += " border-primary bg-primary text-on-primary focus-visible:outline-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark"
	}

	return base
}

// DialogClasses returns the modal dialog container CSS classes
func (cfg Config) DialogClasses() string {
	base := "flex max-w-lg flex-col gap-4 overflow-hidden rounded-radius border border-outline bg-surface text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"
	if cfg.PanelClass != "" {
		base += " " + cfg.PanelClass
	}
	return base
}

func safeJSIdentifier(raw, fallback string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = fallback
	}
	var b strings.Builder
	upperNext := false
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			if b.Len() == 0 && unicode.IsDigit(r) {
				b.WriteString(fallback)
			}
			if upperNext {
				b.WriteRune(unicode.ToUpper(r))
				upperNext = false
			} else {
				b.WriteRune(r)
			}
			continue
		}
		if b.Len() > 0 {
			upperNext = true
		}
	}
	if b.Len() == 0 {
		return fallback
	}
	return b.String()
}

// HeaderClasses returns the dialog header CSS classes
func (cfg Config) HeaderClasses() string {
	if cfg.AlertMode {
		return "flex items-center justify-between border-b border-outline bg-surface-alt/60 px-4 py-2 dark:border-outline-dark dark:bg-surface-dark/20"
	}
	return "flex items-center justify-between border-b border-outline bg-surface-alt/60 p-4 dark:border-outline-dark dark:bg-surface-dark/20"
}

// IconBadgeClasses returns the alert icon badge CSS classes
func (cfg Config) IconBadgeClasses() string {
	base := "flex items-center justify-center rounded-full p-1"
	switch cfg.Variant {
	case Success:
		base += " bg-success/20 text-success"
	case Info:
		base += " bg-info/20 text-info"
	case Warning:
		base += " bg-warning/20 text-warning"
	case Danger:
		base += " bg-danger/20 text-danger"
	}
	return base
}

// AlertCTAClasses returns the full-width CTA button classes for alert modals
func (cfg Config) AlertCTAClasses() string {
	base := "w-full whitespace-nowrap rounded-radius border px-4 py-2 text-center text-sm font-semibold tracking-wide transition hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 active:opacity-100 active:outline-offset-0"
	switch cfg.Variant {
	case Success:
		base += " border-success bg-success text-on-success focus-visible:outline-success"
	case Info:
		base += " border-info bg-info text-on-info focus-visible:outline-info"
	case Warning:
		base += " border-warning bg-warning text-on-warning focus-visible:outline-warning"
	case Danger:
		base += " border-danger bg-danger text-on-danger focus-visible:outline-danger"
	default:
		base += " border-primary bg-primary text-on-primary focus-visible:outline-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark"
	}
	return base
}
