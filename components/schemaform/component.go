package schemaform

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable schema form fields component.
type Instance struct {
	expressionOverrides expressions.SchemaForm
	cfg                 FieldsConfig
}

// Fields returns a renderable schema form fields component.
func Fields(cfg FieldsConfig) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as schema form fields.
func (Instance) Kind() components.Kind {
	return components.KindSchemaFormFields
}

// Render writes the schema form fields markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{SchemaForm: i.expressionOverrides})
	return fieldsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.SchemaForm) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{SchemaForm: i.expressionOverrides}, expressions.Set{SchemaForm: values}).SchemaForm
	return i
}
