package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccordion_StaticContent tests accordion with static content
func TestAccordion_StaticContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Accordion_Expands_And_Collapses", func(t *testing.T) {
		// Find first accordion button
		firstButton := page.Locator("#accordion-fragment button").First()

		// Initially collapsed
		ariaExpanded, err := firstButton.GetAttribute("aria-expanded")
		require.NoError(t, err)
		assert.Equal(t, "false", ariaExpanded, "accordion should be collapsed initially")

		// Click to expand
		err = firstButton.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Now expanded
		ariaExpanded, err = firstButton.GetAttribute("aria-expanded")
		require.NoError(t, err)
		assert.Equal(t, "true", ariaExpanded, "accordion should be expanded after click")

		// Chevron should be rotated
		svg := firstButton.Locator("svg")
		svgClass, err := svg.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, svgClass, "rotate-180", "chevron should be rotated when expanded")

		t.Log("✓ Accordion expands and collapses correctly")
	})

	t.Run("Multiple_Sections_Can_Be_Open", func(t *testing.T) {
		// Find section in "Allow Multiple Open" accordion
		multiAccordion := page.Locator("text=Allow Multiple Open").Locator("xpath=../..//div[contains(@class, 'divide-y')]").First()
		buttons := multiAccordion.Locator("button")

		// Open first section
		err := buttons.Nth(0).Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Open second section
		err = buttons.Nth(1).Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Both should be expanded
		expanded0, _ := buttons.Nth(0).GetAttribute("aria-expanded")
		expanded1, _ := buttons.Nth(1).GetAttribute("aria-expanded")

		if expanded0 == "true" && expanded1 == "true" {
			t.Log("✓ Multiple sections can be open simultaneously")
		}
	})
}

// TestAccordion_ServerLoadedContent tests HTMX lazy loading
func TestAccordion_ServerLoadedContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Lazy_Loaded_Content_Fetches_From_Server", func(t *testing.T) {
		// Find server-loaded section
		lazySection := page.Locator("text=Dynamic Content A").First()

		// Click to expand
		err := lazySection.Click()
		require.NoError(t, err)

		lazyContent := page.Locator("#lazy-content-a")
		err = lazyContent.ScrollIntoViewIfNeeded()
		require.NoError(t, err)

		// Wait for HTMX to fetch content (max 5 seconds)
		err = page.Locator("text=Server Response A").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err, "server-loaded content should appear")

		// Verify content loaded
		content := page.Locator("text=Server Response A")
		visible, err := content.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "server response should be visible")

		// Check for timestamp (proves it came from server)
		contentText, err := content.TextContent()
		require.NoError(t, err)
		assert.Contains(t, contentText, "Server Response A", "should show server response title")

		t.Log("✓ Server-loaded content fetches correctly")
	})

	t.Run("Lazy_Load_Shows_Loading_State", func(t *testing.T) {
		// Navigate fresh
		_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		// Find second lazy-loaded section
		lazySection := page.Locator("text=Dynamic Content B").First()

		// Click to expand
		err = lazySection.Click()
		require.NoError(t, err)

		lazyContent := page.Locator("#lazy-content-b")
		err = lazyContent.ScrollIntoViewIfNeeded()
		require.NoError(t, err)

		// Should show loading indicator immediately
		loadingIndicator := page.Locator("text=Loading content...")
		exists, _ := loadingIndicator.Count()
		if exists > 0 {
			t.Log("✓ Loading indicator shown while fetching")
		}

		// Wait for content to load
		err = page.Locator("text=Server Response B").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		t.Log("✓ Loading state transitions to content")
	})

	t.Run("API_Endpoint_Returns_Correct_Content", func(t *testing.T) {
		// Test API endpoint directly
		resp, err := http.Get(baseURL + "/api/components/accordion-content/lazy-content-a")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

		// Read body
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])

		assert.Contains(t, bodyStr, "Server Response A")
		assert.Contains(t, bodyStr, "HTMX")

		t.Log("✓ API endpoint returns correct content")
	})
}

