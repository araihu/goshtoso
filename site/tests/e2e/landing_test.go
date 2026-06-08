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

	t.Run("HeroUsesRioWaveLockup", func(t *testing.T) {
		lockup := page.Locator("#hero [data-brand-lockup]")
		visible, err := lockup.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "homepage should render the Rio wave brand lockup")

		src, err := lockup.Locator("img").GetAttribute("src")
		require.NoError(t, err)
		require.Contains(t, src, "goshtoso-icon.svg", "homepage should use the Rio wave icon asset")

		txt, err := lockup.InnerText()
		require.NoError(t, err)
		require.Contains(t, txt, "Goshtoso")
		require.Contains(t, txt, "Go UI components for server-rendered apps")
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
			"/examples/profile", "/examples/ticker", "/examples/expense", "/examples/wizard",
		} {
			loc := page.Locator("#examples a[href='" + route + "']")
			cnt, err := loc.Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, cnt, 1, "gallery should link to "+route)
		}
	})

	t.Run("ExampleGalleryCardsUseImages", func(t *testing.T) {
		cards := page.Locator("#examples a[data-example-card]")
		count, err := cards.Count()
		require.NoError(t, err)
		require.Equal(t, 7, count, "gallery should render one Goshtoso card per example app")

		images := page.Locator("#examples a[data-example-card] img")
		imageCount, err := images.Count()
		require.NoError(t, err)
		require.Equal(t, count, imageCount, "each example card should include a generated image")

		for i := 0; i < imageCount; i++ {
			img := images.Nth(i)
			src, err := img.GetAttribute("src")
			require.NoError(t, err)
			require.Contains(t, src, "/assets/images/homepage/examples/", "example card image should be served from embedded assets")
			alt, err := img.GetAttribute("alt")
			require.NoError(t, err)
			require.NotEmpty(t, alt, "example card image should have alt text")
		}
	})

	t.Run("ExampleGalleryFeaturesWizard", func(t *testing.T) {
		featured := page.Locator("#examples a[data-featured-example-card][href='/examples/wizard']")
		visible, err := featured.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "wizard should be the featured example card")

		supporting := page.Locator("#examples a[data-supporting-example-card]")
		count, err := supporting.Count()
		require.NoError(t, err)
		require.Equal(t, 6, count, "remaining examples should render as supporting cards")

		actions := page.Locator("#examples a[data-example-card] >> text=View app")
		actionCount, err := actions.Count()
		require.NoError(t, err)
		require.Equal(t, 7, actionCount, "each example card should expose an explicit View app affordance")
	})

	t.Run("FeaturedExampleIsCompactLeadStrip", func(t *testing.T) {
		featured := page.Locator("#examples a[data-featured-example-card]")
		supporting := page.Locator("#examples a[data-supporting-example-card]").First()

		featuredBox, err := featured.BoundingBox()
		require.NoError(t, err)
		supportingBox, err := supporting.BoundingBox()
		require.NoError(t, err)

		require.Greater(t, featuredBox.Width, supportingBox.Width*2, "featured example should read as a compact lead strip above the grid")
		require.Less(t, featuredBox.Height, supportingBox.Height*0.9, "featured example should not become a massive billboard card")
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
