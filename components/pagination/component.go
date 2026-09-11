package pagination

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable pagination component.
type Instance struct {
	expressionOverrides expressions.Pagination
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Pagination: i.expressionOverrides})
	return paginationTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Pagination) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Pagination: i.expressionOverrides}, expressions.Set{Pagination: values}).Pagination
	return i
}
