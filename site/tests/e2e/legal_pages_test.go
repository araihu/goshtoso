//go:build e2e && full

package e2e

import (
	"testing"

	"github.com/araihu/goshtoso/assets"
	"github.com/mxschmitt/playwright-go"
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

func TestBrowserStorageNoticeOffersChoice(t *testing.T) {
	page := newIsolatedPage(t)

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{
		Name: "Browser storage",
	}).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Allow browser storage",
	}).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Use without storage",
	}).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, page.GetByText("Some examples use cookies and IndexedDB to persist local demo state.").WaitFor(
		playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
}

func TestBrowserStorageNoticeCanDisableDemoCookies(t *testing.T) {
	page := newIsolatedPage(t)

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Use without storage",
	}).Click())

	pref, err := page.Evaluate(`() => document.cookie`, nil)
	require.NoError(t, err)
	require.Contains(t, pref, "gt_storage=denied")

	_, err = page.Goto(baseURL + "/examples/todo")
	require.NoError(t, err)
	cookies, err := page.Evaluate(`() => document.cookie`, nil)
	require.NoError(t, err)
	require.NotContains(t, cookies, "gt_todo=")
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

func TestAttributionsDisplaysCanonicalRuntimePinsAndEmbeddedPaths(t *testing.T) {
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)

	_, err := page.Goto(baseURL + "/attributions")
	require.NoError(t, err)

	for _, text := range []string{
		"CDN-first",
		"assets.DefaultRuntimeManifest()",
		assets.AlpineVersion(),
		assets.AlpineJSURL,
		assets.HTMXVersion(),
		assets.HTMXURL,
	} {
		require.NoError(t, page.GetByText(text).First().WaitFor(
			playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}), text)
	}
}

func TestPrivacyPageExplainsBrowserStorageChoices(t *testing.T) {
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)

	_, err := page.Goto(baseURL + "/privacy")
	require.NoError(t, err)

	for _, text := range []string{
		"Browser storage",
		"localStorage",
		"cookies such as gt_todo",
		"IndexedDB",
		"Use without storage",
	} {
		require.NoError(t, page.GetByText(text).First().WaitFor(
			playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}), text)
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
