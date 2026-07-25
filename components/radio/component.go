package radio

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable radio component.
type Instance struct {
	cfg Config
}

// Radio returns a renderable radio component.
func Radio(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a radio.
func (Instance) Kind() components.Kind {
	return components.KindRadio
}

// Render writes the radio markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return radioTemplate(i.cfg).Render(ctx, w)
}

// BarInstance is a renderable radio bar component.
type BarInstance struct{}

// RadioBar returns a renderable radio bar component.
func RadioBar() BarInstance {
	return BarInstance{}
}

// Kind identifies the component as a radio bar.
func (BarInstance) Kind() components.Kind {
	return components.KindRadioBar
}

// Render writes the radio bar markup.
func (BarInstance) Render(ctx context.Context, w io.Writer) error {
	return radioBarTemplate().Render(ctx, w)
}

// GroupInstance is a renderable radio group component.
type GroupInstance struct {
	cfg GroupConfig
}

// RadioGroup returns a renderable radio group component.
func RadioGroup(cfg GroupConfig) GroupInstance {
	return GroupInstance{cfg: cfg}
}

// Kind identifies the component as a radio group.
func (GroupInstance) Kind() components.Kind {
	return components.KindRadioGroup
}

// Render writes the radio group markup.
func (i GroupInstance) Render(ctx context.Context, w io.Writer) error {
	return radioGroupTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = BarInstance{}
	_ components.Component = GroupInstance{}
)
