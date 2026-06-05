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
		sizes := []string{"size-4", "size-5", "size-8", "size-12"}
		for _, size := range sizes {
			locator := page.Locator("#spinner-sizes svg." + size + ".motion-safe\\:animate-spin")
			count, err := locator.Count()
			require.NoError(t, err)
			assert.Equal(t, 1, count, "expected one spinner with %s", size)
		}
	})
}
