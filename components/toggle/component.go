package toggle

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable toggle component.
type Instance struct {
	cfg Config
}

// Toggle returns a renderable toggle component.
func Toggle(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a toggle.
func (Instance) Kind() components.Kind {
	return components.KindToggle
}

// Render writes the toggle markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return toggleTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
