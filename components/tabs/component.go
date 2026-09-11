package tabs

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable tabs component.
type Instance struct {
	expressionOverrides expressions.Tabs
	cfg                 Config
}

// Tabs returns a renderable tabs component.
func Tabs(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as tabs.
func (Instance) Kind() components.Kind {
	return components.KindTabs
}

// Render writes the tabs markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Tabs: i.expressionOverrides})
	return tabsTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Tabs) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Tabs: i.expressionOverrides}, expressions.Set{Tabs: values}).Tabs
	return i
}
