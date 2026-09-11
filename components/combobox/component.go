package combobox

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable combobox component.
type Instance struct {
	expressionOverrides expressions.Combobox
	cfg                 Config
	state               State
}

// Combobox returns a renderable combobox component.
func Combobox(cfg Config, state State) Instance {
	return Instance{cfg: cfg, state: state}
}

// Kind identifies the component as a combobox.
func (Instance) Kind() components.Kind {
	return components.KindCombobox
}

// Render writes the combobox markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Combobox: i.expressionOverrides})
	return comboboxTemplate(i.cfg, i.state).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Combobox) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Combobox: i.expressionOverrides}, expressions.Set{Combobox: values}).Combobox
	return i
}
