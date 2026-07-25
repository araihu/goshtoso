package pagination

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable pagination component.
type Instance struct {
	cfg Config
}

// Pagination returns a renderable pagination component.
func Pagination(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as pagination.
func (Instance) Kind() components.Kind {
	return components.KindPagination
}

// Render writes the pagination markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return paginationTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
