package splitbutton

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable split button component.
type Instance struct {
	cfg Config
}

// SplitButton returns a primary action with an adjacent menu trigger.
func SplitButton(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a split button.
func (Instance) Kind() components.Kind {
	return components.KindSplitButton
}

// Render writes the split button markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return splitButtonTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
