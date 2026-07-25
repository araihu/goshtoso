package checkbox

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable checkbox component.
type Instance struct {
	cfg Config
}

// Checkbox returns a renderable checkbox component.
func Checkbox(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a checkbox.
func (Instance) Kind() components.Kind {
	return components.KindCheckbox
}

// Render writes the checkbox markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return checkboxTemplate(i.cfg).Render(ctx, w)
}

// GroupInstance is a renderable checkbox group component.
type GroupInstance struct {
	cfg GroupConfig
}

// CheckboxGroup returns a renderable checkbox group component.
func CheckboxGroup(cfg GroupConfig) GroupInstance {
	return GroupInstance{cfg: cfg}
}

// Kind identifies the component as a checkbox group.
func (GroupInstance) Kind() components.Kind {
	return components.KindCheckboxGroup
}

// Render writes the checkbox group markup.
func (i GroupInstance) Render(ctx context.Context, w io.Writer) error {
	return checkboxGroupTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = GroupInstance{}
)
