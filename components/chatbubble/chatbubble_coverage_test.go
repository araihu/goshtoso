package chatbubble

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, c.Render(context.Background(), &buf))
	return buf.String()
}

func TestCoverageConfigHelpers(t *testing.T) {
	sent := Config{Side: Sent, RootClass: "extra-row"}
	assert.True(t, sent.IsMine())
	assert.Equal(t, "true", sent.DataMine())
	assert.Contains(t, sent.RowClasses(), "mt-4")
	assert.Contains(t, sent.RowClasses(), "extra-row")
	assert.Contains(t, sent.BubbleClasses(), "group-data-[mine=true]:bg-primary")

	received := Config{Side: Received, Grouped: true}
	assert.False(t, received.IsMine())
	assert.Equal(t, "false", received.DataMine())
	assert.Contains(t, received.RowClasses(), "mt-0.5")

	assert.True(t, Config{ShowAvatar: true}.HasAvatar())
	assert.False(t, Config{ShowAvatar: true, Grouped: true}.HasAvatar())
	assert.True(t, Config{SenderName: "Ada"}.HasHeader())
	assert.False(t, Config{SenderName: "Ada", Grouped: true}.HasHeader())
}

func TestCoverageChatBubbleRendersAutoSenderRootAttrsAndEscapedMessage(t *testing.T) {
	html := render(t, ChatBubble(Config{
		Side:           Auto,
		Sender:         "ada@example",
		SenderName:     "Ada",
		Message:        `<script>alert("x")</script>`,
		Timestamp:      "11:32",
		AvatarInitials: "AL",
		AvatarVariant:  "info",
		ShowAvatar:     true,
		IsBot:          true,
		RootClass:      "audit-row",
		RootAttrs: templ.Attributes{
			"data-testid": "bubble",
		},
	}))

	assert.Contains(t, html, `data-mine="false"`)
	assert.Contains(t, html, `data-sender="ada@example"`)
	assert.Contains(t, html, `data-testid="bubble"`)
	assert.Contains(t, html, "audit-row")
	assert.Contains(t, html, "Ada")
	assert.Contains(t, html, "11:32")
	assert.Contains(t, html, "BOT")
	assert.Contains(t, html, "AL")
	assert.Contains(t, html, `&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;`)
	assert.NotContains(t, html, `<script>alert`)
}

func TestCoverageChatBubbleSentStatusAndGroupedSuppressions(t *testing.T) {
	sent := render(t, ChatBubble(Config{
		Side:       Sent,
		SenderName: "Me",
		Message:    "On my way",
		Status:     StatusSeen,
	}))

	assert.Contains(t, sent, `data-mine="true"`)
	assert.Contains(t, sent, "On my way")
	assert.Contains(t, sent, "Seen")
	assert.Contains(t, sent, "group-data-[mine=true]:hidden")

	grouped := render(t, ChatBubble(Config{
		Side:           Received,
		SenderName:     "Ada",
		Message:        "Follow-up",
		ShowAvatar:     true,
		AvatarInitials: "AL",
		Grouped:        true,
	}))

	assert.Contains(t, grouped, "Follow-up")
	assert.NotContains(t, grouped, ">Ada<")
	assert.NotContains(t, grouped, ">AL<")
}

func TestCoverageTypingIndicator(t *testing.T) {
	html := render(t, TypingIndicator(Config{
		ShowAvatar:     true,
		AvatarInitials: "AL",
		AvatarVariant:  "info",
		RootClass:      "typing-row",
	}))

	assert.Contains(t, html, `aria-label="typing"`)
	assert.Contains(t, html, `aria-live="polite"`)
	assert.Contains(t, html, `data-mine="false"`)
	assert.Contains(t, html, "typing-row")
	assert.Equal(t, 3, strings.Count(html, "animate-bounce"))
	assert.Contains(t, html, "AL")
}
