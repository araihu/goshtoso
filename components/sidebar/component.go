package sidebar

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable sidebar component.
type Instance struct {
	expressionOverrides expressions.Sidebar
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Sidebar: i.expressionOverrides})
	return sidebarTemplate(i.cfg).Render(ctx, w)
}

// OverlayInstance is a renderable sidebar overlay.
type OverlayInstance struct {
	expressionOverrides expressions.Sidebar
	cfg                 OverlayConfig
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
	ctx = expressions.With(ctx, expressions.Set{Sidebar: i.expressionOverrides})
	return overlayTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = OverlayInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Sidebar) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Sidebar: i.expressionOverrides}, expressions.Set{Sidebar: values}).Sidebar
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i OverlayInstance) WithExpressions(values expressions.Sidebar) OverlayInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Sidebar: i.expressionOverrides}, expressions.Set{Sidebar: values}).Sidebar
	return i
}
