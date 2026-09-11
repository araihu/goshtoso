package structuredinput

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable structured input component.
type Instance struct {
	expressionOverrides expressions.StructuredInput
	cfg                 Config
}

// StructuredInput returns a renderable structured input component.
func StructuredInput(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a structured input.
func (Instance) Kind() components.Kind {
	return components.KindStructuredInput
}

// Render writes the structured input markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{StructuredInput: i.expressionOverrides})
	return structuredInputTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.StructuredInput) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{StructuredInput: i.expressionOverrides}, expressions.Set{StructuredInput: values}).StructuredInput
	return i
}
