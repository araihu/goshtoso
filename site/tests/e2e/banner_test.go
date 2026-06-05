package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBannerComponentDemoVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/banner", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	t.Run("simple banner dismisses only itself", func(t *testing.T) {
		simple := page.Locator("#banner-simple")
		bannerEl := simple.Locator("[role='banner']")
		require.NoError(t, bannerEl.GetByText("Limited Time Offer! Explore exclusive deals & savings", playwright.LocatorGetByTextOptions{
			Exact: playwright.Bool(true),
		}).WaitFor())

		require.NoError(t, simple.Locator("button[aria-label='dismiss banner']").Click())
		require.NoError(t, bannerEl.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))

		require.NoError(t, page.Locator("#banner-persistent [role='banner']").WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
	})

	t.Run("persistent and CTA banners expose expected controls", func(t *testing.T) {
		persistent := page.Locator("#banner-persistent")
		require.NoError(t, persistent.GetByText("This banner cannot be dismissed by users", playwright.LocatorGetByTextOptions{
			Exact: playwright.Bool(true),
		}).WaitFor())

		dismissCount, err := persistent.Locator("button[aria-label='dismiss banner']").Count()
		require.NoError(t, err)
		assert.Zero(t, dismissCount)

		cta := page.Locator("#banner-cta")
		require.NoError(t, cta.GetByText("Get Fit Anywhere, Anytime 💪", playwright.LocatorGetByTextOptions{
			Exact: playwright.Bool(true),
		}).WaitFor())
		require.NoError(t, cta.Locator("button").Filter(playwright.LocatorFilterOptions{HasText: "Start free trial"}).WaitFor())
	})

	t.Run("semantic variants render user-facing copy", func(t *testing.T) {
		variants := page.Locator("#banner-variants")
		for _, text := range []string{
			"Default variant banner",
			"Primary variant for promotions",
			"Info variant for general information",
			"Success! Operation completed successfully",
			"Warning: Please review your settings",
			"Error: Something went wrong",
		} {
			require.NoError(t, variants.GetByText(text, playwright.LocatorGetByTextOptions{
				Exact: playwright.Bool(true),
			}).WaitFor())
		}
	})

	t.Run("cookie banner actions remain scoped to the preview", func(t *testing.T) {
		stage := page.Locator("#banner-cookie > div").First()
		dialog := page.Locator("#banner-cookie [role='dialog']").First()
		require.NoError(t, stage.WaitFor())
		require.NoError(t, dialog.WaitFor())
		require.NoError(t, dialog.Locator("h3").Filter(playwright.LocatorFilterOptions{HasText: "Cookie Time!"}).WaitFor())
		require.NoError(t, dialog.GetByText("We use cookies to make your experience sweet and crispy. For more information, please read our Privacy Policy.", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())

		classAttr, err := dialog.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "absolute bottom-4")
		require.False(t, strings.Contains(classAttr, "fixed bottom-4"), "demo cookie banner should not be viewport fixed")

		stageBox, err := stage.BoundingBox()
		require.NoError(t, err)
		dialogBox, err := dialog.BoundingBox()
		require.NoError(t, err)

		require.GreaterOrEqual(t, dialogBox.X, stageBox.X)
		require.GreaterOrEqual(t, dialogBox.Y, stageBox.Y)
		require.LessOrEqual(t, dialogBox.X+dialogBox.Width, stageBox.X+stageBox.Width)
		require.LessOrEqual(t, dialogBox.Y+dialogBox.Height, stageBox.Y+stageBox.Height)

		require.NoError(t, dialog.Locator("button").Filter(playwright.LocatorFilterOptions{HasText: "No, thank you"}).Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))

		visibleBanners, err := page.Locator("#banner-variants [role='banner']:visible").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, visibleBanners)
	})
}
