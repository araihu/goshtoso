package structuredinput

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable structured input component.
type Instance struct {
	cfg Config
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
	return structuredInputTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
