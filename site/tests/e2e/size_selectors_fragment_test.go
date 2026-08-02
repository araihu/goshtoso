//go:build e2e && full

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSizeSelectors_FragmentNavigationClicksKeepContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{document.cookie='gt_storage=allowed; Path=/; SameSite=Lax';localStorage.setItem('theme','minimal')}catch(e){}"),
	}))

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	cases := []struct {
		href      string
		fragment  string
		selector  string
		optionFor string
		selected  string
		preview   string
		heading   string
	}{
		{"/components/badge", "#badge-fragment", "badge", "badge-size-lg", "lg", "badge-size-preview-lg", "Badge"},
		{"/components/button", "#button-docs-fragment", "button", "button-size-xl", "xl", "button-size-preview-xl", "Button"},
		{"/components/kbd", "#kbd-fragment", "kbd", "kbd-size-lg", "lg", "kbd-size-preview-lg", "KBD"},
		{"/components/radio", "#radio-fragment", "radio", "radio-size-selector-xl", "xl", "radio-size-preview-xl", "Radio"},
		{"/components/rating", "#rating-fragment", "rating", "rating-size-xl", "xl", "rating-size-preview-xl", "Rating"},
		{"/components/spinner", "#spinner-fragment", "spinner", "spinner-size-xl", "xl", "spinner-size-preview-xl", "Spinner"},
	}

	for _, tc := range cases {
		t.Run(tc.selector, func(t *testing.T) {
			_, err := page.Goto(baseURL+"/getting-started", playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)
			require.NoError(t, waitForAlpine(page))

			link := page.Locator("a[href='" + tc.href + "']").First()
			require.NoError(t, link.ScrollIntoViewIfNeeded())
			require.NoError(t, link.Click())
			require.NoError(t, page.Locator(tc.fragment).WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateAttached,
			}))
			_, err = page.Evaluate(`() => {
				window.__sizeSelectorHtmxEvents = [];
				if (!window.__sizeSelectorHtmxProbeInstalled) {
					window.__sizeSelectorHtmxProbeInstalled = true;
					[
						'htmx:beforeRequest',
						'htmx:configRequest',
						'htmx:beforeSwap',
						'htmx:afterSwap',
						'htmx:responseError',
					].forEach((name) => {
						document.body.addEventListener(name, (event) => {
							const detail = event.detail || {};
							const pathInfo = detail.pathInfo || {};
							window.__sizeSelectorHtmxEvents.push({
								name,
								eventTarget: event.target && event.target.id,
								detailTarget: detail.target && detail.target.id,
								detailElement: detail.elt && detail.elt.id,
								path: pathInfo.requestPath || '',
							});
						});
					});
				}
			}`, nil)
			require.NoError(t, err)

			require.NoError(t, page.Locator("label[for='"+tc.optionFor+"']").Click())
			require.NoError(t, page.Locator("[data-testid='"+tc.selector+"-size-selected']").Filter(playwright.LocatorFilterOptions{
				HasText: tc.selected,
			}).WaitFor())
			require.NoError(t, page.Locator("[data-testid='"+tc.preview+"']").WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}))
			require.NoError(t, page.Locator("#main-content").Filter(playwright.LocatorFilterOptions{
				HasText: tc.heading,
			}).WaitFor())
			outerScrollStable, err := page.Evaluate("() => window.scrollY <= 1", nil)
			require.NoError(t, err)
			require.True(t, outerScrollStable.(bool), "size selector click should not meaningfully scroll the outer document")
			events, err := page.Evaluate("() => window.__sizeSelectorHtmxEvents", nil)
			require.NoError(t, err)
			require.Empty(t, events, "size selector click should not trigger HTMX")
		})
	}

	require.Empty(t, pageErrors, "no uncaught JS exceptions after fragment selector clicks: %v", pageErrors)
	require.Empty(t, filterIgnorable(consoleErrors), "no console errors after fragment selector clicks: %s", strings.Join(consoleErrors, "; "))
}

func TestSizeSelector_DirectBadgeMinimalClickKeepsContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{document.cookie='gt_storage=allowed; Path=/; SameSite=Lax';localStorage.setItem('theme','minimal')}catch(e){}"),
	}))

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/badge", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	require.NoError(t, page.Locator("#sizes").ScrollIntoViewIfNeeded())
	outerScrollStable, err := page.Evaluate("() => window.scrollY <= 1", nil)
	require.NoError(t, err)
	require.True(t, outerScrollStable.(bool), "scrolling to the size section should stay inside the app scroller")
	require.NoError(t, page.Locator("label[for='badge-size-lg']").Click())
	require.NoError(t, page.Locator("[data-testid='badge-size-selected']").Filter(playwright.LocatorFilterOptions{
		HasText: "lg",
	}).WaitFor())
	require.NoError(t, page.Locator("[data-testid='badge-size-preview-lg']").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, page.Locator("#main-content").Filter(playwright.LocatorFilterOptions{
		HasText: "Badge",
	}).WaitFor())
	outerScrollStable, err = page.Evaluate("() => window.scrollY <= 1", nil)
	require.NoError(t, err)
	require.True(t, outerScrollStable.(bool), "size selector click should not meaningfully scroll the outer document")

	require.Empty(t, pageErrors, "no uncaught JS exceptions after direct size click: %v", pageErrors)
	require.Empty(t, filterIgnorable(consoleErrors), "no console errors after direct size click: %s", strings.Join(consoleErrors, "; "))
}
