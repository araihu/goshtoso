//go:build e2e && (full || radio)

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRadioCoverage_FragmentNavNoConsoleErrors lands elsewhere first, then
// navigates to the radio demo via the sidebar (an HTMX fragment swap) and
// asserts no uncaught JS exceptions or console errors. This mirrors the
// example/console-error conventions used by the other component suites and
// exercises the demo's Alpine registration after a fragment swap.
func TestRadioCoverage_FragmentNavNoConsoleErrors(t *testing.T) {
	page := newPage(t, sharedBrowser)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/components/radio']").First().Click())

	// Wait for the radio fragment to land.
	require.NoError(t, page.Locator("[data-testid='radio-default-group']").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(3000),
		}))

	// Alpine-driven showcase must rebind after the fragment swap: clicking a
	// segmented size flips the visible `selected` text.
	require.NoError(t, page.Locator("label[for='r-a-lg']").Click())
	_, err = page.WaitForFunction(
		`() => document.querySelector("[data-testid='radio-alpine-out'] span").textContent === 'lg'`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "Alpine showcase must rebind after fragment nav")

	require.Empty(t, pageErrors, "no uncaught JS exceptions on fragment-nav radio page: %v", pageErrors)
	require.Empty(t, consoleErrors, "no unexpected console errors on fragment-nav radio page: %v", consoleErrors)
}

// TestRadioCoverage_TriggerMarkupClean is a browser-level regression guard for
// the `} else if cfg.HTMX.HasHxVerb()` attribute bug: every radio <input> on
// the demo must carry a real hx-trigger="change" (the HTMX showcase radios set
// a verb) and no input may leak a stray ` else` token into its attributes.
func TestRadioCoverage_TriggerMarkupClean(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	// The HTMX showcase radios should render a concrete hx-trigger.
	trigger, err := page.Locator("#r-h-md").GetAttribute("hx-trigger")
	require.NoError(t, err)
	assert.Equal(t, "change", trigger, "HTMX radio must default hx-trigger to change")

	// No radio input anywhere should serialize a leaked ` else` attribute token.
	outer, err := page.Locator("[data-testid='radio-htmx-group']").Evaluate("el => el.outerHTML", nil)
	require.NoError(t, err)
	html, ok := outer.(string)
	require.True(t, ok, "expected string outerHTML")
	assert.NotContains(t, html, " else", "hx-trigger else-if must not leak a literal else into the DOM")
	assert.False(t, strings.Contains(html, "else hx-trigger"),
		"both trigger branches must not render together")
}
