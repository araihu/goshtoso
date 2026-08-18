//go:build e2e && (full || navbar)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
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
	secondaryNav := page.Locator("#navbar-secondary nav[aria-label='secondary navigation']")
	require.NoError(t, secondaryNav.WaitFor())
	currentSecondary := secondaryNav.Locator("a[aria-current='page']")
	require.Equal(t, 1, mustCount(t, currentSecondary))
	require.Equal(t, "Overview", mustText(t, currentSecondary))

	secondaryTrigger := page.Locator("#navbar-secondary-actions [data-popover-trigger] button")
	secondaryPanel := page.Locator("#navbar-secondary-actions [data-popover-panel]")
	require.NoError(t, secondaryTrigger.Click())
	require.NoError(t, secondaryPanel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	secondaryTriggerClass, err := secondaryTrigger.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, secondaryTriggerClass, "border-primary")
	assert.Contains(t, secondaryTriggerClass, "bg-surface-dark-alt/10")
	require.Equal(t, "Open action", mustText(t, secondaryPanel.Locator("[role='menuitem']")))
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, secondaryPanel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	secondaryTriggerClass, err = secondaryTrigger.GetAttribute("class")
	require.NoError(t, err)
	assert.NotContains(t, secondaryTriggerClass, "border-primary")
	assert.NotContains(t, secondaryTriggerClass, "bg-surface-dark-alt/10")

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

func TestNavbarUserMenuReducedMotionShowsWithoutVisualTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))
	_, err := page.Goto(baseURL+"/components/navbar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	userNav := page.Locator("#navbar-user nav[aria-label='main navigation']")
	avatarButton := userNav.Locator("button[aria-label='user menu']")
	require.NoError(t, avatarButton.Hover())
	_, err = page.WaitForFunction(`() => {
		const button = document.querySelector("#navbar-user button[aria-label='user menu']");
		if (!button) return false;
		const style = getComputedStyle(button);
		return style.transitionProperty === 'none' &&
			(style.transform === 'none' || style.transform === 'matrix(1, 0, 0, 1, 0, 0)');
	}`, nil)
	require.NoError(t, err)

	require.NoError(t, avatarButton.Click())
	_, err = page.WaitForFunction(`() => {
		const menu = document.querySelector("#navbar-user ul[role='menu']");
		if (!menu || getComputedStyle(menu).display === 'none') return false;
		const style = getComputedStyle(menu);
		return style.transitionProperty === 'none' && style.opacity === '1';
	}`, nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("main").Click(playwright.LocatorClickOptions{
		Position: &playwright.Position{X: 10, Y: 10},
	}))
	_, err = page.WaitForFunction(`() => {
		const menu = document.querySelector("#navbar-user ul[role='menu']");
		return !!menu && getComputedStyle(menu).display === 'none';
	}`, nil)
	require.NoError(t, err)
}
