package dropdown

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable dropdown component.
type Instance struct {
	cfg Config
}

// Dropdown returns a renderable dropdown component.
func Dropdown(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a dropdown.
func (Instance) Kind() components.Kind {
	return components.KindDropdown
}

// Render writes the dropdown markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return dropdownTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
