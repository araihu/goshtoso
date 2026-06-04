// Package palette renders a generic color picker grid: a Tailwind hue×shade
// matrix plus optional white/black swatches, a Reset action, and a hex input.
// It has no trigger of its own — wrap it in a Select shell (Shell: true) to get
// a dropdown.
//
// On every pick the palette dispatches a bubbling `select-close` event whose
// detail is the chosen value ("blue-700" / "white" / "" / "#aabbcc"). A hosting
// Select shell closes on that event; consumers wire side-effects with their own
// `x-on:select-close` listener. If AlpineModel is set, the palette's own root
// assigns it from $event.detail.
package palette

import "github.com/a-h/templ"

// DefaultHues lists Tailwind v4's named hue families in display order.
var DefaultHues = []string{
	"red", "orange", "amber", "yellow", "lime",
	"green", "emerald", "teal", "cyan", "sky",
	"blue", "indigo", "violet", "purple", "fuchsia",
	"pink", "rose", "slate", "gray", "zinc",
	"neutral", "stone",
}

// DefaultShades lists the shade steps for each hue.
var DefaultShades = []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950"}

// Config configures a Palette.
type Config struct {
	// ID is the wrapper element id.
	ID string
	// AlpineModel, when set, is a JS lvalue assigned the picked value (from the
	// dispatched event detail). Quote-free identifier/path expected, e.g.
	// "picked" or "form.color".
	AlpineModel string
	// Hues / Shades override the default Tailwind sets.
	Hues   []string
	Shades []string
	// HideNeutral hides the white/black quick swatches (shown by default).
	HideNeutral bool
	// HideReset hides the Reset action (shown by default).
	HideReset bool
	// ShowHex adds a native color input + hex text field (off by default).
	ShowHex bool
	// Class appends classes to the wrapper.
	RootClass string
	// LazyWhen is an Alpine expression; when non-empty, the swatch grid is
	// wrapped in <template x-if=...> so it mounts only when the expression is
	// truthy (e.g. inside a Select dropdown, pass the dropdown's open
	// expression). Keeps the initial DOM light.
	LazyWhen string
}

func (c Config) hues() []string {
	if len(c.Hues) > 0 {
		return c.Hues
	}
	return DefaultHues
}

func (c Config) shades() []string {
	if len(c.Shades) > 0 {
		return c.Shades
	}
	return DefaultShades
}

// ContainerClasses returns wrapper classes.
func (c Config) ContainerClasses() string {
	base := "p-2 space-y-2"
	if c.RootClass != "" {
		base += " " + c.RootClass
	}
	return base
}

// swatchStyle returns the swatch's inline style as a raw attribute map. Using
// templ.Attributes (a string value) routes through HTML-escaping only, NOT
// templ's CSS sanitizer — so the value is emitted verbatim with no trailing
// semicolon. The hue/shade tokens contain no HTML-special characters.
func swatchStyle(hue, shade string) templ.Attributes {
	return templ.Attributes{"style": "background-color: var(--color-" + hue + "-" + shade + ")"}
}

// modelAssignExpr is the (quote-free → templ-safe) Alpine expression the root
// runs on select-close to set AlpineModel from the event detail. Empty when no
// model is configured.
func (c Config) modelAssignExpr() string {
	if c.AlpineModel == "" {
		return ""
	}
	return c.AlpineModel + " = $event.detail"
}
