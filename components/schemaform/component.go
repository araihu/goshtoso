package schemaform

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable schema form fields component.
type Instance struct {
	cfg FieldsConfig
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
	return fieldsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
