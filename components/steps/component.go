package steps

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable steps component.
type Instance struct {
	expressionOverrides expressions.Steps
	cfg                 Config
}

// Steps returns a renderable steps component.
func Steps(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as steps.
func (Instance) Kind() components.Kind {
	return components.KindSteps
}

// Render writes the steps markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Steps: i.expressionOverrides})
	return stepsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Steps) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Steps: i.expressionOverrides}, expressions.Set{Steps: values}).Steps
	return i
}
