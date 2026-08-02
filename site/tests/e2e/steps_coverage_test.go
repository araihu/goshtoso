//go:build e2e && (full || steps)

package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStepsCoverageDemo loads the steps demo directly and confirms the page
// renders with every variant container present (horizontal, compact, vertical,
// HTMX flow), with no uncaught JS exceptions or console errors.
func TestStepsCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/steps", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	assert.True(t, strings.Contains(mainText, "Steps"), "main should mention Steps")

	// Every demo variant ordered-list container must render.
	for _, id := range []string{
		"#steps-default",
		"#steps-compact",
		"#steps-vertical",
		"#steps-htmx",
	} {
		require.NoError(t, page.Locator(id).WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(3000),
		}), "variant %s did not render", id)
	}

	assert.Empty(t, pageErrors, "uncaught JS exceptions on steps demo")
	assert.Empty(t, consoleErrors, "console errors on steps demo")
}

// TestStepsCoverageOrientationClasses asserts the vertical variant carries the
// vertical list classes from listClasses() while the horizontal variant does
// not, and that the current step exposes aria-current="step".
func TestStepsCoverageOrientationClasses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/steps", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#steps-vertical").WaitFor())
	vertical := mustAttribute(t, page.Locator("#steps-vertical"), "class")
	assert.Contains(t, vertical, "flex-col")

	horizontal := mustAttribute(t, page.Locator("#steps-default"), "class")
	assert.Contains(t, horizontal, "items-center")
	assert.NotContains(t, horizontal, "flex-col")

	// The current step in the default variant must be marked for assistive tech.
	current := page.Locator("#steps-default li[aria-current='step']")
	require.NoError(t, current.WaitFor())
	count, err := current.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "exactly one current step expected")
}

// TestStepsCoverageDarkMode toggles the <html> .dark class and confirms the
// steps list stays rendered and visible (light/dark regression).
func TestStepsCoverageDarkMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/steps", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	list := page.Locator("#steps-default")
	require.NoError(t, list.WaitFor())
	visible, err := list.IsVisible()
	require.NoError(t, err)
	require.True(t, visible, "steps not visible in light mode")

	_, err = page.Evaluate("() => document.documentElement.classList.add('dark')", nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)

	visible, err = list.IsVisible()
	require.NoError(t, err)
	require.True(t, visible, "steps not visible in dark mode")
	assert.Contains(t, mustAttribute(t, list, "class"), "dark:text-on-surface-dark")
}
