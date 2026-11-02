// Package layout arranges arbitrary components and optionally moves direct items.
package layout

import (
	"fmt"

	"github.com/a-h/templ"
)

// Mode selects natural-height container geometry.
type Mode string

const (
	// ModeStack is a single column and the default.
	ModeStack Mode = ""
	// ModeGrid uses responsive equal-width columns.
	ModeGrid Mode = "grid"
	// ModeMasonry measures row spans while preserving DOM order.
	ModeMasonry Mode = "masonry"
)

// Span controls the width of an item within its parent.
type Span string

const (
	// SpanAuto occupies one column.
	SpanAuto Span = ""
	// SpanFullWidth occupies all columns.
	SpanFullWidth Span = "full"
)

// Gap selects spacing between items.
type Gap string

const (
	// GapMedium is the default 16 CSS pixels.
	GapMedium Gap = ""
	// GapSmall is 12 CSS pixels.
	GapSmall Gap = "small"
	// GapLarge is 24 CSS pixels.
	GapLarge Gap = "large"
)

// Config defines a labelled direct-child collection.
type Config struct {
	// ID is the stable document-unique container DOM ID, at most 256 UTF-8 bytes.
	ID string
	// Label names the collection without choosing a heading level.
	Label string
	// Mode selects Stack, Grid or Masonry geometry.
	Mode Mode
	// Columns is the maximum responsive column count, 1 through 4; zero means 1.
	Columns int
	// MinItemWidthRem is the desired minimum track width, 8 through 40; zero means 12.
	MinItemWidthRem float64
	// Gap selects 12, 16 or 24 CSS pixels between items.
	Gap Gap
	// Items renders direct children in canonical consumer order, without sorting.
	Items []Item
	// Reorder opts into generic movement. Nil preserves geometry without handles.
	Reorder *ReorderConfig
	// RootClass appends classes without overriding structural grid properties.
	RootClass string
	// RootAttrs adds supported semantic, data, HTMX and Alpine event attributes.
	RootAttrs templ.Attributes
}

// Item preserves the consumer's identity and arbitrary content.
type Item struct {
	// ID is an opaque stable identity, unique within its movement scope.
	ID string
	// Label names the item and its movement handle.
	Label string
	// Content retains its own renderer, semantics and natural height. Nil is valid.
	Content templ.Component
	// Span chooses one or all parent columns.
	Span Span
	// RootClass appends classes; positive padding, borders and margins are measured.
	RootClass string
	// RootAttrs adds supported attributes, including an optional document-unique id.
	RootAttrs templ.Attributes
}

// ReorderConfig opts into movement. Empty Group permits only local movement.
type ReorderConfig struct {
	// Group permits transfers among containers with the same nonempty value.
	Group string
	// DragFromContent also allows mouse or pen pickup from non-interactive item content.
	// Links, form controls and editable content keep their actions. Touch uses the handle.
	DragFromContent bool
	// Disabled omits handles and movement listeners while preserving geometry.
	Disabled bool
}

// ValidationError identifies invalid public configuration before markup is written.
type ValidationError struct {
	Path string
	Code string
}

func (e ValidationError) Error() string { return "layout: " + e.Path + ": " + e.Code }
func (c Config) enabled() bool          { return c.Reorder != nil && !c.Reorder.Disabled }
func (c Config) group() string {
	if c.Reorder != nil {
		return c.Reorder.Group
	}
	return ""
}
func (c Config) gapPixels() int {
	switch c.Gap {
	case GapSmall:
		return 12
	case GapLarge:
		return 24
	}
	return 16
}

func (c Config) rootStyle() templ.SafeCSS {
	max := c.Columns
	if c.Mode == ModeStack {
		max = 1
	}
	return templ.SafeCSS(fmt.Sprintf("--layout-max:%d;--layout-min:%grem;--layout-gap:%dpx", max, c.MinItemWidthRem, c.gapPixels()))
}

func (c Config) dragFromContent() bool { return c.enabled() && c.Reorder.DragFromContent }
