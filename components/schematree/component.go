package schematree

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable schema tree.
type Instance struct {
	cfg                 Config
	expressionOverrides expressions.SchemaTree
}

// SchemaTree returns a read-only tree with native expandable branches.
func SchemaTree(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the schema tree component.
func (Instance) Kind() components.Kind { return components.KindSchemaTree }

// Render writes the tree using the caller's context and propagates slot errors.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{SchemaTree: instance.expressionOverrides})
	return treeTemplate(instance.cfg).Render(ctx, writer)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped expression overrides.
func (i Instance) WithExpressions(values expressions.SchemaTree) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{SchemaTree: i.expressionOverrides}, expressions.Set{SchemaTree: values}).SchemaTree
	return i
}
