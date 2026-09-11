package splitbutton

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable split button component.
type Instance struct {
	expressionOverrides expressions.SplitButton
	cfg                 Config
}

// SplitButton returns a primary action with an adjacent menu trigger.
func SplitButton(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a split button.
func (Instance) Kind() components.Kind {
	return components.KindSplitButton
}

// Render writes the split button markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{SplitButton: i.expressionOverrides})
	return splitButtonTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.SplitButton) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{SplitButton: i.expressionOverrides}, expressions.Set{SplitButton: values}).SplitButton
	return i
}