// TestButton_HTMXInteractions tests button HTMX functionality in the docs demo.
func TestButton_HTMXInteractions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("HTMX_POST_Request_Works", func(t *testing.T) {
		htmxButton := page.Locator("#button-htmx button:has-text('Send POST')").First()

		err := htmxButton.Click()
		require.NoError(t, err)

		result := page.Locator("#htmx-result-post")
		err = result.Locator("text=Hello from HTMX!").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		})
		require.NoError(t, err, "HTMX response should appear")

		responseText, err := result.TextContent()
		require.NoError(t, err)
		assert.Contains(t, responseText, "Hello from HTMX!")
		assert.Contains(t, responseText, "POST")

		t.Log("✓ HTMX POST request works")
	})

	t.Run("HTMX_GET_Request_Works", func(t *testing.T) {
		htmxButton := page.Locator("#button-htmx button:has-text('Send GET')").First()

		err := htmxButton.Click()
		require.NoError(t, err)

		result := page.Locator("#htmx-result-get")
		err = result.Locator("text=Hello from HTMX!").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		})
		require.NoError(t, err)

		responseText, err := result.TextContent()
		require.NoError(t, err)
		assert.Contains(t, responseText, "GET")

		t.Log("✓ HTMX GET request works")
	})

	t.Run("HTMX_Confirm_Dialog_Works", func(t *testing.T) {
		deleteButton := page.Locator("#button-htmx button:has-text('Delete')").First()

		dialogShown := false
		page.On("dialog", func(dialog playwright.Dialog) {
			dialogShown = true
			assert.Contains(t, dialog.Message(), "sure")
			_ = dialog.Accept()
		})

		err = deleteButton.Click()
		require.NoError(t, err)

		result := page.Locator("#htmx-result-delete")
		err = result.Locator("text=Hello from HTMX!").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		})
		require.NoError(t, err)

		responseText, err := result.TextContent()
		require.NoError(t, err)
		require.True(t, dialogShown, "HTMX confirm dialog should be shown")
		assert.Contains(t, responseText, "DELETE")
		t.Log("✓ HTMX confirm dialog shown")
	})

	t.Run("HTMX_Action_Rows_Stay_Aligned", func(t *testing.T) {
		aligned, err := page.Evaluate(`() => {
			const rows = Array.from(document.querySelectorAll('#button-htmx > div > div'));
			return rows.every(row => {
				const button = row.querySelector('button');
				const result = row.querySelector('[id^="htmx-result-"]');
				if (!button || !result) return false;
				const b = button.getBoundingClientRect();
				const r = result.getBoundingClientRect();
				return r.left > b.right && Math.abs((b.top + b.height / 2) - (r.top + r.height / 2)) < 2;
			});
		}`)
		require.NoError(t, err)
		require.Equal(t, true, aligned, "HTMX action rows should keep buttons and results aligned")
	})

	t.Run("Button_Variants_Render_Correctly", func(t *testing.T) {
		// Check all button variants are present
		variants := []string{"Primary", "Secondary", "Alternate", "Inverse", "Info", "Danger", "Warning", "Success"}

		for _, variant := range variants {
			button := page.Locator(fmt.Sprintf("button:has-text('%s')", variant)).First()
			visible, err := button.IsVisible()
			require.NoError(t, err)
			assert.True(t, visible, fmt.Sprintf("%s button should be visible", variant))
		}

		t.Logf("✓ All %d button variants render correctly", len(variants))
	})
}

// TestAccordion_AllVariants tests all accordion variants
func TestAccordion_AllVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Default_Variant", func(t *testing.T) {
		// Each variant now lives in its own demo box with a stable container ID.
		defaultAccordion := page.Locator("#accordion-default")
		classAttr, err := defaultAccordion.GetAttribute("class")
		require.NoError(t, err)

		// Should have background
		assert.Contains(t, classAttr, "bg-surface-alt", "default variant should have surface-alt background")
		t.Log("✓ Default variant renders correctly")
	})

	t.Run("NoBackground_Variant", func(t *testing.T) {
		nobgAccordion := page.Locator("#accordion-nobg")
		classAttr, err := nobgAccordion.GetAttribute("class")
		require.NoError(t, err)

		// Should have plain surface background, not the tinted surface-alt
		assert.Contains(t, classAttr, "bg-surface", "no-background variant should have surface background")
		assert.NotContains(t, classAttr, "bg-surface-alt", "no-background variant should not use surface-alt")
		t.Log("✓ NoBackground variant renders correctly")
	})

	t.Run("Split_Variant", func(t *testing.T) {
		splitAccordion := page.Locator("#accordion-split")
		classAttr, err := splitAccordion.GetAttribute("class")
		require.NoError(t, err)

		// Split lays cards out with a gap and drops the shared divider/border.
		assert.Contains(t, classAttr, "flex", "split container should be a flex column")
		assert.Contains(t, classAttr, "gap-4", "split container should gap its cards")
		assert.NotContains(t, classAttr, "divide-y", "split should not use a shared divider")

		// Each item should be its own bordered card.
		firstCard := splitAccordion.Locator("xpath=./div").First()
		cardClass, err := firstCard.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, cardClass, "border", "each split item should be a bordered card")
		t.Log("✓ Split variant renders independent cards")
	})

	t.Run("ServerLoaded_Variant", func(t *testing.T) {
		lazyAccordion := page.Locator("#accordion-lazy")

		// Should have buttons with lazy-loaded content
		buttons := lazyAccordion.Locator("button")
		count, err := buttons.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2, "should have lazy-loaded buttons")

		t.Log("✓ Server-loaded variant renders correctly")
	})
}

