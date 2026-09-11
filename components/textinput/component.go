package textinput

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable text input component.
type Instance struct {
	expressionOverrides expressions.TextInput
	cfg                 Config
}

// TextInput returns a renderable text input component.
func TextInput(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a text input.
func (Instance) Kind() components.Kind {
	return components.KindTextInput
}

// Render writes the text input markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{TextInput: i.expressionOverrides})
	return textInputTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.TextInput) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{TextInput: i.expressionOverrides}, expressions.Set{TextInput: values}).TextInput
	return i
}
