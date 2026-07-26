package pageheader

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable page-header component.
type Instance struct {
	cfg Config
}

// PageHeader returns a renderable page header.
func PageHeader(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a page header.
func (Instance) Kind() components.Kind {
	return components.KindPageHeader
}

// Render writes the page-header markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	return pageHeaderTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
