//go:build e2e && (full || range)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRangeCoverageDemo complements range_test.go with a clean-load check (no JS
// exceptions / console errors), keyboard-driven live value updates, icon
// decoration semantics, and a dark-mode regression. filterIgnorable is defined
// in alert_coverage_test.go; waitForAlpine in modal_test.go.
func TestRangeCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/range", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#range-fragment").WaitFor())
	require.NoError(t, waitForAlpine(page))

	t.Run("heading and all variant frames render", func(t *testing.T) {
		require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
			HasText: "Range",
		}).First().WaitFor())

		for _, id := range []string{
			"#range-default", "#range-ticks", "#range-icons",
			"#range-value", "#range-disabled",
		} {
			require.NoError(t, page.Locator(id).WaitFor(), "missing variant frame %s", id)
		}
	})

	t.Run("keyboard arrow drives the live value badge", func(t *testing.T) {
		input := page.Locator("#rangeValue")
		require.NoError(t, input.WaitFor())

		// Start from a known value, then step up once with the keyboard. The
		// Alpine x-model badge should track the native input live.
		require.NoError(t, input.Fill("50"))
		_, err := page.WaitForFunction(
			"() => document.querySelector('#range-value [x-text=\"currentVal\"]').textContent === '50'", nil)
		require.NoError(t, err)

		require.NoError(t, input.Focus())
		require.NoError(t, page.Keyboard().Press("ArrowRight"))

		_, err = page.WaitForFunction(
			"() => document.querySelector('#range-value [x-text=\"currentVal\"]').textContent === '51'", nil)
		require.NoError(t, err)

		badge, err := page.Locator("#range-value [x-text='currentVal']").TextContent()
		require.NoError(t, err)
		assert.Equal(t, "51", badge)
	})

	t.Run("icons are aria-hidden decorations around the native input", func(t *testing.T) {
		decorations := page.Locator("#range-icons span[aria-hidden='true']")
		count, err := decorations.Count()
		require.NoError(t, err)
		assert.Equal(t, 2, count, "leading and trailing icon wrappers")

		// The native range input stays the single interactive control.
		inputs, err := page.Locator("#range-icons input[type='range']").Count()
		require.NoError(t, err)
		assert.Equal(t, 1, inputs)
	})

	t.Run("disabled range blocks interaction", func(t *testing.T) {
		disabled, err := page.Locator("#rangeDisabled").IsDisabled()
		require.NoError(t, err)
		assert.True(t, disabled)
	})

	t.Run("sliders survive a dark-mode toggle", func(t *testing.T) {
		before, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		beforeBool, _ := before.(bool)

		_, err = page.Evaluate(`() => { document.documentElement.classList.toggle('dark'); }`, nil)
		require.NoError(t, err)

		after, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		afterBool, _ := after.(bool)
		assert.NotEqual(t, beforeBool, afterBool, "dark class should toggle")

		require.NoError(t, page.Locator("#rangeDefault").WaitFor())
	})

	require.Empty(t, pageErrors, "uncaught JS exceptions on range demo")
	require.Empty(t, filterIgnorable(consoleErrors), "console errors on range demo")
}
