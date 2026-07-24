package steps

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable steps component.
type Instance struct {
	cfg Config
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
	return stepsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
