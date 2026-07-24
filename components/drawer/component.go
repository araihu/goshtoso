package drawer

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable drawer component.
type Instance struct {
	cfg Config
}

// Drawer returns a renderable drawer component.
func Drawer(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a drawer.
func (Instance) Kind() components.Kind {
	return components.KindDrawer
}

// Render writes the drawer markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return drawerTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