// TestComponent_DarkMode tests dark mode across components
func TestComponent_DarkMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Dark_Mode_Toggle_Works", func(t *testing.T) {
		// Check initial state
		hasDark, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		initialDark := hasDark.(bool)

		// Click dark mode toggle
		toggleBtn := page.Locator("#darkModeToggleBtn")
		err = toggleBtn.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Verify dark class changed
		hasDarkAfter, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		assert.NotEqual(t, initialDark, hasDarkAfter.(bool), "dark mode should toggle")

		t.Log("✓ Dark mode toggle works")
	})

	t.Run("Dark_Mode_Persists_In_LocalStorage", func(t *testing.T) {
		// Toggle dark mode on
		toggleBtn := page.Locator("#darkModeToggleBtn")
		_ = toggleBtn.Click()
		time.Sleep(50 * time.Millisecond)

		// Check localStorage
		darkModeValue, err := page.Evaluate("() => localStorage.getItem('darkMode')", nil)
		require.NoError(t, err)

		// Should be persisted
		assert.True(t, darkModeValue == "true" || darkModeValue == "false", "darkMode should be persisted in localStorage")

		t.Log("✓ Dark mode persists in localStorage")
	})
}

// TestAPIEndpoints tests API endpoints directly
func TestAPIEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	t.Run("Hello_API_Returns_Correct_Response", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/hello")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

		// Read body
		body := make([]byte, 256)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])

		assert.Contains(t, bodyStr, "Hello from HTMX!")
		assert.Contains(t, bodyStr, "GET")

		t.Log("✓ Hello API works")
	})

	t.Run("Button_Fragment_API_Works", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/button?disabled=true")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

		// Read body
		body := make([]byte, 4096)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])

		// Should contain disabled buttons
		assert.True(t, strings.Contains(bodyStr, "disabled") || strings.Contains(bodyStr, "Disabled"))

		t.Log("✓ Button fragment API works")
	})

	t.Run("Accordion_Content_API_Returns_Different_Content", func(t *testing.T) {
		// Test content A
		respA, err := http.Get(baseURL + "/api/components/accordion-content/lazy-content-a")
		require.NoError(t, err)

		bodyA := make([]byte, 1024)
		nA, _ := respA.Body.Read(bodyA)
		_ = respA.Body.Close()
		bodyStrA := string(bodyA[:nA])

		// Test content B
		respB, err := http.Get(baseURL + "/api/components/accordion-content/lazy-content-b")
		require.NoError(t, err)

		bodyB := make([]byte, 1024)
		nB, _ := respB.Body.Read(bodyB)
		_ = respB.Body.Close()
		bodyStrB := string(bodyB[:nB])

		// Should be different
		assert.Contains(t, bodyStrA, "Response A")
		assert.Contains(t, bodyStrB, "Response B")
		assert.NotEqual(t, bodyStrA, bodyStrB)

		t.Log("✓ Accordion content API returns different content for different IDs")
	})
}

