package actiongroup

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable ActionGroup component.
type Instance struct {
	cfg Config
}

// ActionGroup returns a responsive action group.
func ActionGroup(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an action group.
func (Instance) Kind() components.Kind {
	return components.KindActionGroup
}

// Render writes the action group markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return actionGroupTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
