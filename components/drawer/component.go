package drawer

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable drawer component.
type Instance struct {
	expressionOverrides expressions.Drawer
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Drawer: i.expressionOverrides})
	return drawerTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Drawer) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Drawer: i.expressionOverrides}, expressions.Set{Drawer: values}).Drawer
	return i
}
