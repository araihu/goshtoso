package chatbubble

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func expressionStatusLabel(ctx context.Context, status Status) string {
	text := expressions.From(ctx).ChatBubble
	switch status {
	case StatusSending:
		return text.SendingLabel
	case StatusDelivered:
		return text.DeliveredLabel
	case StatusSeen:
		return text.SeenLabel
	default:
		return string(status)
	}
}
