package accordion

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable accordion component.
type Instance struct {
	cfg AccordionConfig
}

// Accordion returns a renderable accordion component.
func Accordion(cfg AccordionConfig) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an accordion.
func (Instance) Kind() components.Kind {
	return components.KindAccordion
}

// Render writes the accordion markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return accordionTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
