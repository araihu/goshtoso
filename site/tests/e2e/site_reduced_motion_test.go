//go:build e2e && sitemotion

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestPublicDemoSurfacesHonorReducedMotion(t *testing.T) {
	cleanupServer := setupServer(t)
	defer cleanupServer()

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))

	t.Run("landing", func(t *testing.T) {
		navigateSiteMotionRoute(t, page, "/")
		cta := page.Locator("[data-primary-cta]")
		require.NoError(t, cta.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
		assertReducedMotionTransitions(t, page, "[class*='transition']")
		assertReducedMotionAlpineTransitions(t, page)

		image := page.Locator("[data-product-card] img").First()
		require.NoError(t, image.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
		require.NoError(t, image.Hover())
		assertIdentityTransform(t, image, "landing product image")
	})

	t.Run("card recipes", func(t *testing.T) {
		navigateSiteMotionRoute(t, page, "/components/card")
		productImage := page.Locator("#card-product img")
		require.NoError(t, productImage.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
		require.NoError(t, page.Locator("#card-product article").Hover())
		testimonialIcon := page.Locator("#card-testimonial > article > svg")
		assertReducedMotionTransitions(t, page, "#card-product img, #card-testimonial > article > svg")
		assertIdentityTransform(t, productImage, "Card product image")
		assertIdentityTransform(t, testimonialIcon, "Card testimonial icon")
	})

	t.Run("icon catalog and modal", func(t *testing.T) {
		navigateSiteMotionRoute(t, page, "/components/icon")
		card := page.Locator("[data-icon-card]").First()
		require.NoError(t, card.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
		icon := card.Locator("span").First()
		require.NoError(t, card.Hover())
		assertReducedMotionTransitions(t, page, "[data-icon-card], [data-icon-card] span")
		assertIdentityTransform(t, icon, "icon catalog preview")

		require.NoError(t, card.Click())
		dialog := page.Locator("[data-testid='icon-picker-dialog']")
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		assertReducedMotionTransitions(t, page, "[data-testid='icon-picker-dialog'], [data-testid='icon-picker-dialog'] > div")
		assertReducedMotionAlpineTransitions(t, page)
	})

	t.Run("theme documentation", func(t *testing.T) {
		navigateSiteMotionRoute(t, page, "/docs/theme")
		tile := page.Locator("button[data-theme-key]").First()
		require.NoError(t, tile.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
		require.NoError(t, tile.Hover())
		assertReducedMotionTransitions(t, page, "[class*='transition']")
		assertReducedMotionAlpineTransitions(t, page)
	})
}

func navigateSiteMotionRoute(t *testing.T, page playwright.Page, route string) {
	t.Helper()
	_, err := page.Goto(baseURL+route, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
}

func assertReducedMotionTransitions(t *testing.T, page playwright.Page, selector string) {
	t.Helper()
	result, err := page.Evaluate(`selector => Array.from(document.querySelectorAll(selector)).every(element => getComputedStyle(element).transitionProperty === "none")`, selector)
	require.NoError(t, err)
	require.Equal(t, true, result, "reduced-motion transitions should be disabled for %s", selector)
}

func assertReducedMotionAlpineTransitions(t *testing.T, page playwright.Page) {
	t.Helper()
	count, err := page.Locator(`[x-transition\:enter]`).Count()
	require.NoError(t, err)
	require.Greater(t, count, 0, "page should expose an Alpine transition contract")
	result, err := page.Evaluate(`() => Array.from(document.querySelectorAll('[x-transition\\:enter], [x-transition\\:leave]')).every(element => getComputedStyle(element).transitionProperty === "none")`, nil)
	require.NoError(t, err)
	require.Equal(t, true, result, "reduced-motion Alpine transitions should be disabled")
}

func assertIdentityTransform(t *testing.T, locator playwright.Locator, label string) {
	t.Helper()
	transform, err := locator.Evaluate("element => getComputedStyle(element).transform", nil)
	require.NoError(t, err)
	require.Contains(t, []string{"none", "matrix(1, 0, 0, 1, 0, 0)"}, transform, "%s should not scale under reduced motion", label)
}
