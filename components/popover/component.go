package popover

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable popover component.
type Instance struct {
	cfg Config
}

// Popover returns a renderable popover primitive.
func Popover(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a popover.
func (Instance) Kind() components.Kind {
	return components.KindPopover
}

// Render writes the popover markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return popoverTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
