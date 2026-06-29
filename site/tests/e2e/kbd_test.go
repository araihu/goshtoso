package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKbdComponentDemoVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/kbd", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#kbd-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	t.Run("frequently used and inline keys render semantic kbd elements", func(t *testing.T) {
		frequentlyUsed := page.Locator("#kbd-frequently-used")
		for _, key := range []string{"Tab", "Shift", "Space", "Ctrl", "Command", "Alt", "Enter", "Esc", "Caps Lock"} {
			require.NoError(t, frequentlyUsed.GetByText(key, playwright.LocatorGetByTextOptions{
				Exact: playwright.Bool(true),
			}).WaitFor())
		}

		inline := page.Locator("#kbd-inline")
		require.NoError(t, inline.Locator("kbd").Filter(playwright.LocatorFilterOptions{HasText: "Tab"}).WaitFor())
		require.NoError(t, inline.Locator("kbd").Filter(playwright.LocatorFilterOptions{HasText: "Space"}).WaitFor())
		inlineText, err := inline.Locator("p").TextContent()
		require.NoError(t, err)
		assert.Equal(t, "To Tab or to Space, that is the question.", inlineText)
	})

	t.Run("icon keys expose accessible labels and hide glyphs", func(t *testing.T) {
		for _, label := range []string{"Command", "Shift", "Option", "Control", "Tab", "Up", "Down", "Left", "Right"} {
			key := page.Locator("#kbd-icons kbd[aria-label='" + label + "']")
			require.NoError(t, key.WaitFor())
			require.NoError(t, key.Locator("span[aria-hidden='true'] svg").WaitFor())
			require.NoError(t, key.Locator(".sr-only").Filter(playwright.LocatorFilterOptions{HasText: label}).WaitFor())
		}
	})

	t.Run("alphabet number and function sections render complete key sets", func(t *testing.T) {
		alphabetCount, err := page.Locator("#kbd-alphabet kbd").Count()
		require.NoError(t, err)
		assert.Equal(t, 26, alphabetCount)
		alphabet := page.Locator("#kbd-alphabet")
		require.NoError(t, alphabet.GetByText("A", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())
		require.NoError(t, alphabet.GetByText("Z", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())

		numberCount, err := page.Locator("#kbd-numbers kbd").Count()
		require.NoError(t, err)
		assert.Equal(t, 10, numberCount)
		require.NoError(t, page.Locator("#kbd-numbers").GetByText("0", playwright.LocatorGetByTextOptions{Exact: playwright.Bool(true)}).WaitFor())

		functions := page.Locator("#kbd-functions")
		for _, key := range []string{"F1", "F8", "F12"} {
			require.NoError(t, functions.GetByText(key, playwright.LocatorGetByTextOptions{
				Exact: playwright.Bool(true),
			}).WaitFor())
		}
	})

	t.Run("size variants keep distinct sizing hooks", func(t *testing.T) {
		sizes := []struct {
			label string
			class string
		}{
			{"xs", "min-h-5"},
			{"sm", "min-h-6"},
			{"md", "min-h-7"},
			{"lg", "min-h-9"},
		}
		for _, size := range sizes {
			require.NoError(t, page.Locator("label[for='kbd-size-"+size.label+"']").Click())
			require.NoError(t, page.Locator("[data-testid='kbd-size-selected']").Filter(playwright.LocatorFilterOptions{
				HasText: size.label,
			}).WaitFor())
			require.NoError(t, page.Locator("[data-testid='kbd-size-preview-"+size.label+"']").WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}))
			count, err := page.Locator("[data-testid='kbd-size-preview-" + size.label + "'] kbd." + size.class).Count()
			require.NoError(t, err)
			assert.Equal(t, 1, count, "expected one visible %s key", size.label)
		}
	})
}
