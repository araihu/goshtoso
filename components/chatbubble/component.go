package chatbubble

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable chat bubble component.
type Instance struct {
	cfg Config
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
	return chatBubbleTemplate(i.cfg).Render(ctx, w)
}

// TypingIndicatorInstance is a renderable typing indicator component.
type TypingIndicatorInstance struct {
	cfg Config
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
	return typingIndicatorTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = TypingIndicatorInstance{}
)
