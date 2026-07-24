package tooltip

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable tooltip component.
type Instance struct {
	id      string
	label   string
	options []Option
}

// Tooltip returns a renderable tooltip component.
func Tooltip(id, label string, options ...Option) Instance {
	return Instance{id: id, label: label, options: options}
}

// Kind identifies the component as a tooltip.
func (Instance) Kind() components.Kind {
	return components.KindTooltip
}

// Render writes the tooltip markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return tooltipTemplate(i.id, i.label, i.options...).Render(ctx, w)
}

var _ components.Component = Instance{}
