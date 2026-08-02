//go:build e2e && (full || range)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func navigateToRangeDemo(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+"/components/range", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#range-fragment").WaitFor())
}

func TestRange_DefaultAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	navigateToRangeDemo(t, page)

	input := page.Locator("#rangeDefault")
	require.NoError(t, input.WaitFor())

	inputType, err := input.GetAttribute("type")
	require.NoError(t, err)
	require.Equal(t, "range", inputType)

	value, err := input.GetAttribute("value")
	require.NoError(t, err)
	require.Equal(t, "20", value)

	min, err := input.GetAttribute("min")
	require.NoError(t, err)
	require.Equal(t, "0", min)

	max, err := input.GetAttribute("max")
	require.NoError(t, err)
	require.Equal(t, "100", max)
}

func TestRange_TicksAndIconsRender(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	navigateToRangeDemo(t, page)

	ticks := page.Locator("#range-ticks [aria-hidden='true'] span")
	count, err := ticks.Count()
	require.NoError(t, err)
	require.GreaterOrEqual(t, count, 21)

	icons := page.Locator("#range-icons svg")
	iconCount, err := icons.Count()
	require.NoError(t, err)
	require.Equal(t, 2, iconCount)
}

func TestRange_LiveValueUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	navigateToRangeDemo(t, page)
	_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	input := page.Locator("#rangeValue")
	badge := page.Locator("#range-value [x-text='currentVal']")

	require.NoError(t, input.Fill("55"))
	_, err = page.WaitForFunction("() => document.querySelector('#range-value [x-text=\"currentVal\"]').textContent === '55'", nil)
	require.NoError(t, err)

	text, err := badge.TextContent()
	require.NoError(t, err)
	require.Equal(t, "55", text)
}

func TestRange_Disabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	navigateToRangeDemo(t, page)

	disabled, err := page.Locator("#rangeDisabled").IsDisabled()
	require.NoError(t, err)
	require.True(t, disabled)
}
