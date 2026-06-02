package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestChatBubble_DemoPage loads /components/chatbubble and asserts the docs
// fragment renders, both a Sent (data-mine='true') and a received
// (data-mine='false') bubble exist, and the sidebar link is present.
func TestChatBubble_DemoPage(t *testing.T) {
	page := newIsolatedPage(t)

	_, err := page.Goto(baseURL + "/components/chatbubble")
	require.NoError(t, err)

	// The docs fragment anchor proves a 200-equivalent demo page (not a 404 shell).
	require.NoError(t, page.Locator("#chatbubble-fragment").First().WaitFor(
		playwright.LocatorWaitForOptions{Timeout: playwright.Float(5000)}))

	// At least one Sent demo bubble (static data-mine='true').
	mineCount, err := page.Locator("#chatbubble-fragment [data-mine='true']").Count()
	require.NoError(t, err)
	require.Positive(t, mineCount, "demo page should have at least one Sent (data-mine='true') bubble")

	// At least one received bubble (data-mine='false').
	recvCount, err := page.Locator("#chatbubble-fragment [data-mine='false']").Count()
	require.NoError(t, err)
	require.Positive(t, recvCount, "demo page should have at least one received (data-mine='false') bubble")

	// Sidebar link is present.
	require.NoError(t, page.Locator("a[href='/components/chatbubble']").First().WaitFor(
		playwright.LocatorWaitForOptions{Timeout: playwright.Float(5000)}))
}
