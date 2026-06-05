package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestTooltipCoverageDemo complements the legacy tooltip E2E tests with a
// deterministic pass over the documented demo variants, Alpine click state, and
// console/page-error checks.
func TestTooltipCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/tooltip", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForTooltipAlpine(page))

	require.NoError(t, page.Locator("#tooltip-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	require.Contains(t, mainText, "Tooltip")

	t.Run("hover positions render with ARIA wiring", func(t *testing.T) {
		positions := map[string]string{
			"demoTop":    "bottom-full",
			"demoBottom": "top-full",
			"demoLeft":   "right-full",
			"demoRight":  "left-full",
		}
		for id, wantClass := range positions {
			trigger := page.Locator("[aria-describedby='" + id + "']").First()
			require.NoError(t, trigger.ScrollIntoViewIfNeeded())
			require.Equal(t, id, tooltipAttribute(t, trigger, "aria-describedby"))

			tip := page.Locator("#" + id)
			require.Equal(t, "tooltip", tooltipAttribute(t, tip, "role"))
			require.Contains(t, tooltipAttribute(t, tip, "class"), wantClass)
		}
	})

	t.Run("rich tooltip exposes heading and description text", func(t *testing.T) {
		rich := page.Locator("#richTop")
		require.NoError(t, rich.ScrollIntoViewIfNeeded())
		require.Contains(t, tooltipAttribute(t, rich, "class"), "flex w-64 flex-col")
		require.NoError(t, rich.Locator("span.text-sm.font-medium").WaitFor())
		require.NoError(t, rich.Locator("p.text-balance").WaitFor())

		text, err := rich.TextContent()
		require.NoError(t, err)
		require.Contains(t, text, "Tooltip top")
		require.Contains(t, text, "A rich tooltip that contains longer text")
	})

	t.Run("click trigger toggles live Alpine visibility", func(t *testing.T) {
		trigger := page.Locator("[aria-describedby='clickTop']").First()
		require.NoError(t, trigger.ScrollIntoViewIfNeeded())

		visible, err := page.Evaluate("() => getComputedStyle(document.querySelector('#clickTop')).display !== 'none'", nil)
		require.NoError(t, err)
		require.False(t, visible.(bool), "click tooltip should start hidden")

		require.NoError(t, trigger.Click())
		_, err = page.WaitForFunction(
			"() => getComputedStyle(document.querySelector('#clickTop')).display !== 'none'",
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)

		require.NoError(t, page.Locator("main").Click(playwright.LocatorClickOptions{
			Position: &playwright.Position{X: 10, Y: 10},
		}))
		_, err = page.WaitForFunction(
			"() => getComputedStyle(document.querySelector('#clickTop')).display === 'none'",
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
	})

	require.Empty(t, jsErrors, "no JS console/page errors on tooltip demo: %v", jsErrors)
}

func waitForTooltipAlpine(page playwright.Page) error {
	_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	return err
}

func tooltipAttribute(t *testing.T, loc playwright.Locator, name string) string {
	t.Helper()

	value, err := loc.GetAttribute(name)
	require.NoError(t, err)
	return value
}
