package navbar

import (
	"context"
	"io"

	"github.com/a-h/templ"
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

// SecondaryRow renders only the secondary navbar row.
func SecondaryRow(cfg SecondaryConfig) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := cfg.Validate(); err != nil {
			return err
		}
		if !cfg.hasContent() {
			return nil
		}
		return secondaryRowTemplate(cfg).Render(ctx, w)
	})
}

// Kind identifies the component as a navbar.
func (Instance) Kind() components.Kind {
	return components.KindNavbar
}

// Render writes the navbar markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	if err := i.cfg.Validate(); err != nil {
		return err
	}
	return navbarTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
