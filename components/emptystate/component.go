package emptystate

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable empty-state component.
type Instance struct {
	cfg Config
}

// EmptyState returns a renderable empty state.
func EmptyState(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an empty state.
func (Instance) Kind() components.Kind {
	return components.KindEmptyState
}

// Render writes the empty-state markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	return emptyStateTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
