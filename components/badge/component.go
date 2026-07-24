package badge

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable badge component.
type Instance struct {
	cfg Config
}

// Badge returns a renderable badge component.
func Badge(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a badge.
func (Instance) Kind() components.Kind {
	return components.KindBadge
}

// Render writes the badge markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return badgeTemplate(i.cfg).Render(ctx, w)
}

// NotificationBadgeInstance is a renderable notification count badge.
type NotificationBadgeInstance struct {
	count int
}

// NotificationBadge returns a renderable notification count badge.
func NotificationBadge(count int) NotificationBadgeInstance {
	return NotificationBadgeInstance{count: count}
}

// Kind identifies the component as a notification badge.
func (NotificationBadgeInstance) Kind() components.Kind {
	return components.KindNotificationBadge
}

// Render writes the notification badge markup.
func (i NotificationBadgeInstance) Render(ctx context.Context, w io.Writer) error {
	return notificationBadgeTemplate(i.count).Render(ctx, w)
}

// NotificationDotInstance is a renderable notification dot.
type NotificationDotInstance struct{}

// NotificationDot returns a renderable notification dot.
func NotificationDot() NotificationDotInstance {
	return NotificationDotInstance{}
}

// Kind identifies the component as a notification dot.
func (NotificationDotInstance) Kind() components.Kind {
	return components.KindNotificationDot
}

// Render writes the notification dot markup.
func (NotificationDotInstance) Render(ctx context.Context, w io.Writer) error {
	return notificationDotTemplate().Render(ctx, w)
}

// AnimatingDotInstance is a renderable animated notification dot.
type AnimatingDotInstance struct {
	tone Tone
}

// AnimatingDot returns a renderable animated notification dot.
func AnimatingDot(tone Tone) AnimatingDotInstance {
	return AnimatingDotInstance{tone: tone}
}

// Kind identifies the component as an animated notification dot.
func (AnimatingDotInstance) Kind() components.Kind {
	return components.KindAnimatingDot
}

// Render writes the animated notification dot markup.
func (i AnimatingDotInstance) Render(ctx context.Context, w io.Writer) error {
	return animatingDotTemplate(i.tone).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = NotificationBadgeInstance{}
	_ components.Component = NotificationDotInstance{}
	_ components.Component = AnimatingDotInstance{}
)
