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
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

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
// runs on select-close to set Alpine.Model from the event detail. Empty when no
// model is configured.
func (c Config) modelAssignExpr() string {
	if c.Alpine == nil || c.Alpine.Model == "" {
		return ""
	}
	return c.Alpine.Model + " = $event.detail"
}

func (c Config) alpineData() string {
	var sb strings.Builder
	sb.WriteString(`{
			hovered: '',
			selectedHex: '#000000',
			hexInput: '',
			hexInvalid: false,
			swatchHues: `)
	sb.WriteString(jsStringArray(c.hues()))
	sb.WriteString(`,
			swatchShades: `)
	sb.WriteString(jsStringArray(c.shades()))
	sb.WriteString(`,
			hideNeutral: `)
	if c.HideNeutral {
		sb.WriteString("true")
	} else {
		sb.WriteString("false")
	}
	sb.WriteString(`,
			pick(value, el) {
				this.syncHex(value, el)
				this.$dispatch('select-close', value)
			},
			commitHex(value) {
				const hex = this.normalizeHex(value)
				if (!hex) {
					this.hexInvalid = true
					return
				}
				this.hexInvalid = false
				this.selectedHex = hex
				this.hexInput = hex
				this.$dispatch('select-close', hex)
			},
			previewHex(value) {
				this.hexInput = value
				const hex = this.normalizeHex(value)
				this.hexInvalid = this.hexInput !== '' && !hex
				if (hex) {
					this.selectedHex = hex
					this.hexInput = hex
					this.$dispatch('select-close', hex)
				}
			},
			syncHex(value, el) {
				this.hexInvalid = false
				if (!value) {
					this.selectedHex = '#000000'
					this.hexInput = ''
					return
				}
				if (value[0] === '#') {
					const hex = this.normalizeHex(value)
					if (hex) {
						this.selectedHex = hex
						this.hexInput = hex
					}
					return
				}
				if (value === 'white') {
					this.selectedHex = '#ffffff'
					this.hexInput = '#ffffff'
					return
				}
				if (value === 'black') {
					this.selectedHex = '#000000'
					this.hexInput = '#000000'
					return
				}
				if (el) {
					const hex = this.colorToHex(getComputedStyle(el).backgroundColor)
					this.selectedHex = hex
					this.hexInput = hex
				}
			},
			normalizeHex(value) {
				const raw = (value || '').trim()
				const short = raw.match(/^#([0-9a-fA-F]{3})$/)
				if (short) {
					return '#' + short[1].split('').map(ch => ch + ch).join('').toLowerCase()
				}
				const full = raw.match(/^#([0-9a-fA-F]{6})$/)
				return full ? '#' + full[1].toLowerCase() : ''
			},
			colorToHex(color) {
				if (!this._ctx) {
					const canvas = document.createElement('canvas')
					canvas.width = 1
					canvas.height = 1
					this._ctx = canvas.getContext('2d', { willReadFrequently: true })
				}
				const ctx = this._ctx
				ctx.clearRect(0, 0, 1, 1)
				ctx.fillStyle = '#000000'
				ctx.fillStyle = color
				ctx.fillRect(0, 0, 1, 1)
				const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data
				return '#' + [r, g, b].map(n => n.toString(16).padStart(2, '0')).join('')
			},
			escapeAttr(value) {
				return String(value)
					.replace(/&/g, '&amp;')
					.replace(/"/g, '&quot;')
					.replace(/</g, '&lt;')
					.replace(/>/g, '&gt;')
			},
			swatchButton(cls, classes, style) {
				const safeCls = this.escapeAttr(cls)
				const styleAttr = style ? ' style="' + this.escapeAttr(style) + '"' : ''
				return '<button type="button" data-cls="' + safeCls + '" class="' + classes + '"' + styleAttr + ' title="' + safeCls + '"></button>'
			},
			swatchGridHTML() {
				const standard = 'h-5 w-full rounded-sm border border-outline/30 dark:border-outline-dark/30 transition-transform hover:scale-125 hover:ring-2 hover:ring-primary focus:scale-125 dark:hover:ring-primary-dark'
				const neutral = 'h-5 w-full rounded-sm border border-outline/60 transition-transform hover:scale-125 hover:ring-2 hover:ring-primary focus:scale-125 dark:border-outline-dark/60 dark:hover:ring-primary-dark'
				let html = ''
				if (!this.hideNeutral) {
					html += this.swatchButton('white', neutral + ' bg-white', '')
					html += this.swatchButton('black', neutral + ' bg-black', '')
				}
				this.swatchHues.forEach(hue => {
					this.swatchShades.forEach(shade => {
						const cls = hue + '-' + shade
						html += this.swatchButton(cls, standard, 'background-color: var(--color-' + cls + ')')
					})
				})
				return html
			},
			handleSwatchEvent(event, action) {
				const button = event.target.closest('button[data-cls]')
				if (!button || !event.currentTarget.contains(button)) return
				if (action === 'pick') {
					this.pick(button.dataset.cls, button)
					return
				}
				this.hovered = button.dataset.cls
			},
		}`)
	return sb.String()
}

func jsStringArray(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = strconv.Quote(value)
	}
	return "[" + strings.Join(quoted, ",") + "]"
}
