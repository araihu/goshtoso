package head

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable full dependency set.
type Instance struct{}

// Dependencies returns the full Goshtoso runtime dependency set.
func Dependencies() Instance {
	return Instance{}
}

// Kind identifies the component as the full dependency set.
func (Instance) Kind() components.Kind {
	return components.KindDependencies
}

// Render writes the full dependency markup.
func (Instance) Render(ctx context.Context, w io.Writer) error {
	return dependenciesTemplate().Render(ctx, w)
}

// MinimalInstance is a renderable minimal dependency set.
type MinimalInstance struct{}

// DependenciesMinimal returns the minimal Goshtoso runtime dependency set.
func DependenciesMinimal() MinimalInstance {
	return MinimalInstance{}
}

// Kind identifies the component as the minimal dependency set.
func (MinimalInstance) Kind() components.Kind {
	return components.KindDependenciesMinimal
}

// Render writes the minimal dependency markup.
func (MinimalInstance) Render(ctx context.Context, w io.Writer) error {
	return dependenciesMinimalTemplate().Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = MinimalInstance{}
)
