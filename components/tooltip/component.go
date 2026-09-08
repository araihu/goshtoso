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

// Help renders a compact question-mark button with explanatory text on click.
// Click again, click outside, or press Escape to dismiss. The label names the button and the tooltip heading.
func Help(id, label, description string) Instance {
	return Tooltip(id, label, WithDescription(description), WithActivation(ActivationClick), WithPosition(PositionRight), WithPortal(true), WithTrigger(helpTrigger(label)))
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
