package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLinkCoverageDemo exercises live browser behavior for the link demo:
// every variant container renders, links are keyboard-focusable, the button
// link keeps its themed focus-visible outline, and the page stays console
// clean across the interactions.
func TestLinkCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/link", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#link-fragment").WaitFor())

	t.Run("all variant containers render", func(t *testing.T) {
		for _, id := range []string{"#link-default", "#link-inline", "#link-icon", "#link-button"} {
			loc := page.Locator(id)
			require.NoError(t, loc.WaitFor())
			anchor := loc.Locator("a").First()
			require.NoError(t, anchor.WaitFor())
			visible, err := anchor.IsVisible()
			require.NoError(t, err)
			assert.True(t, visible, "anchor in %s should be visible", id)
		}
	})

	t.Run("default link is keyboard focusable", func(t *testing.T) {
		defaultLink := page.Locator("#link-default a").First()
		require.NoError(t, defaultLink.Focus())
		focused, err := defaultLink.Evaluate("el => el === document.activeElement", nil)
		require.NoError(t, err)
		assert.Equal(t, true, focused)
	})

	t.Run("button link exposes themed focus outline", func(t *testing.T) {
		buttonLink := page.Locator("#link-button a").First()
		className := mustAttribute(t, buttonLink, "class")
		assert.Contains(t, className, "focus-visible:outline-primary")
		assert.Contains(t, className, "active:opacity-100")

		require.NoError(t, buttonLink.Focus())
		focused, err := buttonLink.Evaluate("el => el === document.activeElement", nil)
		require.NoError(t, err)
		assert.Equal(t, true, focused)
	})

	t.Run("inline link stays within paragraph flow", func(t *testing.T) {
		inlineAnchor := page.Locator("#link-inline p a").First()
		require.NoError(t, inlineAnchor.WaitFor())
		href := mustAttribute(t, inlineAnchor, "href")
		assert.Equal(t, "#", href)
	})

	assert.Empty(t, consoleErrors, "demo page should not log console errors")
}
