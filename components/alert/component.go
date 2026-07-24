package alert

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable alert component.
type Instance struct {
	cfg Config
}

// Alert returns a renderable alert component.
func Alert(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an alert.
func (Instance) Kind() components.Kind {
	return components.KindAlert
}

// Render writes the alert markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return alertTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
