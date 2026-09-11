package appshell

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable application-shell component.
type Instance struct {
	expressionOverrides expressions.AppShell
	cfg                 Config
}

// AppShell returns a renderable application shell.
func AppShell(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as an application shell.
func (Instance) Kind() components.Kind {
	return components.KindAppShell
}

// Render writes the application-shell markup.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{AppShell: instance.expressionOverrides})
	return appShellTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.AppShell) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{AppShell: i.expressionOverrides}, expressions.Set{AppShell: values}).AppShell
	return i
}
