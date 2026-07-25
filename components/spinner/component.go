package spinner

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable spinner component.
type Instance struct {
	cfg Config
}

// Spinner returns a renderable spinner component.
func Spinner(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a spinner.
func (Instance) Kind() components.Kind {
	return components.KindSpinner
}

// Render writes the spinner markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return spinnerTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
