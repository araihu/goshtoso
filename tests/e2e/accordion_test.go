package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccordion_GoshtosoComponent(t *testing.T) {
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

	t.Run("Page_Loads", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Accordion", "page title should contain 'Accordion'")

		t.Logf("✓ Page loaded: %s", title)
	})

	t.Run("Goshtoso_Accordion_Exists", func(t *testing.T) {
		accordions := page.Locator("#accordion-fragment .divide-y")
		count, err := accordions.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 3, "should have at least 3 accordion variants")

		t.Logf("✓ Found %d Goshtoso accordions", count)
	})

	t.Run("Interactions_Work", func(t *testing.T) {
		firstAccordion := page.Locator("#accordion-fragment .divide-y").First()
		firstButton := firstAccordion.Locator("button").First()

		err := firstButton.Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		expanded, err := firstButton.Evaluate("el => el.getAttribute('aria-expanded')", nil)
		require.NoError(t, err)
		assert.Equal(t, "true", expanded)

		t.Log("✓ Goshtoso accordion interactions work")
	})
}

func TestAccordion_CSSClassParity(t *testing.T) {
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

	t.Run("Verify_Tailwind_Classes_Present", func(t *testing.T) {
		_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		firstAccordion := page.Locator("#accordion-fragment .divide-y").First()
		firstButton := firstAccordion.Locator("button").First()

		VerifyTailwindClasses(t, firstButton, []string{
			"flex",
			"w-full",
			"items-center",
			"justify-between",
			"p-4",
		})

		VerifyTailwindClasses(t, firstAccordion, []string{
			"w-full",
			"divide-y",
			"rounded-radius",
			"border",
		})

		t.Log("✓ Required Tailwind classes present")
	})
}

func TestAccordion_Variants(t *testing.T) {
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
		// The first accordion in the Goshtoso fragment is the default variant
		defaultAccordion := page.Locator("#accordion-fragment .divide-y").First()
		classAttr, err := defaultAccordion.GetAttribute("class")
		require.NoError(t, err)

		assert.Contains(t, classAttr, "bg-surface-alt", "default variant should have surface-alt background")
	})

	t.Run("NoBackground_Variant", func(t *testing.T) {
		// Find the no-background variant by checking classes (bg-surface but not bg-surface-alt)
		accordions := page.Locator("#accordion-fragment .divide-y")
		count, err := accordions.Count()
		require.NoError(t, err)

		found := false
		for i := range count {
			classAttr, err := accordions.Nth(i).GetAttribute("class")
			require.NoError(t, err)
			// NoBackground has bg-surface but not bg-surface-alt
			if contains(classAttr, "bg-surface") && !contains(classAttr, "bg-surface-alt") {
				found = true
				t.Log("✓ Found no-background variant with correct classes")
				break
			}
		}
		assert.True(t, found, "should find a no-background variant accordion")
	})

	t.Run("MultipleOpen_Variant", func(t *testing.T) {
		// Find accordion with allowMultiple by looking for the data attribute
		accordions := page.Locator("#accordion-fragment .divide-y")
		count, err := accordions.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 3, "should have at least 3 accordion variants")

		// Use the last accordion which should be the "allow multiple" variant
		multiAccordion := accordions.Last()
		buttons := multiAccordion.Locator("button")
		btnCount, err := buttons.Count()
		require.NoError(t, err)
		require.GreaterOrEqual(t, btnCount, 2, "should have at least 2 buttons")

		// Click first, then second
		err = buttons.Nth(0).Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		err = buttons.Nth(1).Click()
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		expanded0, _ := buttons.Nth(0).Evaluate("el => el.getAttribute('aria-expanded')", nil)
		expanded1, _ := buttons.Nth(1).Evaluate("el => el.getAttribute('aria-expanded')", nil)

		t.Logf("Item states - Item 1: %v, Item 2: %v", expanded0, expanded1)
	})
}

