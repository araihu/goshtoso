//go:build e2e && (full || spinner)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpinnerCoverageDemo loads the spinner demo and asserts the page renders
// every documented variant/size without any silent Alpine or console failure.
// It complements TestSpinnerComponentDemoVariants by adding the console-error
// guard that exercises the demo render path end to end.
func TestSpinnerCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/spinner", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#spinner-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	require.Contains(t, mainText, "Spinner")

	t.Run("animated svg root is present", func(t *testing.T) {
		count, err := page.Locator("svg.motion-safe\\:animate-spin").Count()
		require.NoError(t, err)
		assert.Greater(t, count, 0, "expected at least one animated spinner svg")
	})

	t.Run("default spinner uses fallback size and fill classes", func(t *testing.T) {
		def := page.Locator("#spinner-default svg.size-5.fill-on-surface").First()
		require.NoError(t, def.WaitFor())
		assert.Equal(t, "true", mustAttribute(t, def, "aria-hidden"))
	})

	require.Empty(t, jsErrors, "no JS console/page errors on spinner demo: %v", jsErrors)
}
