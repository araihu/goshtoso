package avatar

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable avatar component.
type Instance struct {
	cfg Config
}

// Avatar returns a renderable avatar component.
func Avatar(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an avatar.
func (Instance) Kind() components.Kind {
	return components.KindAvatar
}

// Render writes the avatar markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return avatarTemplate(i.cfg).Render(ctx, w)
}

// StackInstance is a renderable avatar stack component.
type StackInstance struct {
	cfg StackConfig
}

// AvatarStack returns a renderable avatar stack component.
func AvatarStack(cfg StackConfig) StackInstance {
	return StackInstance{cfg: cfg}
}

// Kind identifies the component as an avatar stack.
func (StackInstance) Kind() components.Kind {
	return components.KindAvatarStack
}

// Render writes the avatar stack markup.
func (i StackInstance) Render(ctx context.Context, w io.Writer) error {
	return avatarStackTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = StackInstance{}
)
