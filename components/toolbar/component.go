package toolbar

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable toolbar component.
type Instance struct {
	cfg Config
}

// Toolbar returns a renderable toolbar.
func Toolbar(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a toolbar.
func (Instance) Kind() components.Kind {
	return components.KindToolbar
}

// Render writes the toolbar markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	return toolbarTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
