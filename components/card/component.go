package card

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable card component.
type Instance struct {
	cfg Config
}

// Card returns a renderable card component.
func Card(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a card.
func (Instance) Kind() components.Kind {
	return components.KindCard
}

// Render writes the card markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return cardTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
