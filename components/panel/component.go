package panel

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable neutral panel.
type Instance struct {
	cfg Config
}

// Panel returns a renderable neutral application surface.
func Panel(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a panel.
func (Instance) Kind() components.Kind {
	return components.KindPanel
}

// Render writes the panel markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	return panelTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
