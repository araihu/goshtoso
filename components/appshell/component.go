package appshell

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable application-shell component.
type Instance struct {
	cfg Config
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
	return appShellTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
