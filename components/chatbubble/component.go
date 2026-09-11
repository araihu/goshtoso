package chatbubble

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable chat bubble component.
type Instance struct {
	expressionOverrides expressions.ChatBubble
	cfg                 Config
}

// ChatBubble returns a renderable chat bubble component.
func ChatBubble(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a chat bubble.
func (Instance) Kind() components.Kind {
	return components.KindChatBubble
}

// Render writes the chat bubble markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{ChatBubble: i.expressionOverrides})
	return chatBubbleTemplate(i.cfg).Render(ctx, w)
}

// TypingIndicatorInstance is a renderable typing indicator component.
type TypingIndicatorInstance struct {
	expressionOverrides expressions.ChatBubble
	cfg                 Config
}

// TypingIndicator returns a renderable typing indicator component.
func TypingIndicator(cfg Config) TypingIndicatorInstance {
	return TypingIndicatorInstance{cfg: cfg}
}

// Kind identifies the component as a typing indicator.
func (TypingIndicatorInstance) Kind() components.Kind {
	return components.KindTypingIndicator
}

// Render writes the typing indicator markup.
func (i TypingIndicatorInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{ChatBubble: i.expressionOverrides})
	return typingIndicatorTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = TypingIndicatorInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.ChatBubble) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{ChatBubble: i.expressionOverrides}, expressions.Set{ChatBubble: values}).ChatBubble
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i TypingIndicatorInstance) WithExpressions(values expressions.ChatBubble) TypingIndicatorInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{ChatBubble: i.expressionOverrides}, expressions.Set{ChatBubble: values}).ChatBubble
	return i
}
