// Package palette renders a generic color picker grid: a Tailwind hue×shade
// matrix plus optional white/black swatches, a Reset action, and a hex input.
// It has no trigger of its own — wrap it in a Select shell (Shell: true) to get
// a dropdown.
//
// On every pick the palette dispatches a bubbling `select-close` event whose
// detail is the chosen value ("blue-700" / "white" / "" / "#aabbcc"). A hosting
// Select shell closes on that event; consumers wire side-effects with their own
// `x-on:select-close` listener. If Alpine.Model is set, the palette's own root
// assigns it from $event.detail.
package palette

import (
	"encoding/base64"
	"encoding/json"

	"github.com/a-h/templ"
)

// defaultHues lists Tailwind v4's named hue families in display order.
var defaultHues = [...]string{
	"red", "orange", "amber", "yellow", "lime",
	"green", "emerald", "teal", "cyan", "sky",
	"blue", "indigo", "violet", "purple", "fuchsia",
	"pink", "rose", "slate", "gray", "zinc",
	"neutral", "stone",
}

// defaultShades lists the shade steps for each hue.
var defaultShades = [...]string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950"}

// AlpineConfig wires client-side Alpine bindings.
type AlpineConfig struct {
	Model string
}

// Config configures a Palette.
type Config struct {
	// ID is the wrapper element id.
	ID string
	// Alpine wires client-side state.
	Alpine *AlpineConfig
	// Hues / Shades override the default Tailwind sets.
	Hues   []string
	Shades []string
	// HideNeutral hides the white/black quick swatches (shown by default).
	HideNeutral bool
	// HideReset hides the Reset action (shown by default).
	HideReset bool
	// ShowHex adds a native color input + hex text field (off by default).
	ShowHex bool
	// RootClass appends classes to the wrapper.
	RootClass string
	// LazyWhen is an Alpine expression; when non-empty, the swatch grid is
	// generated inside <template x-if=...> only when the expression is truthy
	// (e.g. inside a Select dropdown, pass the dropdown's open expression).
	// This keeps both the initial DOM and the initial HTML payload light.
	LazyWhen string
}

func (c Config) hues() []string {
	if len(c.Hues) > 0 {
		return c.Hues
	}
	return defaultHues[:]
}

func (c Config) shades() []string {
	if len(c.Shades) > 0 {
		return c.Shades
	}
	return defaultShades[:]
}

// ContainerClasses returns wrapper classes.
func (c Config) containerClasses() string {
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
// runs on select-close to set Alpine.Model from the event detail. Empty when no
// model is configured.
func (c Config) modelAssignExpr() string {
	if c.Alpine == nil || c.Alpine.Model == "" {
		return ""
	}
	return c.Alpine.Model + " = $event.detail"
}

func (c Config) swatchesJSON() string {
	data := struct {
		Hues        []string `json:"hues"`
		Shades      []string `json:"shades"`
		HideNeutral bool     `json:"hideNeutral"`
	}{c.hues(), c.shades(), c.HideNeutral}
	encoded, err := json.Marshal(data)
	if err != nil {
		return `{"hues":[],"shades":[],"hideNeutral":false}`
	}
	return string(encoded)
}

func (c Config) swatchesData() string {
	return base64.StdEncoding.EncodeToString([]byte(c.swatchesJSON()))
}
