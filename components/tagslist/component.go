package tagslist

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable tags list component.
type Instance struct {
	expressionOverrides expressions.TagsList
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{TagsList: i.expressionOverrides})
	return tagsListTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.TagsList) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{TagsList: i.expressionOverrides}, expressions.Set{TagsList: values}).TagsList
	return i
}
