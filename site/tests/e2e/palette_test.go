//go:build e2e

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

func TestPalette_PickUpdatesHexControls(t *testing.T) {
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

	swatch := page.Locator(`#demo-palette button[data-cls="blue-700"]`)
	expected, err := swatch.Evaluate(`el => {
		const canvas = document.createElement('canvas')
		canvas.width = 1
		canvas.height = 1
		const ctx = canvas.getContext('2d', { willReadFrequently: true })
		ctx.fillStyle = '#000000'
		ctx.fillStyle = getComputedStyle(el).backgroundColor
		ctx.fillRect(0, 0, 1, 1)
		const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data
		return '#' + [r, g, b].map(n => n.toString(16).padStart(2, '0')).join('')
	}`, nil)
	require.NoError(t, err)
	require.NotEqual(t, "#000000", expected)

	require.NoError(t, swatch.Click())

	_, err = page.WaitForFunction(`expected => {
		const root = document.querySelector('#demo-palette')
		const color = root.querySelector('input[type="color"]')
		const text = root.querySelector('input[type="text"]')
		return color.value === expected && text.value === expected
	}`, expected, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "hex controls should reflect the picked swatch")
}

func TestPalette_HasSingleSelectedPreview(t *testing.T) {
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

	topNeutralCount, err := page.Locator(`#demo-palette > div:first-child button[data-cls="white"], #demo-palette > div:first-child button[data-cls="black"]`).Count()
	require.NoError(t, err)
	require.Equal(t, 0, topNeutralCount)

	previewCount, err := page.Locator(`#demo-palette [data-selected-preview]`).Count()
	require.NoError(t, err)
	require.Equal(t, 1, previewCount)
}

func TestPalette_HexInputCommitsValidTyping(t *testing.T) {
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

	input := page.Locator(`#demo-palette input[type="text"]`)
	require.NoError(t, input.Fill("#123abc"))
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('span[x-text]'))
			.some(s => s.textContent.trim() === '#123abc')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "valid hex typing should commit immediately")

	require.NoError(t, input.Fill("#456def"))
	require.NoError(t, input.Press("Enter"))
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('span[x-text]'))
			.some(s => s.textContent.trim() === '#456def')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "Enter should commit valid hex")

	require.NoError(t, input.Fill("#abc"))
	require.NoError(t, input.Blur())
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('span[x-text]'))
			.some(s => s.textContent.trim() === '#aabbcc')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "blur should normalize and commit short hex")
}

func TestPalette_HexInputRejectsInvalidValues(t *testing.T) {
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
	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('span[x-text]'))
			.some(s => s.textContent.trim() === 'blue-700')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	_, err = page.Locator(`#demo-palette input[type="text"]`).Evaluate(`el => {
		el.value = 'not-a-hex'
		el.dispatchEvent(new Event('input', { bubbles: true }))
		el.dispatchEvent(new Event('change', { bubbles: true }))
	}`, nil)
	require.NoError(t, err)

	valid, err := page.Locator(`#demo-palette input[type="text"]`).Evaluate(`el => el.checkValidity()`, nil)
	require.NoError(t, err)
	require.Equal(t, false, valid)

	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('span[x-text]'))
			.some(s => s.textContent.trim() === 'blue-700')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "invalid hex should not replace the selected model")
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
	expected, err := swatch.Evaluate(`el => {
		const canvas = document.createElement('canvas')
		canvas.width = 1
		canvas.height = 1
		const ctx = canvas.getContext('2d', { willReadFrequently: true })
		ctx.fillStyle = '#000000'
		ctx.fillStyle = getComputedStyle(el).backgroundColor
		ctx.fillRect(0, 0, 1, 1)
		const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data
		return '#' + [r, g, b].map(n => n.toString(16).padStart(2, '0')).join('')
	}`, nil)
	require.NoError(t, err)

	require.NoError(t, swatch.Click())

	// (a) Trigger reflects the raw value (ValueExpr = "shellPicked || 'Pick a color'",
	// no classLabel applied) — so it shows "red-500".
	_, err = page.WaitForFunction(
		`() => document.querySelector('#demo-shell-trigger').textContent.includes('red-500')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "trigger should show picked value red-500")

	chipCount, err := page.Locator(`#demo-shell-trigger [data-shell-selected-preview]`).Count()
	require.NoError(t, err)
	require.Equal(t, 1, chipCount)

	_, err = page.WaitForFunction(`expected => {
		const chip = document.querySelector('#demo-shell-trigger [data-shell-selected-preview]')
		if (!chip) return false
		const canvas = document.createElement('canvas')
		canvas.width = 1
		canvas.height = 1
		const ctx = canvas.getContext('2d', { willReadFrequently: true })
		ctx.fillStyle = '#000000'
		ctx.fillStyle = getComputedStyle(chip).backgroundColor
		ctx.fillRect(0, 0, 1, 1)
		const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data
		const got = '#' + [r, g, b].map(n => n.toString(16).padStart(2, '0')).join('')
		return got === expected
	}`, expected, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "trigger chip should show selected color")

	// (b) The dropdown closed on select-close, so the swatch is hidden again.
	require.NoError(t, swatch.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(2000),
	}))
}

func TestPalette_InShell_HexInputUpdatesTriggerPreview(t *testing.T) {
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

	require.NoError(t, page.Locator("#demo-shell-trigger").Click())
	input := page.Locator(`#demo-shell-palette input[type="text"]`)
	require.NoError(t, input.Fill("#123abc"))

	_, err = page.WaitForFunction(
		`() => document.querySelector('#demo-shell-trigger').textContent.includes('#123abc')`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "trigger should show typed hex")

	_, err = page.WaitForFunction(`() => {
		const chip = document.querySelector('#demo-shell-trigger [data-shell-selected-preview]')
		if (!chip) return false
		const canvas = document.createElement('canvas')
		canvas.width = 1
		canvas.height = 1
		const ctx = canvas.getContext('2d', { willReadFrequently: true })
		ctx.fillStyle = '#000000'
		ctx.fillStyle = getComputedStyle(chip).backgroundColor
		ctx.fillRect(0, 0, 1, 1)
		const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data
		const got = '#' + [r, g, b].map(n => n.toString(16).padStart(2, '0')).join('')
		return got === '#123abc'
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "trigger chip should show typed hex")
}
