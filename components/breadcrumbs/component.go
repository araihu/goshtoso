package breadcrumbs

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable breadcrumbs component.
type Instance struct {
	cfg Config
}

// Breadcrumbs returns a renderable breadcrumbs component.
func Breadcrumbs(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as breadcrumbs.
func (Instance) Kind() components.Kind {
	return components.KindBreadcrumbs
}

// Render writes the breadcrumbs markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return breadcrumbsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
