package kbd

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable keyboard input hint.
type Instance struct {
	text    string
	options []Option
}

// Kbd returns a renderable keyboard input hint.
func Kbd(text string, options ...Option) Instance {
	return Instance{text: text, options: options}
}

// Kind identifies the component as a keyboard input hint.
func (Instance) Kind() components.Kind {
	return components.KindKbd
}

// Render writes the keyboard input hint markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return kbdTemplate(i.text, i.options...).Render(ctx, w)
}

var _ components.Component = Instance{}
