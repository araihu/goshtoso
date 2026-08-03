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

// Metadata renders route-specific title, description, canonical, Open Graph,
// and X/Twitter Card tags from one complete config. Render returns an error and
// writes no tags when required text, image metadata, or absolute HTTPS URLs are
// missing or invalid.
func Metadata(metadata MetadataConfig) templ.Component {
	return metadataComponent{config: metadata}
}

type metadataComponent struct {
	config MetadataConfig
}

func (component metadataComponent) Render(ctx context.Context, w io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	metadata, err := component.config.validated()
	if err != nil {
		return err
	}
	return metadataTemplate(metadata).Render(ctx, w)
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
	if err := cfg.prepare(false); err != nil {
		return err
	}
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
	if err := cfg.prepare(true); err != nil {
		return err
	}
	return dependenciesMinimalTemplate(cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = MinimalInstance{}
)
