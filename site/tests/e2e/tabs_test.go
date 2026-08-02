//go:build e2e && (full || tabs)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTabs_ClickAndKeyboardSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/tabs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	defaultTabs := page.Locator("#tabs-default")
	likes := defaultTabs.GetByRole("tab", playwright.LocatorGetByRoleOptions{Name: "Likes"})

	require.NoError(t, waitForTabSelected(page, "#tabs-default [aria-controls='tabpaneldefaultgroups']", true))
	require.NoError(t, likes.Click())
	require.NoError(t, waitForTabSelected(page, "#tabs-default [aria-controls='tabpaneldefaultlikes']", true))
	require.NoError(t, waitForPanelVisibility(page, "#tabpaneldefaultlikes", true))
	require.NoError(t, waitForPanelVisibility(page, "#tabpaneldefaultgroups", false))
	assert.True(t, mustBeVisible(t, defaultTabs.GetByRole("tabpanel", playwright.LocatorGetByRoleOptions{Name: "Likes"})))
	assert.False(t, mustBeVisible(t, defaultTabs.GetByRole("tabpanel", playwright.LocatorGetByRoleOptions{Name: "Groups"})))

	require.NoError(t, likes.Press("ArrowRight"))
	activeLabel, err := page.Evaluate(`() => document.activeElement?.textContent?.trim()`, nil)
	require.NoError(t, err)
	assert.Contains(t, activeLabel, "Comments")
}

func TestTabs_HTMXLazyPanelLoadsOnlyAfterSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/tabs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	panel := page.Locator("#tabpanelhtmxdetails")
	assert.False(t, mustContainText(t, panel, "Details (Lazy Loaded)"), "lazy panel should not be populated before first activation")

	require.NoError(t, page.Locator("#tabs-htmx").GetByRole("tab", playwright.LocatorGetByRoleOptions{Name: "Details"}).Click())
	require.NoError(t, panel.ScrollIntoViewIfNeeded())
	_, err = page.WaitForFunction(
		`() => document.querySelector('#tabpanelhtmxdetails')?.textContent?.includes('Details (Lazy Loaded)')`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)},
	)
	require.NoError(t, err)
	assert.True(t, mustContainText(t, panel, "Loaded Once"))
}

func TestTabs_URLHashSyncRestoresAndUpdatesSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/tabs#comments", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	hashTabs := page.Locator("#tabs-hash")
	require.NoError(t, waitForTabSelected(page, "#tabs-hash [aria-controls='tabpanelhash-synccomments']", true))
	assert.True(t, mustBeVisible(t, hashTabs.GetByRole("tabpanel", playwright.LocatorGetByRoleOptions{Name: "Comments"})))

	require.NoError(t, hashTabs.GetByRole("tab", playwright.LocatorGetByRoleOptions{Name: "Saved"}).Click())
	require.NoError(t, page.WaitForURL("**/components/tabs#saved"))
}

func waitForPanelVisibility(page playwright.Page, selector string, visible bool) error {
	_, err := page.WaitForFunction(
		`([selector, visible]) => {
			const el = document.querySelector(selector);
			if (!el) return false;
			const isVisible = getComputedStyle(el).display !== 'none' && el.offsetParent !== null;
			return isVisible === visible;
		}`,
		[]any{selector, visible},
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	return err
}
