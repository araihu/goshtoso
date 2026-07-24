package combobox

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable combobox component.
type Instance struct {
	cfg   Config
	state State
}

// Combobox returns a renderable combobox component.
func Combobox(cfg Config, state State) Instance {
	return Instance{cfg: cfg, state: state}
}

// Kind identifies the component as a combobox.
func (Instance) Kind() components.Kind {
	return components.KindCombobox
}

// Render writes the combobox markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return comboboxTemplate(i.cfg, i.state).Render(ctx, w)
}

var _ components.Component = Instance{}
