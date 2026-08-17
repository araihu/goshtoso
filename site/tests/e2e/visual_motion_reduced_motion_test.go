//go:build e2e && visualmotion

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestVisualComponentsHonorReducedMotion(t *testing.T) {
	cleanupServer := setupServer(t)
	defer cleanupServer()

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))

	t.Run("card", func(t *testing.T) {
		_, err := page.Goto(baseURL+"/components/card", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		card := page.Locator("#card-default article")
		require.NoError(t, card.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
		image := card.Locator("img")
		require.NoError(t, card.Hover())

		transitionProperty, err := image.Evaluate("element => getComputedStyle(element).transitionProperty", nil)
		require.NoError(t, err)
		require.Equal(t, "none", transitionProperty, "reduced-motion Card image should not transition")
		transform, err := image.Evaluate("element => getComputedStyle(element).transform", nil)
		require.NoError(t, err)
		require.Contains(t, []string{"none", "matrix(1, 0, 0, 1, 0, 0)"}, transform, "reduced-motion Card image should not scale on hover")
	})

	t.Run("carousel", func(t *testing.T) {
		_, err := page.Goto(baseURL+"/components/carousel", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)
		require.NoError(t, waitForAlpine(page))

		for _, selector := range []string{
			"#carousel-default-c button[aria-label='previous slide']",
			"#carousel-default-c button[aria-label='next slide']",
			"#carousel-default-c button[aria-label='slide 1']",
			"#carousel-autoplay-c button[aria-label='pause carousel']",
			"#carousel-cta-c a",
			"#carousel-card-c button[aria-label='next slide']",
		} {
			control := page.Locator(selector).First()
			require.NoError(t, control.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateAttached,
			}))
			transitionProperty, err := control.Evaluate("element => getComputedStyle(element).transitionProperty", nil)
			require.NoError(t, err)
			require.Equal(t, "none", transitionProperty, "reduced-motion Carousel control %s should not transition", selector)
		}
	})

	t.Run("rating", func(t *testing.T) {
		navigateToRatingDemo(t, page)

		icons := page.Locator("#rating-stars label > span[aria-hidden='true'], #rating-sentiment label > span[aria-hidden='true'], #rating-display span[aria-hidden='true']")
		count, err := icons.Count()
		require.NoError(t, err)
		require.Greater(t, count, 0)
		for index := 0; index < count; index++ {
			transitionProperty, err := icons.Nth(index).Evaluate("element => getComputedStyle(element).transitionProperty", nil)
			require.NoError(t, err)
			require.Equal(t, "none", transitionProperty, "reduced-motion Rating icon %d should not transition", index)
		}

		hovered := page.Locator("#rating-sentiment label[for='rating-sentiment-1']")
		require.NoError(t, hovered.Hover())
		transform, err := hovered.Locator("span[aria-hidden='true']").Evaluate("element => getComputedStyle(element).transform", nil)
		require.NoError(t, err)
		require.Contains(t, []string{"none", "matrix(1, 0, 0, 1, 0, 0)"}, transform, "reduced-motion Rating emoji should not scale on hover")
	})
}
