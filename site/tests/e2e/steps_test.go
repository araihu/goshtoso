//go:build e2e && (full || steps)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSteps_HTMXFlowProgressesAndRegresses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/steps", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	shell := page.Locator("#steps-demo-shell")
	require.NoError(t, shell.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	currentState := func() string {
		value, err := page.Locator("#steps-htmx").GetAttribute("data-current-step")
		require.NoError(t, err)
		return value
	}

	currentStepLabel := func() string {
		text, err := page.Locator("#steps-htmx li[aria-current='step']").TextContent()
		require.NoError(t, err)
		return text
	}

	assert.Equal(t, "2", currentState())
	assert.Contains(t, currentStepLabel(), "Select a plan")

	next := shell.Locator("button").Filter(playwright.LocatorFilterOptions{
		HasText: "Next",
	})
	back := shell.Locator("button").Filter(playwright.LocatorFilterOptions{
		HasText: "Back",
	})

	// The Next/Back buttons live inside #steps-demo-shell, which each click
	// replaces via hx-swap="outerHTML". clickUntil re-fires a click dropped by
	// the htmx rebind race (see its doc) so the regression step can't be lost.
	clickUntil(t, page, next, `() => document.querySelector('#steps-htmx')?.getAttribute('data-current-step') === '3'`)
	assert.Contains(t, currentStepLabel(), "Checkout")

	clickUntil(t, page, back, `() => document.querySelector('#steps-htmx')?.getAttribute('data-current-step') === '2'`)
	assert.Contains(t, currentStepLabel(), "Select a plan")

	completedCount, err := page.Locator("#steps-htmx svg").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, completedCount, "regressing should restore single completed step")
}
