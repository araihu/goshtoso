package skeleton

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable skeleton component.
type Instance struct {
	expressionOverrides expressions.Skeleton
	cfg                 Config
}

// Skeleton returns a renderable skeleton loading state.
func Skeleton(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a skeleton.
func (Instance) Kind() components.Kind {
	return components.KindSkeleton
}

// Render writes the skeleton markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Skeleton: instance.expressionOverrides})
	return skeletonTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Skeleton) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Skeleton: i.expressionOverrides}, expressions.Set{Skeleton: values}).Skeleton
	return i
}
