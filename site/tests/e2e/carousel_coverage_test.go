//go:build e2e && (full || carousel)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarouselCoverageDemoVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	var jsErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})
	page.On("pageerror", func(err error) {
		jsErrors = append(jsErrors, err.Error())
	})

	_, err := page.Goto(baseURL+"/components/carousel", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil)
	require.NoError(t, err)

	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{
		Name:  "Carousel",
		Exact: new(true),
	}).WaitFor())
	for _, selector := range []string{
		"#carousel-default-c",
		"#carousel-text-c",
		"#carousel-cta-c",
		"#carousel-autoplay-c",
		"#carousel-aspect-c",
		"#carousel-touch-c",
		"#carousel-card-c",
		"#carousel-htmx",
	} {
		require.NoErrorf(t, page.Locator(selector).WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}), "carousel example %s should be attached", selector)
	}

	require.NoError(t, page.Locator("#carousel-text-c").GetByText("Front end developers", playwright.LocatorGetByTextOptions{
		Exact: new(true),
	}).WaitFor())
	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{
		Name: "Get Started",
	}).WaitFor())

	aspectClass, err := page.Locator("#carousel-aspect-c > div").First().GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, aspectClass, "aspect-3/1")

	require.NoError(t, waitForCarouselIndex(page, "#carousel-touch-c", 1))
	indexAfterSwipe, err := page.Locator("#carousel-touch-c").Evaluate(`el => {
		const data = Alpine.$data(el);
		data.handleTouchStart({touches: [{clientX: 240}]});
		data.handleTouchMove({touches: [{clientX: 120}]});
		data.handleTouchEnd();
		return data.currentSlideIndex;
	}`, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 2, indexAfterSwipe)

	require.NoError(t, page.Locator("#carousel-card-c").GetByText("Abstract Blue Series", playwright.LocatorGetByTextOptions{
		Exact: new(true),
	}).WaitFor())
	require.NoError(t, page.Locator("#carousel-card-c button[aria-label='next slide']").Click())
	require.NoError(t, page.Locator("#carousel-card-c").GetByText("Sunset Collection", playwright.LocatorGetByTextOptions{
		Exact: new(true),
	}).WaitFor())

	loadedTitle := page.Locator("#carousel-htmx h3").Filter(playwright.LocatorFilterOptions{
		HasText: "Loaded via HTMX",
	})
	require.NoError(t, loadedTitle.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}))

	require.Empty(t, jsErrors, "no JS console/page errors on carousel demo: %v", jsErrors)
}
