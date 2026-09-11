package fileinput

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable file input component.
type Instance struct {
	expressionOverrides expressions.FileInput
	cfg                 Config
}

// FileInput returns a renderable file input component.
func FileInput(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a file input.
func (Instance) Kind() components.Kind {
	return components.KindFileInput
}

// Render writes the file input markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{FileInput: i.expressionOverrides})
	return fileInputTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.FileInput) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{FileInput: i.expressionOverrides}, expressions.Set{FileInput: values}).FileInput
	return i
}
