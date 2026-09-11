package palette

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable palette component.
type Instance struct {
	expressionOverrides expressions.Palette
	cfg                 Config
}

// Palette returns a renderable palette component.
func Palette(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a palette.
func (Instance) Kind() components.Kind {
	return components.KindPalette
}

// Render writes the palette markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Palette: i.expressionOverrides})
	return paletteTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Palette) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Palette: i.expressionOverrides}, expressions.Set{Palette: values}).Palette
	return i
}
