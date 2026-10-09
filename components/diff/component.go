package diff

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/expressions"
)

// Instance is a renderable comparison viewer.
type Instance struct {
	cfg                 Config
	expressionOverrides *expressions.Diff
}

// Diff returns a server-rendered comparison of precomputed rows.
// It requires no JavaScript and can be replaced as an ordinary HTMX fragment.
func Diff(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a diff viewer.
func (Instance) Kind() components.Kind { return components.KindDiff }

// Render writes escaped comparison markup in before-then-after order.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	if i.expressionOverrides != nil {
		ctx = expressions.With(ctx, expressions.Set{Diff: *i.expressionOverrides})
	}
	return diffTemplate(i.cfg).Render(ctx, w)
}

// WithExpressions returns a copy with component-scoped text overrides.
// Explicit config text takes precedence; empty expressions inherit defaults.
func (i Instance) WithExpressions(values expressions.Diff) Instance {
	var current expressions.Diff
	if i.expressionOverrides != nil {
		current = *i.expressionOverrides
	}
	merged := expressions.Merge(expressions.Set{Diff: current}, expressions.Set{Diff: values}).Diff
	i.expressionOverrides = &merged
	return i
}

var _ components.Component = Instance{}
