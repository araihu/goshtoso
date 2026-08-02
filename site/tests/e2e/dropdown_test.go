//go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDropdown_ClickVariant tests click-to-open dropdown
func TestDropdown_ClickVariant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/dropdown", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Click_Opens_Dropdown", func(t *testing.T) {
		// Find the click dropdown trigger button
		button := page.Locator("#dropdown-click button").First()

		// Wait for Alpine.js to process x-show/x-cloak
		_, _ = page.WaitForFunction("() => { const m = document.querySelector('#dropdown-click [role=\"menu\"]'); return m && (m.style.display === 'none' || m.offsetParent === null); }", nil, playwright.PageWaitForFunctionOptions{
			Timeout: playwright.Float(3000),
		})

		// Check visibility via JS (more reliable than Playwright for Alpine.js x-show)
		hidden, err := page.Evaluate("() => { const m = document.querySelector('#dropdown-click [role=\"menu\"]'); return m ? m.offsetParent === null : true; }", nil)
		require.NoError(t, err)
		assert.True(t, hidden.(bool), "dropdown menu should be hidden initially")

		// Click to open
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		// Menu should now be visible
		menu := page.Locator("#dropdown-click [role='menu']")
		visible, err := menu.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "dropdown menu should be visible after click")

		// Click again to close
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		visible, err = menu.IsVisible()
		require.NoError(t, err)
		assert.False(t, visible, "dropdown menu should be hidden after second click")

		t.Log("Click dropdown opens and closes correctly")
	})

	t.Run("Click_Dropdown_Has_Menu_Items", func(t *testing.T) {
		button := page.Locator("#dropdown-click button").First()
		err := button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Check menu items
		items := page.Locator("#dropdown-click [role='menuitem']")
		count, err := items.Count()
		require.NoError(t, err)
		assert.Equal(t, 4, count, "should have 4 menu items")

		// Verify item labels
		firstItem, err := items.Nth(0).TextContent()
		require.NoError(t, err)
		assert.Contains(t, firstItem, "Dashboard")

		// Close
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		t.Log("Click dropdown has correct menu items")
	})

	t.Run("Aria_Expanded_Updates", func(t *testing.T) {
		button := page.Locator("#dropdown-click button").First()
		time.Sleep(150 * time.Millisecond) // Wait for Alpine.js hydration

		// Initially aria-expanded should be false (Alpine.js binding)
		expanded, err := button.Evaluate("el => el.getAttribute('aria-expanded')", nil)
		require.NoError(t, err)
		assert.Equal(t, "false", expanded)

		// Click to open
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		expanded, err = button.Evaluate("el => el.getAttribute('aria-expanded')", nil)
		require.NoError(t, err)
		assert.Equal(t, "true", expanded)

		// Close
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		t.Log("aria-expanded updates correctly")
	})
}

// TestDropdown_WithDividers tests dropdown with sections and dividers
func TestDropdown_WithDividers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/dropdown", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Divider_Dropdown_Has_Sections", func(t *testing.T) {
		button := page.Locator("#dropdown-divider button").First()
		err := button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Should have dividers (divide-y class on menu)
		menu := page.Locator("#dropdown-divider [role='menu']")
		menuClass, err := menu.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, menuClass, "divide-y", "should have divider classes")

		// Should have menu items across sections
		items := page.Locator("#dropdown-divider [role='menuitem']")
		count, err := items.Count()
		require.NoError(t, err)
		assert.Equal(t, 6, count, "should have 6 menu items across 3 sections")

		// Close
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		t.Log("Divider dropdown has correct sections")
	})
}

// TestDropdown_WithIcons tests dropdown with icon items
func TestDropdown_WithIcons(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/dropdown", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Icon_Dropdown_Has_Icons", func(t *testing.T) {
		button := page.Locator("#dropdown-icons button").First()
		err := button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Items should contain SVG icons
		firstItem := page.Locator("#dropdown-icons [role='menuitem']").First()
		svgs := firstItem.Locator("svg")
		svgCount, err := svgs.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, svgCount, 1, "menu items should contain icons")

		// Close
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		t.Log("Icon dropdown has correct icons")
	})
}

// TestDropdown_ContextMenu tests context menu dropdown
func TestDropdown_ContextMenu(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/dropdown", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Context_Menu_Opens_On_Click", func(t *testing.T) {
		button := page.Locator("#dropdown-context button").First()
		err := button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Menu should be visible
		menu := page.Locator("#dropdown-context [role='menu']")
		visible, err := menu.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "context menu should be visible")

		// Should have list items (not anchors)
		items := page.Locator("#dropdown-context [role='menuitem']")
		count, err := items.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 4, "should have context menu items")

		// Close
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		t.Log("Context menu opens and shows items correctly")
	})

	t.Run("Context_Menu_Has_Shortcuts", func(t *testing.T) {
		button := page.Locator("#dropdown-context button").First()
		err := button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Should show shortcut labels
		undoItem := page.Locator("#dropdown-context [role='menuitem']").First()
		text, err := undoItem.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Undo", "first item should be Undo")
		assert.Contains(t, text, "Z", "first item should show Z shortcut")

		// Close
		err = button.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		t.Log("Context menu shows keyboard shortcuts")
	})
}

