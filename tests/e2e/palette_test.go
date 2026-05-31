package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestPalette_PickSetsModel clicks a standalone palette swatch and asserts the
// bound Alpine model (rendered in the "Selected:" text) updates to the chosen
// value.
func TestPalette_PickSetsModel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/palette", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	require.NoError(t, page.Locator(`#demo-palette button[data-cls="blue-700"]`).Click())

	// The standalone preview shows `Selected: <span x-text="picked || '—'">`.
	// After the pick the model is blue-700, so that span should read blue-700.
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('span[x-text]'))
			.some(s => s.textContent.trim() === 'blue-700')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "Selected value text should reflect blue-700")
}

// TestPalette_InShell_OpensPicksCloses verifies the Select-shell-hosted palette:
// swatches hidden until the trigger opens the dropdown, picking sets the shell
// value and closes the dropdown via the bubbling select-close event.
func TestPalette_InShell_OpensPicksCloses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/palette", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	swatch := page.Locator(`#demo-shell-palette button[data-cls="red-500"]`)

	// Dropdown is collapsed initially, so the swatch is not visible.
	require.NoError(t, swatch.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(2000),
	}))

	// Open the shell.
	require.NoError(t, page.Locator("#demo-shell-trigger").Click())
	require.NoError(t, swatch.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(2000),
	}))

	// Pick red-500.
	require.NoError(t, swatch.Click())

	// (a) Trigger reflects the raw value (ValueExpr = "shellPicked || 'Pick a color'",
	// no classLabel applied) — so it shows "red-500".
	_, err = page.WaitForFunction(
		`() => document.querySelector('#demo-shell-trigger').textContent.includes('red-500')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "trigger should show picked value red-500")

	// (b) The dropdown closed on select-close, so the swatch is hidden again.
	require.NoError(t, swatch.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(2000),
	}))
}
