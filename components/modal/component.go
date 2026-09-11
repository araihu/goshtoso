package modal

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable modal component.
type Instance struct {
	expressionOverrides expressions.Modal
	cfg                 Config
}

// Modal returns a renderable modal component.
func Modal(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a modal.
func (Instance) Kind() components.Kind {
	return components.KindModal
}

// Render writes the modal markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Modal: i.expressionOverrides})
	return modalTemplate(i.cfg).Render(ctx, w)
}

// AlertDialogInstance is a renderable alert dialog component.
type AlertDialogInstance struct {
	expressionOverrides expressions.Modal
	cfg                 AlertDialogConfig
}

// AlertDialog returns a renderable alert dialog component.
func AlertDialog(cfg AlertDialogConfig) AlertDialogInstance {
	return AlertDialogInstance{cfg: cfg}
}

// Kind identifies the component as an alert dialog.
func (AlertDialogInstance) Kind() components.Kind {
	return components.KindAlertDialog
}

// Render writes the alert dialog markup.
func (i AlertDialogInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Modal: i.expressionOverrides})
	return alertDialogTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = AlertDialogInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Modal) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Modal: i.expressionOverrides}, expressions.Set{Modal: values}).Modal
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i AlertDialogInstance) WithExpressions(values expressions.Modal) AlertDialogInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Modal: i.expressionOverrides}, expressions.Set{Modal: values}).Modal
	return i
}
