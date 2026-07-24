package rangeinput

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable range component.
type Instance struct {
	cfg Config
}

// Range returns a renderable range component.
func Range(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a range.
func (Instance) Kind() components.Kind {
	return components.KindRange
}

// Render writes the range markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return rangeTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
