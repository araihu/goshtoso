package codeblock

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable code block component.
type Instance struct {
	expressionOverrides *expressions.CodeBlock
	cfg                 Config
}

// CodeBlock returns a renderable code block component.
func CodeBlock(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a code block.
func (Instance) Kind() components.Kind {
	return components.KindCodeBlock
}

// Render writes the code block markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	if i.expressionOverrides != nil {
		ctx = expressions.With(ctx, expressions.Set{CodeBlock: *i.expressionOverrides})
	}
	return codeBlockTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.CodeBlock) Instance {
	var current expressions.CodeBlock
	if i.expressionOverrides != nil {
		current = *i.expressionOverrides
	}
	merged := expressions.Merge(expressions.Set{CodeBlock: current}, expressions.Set{CodeBlock: values}).CodeBlock
	i.expressionOverrides = &merged
	return i
}
