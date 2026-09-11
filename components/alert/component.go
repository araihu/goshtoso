package alert

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable alert component.
type Instance struct {
	expressionOverrides expressions.Alert
	cfg                 Config
}

// Alert returns a renderable alert component.
func Alert(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an alert.
func (Instance) Kind() components.Kind {
	return components.KindAlert
}

// Render writes the alert markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Alert: i.expressionOverrides})
	return alertTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Alert) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Alert: i.expressionOverrides}, expressions.Set{Alert: values}).Alert
	return i
}
