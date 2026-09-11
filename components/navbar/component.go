package navbar

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable navbar component.
type Instance struct {
	expressionOverrides expressions.Navbar
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Navbar: i.expressionOverrides})
	if err := i.cfg.Validate(); err != nil {
		return err
	}
	return navbarTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Navbar) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Navbar: i.expressionOverrides}, expressions.Set{Navbar: values}).Navbar
	return i
}
