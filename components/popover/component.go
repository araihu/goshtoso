package popover

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable popover component.
type Instance struct {
	expressionOverrides expressions.Popover
	cfg                 Config
}

// Popover returns a renderable popover primitive.
func Popover(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a popover.
func (Instance) Kind() components.Kind {
	return components.KindPopover
}

// Render writes the popover markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Popover: i.expressionOverrides})
	return popoverTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Popover) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Popover: i.expressionOverrides}, expressions.Set{Popover: values}).Popover
	return i
}
