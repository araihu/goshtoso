package link

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable link component.
type Instance struct {
	href    string
	options []Option
}

// Link returns a renderable link component.
func Link(href string, options ...Option) Instance {
	return Instance{href: href, options: options}
}

// Kind identifies the component as a link.
func (Instance) Kind() components.Kind {
	return components.KindLink
}

// Render writes the link markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return linkTemplate(i.href, i.options...).Render(ctx, w)
}

var _ components.Component = Instance{}
