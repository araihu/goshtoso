//go:build e2e && (full || link)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkComponentDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/link", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#link-fragment").WaitFor())

	t.Run("default and inline links expose hrefs and visible text", func(t *testing.T) {
		defaultLink := page.Locator("#link-default a").First()
		require.NoError(t, defaultLink.GetByText("Read if bored", playwright.LocatorGetByTextOptions{
			Exact: new(true),
		}).WaitFor())
		assert.Equal(t, "#", mustAttribute(t, defaultLink, "href"))

		className := mustAttribute(t, defaultLink, "class")
		assert.Contains(t, className, "text-primary")

		inline := page.Locator("#link-inline")
		require.NoError(t, inline.GetByText("Follow us on", playwright.LocatorGetByTextOptions{Exact: new(false)}).WaitFor())
		require.NoError(t, inline.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: "social media"}).WaitFor())
		assert.Equal(t, "#", mustAttribute(t, inline.Locator("a").First(), "href"))
	})

	t.Run("icon link keeps text and decorative icon together", func(t *testing.T) {
		iconLink := page.Locator("#link-icon a").First()
		require.NoError(t, iconLink.GetByText("about our company", playwright.LocatorGetByTextOptions{
			Exact: new(true),
		}).WaitFor())

		iconCount, err := iconLink.Locator("svg[aria-hidden='true']").Count()
		require.NoError(t, err)
		assert.Equal(t, 1, iconCount)

		className := mustAttribute(t, iconLink, "class")
		assert.Contains(t, className, "inline-flex")
		assert.Contains(t, className, "gap-1.5")
	})

	t.Run("button appearance preserves link semantics and primary action styling", func(t *testing.T) {
		buttonLink := page.Locator("#link-button a").First()
		require.NoError(t, buttonLink.GetByText("I'm a link", playwright.LocatorGetByTextOptions{
			Exact: new(true),
		}).WaitFor())
		assert.Empty(t, mustAttribute(t, buttonLink, "role"))

		buttonClass := mustAttribute(t, buttonLink, "class")
		assert.Contains(t, buttonClass, "bg-primary")
		assert.Contains(t, buttonClass, "rounded-2xl")
	})
}
