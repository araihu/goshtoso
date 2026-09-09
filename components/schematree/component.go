package schematree

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable schema tree.
type Instance struct{ cfg Config }

// SchemaTree returns a read-only tree with native expandable branches.
func SchemaTree(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the schema tree component.
func (Instance) Kind() components.Kind { return components.KindSchemaTree }

// Render writes the tree using the caller's context and propagates slot errors.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	return treeTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}
