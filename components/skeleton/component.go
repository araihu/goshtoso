package skeleton

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable skeleton component.
type Instance struct {
	cfg Config
}

// Skeleton returns a renderable skeleton loading state.
func Skeleton(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a skeleton.
func (Instance) Kind() components.Kind {
	return components.KindSkeleton
}

// Render writes the skeleton markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	return skeletonTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
