package textarea

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable textarea component.
type Instance struct {
	cfg Config
}

// Textarea returns a renderable textarea component.
func Textarea(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a textarea.
func (Instance) Kind() components.Kind {
	return components.KindTextarea
}

// Render writes the textarea markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return textareaTemplate(i.cfg).Render(ctx, w)
}

// WithActionsInstance is a renderable textarea with actions component.
type WithActionsInstance struct {
	cfg Config
}

// TextareaWithActions returns a renderable textarea with actions component.
func TextareaWithActions(cfg Config) WithActionsInstance {
	return WithActionsInstance{cfg: cfg}
}

// Kind identifies the component as a textarea with actions.
func (WithActionsInstance) Kind() components.Kind {
	return components.KindTextareaWithActions
}

// Render writes the textarea with actions markup.
func (i WithActionsInstance) Render(ctx context.Context, w io.Writer) error {
	return textareaWithActionsTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = WithActionsInstance{}
)
