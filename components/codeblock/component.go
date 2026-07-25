package codeblock

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable code block component.
type Instance struct {
	cfg Config
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
	return codeBlockTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
