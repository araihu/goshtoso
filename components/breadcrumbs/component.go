package breadcrumbs

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable breadcrumbs component.
type Instance struct {
	expressionOverrides expressions.Breadcrumbs
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Breadcrumbs: i.expressionOverrides})
	return breadcrumbsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Breadcrumbs) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Breadcrumbs: i.expressionOverrides}, expressions.Set{Breadcrumbs: values}).Breadcrumbs
	return i
}
