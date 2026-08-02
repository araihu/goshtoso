// Package inlinecode renders short inline code fragments inside prose.
package inlinecode

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable inline code fragment.
type Instance struct {
	text    string
	options []Option
}

// InlineCode returns a semantic inline code fragment.
func InlineCode(text string, options ...Option) Instance {
	return Instance{text: text, options: options}
}

// Kind identifies the component as inline code.
func (Instance) Kind() components.Kind {
	return components.KindInlineCode
}

// Render writes the inline code markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return inlineCodeTemplate(i.text, i.options...).Render(ctx, w)
}

var _ components.Component = Instance{}
