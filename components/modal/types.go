package modal

import (
	"strings"
	"unicode"
)

// Tone represents alert-dialog semantic color treatments.
type Tone string

const (
	ToneDefault Tone = "default"
	ToneSuccess Tone = "success"
	ToneInfo    Tone = "info"
	ToneWarning Tone = "warning"
	ToneDanger  Tone = "danger"
)

// HTMXConfig holds HTMX attributes for a modal action button.
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

// ButtonAction holds optional HTMX and Alpine.js actions for a button
type ButtonAction struct {
	// OnClick is a custom Alpine.js expression (appended after modal close)
	OnClick string
	// HTMX configures server-side action behavior.
	HTMX *HTMXConfig
}

// Config holds configuration for the modal component
type Config struct {
	// ID is a unique identifier used for aria-labelledby (required)
	ID string
	// Title is the modal heading
	Title string
	// Body is the modal body text
	Body string
	// TriggerLabel is the trigger button label
	TriggerLabel string
	// PrimaryLabel is the primary action button label
	PrimaryLabel string
	// PrimaryAction holds optional HTMX/JS actions for the primary button
	PrimaryAction *ButtonAction
	// SecondaryLabel is the secondary/dismiss button label.
	SecondaryLabel string
	// SecondaryAction holds optional HTMX/JS actions for the secondary button
	SecondaryAction *ButtonAction
	// PanelClass allows additional CSS classes on the dialog.
	PanelClass string
}

// AlertDialogConfig holds configuration for an alert dialog.
type AlertDialogConfig struct {
	// ID is a unique identifier used for aria-labelledby.
	ID string
	// Title is the alert-dialog heading.
	Title string
	// Body is the alert-dialog body text.
	Body string
	// TriggerLabel is the trigger button label.
	TriggerLabel string
	// ActionLabel is the single action button label.
	ActionLabel string
	// Action holds optional HTMX/JS actions for the action button.
	Action *ButtonAction
	// Tone determines the semantic color treatment.
	Tone Tone
	// PanelClass allows additional CSS classes on the dialog.
	PanelClass string
}

func (cfg Config) stateVar() string {
	return safeJSIdentifier(cfg.ID, "modal") + "IsOpen"
}

func (cfg Config) titleID() string {
	base := cfg.ID
	if base == "" {
		base = "modal"
	}
	return base + "Title"
}

func (cfg AlertDialogConfig) stateVar() string {
	return safeJSIdentifier(cfg.ID, "alertDialog") + "IsOpen"
}

func (cfg AlertDialogConfig) titleID() string {
	base := cfg.ID
	if base == "" {
		base = "alertDialog"
	}
	return base + "Title"
}

// TriggerClasses returns the trigger button CSS classes
func (cfg Config) triggerClasses() string {
	return "whitespace-nowrap rounded-radius border border-primary dark:border-primary-dark bg-primary px-4 py-2 text-center text-sm font-medium tracking-wide text-on-primary transition motion-reduce:transition-none hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:opacity-100 active:outline-offset-0 dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark"
}

func (cfg AlertDialogConfig) triggerClasses() string {
	base := "w-36 whitespace-nowrap rounded-radius border px-4 py-2 text-center text-sm font-medium tracking-wide transition motion-reduce:transition-none hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 active:opacity-100 active:outline-offset-0"

	switch cfg.Tone {
	case ToneSuccess:
		base += " border-success bg-success text-on-success focus-visible:outline-success"
	case ToneInfo:
		base += " border-info bg-info text-on-info focus-visible:outline-info"
	case ToneWarning:
		base += " border-warning bg-warning text-on-warning focus-visible:outline-warning"
	case ToneDanger:
		base += " border-danger bg-danger text-on-danger focus-visible:outline-danger"
	default:
		base += " border-primary bg-primary text-on-primary focus-visible:outline-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark"
	}

	return base
}

// DialogClasses returns the modal dialog container CSS classes
func (cfg Config) dialogClasses() string {
	return dialogClasses(cfg.PanelClass)
}

func (cfg AlertDialogConfig) dialogClasses() string {
	return dialogClasses(cfg.PanelClass)
}

func dialogClasses(panelClass string) string {
	base := "flex max-w-lg flex-col gap-4 overflow-hidden rounded-radius border border-outline bg-surface text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"
	if panelClass != "" {
		base += " " + panelClass
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
func (cfg Config) headerClasses() string {
	return "flex items-center justify-between border-b border-outline bg-surface-alt/60 p-4 dark:border-outline-dark dark:bg-surface-dark/20"
}

func (cfg AlertDialogConfig) headerClasses() string {
	return "flex items-center justify-between border-b border-outline bg-surface-alt/60 px-4 py-2 dark:border-outline-dark dark:bg-surface-dark/20"
}

func (cfg AlertDialogConfig) iconBadgeClasses() string {
	base := "flex items-center justify-center rounded-full p-1"
	switch cfg.Tone {
	case ToneSuccess:
		base += " bg-success/20 text-success"
	case ToneInfo:
		base += " bg-info/20 text-info"
	case ToneWarning:
		base += " bg-warning/20 text-warning"
	case ToneDanger:
		base += " bg-danger/20 text-danger"
	}
	return base
}

func (cfg AlertDialogConfig) actionClasses() string {
	base := "w-full whitespace-nowrap rounded-radius border px-4 py-2 text-center text-sm font-semibold tracking-wide transition motion-reduce:transition-none hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 active:opacity-100 active:outline-offset-0"
	switch cfg.Tone {
	case ToneSuccess:
		base += " border-success bg-success text-on-success focus-visible:outline-success"
	case ToneInfo:
		base += " border-info bg-info text-on-info focus-visible:outline-info"
	case ToneWarning:
		base += " border-warning bg-warning text-on-warning focus-visible:outline-warning"
	case ToneDanger:
		base += " border-danger bg-danger text-on-danger focus-visible:outline-danger"
	default:
		base += " border-primary bg-primary text-on-primary focus-visible:outline-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark dark:focus-visible:outline-primary-dark"
	}
	return base
}
