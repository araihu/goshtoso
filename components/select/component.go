package selectfield

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable select component.
type Instance struct {
	cfg Config
}

// Select returns a renderable select component.
func Select(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a select.
func (Instance) Kind() components.Kind {
	return components.KindSelect
}

// Render writes the select markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return selectTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
