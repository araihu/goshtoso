package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestLanding_HeroAndStructure loads the homepage ("/") and asserts the
// redesigned hero — headline + primary CTA — is present. The page is a
// standalone document (not the demo shell).
func TestLanding_HeroAndStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("HeroHeadline", func(t *testing.T) {
		h1 := page.Locator("#hero h1")
		txt, err := h1.InnerText()
		require.NoError(t, err)
		require.Contains(t, txt, "Build interactive UIs in Go")
	})

	t.Run("BrowseComponentsCTA", func(t *testing.T) {
		cta := page.Locator("#hero a[href='/components/button']")
		visible, err := cta.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "Browse components CTA should be visible")
	})
}
