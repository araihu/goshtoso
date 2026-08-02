//go:build e2e && full

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFilterIgnorableOnlyDropsIntentionalAvatarFixtureNoise(t *testing.T) {
	fixtureURL := baseURL + "/assets/images/does-not-exist-404.png"
	messages := []string{
		"HTTP response: 404 Not Found: " + fixtureURL,
		"console error: Failed to load resource: the server responded with a status of 404 (Not Found) [url=" + fixtureURL + " line=0 column=0]",
		"HTTP response: 404 Not Found: " + baseURL + "/components/missing",
		"page error: component 404",
		"console error: favicon component failed [url=" + baseURL + "/components/avatar line=10 column=2]",
		"HTTP response: 500 Internal Server Error: " + fixtureURL,
		"request failed: GET " + fixtureURL + ": net::ERR_FAILED",
	}

	require.Equal(t, messages[2:], filterIgnorable(messages))
}

func TestWaitForPageSettledCapturesDelayedConsoleError(t *testing.T) {
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)

	require.NoError(t, page.SetContent(
		`<script>
document.addEventListener("DOMContentLoaded", () => {
	setTimeout(() => console.error("delayed component failure"), 150);
});
</script>`,
		playwright.PageSetContentOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		},
	))

	failures.mu.Lock()
	immediate := append([]string(nil), failures.messages...)
	failures.mu.Unlock()
	require.Empty(t, immediate, "the old immediate snapshot must miss the delayed error")

	waitForPageSettled(t, page)

	failures.mu.Lock()
	messages := append([]string(nil), failures.messages...)
	failures.mu.Unlock()
	joined := strings.Join(messages, "\n")
	require.Contains(t, joined, "console error: delayed component failure")
	require.Contains(t, joined, "[url=", "collector messages must include console source locations")
}
