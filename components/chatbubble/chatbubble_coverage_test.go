package chatbubble

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/avatar"
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
	assert.True(t, sent.isMine())
	assert.Equal(t, "true", sent.dataMine())
	assert.Contains(t, sent.rowClasses(), "mt-4")
	assert.Contains(t, sent.rowClasses(), "extra-row")
	assert.Contains(t, sent.bubbleClasses(), "group-data-[mine=true]:bg-primary")

	received := Config{Side: Received, Grouped: true}
	assert.False(t, received.isMine())
	assert.Equal(t, "false", received.dataMine())
	assert.Contains(t, received.rowClasses(), "mt-0.5")

	assert.True(t, Config{ShowAvatar: true}.hasAvatar())
	assert.False(t, Config{ShowAvatar: true, Grouped: true}.hasAvatar())
	assert.True(t, Config{SenderName: "Ada"}.hasHeader())
	assert.False(t, Config{SenderName: "Ada", Grouped: true}.hasHeader())
}

func TestCoverageChatBubbleRendersAutoSenderRootAttrsAndEscapedMessage(t *testing.T) {
	html := render(t, ChatBubble(Config{
		Side:           Auto,
		Sender:         "ada@example",
		SenderName:     "Ada",
		Message:        `<script>alert("x")</script>`,
		Timestamp:      "11:32",
		AvatarInitials: "AL",
		AvatarTone:     avatar.ToneInfo,
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
		AvatarTone:     avatar.ToneInfo,
		RootClass:      "typing-row",
	}))

	assert.Contains(t, html, `aria-label="typing"`)
	assert.Contains(t, html, `aria-live="polite"`)
	assert.Contains(t, html, `data-mine="false"`)
	assert.Contains(t, html, "typing-row")
	assert.Equal(t, 3, strings.Count(html, "animate-bounce"))
	assert.Contains(t, html, "AL")
}
