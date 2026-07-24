package textinput

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable text input component.
type Instance struct {
	cfg Config
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
	return textInputTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