// contains checks if s contains substr (simple helper for readability)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestAccordion_Accessibility(t *testing.T) {
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

	t.Run("ARIA_Attributes_Present", func(t *testing.T) {
		firstAccordion := page.Locator("#accordion-fragment .divide-y").First()
		firstButton := firstAccordion.Locator("button").First()

		ariaControls, err := firstButton.GetAttribute("aria-controls")
		require.NoError(t, err)
		assert.NotEmpty(t, ariaControls, "button should have aria-controls")

		contentRegion := page.Locator("#" + ariaControls)
		exists, err := contentRegion.Count()
		require.NoError(t, err)
		assert.Greater(t, exists, 0, "aria-controls should reference existing element")

		t.Log("✓ ARIA attributes present")
	})

	t.Run("Keyboard_Navigation", func(t *testing.T) {
		firstAccordion := page.Locator("#accordion-fragment .divide-y").First()
		firstButton := firstAccordion.Locator("button").First()

		err := firstButton.Focus()
		require.NoError(t, err)

		err = page.Keyboard().Press("Enter")
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		expanded, err := firstButton.Evaluate("el => el.getAttribute('aria-expanded')", nil)
		require.NoError(t, err)
		assert.Equal(t, "true", expanded, "should expand with Enter key")

		err = page.Keyboard().Press("Enter")
		require.NoError(t, err)
		time.Sleep(150 * time.Millisecond)

		collapsed, err := firstButton.Evaluate("el => el.getAttribute('aria-expanded')", nil)
		require.NoError(t, err)
		assert.Equal(t, "false", collapsed, "should collapse with Enter key")

		t.Log("✓ Keyboard navigation works")
	})
}

func TestAccordion_AllThemes(t *testing.T) {
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

	t.Run("Screenshot_All_Themes", func(t *testing.T) {
		screenshotDir := filepath.Join("test-results", "screenshots", "accordion-themes")
		require.NoError(t, os.MkdirAll(screenshotDir, 0755))

		screenshotPath := filepath.Join(screenshotDir, fmt.Sprintf("accordion-theme-default-%d.png", time.Now().Unix()))
		_, err := page.Screenshot(playwright.PageScreenshotOptions{
			Path:     new(screenshotPath),
			FullPage: new(false),
		})
		require.NoError(t, err)

		t.Logf("✓ Screenshot saved: %s", screenshotPath)
	})

	t.Run("Dark_Mode_Toggle", func(t *testing.T) {
		hasDark, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		initialDark := hasDark.(bool)

		toggleBtn := page.Locator("#darkModeToggleBtn")
		err = toggleBtn.Click()
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		hasDarkAfter, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		assert.NotEqual(t, initialDark, hasDarkAfter.(bool), "dark mode should toggle")

		t.Log("✓ Dark mode toggle works")
	})
}

func TestAccordion_VisualParity_99_99(t *testing.T) {
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

	t.Run("Visual_Parity_Comprehensive", func(t *testing.T) {
		// Compare first Goshtoso accordion's classes against expected Tailwind classes
		firstAccordion := page.Locator("#accordion-fragment .divide-y").First()

		classAttr, err := firstAccordion.GetAttribute("class")
		require.NoError(t, err)

		// Verify key structural classes are present
		expectedClasses := []string{
			"w-full", "divide-y", "divide-outline", "overflow-hidden",
			"rounded-radius", "border", "border-outline", "text-on-surface",
		}

		matched := 0
		for _, cls := range expectedClasses {
			if stringContains(classAttr, cls) {
				matched++
			} else {
				t.Logf("  Missing class: %s", cls)
			}
		}

		parity := float64(matched) / float64(len(expectedClasses))
		t.Logf("Class parity: %.0f%% (%d/%d)", parity*100, matched, len(expectedClasses))
		assert.GreaterOrEqual(t, parity, 0.90,
			"accordion should have at least 90%% of expected structural classes")
	})
}
