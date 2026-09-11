package actiongroup

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable ActionGroup component.
type Instance struct {
	expressionOverrides expressions.ActionGroup
	cfg                 Config
}

// ActionGroup returns a responsive action group.
func ActionGroup(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an action group.
func (Instance) Kind() components.Kind {
	return components.KindActionGroup
}

// Render writes the action group markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{ActionGroup: i.expressionOverrides})
	return actionGroupTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.ActionGroup) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{ActionGroup: i.expressionOverrides}, expressions.Set{ActionGroup: values}).ActionGroup
	return i
}
