package themeeditor

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/themes"
)

type Instance struct{ cfg Config }

func ThemeEditor(cfg Config) Instance {
	if cfg.InputName == "" {
		cfg.InputName = "definition"
	}
	if cfg.Fonts == nil {
		cfg.Fonts = themes.DefaultFonts()
	}
	return Instance{cfg}
}
func (Instance) Kind() components.Kind { return components.KindThemeEditor }
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	if err := themes.Validate(i.cfg.Initial); err != nil {
		return err
	}
	for _, item := range i.cfg.Catalog {
		if err := themes.Validate(item.Definition); err != nil {
			return err
		}
	}
	return editorTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
