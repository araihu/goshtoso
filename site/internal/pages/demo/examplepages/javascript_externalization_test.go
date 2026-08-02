package examplepages_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	chatdomain "github.com/araihu/goshtoso/site/internal/examples/chat"
	profiledomain "github.com/araihu/goshtoso/site/internal/examples/profile"
	chatpage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/chat"
	profilepage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/profile"
	tickerpage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/ticker"
	"github.com/stretchr/testify/require"
)

func TestComplexExampleProvidersUseExternalJavaScript(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		inlineJS string
		view     templ.Component
	}{
		{
			name:     "chat",
			provider: `id="chat-fragment"`,
			inlineJS: "__gtChatInit",
			view:     chatpage.ChatApp(chatdomain.Identity{Nick: "Guest", Color: "primary"}),
		},
		{
			name:     "profile",
			provider: `x-data="profileImages"`,
			inlineJS: "Alpine.data('profileImages'",
			view:     profilepage.ProfileApp(profiledomain.State{}),
		},
		{
			name:     "ticker",
			provider: `x-data="tickerPane()"`,
			inlineJS: "Alpine.data('tickerPane'",
			view:     tickerpage.TickerContent(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			require.NoError(t, test.view.Render(context.Background(), &output))
			html := output.String()
			require.Contains(t, html, test.provider)
			require.NotContains(t, html, test.inlineJS, "owned provider must not be emitted inline")
		})
	}
}
