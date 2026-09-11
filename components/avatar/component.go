package avatar

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable avatar component.
type Instance struct {
	expressionOverrides expressions.Avatar
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Avatar: i.expressionOverrides})
	return avatarTemplate(i.cfg).Render(ctx, w)
}

// StackInstance is a renderable avatar stack component.
type StackInstance struct {
	expressionOverrides expressions.Avatar
	cfg                 StackConfig
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
	ctx = expressions.With(ctx, expressions.Set{Avatar: i.expressionOverrides})
	return avatarStackTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = StackInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Avatar) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Avatar: i.expressionOverrides}, expressions.Set{Avatar: values}).Avatar
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i StackInstance) WithExpressions(values expressions.Avatar) StackInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Avatar: i.expressionOverrides}, expressions.Set{Avatar: values}).Avatar
	return i
}
