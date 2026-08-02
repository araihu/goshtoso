//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNavbarCoverageDemo(t *testing.T) {
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1280, Height: 900},
	})

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})
	page.On("pageerror", func(err error) {
		consoleErrors = append(consoleErrors, err.Error())
	})

	_, err := page.Goto(baseURL+"/components/navbar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}))

	require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{HasText: "Navbar"}).WaitFor())
	require.NoError(t, page.Locator("#navbar-simple nav[aria-label='main navigation']").WaitFor())
	require.NoError(t, page.Locator("#navbar-user nav[aria-label='main navigation']").WaitFor())
	rightSlotCount, err := page.Locator("#navbar-right-slot button[aria-label='Toggle dark mode']").Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, rightSlotCount, 1)

	userNav := page.Locator("#navbar-user nav[aria-label='main navigation']")
	avatarButton := userNav.Locator("button[aria-label='user menu']")
	require.NoError(t, avatarButton.ScrollIntoViewIfNeeded())
	require.NoError(t, avatarButton.WaitFor())

	expanded, err := avatarButton.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "false", expanded)

	require.NoError(t, avatarButton.Click())
	require.NoError(t, userNav.Locator("ul[role='menu']").WaitFor())

	expanded, err = avatarButton.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "true", expanded)
	require.NoError(t, userNav.Locator("ul[role='menu']").Filter(playwright.LocatorFilterOptions{HasText: "Alice Brown"}).WaitFor())

	signOut := userNav.Locator("a[role='menuitem']").Filter(playwright.LocatorFilterOptions{
		HasText: "Sign Out",
	})
	require.NoError(t, signOut.WaitFor())
	signOutClass, err := signOut.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, signOutClass, "text-danger")

	require.NoError(t, page.SetViewportSize(390, 844))
	mobileMenuButton := page.Locator("#navbar-simple > nav > button.sm\\:hidden")
	require.NoError(t, mobileMenuButton.ScrollIntoViewIfNeeded())
	require.NoError(t, mobileMenuButton.WaitFor())

	expanded, err = mobileMenuButton.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "false", expanded)

	require.NoError(t, mobileMenuButton.Click())
	expanded, err = mobileMenuButton.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "true", expanded)
	mobileMenu := page.Locator("#navbar-simple ul.fixed").First()
	require.NoError(t, mobileMenu.Filter(playwright.LocatorFilterOptions{HasText: "Products"}).WaitFor())
	require.NoError(t, mobileMenu.Filter(playwright.LocatorFilterOptions{HasText: "Login"}).WaitFor())

	assert.Empty(t, consoleErrors, "console errors on navbar demo")
}
