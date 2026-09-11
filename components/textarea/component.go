package textarea

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable textarea component.
type Instance struct {
	expressionOverrides expressions.Textarea
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Textarea: i.expressionOverrides})
	return textareaTemplate(i.cfg).Render(ctx, w)
}

// WithActionsInstance is a renderable textarea with actions component.
type WithActionsInstance struct {
	expressionOverrides expressions.Textarea
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Textarea: i.expressionOverrides})
	return textareaWithActionsTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = WithActionsInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Textarea) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Textarea: i.expressionOverrides}, expressions.Set{Textarea: values}).Textarea
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i WithActionsInstance) WithExpressions(values expressions.Textarea) WithActionsInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Textarea: i.expressionOverrides}, expressions.Set{Textarea: values}).Textarea
	return i
}
