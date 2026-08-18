//go:build e2e && (full || badge)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
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
					Exact: new(true),
				}).WaitFor())
			}
		}
	})

	t.Run("preview frame paints a non-clipping border", func(t *testing.T) {
		shell := page.Locator("#badge-solid").Locator("xpath=ancestor::div[contains(@class, 'component-page__preview')][1]")
		className, err := shell.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, className, "border-outline")

		borderVisible, err := shell.Evaluate(`el => {
			const frame = getComputedStyle(el)
			return frame.borderTopWidth === "1px" &&
				frame.borderTopStyle === "solid" &&
				frame.borderTopColor !== "rgba(0, 0, 0, 0)" &&
				frame.borderTopColor !== "transparent"
		}`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, borderVisible)

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

		require.NoError(t, page.Locator("#badge-notification").GetByText("99", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, page.Locator("#badge-notification").GetByText("5", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

		dotCount, err := page.Locator("#badge-notification span.size-3").Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, dotCount, 1)
	})

	t.Run("animating dots and size selector render stable class hooks", func(t *testing.T) {
		animating, err := page.Locator("#badge-animating span[aria-label='notification']").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, animating)

		cases := []struct {
			size  string
			label string
		}{
			{"sm", "Small"},
			{"md", "Medium"},
			{"lg", "Large"},
		}
		for _, tc := range cases {
			require.NoError(t, page.Locator("label[for='badge-size-"+tc.size+"']").Click())
			require.NoError(t, page.Locator("[data-testid='badge-size-selected']").Filter(playwright.LocatorFilterOptions{
				HasText: tc.size,
			}).WaitFor())
			require.NoError(t, page.Locator("[data-testid='badge-size-preview-"+tc.size+"']").WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}))
			visible, err := page.Locator("[data-testid='badge-size-preview-"+tc.size+"']").GetByText(tc.label, playwright.LocatorGetByTextOptions{
				Exact: new(true),
			}).IsVisible()
			require.NoError(t, err)
			assert.True(t, visible, "expected %s badge to be visible", tc.label)
		}
	})
}
