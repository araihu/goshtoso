package icon

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable sprite icon.
type Instance struct {
	cfg Config
}

// Icon returns a renderable sprite icon.
func Icon(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an icon.
func (Instance) Kind() components.Kind {
	return components.KindIcon
}

// Render writes icon markup after validating its symbol reference.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	if err := instance.cfg.validate(); err != nil {
		return err
	}
	return iconTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
