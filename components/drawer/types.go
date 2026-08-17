package drawer

import (
	"strings"
	"unicode"
)

// Side is the edge the drawer slides from.
type Side string

const (
	SideRight  Side = "right"
	SideLeft   Side = "left"
	SideTop    Side = "top"
	SideBottom Side = "bottom"
)

// Width is a preset panel width.
type Width string

const (
	WidthSM   Width = "sm"   // 320px
	WidthMD   Width = "md"   // 420px (default)
	WidthLG   Width = "lg"   // 560px
	WidthXL   Width = "xl"   // 720px
	WidthFull Width = "full" // 100vw on mobile, capped on desktop
)

// Height is a preset panel height for top and bottom drawers.
type Height string

const (
	HeightSM   Height = "sm"   // 320px
	HeightMD   Height = "md"   // 420px (default)
	HeightLG   Height = "lg"   // 560px
	HeightXL   Height = "xl"   // 720px
	HeightFull Height = "full" // 100vh on mobile, capped on desktop
)

// Config holds the drawer configuration. The drawer is opened/closed via Alpine
// state whose name is derived from ID ("{ID}IsOpen"). To open from outside,
// dispatch a window event: `drawer:open` with `{detail: {id: "<ID>"}}`. The
// drawer listens and matches on ID.
//
// For HTMX-driven flows, return an HX-Trigger header: `{"drawer:open": {"id": "<ID>"}}`.
type Config struct {
	// ID uniquely identifies the drawer. Required. Used for the Alpine state
	// var name (`{ID}IsOpen`) and for the aria-labelledby target (`{ID}Title`).
	ID string

	// Title is the drawer heading. Required for accessibility.
	Title string

	// Side the drawer slides in from. Default: SideRight.
	Side Side

	// Width preset. Default: WidthMD.
	// Applies to left and right drawers.
	Width Width

	// Height preset. Default: HeightMD.
	// Applies to top and bottom drawers.
	Height Height

	// BodyID is the id attribute of the inner content container. Exposed so
	// HTMX targets can swap content directly: hx-target="#{BodyID}".
	// Default: "{ID}-body".
	BodyID string

	// Persistent disables click-backdrop and Esc-to-close. Default: false.
	Persistent bool

	// PanelClass allows extra CSS classes on the panel (not the overlay).
	PanelClass string
}

// stateVar returns the Alpine state-variable name for this drawer's open bit.
func (cfg Config) stateVar() string {
	return safeJSIdentifier(cfg.ID, "drawer") + "IsOpen"
}

func (cfg Config) eventIDLiteral() string {
	return jsStringSingle(cfg.ID)
}

// GetBodyID returns the resolved body slot id.
func (cfg Config) getBodyID() string {
	if cfg.BodyID != "" {
		return cfg.BodyID
	}
	return cfg.ID + "-body"
}

// TitleID returns the id used on the drawer's <h2> for aria-labelledby.
func (cfg Config) titleID() string {
	return cfg.ID + "Title"
}

// OverlayClasses returns classes for the backdrop overlay.
func (cfg Config) overlayClasses() string {
	return "fixed inset-0 z-40 bg-backdrop/40 dark:bg-backdrop/60"
}

// PanelClasses returns classes for the sliding panel.
func (cfg Config) panelClasses() string {
	base := "fixed z-50 flex flex-col bg-surface dark:bg-surface-dark border-outline dark:border-outline-dark shadow-elevation-raised"

	switch cfg.Side {
	case SideLeft:
		base += " left-0 top-0 bottom-0 border-r"
		base += cfg.widthClasses()
	case SideTop:
		base += " left-0 right-0 top-0 border-b"
		base += cfg.heightClasses()
	case SideBottom:
		base += " left-0 right-0 bottom-0 border-t"
		base += cfg.heightClasses()
	default:
		base += " right-0 top-0 bottom-0 border-l"
		base += cfg.widthClasses()
	}

	if cfg.PanelClass != "" {
		base += " " + cfg.PanelClass
	}
	return base
}

func (cfg Config) widthClasses() string {
	switch cfg.Width {
	case WidthSM:
		return " w-full max-w-[320px]"
	case WidthLG:
		return " w-full max-w-[560px]"
	case WidthXL:
		return " w-full max-w-[720px]"
	case WidthFull:
		return " w-full max-w-full md:max-w-[90vw]"
	default:
		return " w-full max-w-[420px]"
	}
}

func (cfg Config) heightClasses() string {
	switch cfg.Height {
	case HeightSM:
		return " h-full max-h-[320px]"
	case HeightLG:
		return " h-full max-h-[560px]"
	case HeightXL:
		return " h-full max-h-[720px]"
	case HeightFull:
		return " h-full max-h-full md:max-h-[90vh]"
	default:
		return " h-full max-h-[420px]"
	}
}

// EnterStart returns the Alpine transition enter-start classes.
func (cfg Config) enterStart() string {
	switch cfg.Side {
	case SideLeft:
		return "-translate-x-full"
	case SideTop:
		return "-translate-y-full"
	case SideBottom:
		return "translate-y-full"
	default:
		return "translate-x-full"
	}
}

// EnterEnd returns the Alpine transition enter-end classes.
func (cfg Config) enterEnd() string {
	if cfg.Side == SideTop || cfg.Side == SideBottom {
		return "translate-y-0"
	}
	return "translate-x-0"
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

func jsStringSingle(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\u2028':
			b.WriteString(`\u2028`)
		case '\u2029':
			b.WriteString(`\u2029`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
