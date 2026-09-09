//go:build e2e && (full || card)

package e2e

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardComponentDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/card", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	assertCardVariants(t, page)
}

func TestCardComponentFragmentThemes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%t", theme, dark), func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				failures := watchPageFailures(page)
				_, err := page.Goto(baseURL + "/components/button")
				require.NoError(t, err)
				_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined' && typeof htmx !== 'undefined'", nil)
				require.NoError(t, err)
				_, err = page.Evaluate(`state => {
                    window.__cardNavigation = true;
                    document.documentElement.setAttribute('data-theme', state.theme);
                    document.documentElement.classList.toggle('dark', state.dark);
                }`, map[string]any{"theme": theme, "dark": dark})
				require.NoError(t, err)
				link := page.Locator(`#componentdocshell-sidebar-content a[href="/components/card"]`)
				clickUntil(t, page, link, "!!document.getElementById('card-title-prefix')")
				preserved, err := page.Evaluate(`state => window.__cardNavigation === true && document.documentElement.getAttribute('data-theme') === state.theme && document.documentElement.classList.contains('dark') === state.dark`, map[string]any{"theme": theme, "dark": dark})
				require.NoError(t, err)
				require.Equal(t, true, preserved, "fragment navigation must preserve the document and theme")
				assertCardVariants(t, page)
				failures.RequireEmpty(t)
			})
		}
	}
}

func assertCardVariants(t *testing.T, page playwright.Page) {
	t.Helper()
	prefixTitle := page.Locator("#card-title-prefix h3")
	require.NoError(t, prefixTitle.Locator("svg").WaitFor())
	prefixLayout, err := prefixTitle.Evaluate(`el => getComputedStyle(el).display`, nil)
	require.NoError(t, err)
	assert.Equal(t, "flex", prefixLayout)
	cardOverflow, err := page.Locator("#card-title-prefix article").Evaluate(`el => getComputedStyle(el).overflow`, nil)
	require.NoError(t, err)
	assert.Equal(t, "visible", cardOverflow)

	defaultCard := page.Locator("#card-default article")
	require.NoError(t, defaultCard.WaitFor())
	assert.Equal(t, "A penguin robot talking with a human", mustAttribute(t, defaultCard.Locator("img"), "alt"))
	require.NoError(t, defaultCard.GetByText("Features", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	require.NoError(t, defaultCard.GetByText("Penguai can teach you Javascript", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

	require.NoError(t, page.Locator("#card-button button").Filter(playwright.LocatorFilterOptions{HasText: "Book Now"}).WaitFor())

	horizontalClass := mustAttribute(t, page.Locator("#card-horizontal article"), "class")
	assert.Contains(t, horizontalClass, "md:grid-cols-8")
	assert.Equal(t, "Man wearing VR goggles", mustAttribute(t, page.Locator("#card-horizontal img"), "alt"))
	require.NoError(t, page.Locator("#card-horizontal").GetByText("AI-Powered VR Goggles Redefine Reality").WaitFor())

	product := page.Locator("#card-product")
	require.NoError(t, product.GetByText("CASIO G-SHOCK GA2100", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	require.NoError(t, product.GetByText("$99.99", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	productRating := product.GetByRole("img", playwright.LocatorGetByRoleOptions{
		Name:  "Rated 3 stars",
		Exact: new(true),
	})
	require.NoError(t, productRating.WaitFor())
	assert.Equal(t, "Rated 3 stars", mustAttribute(t, productRating, "aria-label"))
	require.NoError(t, product.Locator("button").Filter(playwright.LocatorFilterOptions{HasText: "Add to Cart"}).WaitFor())

	pricing := page.Locator("#card-pricing")
	require.NoError(t, pricing.GetByText("Premium", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	require.NoError(t, pricing.GetByText("Unlimited access to all courses").WaitFor())
	require.NoError(t, pricing.Locator("button").Filter(playwright.LocatorFilterOptions{HasText: "Start your free trial"}).WaitFor())

	testimonial := page.Locator("#card-testimonial")
	require.NoError(t, testimonial.GetByText("Bob Johnson", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	require.NoError(t, testimonial.GetByText("CEO - TechNova", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
	testimonialRating := testimonial.GetByRole("img", playwright.LocatorGetByRoleOptions{
		Name:  "Rated 4 stars",
		Exact: new(true),
	})
	require.NoError(t, testimonialRating.WaitFor())
	assert.Equal(t, "Rated 4 stars", mustAttribute(t, testimonialRating, "aria-label"))
}
