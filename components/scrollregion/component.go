package scrollregion

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is the renderable Scroll Region value returned behind the original
// templ.Component constructor signature.
type Instance struct {
	cfg Config
}

// Kind identifies the component as a Scroll Region.
func (Instance) Kind() components.Kind {
	return components.KindScrollRegion
}

// Render writes the Scroll Region markup.
func (i Instance) Render(ctx context.Context, writer io.Writer) error {
	return scrollRegionTemplate(i.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
