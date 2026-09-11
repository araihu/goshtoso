package toolbar

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable toolbar component.
type Instance struct {
	expressionOverrides expressions.Toolbar
	cfg                 Config
}

// Toolbar returns a renderable toolbar.
func Toolbar(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a toolbar.
func (Instance) Kind() components.Kind {
	return components.KindToolbar
}

// Render writes the toolbar markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Toolbar: instance.expressionOverrides})
	return toolbarTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Toolbar) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Toolbar: i.expressionOverrides}, expressions.Set{Toolbar: values}).Toolbar
	return i
}
