//go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTooltip_DefaultHover tests that hover tooltips render and show on hover
func TestTooltip_DefaultHover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/tooltip", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Page_Loads", func(t *testing.T) {
		title := page.Locator("h1")
		text, err := title.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Tooltip")
		t.Log("Tooltip demo page loads correctly")
	})

	t.Run("Default_Tooltips_Exist", func(t *testing.T) {
		// Check that tooltip elements exist with role="tooltip"
		tooltips := page.Locator("[role='tooltip']")
		count, err := tooltips.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 4, "should have at least 4 tooltip elements")
		t.Log("Default tooltip elements exist")
	})

	t.Run("Tooltip_Has_Aria_Describedby", func(t *testing.T) {
		// Check that trigger buttons have aria-describedby
		trigger := page.Locator("[aria-describedby='demoTop']")
		count, err := trigger.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have a trigger with aria-describedby='demoTop'")
		t.Log("Tooltip trigger has correct aria-describedby")
	})

	t.Run("Hover_Tooltip_Initially_Hidden", func(t *testing.T) {
		// Default tooltip should be invisible (opacity-0)
		tooltip := page.Locator("#demoTop")
		classes, err := tooltip.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classes, "opacity-0", "tooltip should be initially hidden with opacity-0")
		t.Log("Hover tooltip is initially hidden")
	})

	t.Run("Hover_Shows_Tooltip", func(t *testing.T) {
		// Hover over the trigger button
		trigger := page.Locator("[aria-describedby='demoTop']")
		err := trigger.Hover()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Tooltip should become visible via peer-hover
		tooltip := page.Locator("#demoTop")
		visible, err := tooltip.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "tooltip should be visible after hover")
		t.Log("Tooltip becomes visible on hover")
	})
}

// TestTooltip_RichTooltip tests tooltips with description
func TestTooltip_RichTooltip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/tooltip", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Rich_Tooltip_Has_Title_And_Description", func(t *testing.T) {
		tooltip := page.Locator("#richTop")
		count, err := tooltip.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "rich tooltip should exist")

		// Check it contains a title span and a description paragraph
		titleSpan := tooltip.Locator("span.text-sm.font-medium")
		titleCount, err := titleSpan.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, titleCount, "rich tooltip should have a title span")

		descP := tooltip.Locator("p.text-balance")
		descCount, err := descP.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, descCount, "rich tooltip should have a description paragraph")

		t.Log("Rich tooltip has title and description")
	})
}

// TestTooltip_ClickTooltip tests click-triggered tooltips
func TestTooltip_ClickTooltip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/tooltip", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond) // Wait for Alpine.js

	t.Run("Click_Tooltip_Initially_Hidden", func(t *testing.T) {
		tooltip := page.Locator("#clickTop")
		visible, err := tooltip.IsVisible()
		require.NoError(t, err)
		assert.False(t, visible, "click tooltip should be hidden initially (x-cloak)")
		t.Log("Click tooltip is initially hidden")
	})

	t.Run("Click_Shows_Tooltip", func(t *testing.T) {
		// Click the trigger
		trigger := page.Locator("[aria-describedby='clickTop']")
		err := trigger.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Tooltip should be visible
		tooltip := page.Locator("#clickTop")
		visible, err := tooltip.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "click tooltip should be visible after click")
		t.Log("Click tooltip shows on click")
	})

	t.Run("Click_Outside_Hides_Tooltip", func(t *testing.T) {
		// First make sure tooltip is shown
		trigger := page.Locator("[aria-describedby='clickTop']")
		err := trigger.Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		// Click outside (use body to ensure it's a valid target)
		_ = page.Locator("body").Click(playwright.LocatorClickOptions{
			Position: &playwright.Position{X: 10, Y: 10},
		})
		time.Sleep(150 * time.Millisecond)

		// Tooltip should be hidden
		tooltip := page.Locator("#clickTop")
		visible, err := tooltip.IsVisible()
		require.NoError(t, err)
		assert.False(t, visible, "click tooltip should be hidden after clicking outside")
		t.Log("Click tooltip hides on outside click")
	})
}

// TestTooltip_Positions tests that all position variants render
func TestTooltip_Positions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/tooltip", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	positions := []struct {
		id       string
		expected string
	}{
		{"demoTop", "bottom-full"},
		{"demoBottom", "top-full"},
		{"demoLeft", "right-full"},
		{"demoRight", "left-full"},
	}

	for _, pos := range positions {
		t.Run("Position_"+pos.id, func(t *testing.T) {
			tooltip := page.Locator("#" + pos.id)
			classes, err := tooltip.GetAttribute("class")
			require.NoError(t, err)
			assert.Contains(t, classes, pos.expected, "tooltip %s should have %s class", pos.id, pos.expected)
			t.Logf("Tooltip %s has correct position class %s", pos.id, pos.expected)
		})
	}
}
