//go:build e2e && full

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestLanding_HeroAndStructure loads the homepage ("/") and asserts the
// redesigned hero — brand heading + primary CTA — is present. The page is a
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
	playground := page.FrameLocator("#theme-playground-frame")
	charts := page.FrameLocator("#charts-showcase-frame-line-3d")

	t.Run("HeroHeadline", func(t *testing.T) {
		h1 := page.Locator("#hero h1")
		txt, err := h1.InnerText()
		require.NoError(t, err)
		require.Equal(t, "Build server-rendered Go interfaces with typed components.", txt)
		body, err := page.Locator("body").InnerText()
		require.NoError(t, err)
		require.NotContains(t, body, "Build interactive UIs in Go")
	})

	t.Run("ProductSpecificFirstViewport", func(t *testing.T) {
		preview := page.Locator("#hero [data-hero-preview]")
		visible, err := preview.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "hero should show a real Goshtoso component surface")

		text, err := preview.InnerText()
		require.NoError(t, err)
		require.Contains(t, text, "Ready to ship")
		require.Contains(t, text, "templ DeployPanel()")
		require.Contains(t, text, "Typed props in, accessible HTML out")
		require.Contains(t, text, "Deploy")
	})

	t.Run("HomepageNavigation", func(t *testing.T) {
		nav := page.Locator("#hero nav[aria-label='Primary navigation']")
		visible, err := nav.IsVisible()
		require.NoError(t, err)
		require.True(t, visible)
		count, err := nav.Locator("a[href='/getting-started']").Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)
		github := nav.Locator("a[aria-label='GitHub repository'][href='https://github.com/araihu/goshtoso']")
		visible, err = github.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "GitHub repository action should be visible in the topbar")
		count, err = github.Locator("svg[aria-hidden='true']").Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)
		count, err = page.Locator("#hero-content a[href='https://github.com/araihu/goshtoso']").Count()
		require.NoError(t, err)
		require.Zero(t, count, "GitHub should not be duplicated in hero actions")
	})

	t.Run("MainLandmarkAndSkipLink", func(t *testing.T) {
		main := page.Locator("main#main-content")
		count, err := main.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "standalone homepage should expose one main landmark")

		skip := page.Locator("a[href='#hero-content']")
		count, err = skip.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "keyboard users should be able to skip the homepage navigation")
		require.NoError(t, skip.Focus())
		require.NoError(t, skip.Press("Enter"))
		activeID, err := page.Evaluate("() => document.activeElement && document.activeElement.id", nil)
		require.NoError(t, err)
		require.Equal(t, "hero-content", activeID, "skip link should move keyboard focus beyond the site navigation")
	})

	t.Run("HomepageHasNoRemoteFontDependency", func(t *testing.T) {
		count, err := page.Locator("link[href*='fonts.googleapis.com'], link[href*='fonts.gstatic.com']").Count()
		require.NoError(t, err)
		require.Zero(t, count, "homepage typography should work without a third-party font request")
	})

	t.Run("GettingStartedCTA", func(t *testing.T) {
		cta := page.Locator("#hero a[data-primary-cta][href='/getting-started']")
		visible, err := cta.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "getting started CTA should be visible")
	})

	t.Run("InstallCommandCopiesPasteReadyCommand", func(t *testing.T) {
		_, err := page.Evaluate(`() => {
			window.__landingCopied = null;
			Object.defineProperty(navigator, 'clipboard', {
				configurable: true,
				value: {
					writeText: (text) => {
						window.__landingCopied = text;
						return Promise.resolve();
					},
				},
			});
		}`, nil)
		require.NoError(t, err)

		copyButton := page.Locator(`#hero-content button[aria-label='Copy Install Goshtoso code']`)
		count, err := copyButton.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.NoError(t, copyButton.Click())

		copied, err := page.Evaluate(`() => window.__landingCopied`, nil)
		require.NoError(t, err)
		require.Equal(t, "go get github.com/araihu/goshtoso@latest", copied)
		require.NoError(t, copyButton.Locator("text=Copied!").WaitFor())
	})

	t.Run("HeroUsesV11GoshtosoLockup", func(t *testing.T) {
		lockup := page.Locator("#hero [data-brand-lockup]")
		visible, err := lockup.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "homepage should render the Goshtoso V11 brand lockup")

		src, err := lockup.Locator("img").GetAttribute("src")
		require.NoError(t, err)
		require.Contains(t, src, "goshtoso-logo.svg", "homepage should use the canonical V11 logo asset")

		txt, err := lockup.InnerText()
		require.NoError(t, err)
		require.Contains(t, txt, "Server-rendered Go UI")
	})

	t.Run("BrandFollowsDarkMode", func(t *testing.T) {
		hadDark, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		_, err = page.Evaluate("() => document.documentElement.classList.remove('dark')", nil)
		require.NoError(t, err)
		colorScheme, err := page.Evaluate("() => getComputedStyle(document.documentElement).colorScheme", nil)
		require.NoError(t, err)
		require.Equal(t, "light", colorScheme)

		_, err = page.Evaluate("() => document.documentElement.classList.add('dark')", nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, cleanupErr := page.Evaluate("dark => document.documentElement.classList.toggle('dark', dark)", hadDark)
			require.NoError(t, cleanupErr)
		})
		colorScheme, err = page.Evaluate("() => getComputedStyle(document.documentElement).colorScheme", nil)
		require.NoError(t, err)
		require.Equal(t, "dark", colorScheme, "v11 brand SVGs inherit the .dark color scheme")
	})

	t.Run("DefaultsToAraiHuTheme", func(t *testing.T) {
		got, err := page.Evaluate("() => document.documentElement.getAttribute('data-theme')", nil)
		require.NoError(t, err)
		require.Equal(t, "araihu", got, "homepage should default to the Arai Hû theme")
	})

	t.Run("AraiHuThemeSegmentAvailable", func(t *testing.T) {
		segment := playground.Locator("#home-theme-picker label:has(input[data-theme-key='araihu'])")
		visible, err := segment.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "Arai Hû should be available in the homepage theme picker")
	})

	t.Run("PlaygroundCopyStaysFocused", func(t *testing.T) {
		body, err := page.Locator("body").InnerText()
		require.NoError(t, err)
		require.NotContains(t, body, "Every control below is a real Goshtoso component")

		frameBody, err := playground.Locator("body").InnerText()
		require.NoError(t, err)
		require.Contains(t, frameBody, "Live theme preview")
		require.NotContains(t, frameBody, "Customer workspace")
		require.NotContains(t, frameBody, "Explore all")
	})

	t.Run("HowItWorksIncludesStaticSites", func(t *testing.T) {
		text, err := page.Locator("#how-it-works").InnerText()
		require.NoError(t, err)
		require.Contains(t, text, "Publish static sites")
		require.Contains(t, text, "serve them from any static host")
	})

	t.Run("ExtensionsExposeCharts", func(t *testing.T) {
		extensions := page.Locator("#extensions")
		text, err := extensions.InnerText()
		require.NoError(t, err)
		require.Contains(t, text, "Goshtoso Charts")
		require.NotContains(t, text, "Goshtoso App Shells")
		require.NotContains(t, text, "only with the preview below")
		require.NoError(t, page.Locator("#charts-showcase-frame-line-3d").WaitFor())
		require.Zero(t, mustCount(t, page.Locator("#charts-showcase-loader")), "HTMX load should replace the placeholder before scrolling")
		require.Equal(t, 1, mustCount(t, extensions.Locator(`a[href="https://github.com/araihu/goshtoso-charts"]`)))
		require.Equal(t, 1, mustCount(t, extensions.Locator(`a[href="https://charts.goshtoso.araihu.com"]`)))
		require.Zero(t, mustCount(t, extensions.Locator(`a[href="/modules/app-shells"]`)))
	})

	t.Run("ChartsShowcaseRendersAndFits", func(t *testing.T) {
		require.NoError(t, page.Locator("#charts-showcase-frame-line-3d").WaitFor())
		require.Equal(t, "eager", mustAttribute(t, page.Locator("#charts-showcase-frame-line-3d"), "loading"))
		require.Equal(t, "no", mustAttribute(t, page.Locator("#charts-showcase-frame-line-3d"), "scrolling"))
		require.NoError(t, charts.Locator("canvas").First().WaitFor())
		require.Equal(t, 1, mustCount(t, charts.Locator("[data-showcase-chart]")))
		require.Equal(t, 1, mustCount(t, charts.Locator(`[data-showcase-component="line-3d"]`)))
		require.Zero(t, mustCount(t, charts.Locator("[data-chart-carousel-slide]")))
		require.GreaterOrEqual(t, mustCount(t, charts.Locator("canvas")), 1)
		require.Equal(t, "Charts for every use case", mustText(t, charts.Locator("#charts-showcase-title")))
		require.Zero(t, mustCount(t, charts.Locator("figcaption")))
		require.Zero(t, mustCount(t, charts.GetByText("Typed Go configs, local runtime, theme-aware output.")))
		hidden, err := charts.Locator("[data-chart-fallback]").IsHidden()
		require.NoError(t, err)
		require.True(t, hidden, "successful chart rendering should hide fallback")
		fits, err := charts.Locator("html").Evaluate(`el => el.scrollWidth <= el.clientWidth && el.scrollHeight <= el.clientHeight`, nil)
		require.NoError(t, err)
		require.Equal(t, true, fits, "chart iframe should not create nested scrollbars")
		animations := charts.Locator(`[data-goshtoso-charts-explicit-animation="false"]`)
		require.Equal(t, 1, mustCount(t, animations), "showcase chart must disable initial animation")
	})

	t.Run("PlaygroundFitsWithoutNestedScroll", func(t *testing.T) {
		_, err := page.WaitForFunction(`() => {
			const frame = document.querySelector('#theme-playground-frame');
			if (!frame || !frame.contentDocument || !frame.contentDocument.body) return false;
			const contentHeight = Math.max(
				frame.contentDocument.documentElement.scrollHeight,
				frame.contentDocument.body.scrollHeight,
			);
			return frame.clientHeight >= contentHeight &&
				getComputedStyle(frame.contentDocument.body).overflowY === 'hidden';
		}`, nil)
		require.NoError(t, err, "playground iframe should grow to its content without nested scrolling")
	})

	t.Run("GoshtosoThemeSegmentAvailable", func(t *testing.T) {
		segment := playground.Locator("#home-theme-picker label:has(input[data-theme-key='goshtoso'])")
		visible, err := segment.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "Goshtoso should be available in the homepage theme picker")
	})

	t.Run("ThemeSegmentSwitchesTheme", func(t *testing.T) {
		parentTheme, err := page.Evaluate("() => document.documentElement.getAttribute('data-theme')", nil)
		require.NoError(t, err)

		require.NoError(t, playground.Locator("#home-theme-picker label:has(input[data-theme-key='dracula'])").Click())
		got, err := playground.Locator("html").GetAttribute("data-theme")
		require.NoError(t, err)
		require.Equal(t, "dracula", got, "clicking a segment should set data-theme on <html>")
		checked, err := playground.Locator("#home-theme-picker input[data-theme-key='dracula']").IsChecked()
		require.NoError(t, err)
		require.True(t, checked, "selected theme segment should be checked")

		parentThemeAfter, err := page.Evaluate("() => document.documentElement.getAttribute('data-theme')", nil)
		require.NoError(t, err)
		require.Equal(t, parentTheme, parentThemeAfter, "playground theme must not restyle homepage")

		stored, err := page.Evaluate("() => localStorage.getItem('theme')", nil)
		require.NoError(t, err)
		require.Nil(t, stored, "playground theme must not persist into homepage storage")

		require.NoError(t, page.Locator("button", playwright.PageLocatorOptions{HasText: "Allow browser storage"}).Click())
		require.NoError(t, playground.Locator("#home-theme-picker label:has(input[data-theme-key='minimal'])").Click())
		stored, err = page.Evaluate("() => localStorage.getItem('theme')", nil)
		require.NoError(t, err)
		require.Nil(t, stored, "playground must remain non-persistent after storage consent")
	})

	t.Run("LiveTableLoadsRows", func(t *testing.T) {
		// The table is intentionally lazy. Bring it into the viewport, then prove
		// the HTMX response replaced the loading placeholder with real cells.
		require.NoError(t, playground.Locator("#home-table").ScrollIntoViewIfNeeded())
		require.NoError(t, playground.Locator("#home-table tbody").Locator("text=Sarah Adams").WaitFor(), "live HTMX table should populate rows")
		text, err := playground.Locator("#home-table tbody").InnerText()
		require.NoError(t, err)
		require.NotContains(t, text, "Loading...")
		require.Contains(t, text, "Sarah Adams")
	})

	t.Run("PlaygroundOmitsServerEndpointCopy", func(t *testing.T) {
		proofCount, err := playground.Locator("#playground-server-proof").Count()
		require.NoError(t, err)
		require.Zero(t, proofCount)

		text, err := playground.Locator("body").InnerText()
		require.NoError(t, err)
		require.NotContains(t, text, "Table rows are lazy-loaded from Go through HTMX.")
		require.NotContains(t, text, "/api/components/table/rows")
	})

	t.Run("ProductGalleryLinksToLiveAraiHuSites", func(t *testing.T) {
		for _, productURL := range []string{
			"https://manja.araihu.com/",
			"https://x9.araihu.com/",
			"https://paje.araihu.com/en/",
			"https://araihu.com/",
		} {
			product := page.Locator("#examples a[data-product-card][href='" + productURL + "']")
			count, err := product.Count()
			require.NoError(t, err)
			require.Equal(t, 1, count, "gallery should link once to "+productURL)
			require.Equal(t, "_blank", mustAttribute(t, product, "target"))
			require.Contains(t, mustAttribute(t, product, "rel"), "noopener")
		}
		require.Zero(t, mustCount(t, page.Locator("#examples a[href^='/examples/']")), "homepage gallery should not repeat built-in demos")
		require.Zero(t, mustCount(t, page.Locator(`#examples a:has-text("Visit Arai Hû")`)), "the gallery heading should not repeat the Arai Hû product link")
		require.Equal(t, 3, mustCount(t, page.Locator(`#examples >> text="Work in progress"`)))
		examplesText, err := page.Locator("#examples").InnerText()
		require.NoError(t, err)
		require.Contains(t, examplesText, "agent-driven code changes")
		require.NotContains(t, examplesText, "agent-piloted")
	})

	t.Run("ProductGalleryUsesCapturedSiteImages", func(t *testing.T) {
		cards := page.Locator("#examples a[data-product-card]")
		count, err := cards.Count()
		require.NoError(t, err)
		require.Equal(t, 4, count, "gallery should render one card per Arai Hû product site")

		images := page.Locator("#examples a[data-product-card] img")
		imageCount, err := images.Count()
		require.NoError(t, err)
		require.Equal(t, count, imageCount)
		for i := range imageCount {
			img := images.Nth(i)
			require.Contains(t, mustAttribute(t, img, "src"), "/assets/images/homepage/products/")
			require.NotEmpty(t, mustAttribute(t, img, "alt"))
		}

		actions := page.Locator("#examples a[data-product-card] >> text=Visit site")
		actionCount, err := actions.Count()
		require.NoError(t, err)
		require.Equal(t, 4, actionCount)
	})

	t.Run("MobileKeepsHowItWorksBeforeProductGallery", func(t *testing.T) {
		mobile := newPage(t, browser, playwright.BrowserNewPageOptions{
			Viewport: &playwright.Size{Width: 390, Height: 844},
		})
		_, err := mobile.Goto(baseURL+"/", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		how := mobile.Locator("#how-it-works")
		count, err := how.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count, "homepage should expose a targetable How it works section")

		howBox, err := how.BoundingBox()
		require.NoError(t, err)
		product := mobile.Locator("#examples a[data-product-card]").First()
		productBox, err := product.BoundingBox()
		require.NoError(t, err)
		require.Less(t, howBox.Y, productBox.Y, "mobile should teach the mental model before the product gallery")
		require.LessOrEqual(t, productBox.Width, 350.0, "product cards should fit the mobile content width")
	})

	t.Run("SmallScreensCenterProductGrid", func(t *testing.T) {
		small := newPage(t, browser, playwright.BrowserNewPageOptions{
			Viewport: &playwright.Size{Width: 576, Height: 779},
		})
		_, err := small.Goto(baseURL+"/", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		require.NoError(t, err)

		examplesBox, err := small.Locator("#examples").BoundingBox()
		require.NoError(t, err)
		products := small.Locator("#examples a[data-product-card] article")
		count, err := products.Count()
		require.NoError(t, err)
		require.Equal(t, 4, count)
		firstBox, err := products.Nth(0).BoundingBox()
		require.NoError(t, err)
		secondBox, err := products.Nth(1).BoundingBox()
		require.NoError(t, err)

		examplesCenter := examplesBox.X + examplesBox.Width/2
		firstRowCenter := (firstBox.X + secondBox.X + secondBox.Width) / 2
		require.InDelta(t, examplesCenter, firstRowCenter, 2.0, "product grid should be horizontally centered on small screens")
	})

	t.Run("FooterProductAndOrganization", func(t *testing.T) {
		require.Zero(t, mustCount(t, page.Locator("#stack-strip")), "stack strip should be removed")
		body, err := page.Locator("body").InnerText()
		require.NoError(t, err)
		require.Contains(t, body, "16 themes", "homepage footer should match the theme picker count")
		org := page.Locator("footer a[href='https://araihu.com']")
		visible, err := org.IsVisible()
		require.NoError(t, err)
		require.True(t, visible, "footer should attribute Goshtoso to Arai Hû")
		text, err := org.InnerText()
		require.NoError(t, err)
		require.Equal(t, "Arai Hû", text)
	})
}

func TestChartsShowcase_MobileFits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 390, Height: 592}})
	_, err := page.Goto(baseURL+"/playground/extensions/charts?variant=line-3d", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateCommit})
	require.NoError(t, err)
	require.NoError(t, page.Locator("canvas").First().WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(15000)}))
	require.Zero(t, mustCount(t, page.Locator("[data-chart-carousel-slide]")))
	fits, err := page.Locator("html").Evaluate(`el => el.scrollWidth <= el.clientWidth && el.scrollHeight <= el.clientHeight`, nil)
	require.NoError(t, err)
	require.Equal(t, true, fits, "mobile chart iframe should not create nested scrollbars")
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
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#charts-showcase-frame-line-3d").WaitFor())
	require.NoError(t, page.FrameLocator("#charts-showcase-frame-line-3d").Locator("canvas").First().WaitFor())

	title, err := page.Title()
	require.NoError(t, err)
	require.Contains(t, title, "Goshtoso")

	require.Empty(t, jsErrors, "no JS console/page errors on homepage: %v", jsErrors)
}
