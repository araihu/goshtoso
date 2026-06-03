package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestLanding_HeroAndStructure loads the homepage ("/") and asserts the
// redesigned hero — headline + primary CTA — is present. The page is a
// standalone document (not the demo shell).
func TestLanding_HeroAndStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("HeroHeadline", func(t *testing.T) {
		h1 := page.Locator("#hero h1")
		txt, err := h1.InnerText()
		require.NoError(t, err)
		require.Contains(t, txt, "Build interactive UIs in Go")
	})

	t.Run("BrowseComponentsCTA", func(t *testing.T) {
		cta := page.Locator("#hero a[href='/components/accordion']")
		visible, err := cta.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "Browse components CTA should be visible")
	})

	t.Run("DefaultsToGoshtosoTheme", func(t *testing.T) {
		got, err := page.Evaluate("() => document.documentElement.getAttribute('data-theme')", nil)
		require.NoError(t, err)
		require.Equal(t, "goshtoso", got, "homepage should default to the Goshtoso theme")
	})

	t.Run("GoshtosoThemeSegmentAvailable", func(t *testing.T) {
		segment := page.Locator("#home-theme-picker label:has(input[data-theme-key='goshtoso'])")
		visible, err := segment.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "Goshtoso should be available in the homepage theme picker")
	})

	t.Run("ThemeSegmentSwitchesTheme", func(t *testing.T) {
		_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
		require.NoError(t, err)
		// click the Dracula segment
		require.NoError(t, page.Locator("#home-theme-picker label:has(input[data-theme-key='dracula'])").Click())
		got, err := page.Evaluate("() => document.documentElement.getAttribute('data-theme')", nil)
		require.NoError(t, err)
		require.Equal(t, "dracula", got, "clicking a segment should set data-theme on <html>")
		checked, err := page.Locator("#home-theme-picker input[data-theme-key='dracula']").IsChecked()
		require.NoError(t, err)
		require.True(t, checked, "selected theme segment should be checked")
	})

	t.Run("LiveTableLoadsRows", func(t *testing.T) {
		// lazy table fetches rows via HTMX on load
		_, err := page.WaitForFunction(
			"() => document.querySelectorAll('#home-table tbody tr').length > 0", nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)},
		)
		require.NoError(t, err, "live HTMX table should populate rows")
	})

	t.Run("ExampleGalleryLinks", func(t *testing.T) {
		for _, route := range []string{
			"/examples/todo", "/examples/chat", "/examples/logs",
			"/examples/profile", "/examples/ticker",
		} {
			loc := page.Locator("#examples a[href='" + route + "']")
			cnt, err := loc.Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, cnt, 1, "gallery should link to "+route)
		}
	})

	t.Run("StackStripCondensed", func(t *testing.T) {
		strip := page.Locator("#stack-strip")
		visible, err := strip.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "condensed stack strip should be present")
		// it is a single strip, not six cards: at most a handful of links
		links := page.Locator("#stack-strip a")
		cnt, err := links.Count()
		require.NoError(t, err)
		require.LessOrEqual(t, cnt, 6, "stack strip should be condensed, not a card grid")
	})

	t.Run("FooterThemeCount", func(t *testing.T) {
		body, err := page.Locator("body").InnerText()
		require.NoError(t, err)
		require.Contains(t, body, "16 themes", "homepage footer should match the theme picker count")
	})
}

// TestLanding_NoConsoleErrors loads the homepage and asserts no JS console or
// page errors — the silent-Alpine-failure guard (broken x-data fails quietly).
func TestLanding_NoConsoleErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	title, err := page.Title()
	require.NoError(t, err)
	require.Contains(t, title, "Goshtoso")

	require.Empty(t, jsErrors, "no JS console/page errors on homepage: %v", jsErrors)
}
