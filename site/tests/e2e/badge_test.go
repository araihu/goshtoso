package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadgeComponentDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/badge", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("solid and soft variants render semantic labels", func(t *testing.T) {
		for _, id := range []string{"#badge-solid", "#badge-soft"} {
			for _, label := range []string{"Default", "Primary", "Secondary", "Info", "Success", "Warning", "Danger"} {
				require.NoError(t, page.Locator(id).GetByText(label, playwright.LocatorGetByTextOptions{
					Exact: playwright.Bool(true),
				}).WaitFor())
			}
		}
	})

	t.Run("preview frame paints non-clipping border overlay", func(t *testing.T) {
		shell := page.Locator("#badge-solid").Locator("xpath=ancestor::div[contains(@class, 'border-outline') and contains(@class, 'rounded-radius')][1]")
		className, err := shell.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, className, "after:border-outline")

		overflow, err := shell.Evaluate("el => getComputedStyle(el).overflow", nil)
		require.NoError(t, err)
		assert.Equal(t, "visible", overflow)
	})

	t.Run("icon and indicator badges keep labels visible", func(t *testing.T) {
		for _, label := range []string{"Penguin", "Filter", "Verified", "Active", "Warning", "Error"} {
			badge := page.Locator("#badge-icons span").Filter(playwright.LocatorFilterOptions{HasText: label}).First()
			require.NoError(t, badge.WaitFor())

			iconCount, err := badge.Locator("svg").Count()
			require.NoError(t, err)
			assert.Greater(t, iconCount, 0, "expected leading icon for %s", label)
		}

		indicatorCount, err := page.Locator("#badge-indicators span[aria-hidden='true']").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, indicatorCount)
	})

	t.Run("notification badges and dots expose counts and buttons", func(t *testing.T) {
		require.NoError(t, page.Locator("#badge-notification button[aria-label='notifications']").WaitFor())
		require.NoError(t, page.Locator("#badge-notification button[aria-label='messages']").WaitFor())
		require.NoError(t, page.Locator("#badge-notification button[aria-label='alerts']").WaitFor())

		require.NoError(t, page.Locator("#badge-notification").GetByText("99", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
		require.NoError(t, page.Locator("#badge-notification").GetByText("5", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())

		dotCount, err := page.Locator("#badge-notification span.size-3").Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, dotCount, 1)
	})

	t.Run("animating dots and sizes render stable class hooks", func(t *testing.T) {
		animating, err := page.Locator("#badge-animating span[aria-label='notification']").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, animating)

		for _, label := range []string{"Small", "Medium", "Large"} {
			require.NoError(t, page.Locator("#badge-sizes").GetByText(label, playwright.LocatorGetByTextOptions{
				Exact: playwright.Bool(true),
			}).WaitFor())
		}
	})
}
