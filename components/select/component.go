package selectfield

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable select component.
type Instance struct {
	expressionOverrides expressions.Select
	cfg                 Config
}

// Select returns a renderable select component. The hidden submission input
// uses cfg.ID; external draft restoration can set that input's value and
// dispatch a bubbling input or change event to synchronize the visible option.
func Select(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a select.
func (Instance) Kind() components.Kind {
	return components.KindSelect
}

// Render writes the select markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Select: i.expressionOverrides})
	return selectTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Select) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Select: i.expressionOverrides}, expressions.Set{Select: values}).Select
	return i
}
