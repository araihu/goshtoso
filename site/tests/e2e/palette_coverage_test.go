package e2e

import (
	"sync"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// gotoPalette navigates to the palette demo and waits for Alpine to boot.
func gotoPalette(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+"/components/palette", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// TestPalette_ResetClearsModel picks a swatch, then clicks Reset and asserts the
// bound model returns to the empty placeholder ("—"). Covers the Reset
// pick(empty, null) branch.
func TestPalette_ResetClearsModel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoPalette(t, page)

	require.NoError(t, page.Locator(`#demo-palette button[data-cls="blue-700"]`).Click())
	_, err := page.WaitForFunction(
		`() => document.querySelector('#palette-standalone p span[x-text]').textContent.trim() === 'blue-700'`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "pick should set the model to blue-700")

	// Reset is the button rendered next to the hovered label in the standalone root.
	require.NoError(t, page.Locator(`#demo-palette button:has-text("Reset")`).Click())
	_, err = page.WaitForFunction(
		`() => document.querySelector('#palette-standalone p span[x-text]').textContent.trim() === '—'`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "Reset should clear the model back to the placeholder")

	// The hex text input should also be cleared by syncHex('').
	val, err := page.Locator(`#demo-palette input[type="text"]`).InputValue()
	require.NoError(t, err)
	require.Equal(t, "", val)
}

// TestPalette_HoverShowsLabel hovers a swatch and asserts the hovered readout
// (x-text="hovered || 'Pick a color'") reflects the swatch data-cls.
func TestPalette_HoverShowsLabel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoPalette(t, page)

	label := page.Locator(`#demo-palette span[x-text="hovered || 'Pick a color'"]`)
	require.NoError(t, label.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(2000),
	}))

	require.NoError(t, page.Locator(`#demo-palette button[data-cls="green-500"]`).Hover())
	_, err := page.WaitForFunction(
		`() => document.querySelector('#demo-palette span[x-text="hovered || \'Pick a color\'"]').textContent.trim() === 'green-500'`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "hovering a swatch should update the hovered label")
}

// TestPalette_NoConsoleErrors loads the demo, exercises a pick and a hex commit,
// and asserts no console errors were emitted (Alpine binding / escaping smoke).
func TestPalette_NoConsoleErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	var mu sync.Mutex
	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			mu.Lock()
			consoleErrors = append(consoleErrors, msg.Text())
			mu.Unlock()
		}
	})

	gotoPalette(t, page)

	require.NoError(t, page.Locator(`#demo-palette button[data-cls="blue-700"]`).Click())
	_, err := page.WaitForFunction(
		`() => document.querySelector('#palette-standalone p span[x-text]').textContent.trim() === 'blue-700'`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	input := page.Locator(`#demo-palette input[type="text"]`)
	require.NoError(t, input.Fill("#abc"))
	require.NoError(t, input.Blur())
	_, err = page.WaitForFunction(
		`() => document.querySelector('#palette-standalone p span[x-text]').textContent.trim() === '#aabbcc'`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Empty(t, consoleErrors, "palette demo should emit no console errors")
}
