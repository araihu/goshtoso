package emptystate

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable empty-state component.
type Instance struct {
	expressionOverrides expressions.EmptyState
	cfg                 Config
}

// EmptyState returns a renderable empty state.
func EmptyState(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an empty state.
func (Instance) Kind() components.Kind {
	return components.KindEmptyState
}

// Render writes the empty-state markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{EmptyState: instance.expressionOverrides})
	return emptyStateTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.EmptyState) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{EmptyState: i.expressionOverrides}, expressions.Set{EmptyState: values}).EmptyState
	return i
}
