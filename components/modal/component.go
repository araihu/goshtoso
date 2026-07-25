package modal

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable modal component.
type Instance struct {
	cfg Config
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
	return modalTemplate(i.cfg).Render(ctx, w)
}

// AlertDialogInstance is a renderable alert dialog component.
type AlertDialogInstance struct {
	cfg AlertDialogConfig
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
	return alertDialogTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = AlertDialogInstance{}
)
