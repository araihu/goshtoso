//go:build e2e && (full || spinner)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinnerComponentDemoVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/spinner", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#spinner-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	t.Run("default spinner is decorative with stable SVG attributes", func(t *testing.T) {
		defaultSpinner := page.Locator("#spinner-default svg.motion-safe\\:animate-spin").First()
		require.NoError(t, defaultSpinner.WaitFor())
		assert.Equal(t, "true", mustAttribute(t, defaultSpinner, "aria-hidden"))
		assert.Equal(t, "0 0 24 24", mustAttribute(t, defaultSpinner, "viewBox"))
	})

	t.Run("color variants render one spinner per semantic color", func(t *testing.T) {
		variants := []struct {
			name  string
			class string
		}{
			{"primary", "fill-primary"},
			{"secondary", "fill-secondary"},
			{"info", "fill-info"},
			{"success", "fill-success"},
			{"warning", "fill-warning"},
			{"danger", "fill-danger"},
		}

		for _, variant := range variants {
			locator := page.Locator("#spinner-variants svg." + variant.class)
			count, err := locator.Count()
			require.NoError(t, err)
			assert.Equal(t, 1, count, "expected one %s spinner", variant.name)
		}
	})

	t.Run("size variants render each documented size", func(t *testing.T) {
		sizes := []struct {
			label string
			class string
		}{
			{"sm", "size-4"},
			{"md", "size-5"},
			{"lg", "size-8"},
			{"xl", "size-12"},
		}
		for _, size := range sizes {
			require.NoError(t, page.Locator("label[for='spinner-size-"+size.label+"']").Click())
			require.NoError(t, page.Locator("[data-testid='spinner-size-selected']").Filter(playwright.LocatorFilterOptions{
				HasText: size.label,
			}).WaitFor())
			require.NoError(t, page.Locator("[data-testid='spinner-size-preview-"+size.label+"']").WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}))
			locator := page.Locator("[data-testid='spinner-size-preview-" + size.label + "'] svg." + size.class + ".motion-safe\\:animate-spin")
			count, err := locator.Count()
			require.NoError(t, err)
			assert.Equal(t, 1, count, "expected one visible spinner with %s", size.class)
		}
	})
}
