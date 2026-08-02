//go:build e2e && (full || tabs)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTabsCoverageDemoNoConsoleErrors loads the tabs demo page and asserts every
// documented variant container is present and Alpine boots without console
// errors.
func TestTabsCoverageDemoNoConsoleErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/tabs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	assert.True(t, mustContainText(t, page.Locator("main"), "Tabs"), "main should mention Tabs")
	for _, id := range []string{"#tabs-default", "#tabs-icons", "#tabs-badges", "#tabs-htmx", "#tabs-hash"} {
		assert.True(t, mustBeVisible(t, page.Locator(id)), "variant %s should be visible", id)
	}

	assert.Empty(t, consoleErrors, "tabs demo should load without console errors")
}

// TestTabsCoverageBadgeActiveStateSwitches verifies the badge active/inactive
// class binding toggles when switching tabs in the badges variant.
func TestTabsCoverageBadgeActiveStateSwitches(t *testing.T) {
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

	badges := page.Locator("#tabs-badges")

	// First tab (Groups) starts selected; its badge carries the active class.
	require.NoError(t, waitForTabSelected(page, "#tabs-badges [aria-controls='tabpanelbadgegroups']", true))
	assertBadgeActive(t, page, "#tabs-badges [aria-controls='tabpanelbadgegroups'] span", true)
	assertBadgeActive(t, page, "#tabs-badges [aria-controls='tabpanelbadgelikes'] span", false)

	// Activate Likes; active badge class moves with the selection.
	require.NoError(t, badges.GetByRole("tab", playwright.LocatorGetByRoleOptions{Name: "Likes"}).Click())
	require.NoError(t, waitForTabSelected(page, "#tabs-badges [aria-controls='tabpanelbadgelikes']", true))
	assertBadgeActive(t, page, "#tabs-badges [aria-controls='tabpanelbadgelikes'] span", true)
	assertBadgeActive(t, page, "#tabs-badges [aria-controls='tabpanelbadgegroups'] span", false)
}

// assertBadgeActive reads the live class list of a badge span and checks for the
// active marker class (bg-primary/10) that BadgeActiveClasses contributes.
func assertBadgeActive(t *testing.T, page playwright.Page, selector string, active bool) {
	t.Helper()
	cls, err := page.Evaluate(`(s) => document.querySelector(s)?.getAttribute('class') || ''`, selector)
	require.NoError(t, err)
	classStr, ok := cls.(string)
	require.True(t, ok)
	if active {
		assert.Contains(t, classStr, "bg-primary/10", "expected active badge classes for %s", selector)
	} else {
		assert.NotContains(t, classStr, "bg-primary/10", "expected inactive badge classes for %s", selector)
	}
}
