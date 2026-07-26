package head

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable full dependency set.
type Instance struct {
	config config
}

// Dependencies returns the full Goshtoso runtime dependency set.
func Dependencies(options ...Option) Instance {
	return Instance{config: newConfig(options)}
}

// Kind identifies the component as the full dependency set.
func (Instance) Kind() components.Kind {
	return components.KindDependencies
}

// Render writes the full dependency markup.
func (instance Instance) Render(ctx context.Context, w io.Writer) error {
	cfg := instance.config
	if !cfg.initialized {
		cfg = newConfig(nil)
	}
	cfg.nonce = templ.GetNonce(ctx)
	return dependenciesTemplate(cfg).Render(ctx, w)
}

// MinimalInstance is a renderable minimal dependency set.
type MinimalInstance struct {
	config config
}

// DependenciesMinimal returns the minimal Goshtoso runtime dependency set.
func DependenciesMinimal(options ...Option) MinimalInstance {
	return MinimalInstance{config: newConfig(options)}
}

// Kind identifies the component as the minimal dependency set.
func (MinimalInstance) Kind() components.Kind {
	return components.KindDependenciesMinimal
}

// Render writes the minimal dependency markup.
func (instance MinimalInstance) Render(ctx context.Context, w io.Writer) error {
	cfg := instance.config
	if !cfg.initialized {
		cfg = newConfig(nil)
	}
	cfg.nonce = templ.GetNonce(ctx)
	return dependenciesMinimalTemplate(cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = MinimalInstance{}
)