// TestDropdown_PageLoads tests that the dropdown page loads correctly
func TestDropdown_PageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/dropdown", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Page_Title_Contains_Dropdown", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Dropdown")

		t.Log("Dropdown page loads with correct title")
	})

	t.Run("All_Dropdown_Variants_Present", func(t *testing.T) {
		// Check all variant sections are rendered
		clickDropdown := page.Locator("#dropdown-click")
		count, err := clickDropdown.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "click dropdown should be present")

		hoverDropdown := page.Locator("#dropdown-hover")
		count, err = hoverDropdown.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "hover dropdown should be present")

		dividerDropdown := page.Locator("#dropdown-divider")
		count, err = dividerDropdown.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "divider dropdown should be present")

		iconDropdown := page.Locator("#dropdown-icons")
		count, err = iconDropdown.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "icon dropdown should be present")

		contextDropdown := page.Locator("#dropdown-context")
		count, err = contextDropdown.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "context dropdown should be present")

		t.Log("All dropdown variants are present on page")
	})
}

// TestDropdown_ActionsMenu tests the button-item variant: OnClick, Disabled,
// Danger styling, and the icon-only click trigger.
func TestDropdown_ActionsMenu(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/dropdown", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// Wait for Alpine hydration before interacting
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err)

	// openMenu ensures the dropdown is open (idempotent). Clicking an item
	// doesn't auto-close the menu, so a second direct click on the trigger
	// would toggle it closed. Check aria-expanded first.
	openMenu := func(t *testing.T) {
		t.Helper()
		trigger := page.Locator("#dropdown-actions button").First()
		expanded, err := trigger.Evaluate("el => el.getAttribute('aria-expanded')", nil)
		require.NoError(t, err)
		if expanded == "true" {
			return
		}
		require.NoError(t, trigger.Click())
		time.Sleep(150 * time.Millisecond)
	}

	t.Run("IconOnly_Trigger_Has_AriaLabel_And_No_Chevron", func(t *testing.T) {
		trigger := page.Locator("#dropdown-actions button").First()

		aria, err := trigger.Evaluate("el => el.getAttribute('aria-label')", nil)
		require.NoError(t, err)
		assert.Equal(t, "cluster actions", aria, "icon-only trigger should carry aria-label from cfg.Label")

		// The trigger should contain exactly one SVG (the TriggerIcon); no chevron.
		svgCount, err := trigger.Locator("svg").Count()
		require.NoError(t, err)
		assert.Equal(t, 1, svgCount, "icon-only trigger should render only the TriggerIcon (no chevron)")

		text, err := trigger.TextContent()
		require.NoError(t, err)
		assert.Empty(t, text, "icon-only trigger should have no visible label text")
	})

	t.Run("OnClick_Button_Item_Fires_Alpine", func(t *testing.T) {
		openMenu(t)

		editItem := page.Locator("#dropdown-actions-edit")
		tag, err := editItem.Evaluate("el => el.tagName.toLowerCase()", nil)
		require.NoError(t, err)
		assert.Equal(t, "button", tag, "OnClick item must render as <button>")

		require.NoError(t, editItem.Click())
		time.Sleep(150 * time.Millisecond)

		// Alpine state should reflect the click. Read the scoped x-data via the
		// wrapper's __x.$data — the wrapping div is the closest ancestor with x-data.
		state, err := page.Evaluate(`() => {
			const el = document.querySelector('[x-data*="editOpen"]');
			return { editOpen: el._x_dataStack[0].editOpen, editCount: el._x_dataStack[0].editCount };
		}`, nil)
		require.NoError(t, err)
		stateMap := state.(map[string]any)
		assert.Equal(t, true, stateMap["editOpen"], "OnClick should flip editOpen to true")
		assert.EqualValues(t, 1, stateMap["editCount"], "OnClick should increment editCount")
	})

	t.Run("Disabled_Button_Item_Is_Inert", func(t *testing.T) {
		openMenu(t)

		archive := page.Locator("#dropdown-actions-locked")

		disabled, err := archive.Evaluate("el => el.hasAttribute('disabled')", nil)
		require.NoError(t, err)
		assert.Equal(t, true, disabled, "disabled item must have native disabled attribute")

		cls, err := archive.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, cls, "opacity-50", "disabled item should carry opacity-50")
		assert.Contains(t, cls, "cursor-not-allowed", "disabled item should carry cursor-not-allowed")
		assert.Contains(t, cls, "pointer-events-none", "disabled item should carry pointer-events-none")

		title, err := archive.GetAttribute("title")
		require.NoError(t, err)
		assert.Equal(t, "Not available", title, "Tooltip should render as native title attr")
	})

	t.Run("Danger_Item_Has_Danger_Classes", func(t *testing.T) {
		openMenu(t)

		del := page.Locator("#dropdown-actions-delete")
		cls, err := del.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, cls, "text-danger", "danger item should carry text-danger class")
	})
}
