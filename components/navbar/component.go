package navbar

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable navbar component.
type Instance struct {
	cfg Config
}

// Navbar returns a renderable navbar component.
func Navbar(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a navbar.
func (Instance) Kind() components.Kind {
	return components.KindNavbar
}

// Render writes the navbar markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return navbarTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
