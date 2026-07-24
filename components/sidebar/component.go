package sidebar

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable sidebar component.
type Instance struct {
	cfg Config
}

// Sidebar returns a renderable sidebar component.
func Sidebar(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a sidebar.
func (Instance) Kind() components.Kind {
	return components.KindSidebar
}

// Render writes the sidebar markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return sidebarTemplate(i.cfg).Render(ctx, w)
}

// OverlayInstance is a renderable sidebar overlay.
type OverlayInstance struct {
	cfg OverlayConfig
}

// Overlay returns a renderable sidebar overlay.
func Overlay(cfg OverlayConfig) OverlayInstance {
	return OverlayInstance{cfg: cfg}
}

// Kind identifies the component as a sidebar overlay.
func (OverlayInstance) Kind() components.Kind {
	return components.KindSidebarOverlay
}

// Render writes the sidebar overlay markup.
func (i OverlayInstance) Render(ctx context.Context, w io.Writer) error {
	return overlayTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = OverlayInstance{}
)
