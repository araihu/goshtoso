//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
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
