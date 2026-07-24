package tagslist

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable tags list component.
type Instance struct {
	cfg Config
}

// TagsList returns a renderable tags list component.
func TagsList(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a tags list.
func (Instance) Kind() components.Kind {
	return components.KindTagsList
}

// Render writes the tags list markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return tagsListTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
