package themeplayground

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

type Instance struct{ cfg Config }

func ThemePlayground(cfg Config) Instance { return Instance{cfg} }
func (Instance) Kind() components.Kind    { return components.KindThemePlayground }
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return playgroundTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
