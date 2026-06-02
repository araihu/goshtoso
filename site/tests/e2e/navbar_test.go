package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNavbar_DesktopLinksVisible(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	_ = setupServer(t)
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/navbar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// Verify at least one navbar is rendered on the demo page
	navCount, err := page.Locator("nav[aria-label='main navigation']").Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, navCount, 1, "expected at least one navbar on demo page")

	// Verify the first navbar has nav links visible (desktop)
	firstNav := page.Locator("nav[aria-label='main navigation']").First()

	// Check brand is present
	brandVisible, err := firstNav.Locator("a >> text=Penguin").IsVisible()
	require.NoError(t, err)
	assert.True(t, brandVisible, "brand should be visible")

	// Check Products link exists
	productsCount, err := firstNav.Locator("a >> text=Products").Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, productsCount, 1, "Products link should exist")
}

func TestNavbar_AvatarDropdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	_ = setupServer(t)
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/navbar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// Find the navbar with user profile (second or third demo variant)
	// Look for the avatar button
	avatarBtn := page.Locator("button[aria-label='user menu']").First()
	require.NoError(t, avatarBtn.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	// Click avatar to open dropdown
	err = avatarBtn.Click()
	require.NoError(t, err)

	// Verify dropdown shows user info
	dropdown := page.Locator("ul[role='menu']").First()
	require.NoError(t, dropdown.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	// Check user name is visible in dropdown
	nameVisible, err := dropdown.Locator("text=Alice Brown").IsVisible()
	require.NoError(t, err)
	assert.True(t, nameVisible, "user name should be visible in dropdown")

	// Check menu items exist
	menuItemCount, err := dropdown.Locator("a[role='menuitem']").Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, menuItemCount, 1, "dropdown should have at least one menu item")
}

func TestNavbar_MobileMenu(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	_ = setupServer(t)
	_, browser, _ := setupPlaywright(t)

	// Create page with mobile viewport
	page := newPage(t, browser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 375, Height: 812},
	})

	_, err := page.Goto(baseURL+"/components/navbar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// Desktop links should be hidden on mobile
	firstNav := page.Locator("nav[aria-label='main navigation']").First()
	desktopMenu := firstNav.Locator("div.hidden.sm\\:flex")
	isHidden, err := desktopMenu.IsHidden()
	require.NoError(t, err)
	assert.True(t, isHidden, "desktop menu should be hidden on mobile viewport")

	// Find and click hamburger button
	hamburger := firstNav.Locator("button[aria-label='mobile menu']")
	require.NoError(t, hamburger.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	err = hamburger.Click()
	require.NoError(t, err)

	// Wait for mobile menu to appear
	// Mobile menu contains nav links
	mobileLink := page.Locator("ul.fixed >> text=Products").First()
	require.NoError(t, mobileLink.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	mobileLinkVisible, err := mobileLink.IsVisible()
	require.NoError(t, err)
	assert.True(t, mobileLinkVisible, "mobile menu should show nav links")
}
