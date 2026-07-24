package button

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable button component.
type Instance struct {
	options []Option
}

// Button returns a renderable button component.
func Button(options ...Option) Instance {
	return Instance{options: options}
}

// Kind identifies the component as a button.
func (Instance) Kind() components.Kind {
	return components.KindButton
}

// Render writes the button markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return buttonTemplate(i.options...).Render(ctx, w)
}

var _ components.Component = Instance{}
