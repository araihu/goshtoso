package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// legalPages are the non-component "specials" added alongside the global footer:
// each is reachable directly and via the footer's legal nav.
var legalPages = []struct {
	path  string
	href  string
	title string
}{
	{"/attributions", "/attributions", "Attributions"},
	{"/license", "/license", "License"},
	{"/privacy", "/privacy", "Privacy Policy"},
}

// dismissCookieBanner pre-sets the consent flag so the fixed banner can't
// intercept footer/link clicks.
func dismissCookieBanner(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{localStorage.setItem('cookieConsent','v1')}catch(e){}"),
	}))
}

// TestLegalPages_DirectLoad asserts each page loads, shows its heading and the
// global footer, with no console/page errors.
func TestLegalPages_DirectLoad(t *testing.T) {
	for _, p := range legalPages {
		t.Run(p.title, func(t *testing.T) {
			page := newIsolatedPage(t)
			dismissCookieBanner(t, page)

			var jsErrors []string
			page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
			page.On("console", func(m playwright.ConsoleMessage) {
				if m.Type() == "error" {
					jsErrors = append(jsErrors, m.Text())
				}
			})

			_, err := page.Goto(baseURL + p.path)
			require.NoError(t, err)

			require.NoError(t, page.Locator("h1", playwright.PageLocatorOptions{
				HasText: p.title,
			}).First().WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}))

			// Footer present with its three legal links.
			require.NoError(t, page.Locator("footer a[href='/attributions']").WaitFor(
				playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
			for _, l := range []string{"/attributions", "/license", "/privacy"} {
				c, err := page.Locator("footer a[href='" + l + "']").Count()
				require.NoError(t, err)
				require.Equal(t, 1, c, "footer should link to %s", l)
			}

			require.Empty(t, jsErrors, "no JS console/page errors on %s: %v", p.path, jsErrors)
		})
	}
}

// TestLegalPages_FooterNav lands on Getting Started and navigates to each legal
// page via the footer link (htmx fragment swap into #main-content), asserting
// the content swaps in AND the footer re-renders (it lives in the swapped region).
func TestLegalPages_FooterNav(t *testing.T) {
	for _, p := range legalPages {
		t.Run(p.title, func(t *testing.T) {
			page := newIsolatedPage(t)
			dismissCookieBanner(t, page)

			var jsErrors []string
			page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
			page.On("console", func(m playwright.ConsoleMessage) {
				if m.Type() == "error" {
					jsErrors = append(jsErrors, m.Text())
				}
			})

			_, err := page.Goto(baseURL + "/getting-started")
			require.NoError(t, err)
			_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
			require.NoError(t, err)

			require.NoError(t, page.Locator("footer a[href='"+p.href+"']").Click())

			// Content swapped: the page heading appears.
			require.NoError(t, page.Locator("#main-content h1", playwright.PageLocatorOptions{
				HasText: p.title,
			}).First().WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}))
			// Footer survived the swap (re-rendered by the Fragment).
			require.NoError(t, page.Locator("footer a[href='/license']").WaitFor(
				playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))

			require.Empty(t, jsErrors, "no JS console/page errors navigating to %s: %v", p.path, jsErrors)
		})
	}
}
