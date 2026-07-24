package tabs

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable tabs component.
type Instance struct {
	cfg Config
}

// Tabs returns a renderable tabs component.
func Tabs(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as tabs.
func (Instance) Kind() components.Kind {
	return components.KindTabs
}

// Render writes the tabs markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return tabsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
