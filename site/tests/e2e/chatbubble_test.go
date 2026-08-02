//go:build e2e && (full || chatbubble)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatBubbleComponentDemoVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)

	_, err := page.Goto(baseURL+"/components/chatbubble", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#chatbubble-fragment").First().WaitFor(
		playwright.LocatorWaitForOptions{Timeout: playwright.Float(5000)}))

	t.Run("default preview renders sent and received alignment hooks", func(t *testing.T) {
		defaultPreview := page.Locator("#chatbubble-default")
		require.NoError(t, defaultPreview.GetByText("Hi there! How can I assist you today?", playwright.LocatorGetByTextOptions{
			Exact: new(true),
		}).WaitFor())
		require.NoError(t, defaultPreview.GetByText("I accidentally deleted some important files. Can they be recovered?", playwright.LocatorGetByTextOptions{
			Exact: new(true),
		}).WaitFor())

		receivedCount, err := defaultPreview.Locator("[data-mine='false']").Count()
		require.NoError(t, err)
		assert.Equal(t, 2, receivedCount)

		sentCount, err := defaultPreview.Locator("[data-mine='true']").Count()
		require.NoError(t, err)
		assert.Equal(t, 1, sentCount)
	})

	t.Run("timestamp and status previews expose message metadata", func(t *testing.T) {
		timestamp := page.Locator("#chatbubble-timestamp")
		require.NoError(t, timestamp.GetByText("Ada", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, timestamp.GetByText("11:32 AM", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

		sentHeaderVisible, err := timestamp.GetByText("11:33 AM", playwright.LocatorGetByTextOptions{Exact: new(true)}).IsVisible()
		require.NoError(t, err)
		assert.False(t, sentHeaderVisible, "sent bubble headers are intentionally suppressed")

		status := page.Locator("#chatbubble-status")
		require.NoError(t, status.GetByText("Grace", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, status.GetByText("Delivered", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, status.GetByText("Seen", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	})

	t.Run("avatar and grouped previews keep avatars scoped to visible senders", func(t *testing.T) {
		avatar := page.Locator("#chatbubble-avatar")
		require.NoError(t, avatar.GetByText("AL", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, avatar.GetByText("ME", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

		grouped := page.Locator("#chatbubble-grouped")
		for _, message := range []string{
			"I split the work into three slices.",
			"Taking the first one myself.",
			"Can you grab the second?",
		} {
			require.NoError(t, grouped.GetByText(message, playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		}

		adaCount, err := grouped.GetByText("Ada", playwright.LocatorGetByTextOptions{Exact: new(true)}).Count()
		require.NoError(t, err)
		assert.Equal(t, 1, adaCount)

		initialsCount, err := grouped.GetByText("AL", playwright.LocatorGetByTextOptions{Exact: new(true)}).Count()
		require.NoError(t, err)
		assert.Equal(t, 1, initialsCount)
	})

	t.Run("typing indicator exposes live region and animated dots", func(t *testing.T) {
		typing := page.Locator("#chatbubble-typing [aria-label='typing']")
		require.NoError(t, typing.WaitFor())
		assert.Equal(t, "polite", mustAttribute(t, typing, "aria-live"))

		dotCount, err := typing.Locator("span.animate-bounce").Count()
		require.NoError(t, err)
		assert.Equal(t, 3, dotCount)
	})
}
