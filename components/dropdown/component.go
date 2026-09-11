package dropdown

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable dropdown component.
type Instance struct {
	expressionOverrides expressions.Dropdown
	cfg                 Config
}

// Dropdown returns a renderable dropdown component.
func Dropdown(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a dropdown.
func (Instance) Kind() components.Kind {
	return components.KindDropdown
}

// Render writes the dropdown markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Dropdown: i.expressionOverrides})
	return dropdownTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Dropdown) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Dropdown: i.expressionOverrides}, expressions.Set{Dropdown: values}).Dropdown
	return i
}
