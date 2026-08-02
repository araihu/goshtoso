//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCardCoverageFragmentNav navigates to the card demo via the sidebar
// (fragment swap), confirms every variant container renders after the swap,
// and asserts no uncaught JS exceptions or console errors are produced. Card is
// a static component, so the regression risk is render/markup, not interaction.
func TestCardCoverageFragmentNav(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	// Land elsewhere, then navigate via the sidebar fragment link.
	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/components/card']").First().Click())

	// Every demo variant container must land after the swap.
	for _, id := range []string{
		"#card-default",
		"#card-button",
		"#card-horizontal",
		"#card-product",
		"#card-pricing",
		"#card-testimonial",
	} {
		require.NoError(t, page.Locator(id+" article").First().WaitFor(
			playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateAttached,
				Timeout: playwright.Float(3000),
			}), "variant %s did not render after fragment nav", id)
	}

	assert.Empty(t, pageErrors, "uncaught JS exceptions on card demo")
	assert.Empty(t, consoleErrors, "console errors on card demo")
}

// TestCardCoverageHorizontalLayout asserts the horizontal variant carries the
// grid layout classes from ContainerClasses, and the default variant does not.
func TestCardCoverageHorizontalLayout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/card", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#card-horizontal article").WaitFor())
	horizontal := mustAttribute(t, page.Locator("#card-horizontal article"), "class")
	assert.Contains(t, horizontal, "md:grid-cols-8")
	assert.Contains(t, horizontal, "max-w-2xl")

	vertical := mustAttribute(t, page.Locator("#card-default article"), "class")
	assert.Contains(t, vertical, "flex-col")
	assert.NotContains(t, vertical, "md:grid-cols-8")
}

// TestCardCoverageDarkMode toggles the <html> .dark class and confirms the card
// article stays rendered and visible in dark mode (light/dark regression).
func TestCardCoverageDarkMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/card", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	card := page.Locator("#card-default article")
	require.NoError(t, card.WaitFor())
	visible, err := card.IsVisible()
	require.NoError(t, err)
	require.True(t, visible, "card not visible in light mode")

	// Force dark mode on <html> and re-confirm the card renders/visible.
	_, err = page.Evaluate("() => document.documentElement.classList.add('dark')", nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)

	visible, err = card.IsVisible()
	require.NoError(t, err)
	require.True(t, visible, "card not visible in dark mode")
	// Dark utility classes are present on the root regardless of toggle state.
	assert.Contains(t, mustAttribute(t, card, "class"), "dark:bg-surface-dark-alt")
}
