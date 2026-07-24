package palette

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable palette component.
type Instance struct {
	cfg Config
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
	return paletteTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
