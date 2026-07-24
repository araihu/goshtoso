package fileinput

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable file input component.
type Instance struct {
	cfg Config
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
	return fileInputTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