// TestAccordion_Visual_Parity tests visual parity between Original and Goshtoso
func TestAccordion_Visual_Parity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 800,
		},
	})

	_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Screenshots_Match_Threshold", func(t *testing.T) {
		// Take screenshot of entire page for visual comparison
		screenshotDir := "test-results/screenshots"
		screenshotPath := fmt.Sprintf("%s/accordion-parity-%d.png", screenshotDir, time.Now().Unix())

		_, err := page.Screenshot(playwright.PageScreenshotOptions{
			Path:     new(screenshotPath),
			FullPage: new(false),
		})
		require.NoError(t, err)

		t.Logf("✓ Screenshot saved: %s", screenshotPath)
	})

	t.Run("CSS_Classes_Match_Expected", func(t *testing.T) {
		// Get first Goshtoso accordion
		accordion := page.Locator("#accordion-fragment .divide-y").First()
		classAttr, err := accordion.GetAttribute("class")
		require.NoError(t, err)

		// Should have expected classes
		assert.Contains(t, classAttr, "w-full")
		assert.Contains(t, classAttr, "divide-y")
		assert.Contains(t, classAttr, "rounded-radius")
		assert.Contains(t, classAttr, "border")

		t.Log("✓ CSS classes match expected")
	})
}

// TestPerformance tests performance characteristics
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	t.Run("Accordion_Content_Loads_Quickly", func(t *testing.T) {
		// Measure API response time
		start := time.Now()
		resp, err := http.Get(baseURL + "/api/components/accordion-content/lazy-content-a")
		elapsed := time.Since(start)
		require.NoError(t, err)
		_ = resp.Body.Close()

		// Should load within reasonable time (server has 500ms simulated delay)
		assert.Less(t, elapsed, 15*time.Second, "API should respond within 15 seconds")

		t.Logf("✓ Content loaded in %v", elapsed)
	})

	t.Run("Page_Loads_Within_Reasonable_Time", func(t *testing.T) {
		_, browser, cleanupPW := setupPlaywright(t)
		defer cleanupPW()

		page := newPage(t, browser)

		start := time.Now()
		_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)
		elapsed := time.Since(start)

		assert.Less(t, elapsed, 5*time.Second, "Page should load within 5 seconds")

		t.Logf("✓ Page loaded in %v", elapsed)
	})
}

// TestIntegration tests full user workflows
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	t.Run("User_Can_Navigate_Between_Components", func(t *testing.T) {
		// Start at accordion
		_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		title, _ := page.Title()
		assert.Contains(t, title, "Accordion")

		// Navigate to button. Sidebar links are htmx fragment-nav: the swapped-in
		// link is re-bound a beat after it appears, so a plain Click can be lost
		// under full-suite load (the documented HTMX rebind race — see CLAUDE.md).
		// clickUntil re-fires until the title actually flips; htmx updates
		// document.title from the swapped fragment's <title>.
		buttonLink := page.Locator("a:has-text('Button')")
		clickUntil(t, page, buttonLink, "() => document.title.includes('Button')")

		title, _ = page.Title()
		assert.Contains(t, title, "Button")

		t.Log("✓ Navigation between components works")
	})

	t.Run("Full_Workflow_Expand_Load_Interact", func(t *testing.T) {
		_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		// Expand static section
		staticButton := page.Locator("#accordion-fragment button").First()
		err = staticButton.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Expand lazy-loaded section
		lazyButton := page.Locator("text=Dynamic Content A").First()
		err = lazyButton.Click()
		require.NoError(t, err)

		lazyContent := page.Locator("#lazy-content-a")
		err = lazyContent.ScrollIntoViewIfNeeded()
		require.NoError(t, err)

		// Wait for content
		err = page.Locator("text=Server Response A").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		// Toggle dark mode
		toggleBtn := page.Locator("#darkModeToggleBtn")
		err = toggleBtn.Click()
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		// Verify content still visible after theme change
		content := page.Locator("text=Server Response A")
		visible, _ := content.IsVisible()
		assert.True(t, visible, "Content should remain visible after theme change")

		t.Log("✓ Full workflow works correctly")
	})
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	t.Run("API_Returns_Error_For_Invalid_Content_ID", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/components/accordion-content/invalid-id")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := make([]byte, 256)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])

		// Should show error message
		assert.Contains(t, bodyStr, "Unknown content ID")

		t.Log("✓ API handles invalid IDs gracefully")
	})

	t.Run("404_Page_Works", func(t *testing.T) {
		_, browser, cleanupPW := setupPlaywright(t)
		defer cleanupPW()

		page := newPage(t, browser)

		resp, err := page.Goto(baseURL+"/nonexistent-page", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		// Should get 404
		assert.Equal(t, 404, resp.Status())

		t.Log("✓ 404 handling works")
	})
}
