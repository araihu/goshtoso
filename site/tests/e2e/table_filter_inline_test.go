package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTableFilter_InlineVariant pins the `FilterVariantInline` contract:
// same reactive filter behavior, no bordered block, no collapsible toggle,
// no `x-show/x-collapse` wrapper. Designed for modal bodies and toolbar
// strips where the host container already owns chrome.
func TestTableFilter_InlineVariant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(5000)

	_, err := page.Goto(baseURL+"/components/table", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "Alpine.js should be available")

	_, err = page.WaitForFunction(`() => {
		var el = document.querySelector('[x-data="inlineFilteredTableFilters"]');
		if (!el) return false;
		try { return !!Alpine.$data(el); } catch(e) { return false; }
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "inline filter Alpine component should initialize")

	t.Run("NoBorderedWrapper", func(t *testing.T) {
		// The bar-variant wrapper carries the chrome classes; inline must not.
		bar := page.Locator("#inline-filtered-table-filters")
		className, err := bar.GetAttribute("class")
		require.NoError(t, err)
		assert.NotContains(t, className, "border", "inline variant must not render a bordered block")
		assert.NotContains(t, className, "rounded-radius", "inline variant must not render rounded-radius chrome")
		assert.Contains(t, className, "flex", "inline variant must render filters as a flex row")
	})

	t.Run("NoCollapsibleToggle", func(t *testing.T) {
		toggle := page.Locator("#inline-filtered-table-filters button:has-text('Filters')")
		count, err := toggle.Count()
		require.NoError(t, err)
		assert.Equal(t, 0, count, "inline variant must not render the collapsible header button")
	})

	t.Run("SearchStillSwapsTbody", func(t *testing.T) {
		// Inline variant should still fire HTMX on input, same as bar.
		input := page.Locator("#inline-filtered-table-filters input[type='search']")
		require.NoError(t, input.Fill("alice"))
		_, err := input.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
		require.NoError(t, err)

		_, err = page.WaitForFunction(
			`() => {
				var rows = document.querySelectorAll('#inline-filtered-table tbody tr');
				return rows.length === 1;
			}`, nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err, "inline filter input should still swap the tbody on input")

		text, err := page.Locator("#inline-filtered-table tbody").TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Alice Brown")

		// Clean up so other subtests start from baseline.
		require.NoError(t, input.Fill(""))
		_, err = input.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
		require.NoError(t, err)
		_, err = page.WaitForFunction(
			`() => document.querySelectorAll('#inline-filtered-table tbody tr').length === 3`,
			nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
	})
}
