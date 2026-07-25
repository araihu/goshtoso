package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarousel_DefaultNavigationAndIndicators(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/carousel", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForCarouselIndex(page, "#carousel-default-c", 1))

	next := page.Locator("#carousel-default-c button[aria-label='next slide']")
	previous := page.Locator("#carousel-default-c button[aria-label='previous slide']")

	require.NoError(t, next.Click())
	require.NoError(t, waitForCarouselIndex(page, "#carousel-default-c", 2))

	require.NoError(t, previous.Click())
	require.NoError(t, waitForCarouselIndex(page, "#carousel-default-c", 1))

	require.NoError(t, previous.Click())
	require.NoError(t, waitForCarouselIndex(page, "#carousel-default-c", 3))

	require.NoError(t, page.Locator("#carousel-default-c button[aria-label='slide 2']").Click())
	require.NoError(t, waitForCarouselIndex(page, "#carousel-default-c", 2))
}

func TestCarousel_HTMXLoadedSlidesRemainInteractive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/carousel", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	loadedTitle := page.Locator("#carousel-htmx h3").Filter(playwright.LocatorFilterOptions{
		HasText: "Loaded via HTMX",
	})
	require.NoError(t, loadedTitle.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}))

	require.NoError(t, waitForCarouselIndex(page, "#carousel-htmx [x-data]", 1))
	require.NoError(t, page.Locator("#carousel-htmx button[aria-label='next slide']").Click())
	require.NoError(t, waitForCarouselIndex(page, "#carousel-htmx [x-data]", 2))

	_, err = page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('#carousel-htmx h3')).
			some(el => el.textContent.includes('Dynamic Content') && el.offsetParent !== null)`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	assert.NoError(t, err, "HTMX-loaded carousel should keep Alpine navigation behavior")
}

func TestCarousel_AutoplayCanBePaused(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/carousel", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForCarouselIndex(page, "#carousel-autoplay-c", 1))

	pause := page.Locator("#carousel-autoplay-c button[aria-label='pause carousel']")
	pressed, err := pause.Evaluate("el => el.getAttribute('aria-pressed')", nil)
	require.NoError(t, err)
	assert.Equal(t, "false", pressed)

	require.NoError(t, pause.Click())
	pressed, err = pause.Evaluate("el => el.getAttribute('aria-pressed')", nil)
	require.NoError(t, err)
	assert.Equal(t, "true", pressed)
}

func waitForCarouselIndex(page playwright.Page, selector string, index int) error {
	_, err := page.WaitForFunction(
		`([selector, index]) => {
			const el = document.querySelector(selector);
			return !!el && Alpine.$data(el).currentSlideIndex === index;
		}`,
		[]any{selector, index},
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	return err
}
