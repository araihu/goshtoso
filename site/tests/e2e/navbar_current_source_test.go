//go:build e2e && full && goshtoso_current_source

package e2e

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/navbar"
	"github.com/araihu/goshtoso/components/popover"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

const (
	navbarOverviewURL = "/components/navbar?view=overview"
	navbarDetailsURL  = "/components/navbar?view=details"
	navbarRowEndpoint = "/api/components/navbar/secondary?view="
)

type navbarCurrentSourceFixture struct {
	server         *httptest.Server
	nativeRequests atomic.Int64
	rowRequests    atomic.Int64
}

func newNavbarPage(t *testing.T, options playwright.BrowserNewPageOptions) playwright.Page {
	t.Helper()
	return newPage(t, sharedBrowser, options)
}

func navbarPageFailures(t *testing.T, page playwright.Page) *pageFailures {
	t.Helper()
	failures := watchPageFailures(page)
	t.Cleanup(func() { failures.RequireEmpty(t) })
	return failures
}

func TestNavbar_CurrentSourceSecondaryRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	fixture := newNavbarCurrentSourceFixture(t)
	focusPage := newNavbarPage(t, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1280, Height: 800},
	})
	failures := navbarPageFailures(t, focusPage)

	t.Run("native fallback and row-only outerHTML swap", func(t *testing.T) {
		assertNavbarNativeFallback(t, fixture)
		assertNavbarHTMXSwap(t, fixture, focusPage)
	})
	t.Run("history cache hit and miss", func(t *testing.T) {
		assertNavbarHistoryCacheHit(t, fixture)
		assertNavbarHistoryCacheMiss(t, fixture)
	})
	t.Run("unrelated history preserves outside focus", func(t *testing.T) {
		assertNavbarUnrelatedHistoryFocus(t, fixture)
	})
	t.Run("visual and interaction matrix", func(t *testing.T) {
		assertNavbarVisualMatrix(t, fixture)
	})

	waitForPageSettled(t, focusPage)
	failures.RequireEmpty(t)
}

func newNavbarCurrentSourceFixture(t *testing.T) *navbarCurrentSourceFixture {
	t.Helper()
	fixture := &navbarCurrentSourceFixture{}
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /components/navbar", func(writer http.ResponseWriter, request *http.Request) {
		fixture.nativeRequests.Add(1)
		view := navbarView(request)
		scrollable := request.URL.Query().Get("scrollable") == "true"
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(writer, navbarCurrentSourcePage(t, view, scrollable))
	})
	mux.HandleFunc("GET /api/components/navbar/secondary", func(writer http.ResponseWriter, request *http.Request) {
		fixture.rowRequests.Add(1)
		view := navbarView(request)
		if view != "overview" && view != "details" {
			http.Error(writer, "unknown view", http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(writer, renderComponentFragment(t, navbar.SecondaryRow(navbarSecondaryConfig(view, false))))
	})
	mux.HandleFunc("GET /api/navbar/unrelated", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(writer, `<span id="unrelated-result">Unrelated content loaded</span>`)
	})
	fixture.server = httptest.NewServer(mux)
	t.Cleanup(fixture.server.Close)

	response, err := http.Get(fixture.server.URL + navbarRowEndpoint + "details")
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, http.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), `id="navbar-secondary-row"`)
	require.NotContains(t, string(body), "<html")
	require.Equal(t, int64(1), fixture.rowRequests.Load(), "probe must use the row-only endpoint")
	return fixture
}

func navbarView(request *http.Request) string {
	if request.URL.Query().Get("view") == "details" {
		return "details"
	}
	return "overview"
}

func navbarCurrentSourcePage(t *testing.T, view string, scrollable bool) string {
	t.Helper()
	metadata := head.Metadata(head.MetadataConfig{
		Title:        "Navbar secondary row - Goshtoso",
		Description:  "Current-source HTMX proof for Goshtoso secondary navigation.",
		CanonicalURL: "https://goshtoso.araihu.com/components/navbar?view=" + view,
		SiteName:     "Goshtoso",
		Image: head.SocialImage{
			URL:      "https://goshtoso.araihu.com/assets/images/goshtoso-social-card.png",
			MIMEType: "image/png",
			Width:    1200,
			Height:   630,
			Alt:      "Goshtoso Go UI component library preview",
		},
		TwitterCard: head.TwitterCardSummaryLargeImage,
	})
	header := renderComponentFragment(t, metadata)
	dependencies := renderComponentFragment(t, head.Dependencies(head.WithLocalRuntime()))
	nav := renderComponentFragment(t, navbar.Navbar(navbar.Config{
		Brand:     templ.Raw(`<span class="font-bold">Goshtoso</span>`),
		BrandHref: "/",
		Links:     []navbar.NavLink{{Label: "Home", Href: "/"}},
		NavAttrs:  templ.Attributes{"id": "primary-navbar"},
		Secondary: ptr(navbarSecondaryConfig(view, scrollable)),
	}))
	return fmt.Sprintf(`<!doctype html>
<html lang="en" data-theme="goshtoso">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">%s<link rel="stylesheet" href="/assets/styles.css">%s<style>#navbar-secondary-actions [data-popover-panel]{min-width:0}</style></head>
<body class="min-h-[1200px] bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark">
<main class="mx-auto w-full max-w-[1280px] overflow-x-hidden">
<button id="outside-focus" type="button" class="m-2 min-h-11 min-w-11 px-3 py-2" hx-get="/api/navbar/unrelated" hx-target="#unrelated-content" hx-swap="innerHTML" hx-push-url="/other">Outside focus</button>
<button id="navbar-popover-dismiss-outside" type="button" class="m-2 min-h-11 min-w-11 px-3 py-2">Dismiss popover</button>
<div id="navbar-visual-spacer" aria-hidden="true" style="height:256px;pointer-events:none"></div>
<div id="navbar-secondary-host">%s</div>
<div id="unrelated-content" class="p-4">Initial unrelated content</div>
</main>
<script>
(() => {
window.__navbarHistoryEvents = window.__navbarHistoryEvents || [];
window.__navbarFocusByPath = window.__navbarFocusByPath || {};
const historyKey = (path) => {
  try {
    const url = new URL(path || window.location.href, window.location.origin);
    return url.pathname + url.search;
  } catch {
    return window.location.pathname + window.location.search;
  }
};
document.body.addEventListener("focusin", (event) => {
  const target = event.target;
  if (target?.id) window.__navbarFocusByPath[historyKey(window.location.href)] = target.id;
});
const focusCurrent = () => requestAnimationFrame(() => {
  const row = document.getElementById("navbar-secondary-row");
  row?.querySelector('a[aria-current="page"], a[aria-current="location"]')?.focus();
});
const restoreUnrelatedFocus = () => requestAnimationFrame(() => requestAnimationFrame(() => document.getElementById(window.__navbarFocusByPath[historyKey(window.location.href)])?.focus({preventScroll: true})));
document.body.addEventListener("htmx:afterSettle", (event) => {
  if (event.detail?.target?.id === "navbar-secondary-row") focusCurrent();
  else restoreUnrelatedFocus();
  const active = document.activeElement;
  if (active?.id) window.__navbarFocusByPath[historyKey(window.location.href)] = active.id;
});
const isSecondaryHistoryPath = (path) => {
  try {
    const url = new URL(path, window.location.origin);
    return url.pathname === "/components/navbar" &&
      ["overview", "details"].includes(url.searchParams.get("view"));
  } catch {
    return false;
  }
};
document.body.addEventListener("htmx:historyRestore", (event) => {
  const path = event.detail?.path || "";
  window.__navbarHistoryEvents.push({
    path,
    cacheMiss: event.detail?.cacheMiss === true,
  });
  if (isSecondaryHistoryPath(path)) {
    focusCurrent();
  } else {
    restoreUnrelatedFocus();
  }
});
})();
</script>
</body></html>`, header, dependencies, nav)
}

func navbarSecondaryConfig(view string, scrollable bool) navbar.SecondaryConfig {
	leadingLabel := "Context"
	var leadingAttrs templ.Attributes
	if scrollable {
		leadingLabel = "Menu"
		leadingAttrs = templ.Attributes{"class": "mr-2"}
	}
	return navbar.SecondaryConfig{
		Links: []navbar.SecondaryLink{
			{Label: leadingLabel, Href: "#" + strings.ToLower(leadingLabel), LinkAttrs: leadingAttrs},
			{Label: "Overview", Href: navbarOverviewURL, Current: currentForView(view, "overview"), LinkAttrs: navbarHTMXAttrs("overview")},
			{Label: "Details", Href: navbarDetailsURL, Current: currentForView(view, "details"), LinkAttrs: navbarHTMXAttrs("details")},
			{Label: "A long secondary navigation label for overflow", Href: "/components/navbar?view=long", LinkAttrs: navbarHTMXAttrs("overview")},
			{Label: "More", Href: "/components/navbar?view=more", LinkAttrs: navbarHTMXAttrs("overview")},
		},
		Actions: []templ.Component{popover.Popover(popover.Config{
			ID:        "navbar-secondary-actions",
			Label:     "Actions",
			Placement: popover.PlacementBottomEnd,
			Role:      "menu",
			Trigger:   templ.Raw(`<button type="button" aria-haspopup="menu" class="inline-flex min-h-11 min-w-11 items-center gap-2 whitespace-nowrap px-3 py-2">Actions</button>`),
			Content:   templ.Raw(`<div class="flex flex-col py-1.5"><a href="#open-action" role="menuitem" tabindex="0" class="inline-flex min-h-11 min-w-11 items-center whitespace-nowrap px-4 py-2 text-sm">Open action</a></div>`),
		})},
		AriaLabel:  "secondary navigation",
		Scrollable: scrollable,
		RootClass:  "",
		RootAttrs:  templ.Attributes{"id": "navbar-secondary-row"},
	}
}

func currentForView(view, candidate string) navbar.SecondaryCurrent {
	if view == candidate {
		if candidate == "details" {
			return navbar.SecondaryCurrentLocation
		}
		return navbar.SecondaryCurrentPage
	}
	return navbar.SecondaryCurrentNone
}

func navbarHTMXAttrs(view string) templ.Attributes {
	nativeURL := navbarOverviewURL
	if view == "details" {
		nativeURL = navbarDetailsURL
	}
	return templ.Attributes{
		"hx-get":      navbarRowEndpoint + view,
		"hx-target":   "#navbar-secondary-row",
		"hx-swap":     "outerHTML",
		"hx-push-url": nativeURL,
	}
}

func ptr[T any](value T) *T { return &value }

func assertNavbarNativeFallback(t *testing.T, fixture *navbarCurrentSourceFixture) {
	t.Helper()
	context, err := sharedBrowser.NewContext(playwright.BrowserNewContextOptions{
		JavaScriptEnabled: playwright.Bool(false),
		Viewport:          &playwright.Size{Width: 1280, Height: 800},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = context.Close() })
	page, err := context.NewPage()
	require.NoError(t, err)
	t.Cleanup(func() { _ = page.Close() })
	navbarPageFailures(t, page)

	for _, test := range []struct {
		path  string
		label string
	}{
		{path: navbarOverviewURL, label: "Overview"},
		{path: navbarDetailsURL, label: "Details"},
	} {
		_, err := page.Goto(fixture.server.URL+test.path, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		require.NoError(t, err)
		assertNavbarCurrentState(t, page, test.label)
	}
}

func assertNavbarHTMXSwap(t *testing.T, fixture *navbarCurrentSourceFixture, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(fixture.server.URL+navbarOverviewURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	primaryBefore, err := page.Locator("#primary-navbar").Evaluate("element => element.outerHTML", nil)
	require.NoError(t, err)
	assertNavbarCurrentState(t, page, "Overview")

	clickNavbarView(t, page, "details")
	assertNavbarCurrentState(t, page, "Details")
	require.Equal(t, primaryBefore, mustEvaluateString(t, page.Locator("#primary-navbar"), "element => element.outerHTML"), "secondary swap must not replace primary Navbar")
	require.Equal(t, 1, mustCount(t, page.Locator("#navbar-secondary-row")))
	require.Equal(t, 1, mustCount(t, page.Locator("#navbar-secondary-row a[aria-current]")))
	require.Contains(t, mustAttribute(t, page.Locator("#navbar-secondary-row a[aria-current]"), "class"), "border-primary")
	require.Equal(t, "Details", mustText(t, page.Locator("#navbar-secondary-row a[aria-current]")))
	require.Contains(t, mustAttribute(t, page.Locator("#navbar-secondary-row a[aria-current]"), "hx-swap"), "outerHTML")
	assertNavbarActionSeparation(t, page)
	assertNavbarKeyboardInteractions(t, fixture, page)
}

func assertNavbarHistoryCacheHit(t *testing.T, fixture *navbarCurrentSourceFixture) {
	t.Helper()
	page := newNavbarPage(t, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 1280, Height: 800}})
	navbarPageFailures(t, page)
	navigateNavbarFixture(t, page, fixture, navbarOverviewURL)
	clearNavbarHistoryEvents(t, page)
	clickNavbarView(t, page, "details")
	nativeBeforeBack := fixture.nativeRequests.Load()
	rowBeforeBack := fixture.rowRequests.Load()
	goBackNavbar(t, page, navbarOverviewURL, "Overview")
	require.Equal(t, nativeBeforeBack, fixture.nativeRequests.Load(), "cache-hit back must not request the native page")
	require.Equal(t, rowBeforeBack, fixture.rowRequests.Load(), "cache-hit back must not request the row endpoint")
	assertLastNavbarHistoryEvent(t, page, navbarOverviewURL, false)
	goForwardNavbar(t, page, navbarDetailsURL, "Details")
	require.Equal(t, nativeBeforeBack, fixture.nativeRequests.Load(), "cache-hit forward must not request the native page")
	require.Equal(t, rowBeforeBack, fixture.rowRequests.Load(), "cache-hit forward must not request the row endpoint")
	assertLastNavbarHistoryEvent(t, page, navbarDetailsURL, false)
}

func assertNavbarHistoryCacheMiss(t *testing.T, fixture *navbarCurrentSourceFixture) {
	t.Helper()
	page := newNavbarPage(t, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 1280, Height: 800}})
	navbarPageFailures(t, page)
	navigateNavbarFixture(t, page, fixture, navbarOverviewURL)
	_, err := page.Evaluate("() => { htmx.config.historyCacheSize = 0; window.__navbarHistoryEvents = []; }", nil)
	require.NoError(t, err)
	clickNavbarView(t, page, "details")
	nativeBeforeBack := fixture.nativeRequests.Load()
	rowBeforeBack := fixture.rowRequests.Load()
	goBackNavbar(t, page, navbarOverviewURL, "Overview")
	require.Equal(t, nativeBeforeBack+1, fixture.nativeRequests.Load(), "cache-miss back must reload the native page")
	require.Equal(t, rowBeforeBack, fixture.rowRequests.Load(), "cache-miss back must use native history, not the row endpoint")
	assertLastNavbarHistoryEvent(t, page, navbarOverviewURL, true)
	nativeBeforeForward := fixture.nativeRequests.Load()
	rowBeforeForward := fixture.rowRequests.Load()
	goForwardNavbar(t, page, navbarDetailsURL, "Details")
	require.Equal(t, nativeBeforeForward+1, fixture.nativeRequests.Load(), "cache-miss forward must reload the native page")
	require.Equal(t, rowBeforeForward, fixture.rowRequests.Load(), "cache-miss forward must use native history, not the row endpoint")
	assertLastNavbarHistoryEvent(t, page, navbarDetailsURL, true)
	_, err = page.Evaluate("() => { htmx.config.historyCacheSize = 10; }", nil)
	require.NoError(t, err)
}

func assertNavbarUnrelatedHistoryFocus(t *testing.T, fixture *navbarCurrentSourceFixture) {
	t.Helper()
	page := newNavbarPage(t, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 1280, Height: 800}})
	navbarPageFailures(t, page)
	navigateNavbarFixture(t, page, fixture, navbarOverviewURL)
	clearNavbarHistoryEvents(t, page)
	_, err := page.ExpectResponse("**/api/navbar/unrelated", func() error {
		return page.Locator("#outside-focus").Click()
	}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	require.NoError(t, page.WaitForURL("**/other"))
	require.NoError(t, page.Locator("#outside-focus").Focus())
	clickNavbarView(t, page, "details")
	goBackNavbar(t, page, "/other", "")
	_, err = page.WaitForFunction("() => document.activeElement?.id === 'outside-focus'", nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(1500)})
	if err != nil {
		t.Fatalf("unrelated restore must preserve outside focus: actual=%q debug=%s: %v", mustEvaluateString(t, page, "() => document.activeElement?.id"), mustEvaluateString(t, page, "() => JSON.stringify({location: location.href, events: window.__navbarHistoryEvents, focus: window.__navbarFocusByPath, outside: !!document.getElementById('outside-focus')})"), err)
	}
	assertLastNavbarHistoryEvent(t, page, "/other", false)
	goBackNavbar(t, page, navbarOverviewURL, "Overview")
	assertCurrentLinkFocused(t, page)
}

func assertNavbarVisualMatrix(t *testing.T, fixture *navbarCurrentSourceFixture) {
	t.Helper()
	viewports := []struct {
		name   string
		width  int
		height int
		scale2 bool
	}{
		{name: "desktop", width: 1280, height: 800},
		{name: "phone-390", width: 390, height: 844},
		{name: "phone-320", width: 320, height: 800},
		{name: "phone-390-scale-2", width: 390, height: 844, scale2: true},
	}
	for _, viewport := range viewports {
		viewport := viewport
		t.Run(viewport.name, func(t *testing.T) {
			for _, theme := range []struct {
				name  string
				theme string
				dark  bool
			}{
				{name: "goshtoso-light", theme: "goshtoso"}, {name: "goshtoso-dark", theme: "goshtoso", dark: true},
				{name: "minimal-light", theme: "minimal"}, {name: "minimal-dark", theme: "minimal", dark: true},
			} {
				for _, scrollable := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/scrollable=%t", theme.name, scrollable), func(t *testing.T) {
						page := newNavbarPage(t, playwright.BrowserNewPageOptions{
							HasTouch:          playwright.Bool(viewport.scale2),
							Viewport:          &playwright.Size{Width: viewport.width, Height: viewport.height},
							DeviceScaleFactor: floatPtr(boolFloat(viewport.scale2, 2, 1)),
						})
						navbarPageFailures(t, page)
						path := navbarOverviewURL + "&scrollable=" + strconv.FormatBool(scrollable)
						navigateNavbarFixture(t, page, fixture, path)
						if viewport.scale2 {
							emulateNavbarScaleTwo(t, page)
						}
						setNavbarThemeMode(t, page, theme.theme, theme.dark)
						assertNavbarCurrentState(t, page, "Overview")
						assertNavbarVisualGeometry(t, page, scrollable, viewport.scale2)
						if viewport.scale2 {
							contrastPage := newNavbarPage(t, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: viewport.width, Height: viewport.height}})
							navbarPageFailures(t, contrastPage)
							contrastPath := navbarOverviewURL + "&scrollable=" + strconv.FormatBool(scrollable)
							navigateNavbarFixture(t, contrastPage, fixture, contrastPath)
							setNavbarThemeMode(t, contrastPage, theme.theme, theme.dark)
							assertNavbarCurrentState(t, contrastPage, "Overview")
							assertNavbarContrastMatrix(t, contrastPage, theme.dark)
						} else {
							assertNavbarContrastMatrix(t, page, theme.dark)
						}
					})
				}
			}
		})
	}
}

func assertNavbarCurrentState(t *testing.T, page playwright.Page, label string) {
	t.Helper()
	current := page.Locator("#navbar-secondary-row a[aria-current]")
	require.Equal(t, 1, mustCount(t, current))
	require.Equal(t, label, mustText(t, current))
	wantCurrent := "page"
	if label == "Details" {
		wantCurrent = "location"
	}
	require.Equal(t, wantCurrent, mustAttribute(t, current, "aria-current"))
	classes := mustAttribute(t, current, "class")
	require.Contains(t, classes, "border-primary")
	require.Contains(t, classes, "font-semibold")
}

func setNavbarThemeMode(t *testing.T, page playwright.Page, theme string, dark bool) {
	t.Helper()
	_, err := page.Evaluate(`([theme, dark]) => {
		const html = document.documentElement;
		localStorage.setItem('theme', theme);
		html.setAttribute('data-theme', theme);
		html.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
	_, err = page.WaitForFunction("theme => document.documentElement.dataset.theme === theme", theme)
	require.NoError(t, err)
	page.WaitForTimeout(200)
}

func assertNavbarActionSeparation(t *testing.T, page playwright.Page) {
	t.Helper()
	require.Equal(t, 1, mustCount(t, page.Locator("#navbar-secondary-row > nav")))
	require.Equal(t, 1, mustCount(t, page.Locator("#navbar-secondary-row > [data-navbar-actions='true']")))
	require.Equal(t, 0, mustCount(t, page.Locator("#navbar-secondary-row > nav button")))
	actions := page.Locator("#navbar-secondary-row > [data-navbar-actions='true']")
	require.GreaterOrEqual(t, mustCount(t, actions.Locator("button")), 1)
	require.Equal(t, "8px", mustEvaluateString(t, actions, "element => getComputedStyle(element).columnGap"))
}

func assertNavbarKeyboardInteractions(t *testing.T, fixture *navbarCurrentSourceFixture, page playwright.Page) {
	t.Helper()
	navigateNavbarFixture(t, page, fixture, navbarOverviewURL)
	overview := page.Locator("#navbar-secondary-row a[href='" + navbarOverviewURL + "']")
	details := page.Locator("#navbar-secondary-row a[href='" + navbarDetailsURL + "']")
	require.NoError(t, overview.Focus())
	require.NoError(t, page.Keyboard().Press("Tab"))
	require.Equal(t, navbarDetailsURL, mustEvaluateString(t, page, "() => document.activeElement?.getAttribute('href') || ''"), "Tab should move between primitive secondary links")
	require.NoError(t, page.Keyboard().Press("Shift+Tab"))
	require.Equal(t, navbarOverviewURL, mustEvaluateString(t, page, "() => document.activeElement?.getAttribute('href') || ''"), "Shift+Tab should return to the prior primitive link")

	_, err := page.ExpectResponse("**"+navbarRowEndpoint+"details", func() error { return details.Press("Enter") }, playwright.PageExpectResponseOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`href => window.location.pathname + window.location.search === href`, navbarDetailsURL, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	assertNavbarCurrentState(t, page, "Details")
	assertCurrentLinkFocused(t, page)
	navigateNavbarFixture(t, page, fixture, navbarOverviewURL)

	trigger := page.Locator("#navbar-secondary-row [data-navbar-actions='true'] button").First()
	panel := page.Locator("#navbar-secondary-row [role='menu']").First()
	firstItem := panel.Locator("[role='menuitem']").First()
	dismissOutside := page.Locator("#navbar-popover-dismiss-outside")

	// Each keyboard activation starts from a closed popover so the test proves
	// the trigger's own key bindings rather than retaining state from another
	// activation path.
	require.NoError(t, trigger.Focus())
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	require.True(t, mustEvaluateBool(t, firstItem, "element => element === document.activeElement"), "ArrowDown should open and focus the first menu item")
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	require.True(t, mustEvaluateBool(t, trigger, "element => element === document.activeElement"), "Escape should restore trigger focus")

	require.NoError(t, trigger.Focus())
	require.NoError(t, page.Keyboard().Press("Space"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	require.True(t, mustEvaluateBool(t, firstItem, "element => element === document.activeElement"), "Space should focus the first menu item")
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	require.True(t, mustEvaluateBool(t, trigger, "element => element === document.activeElement"), "Space/Escape should restore trigger focus")

	require.NoError(t, trigger.Focus())
	require.NoError(t, page.Keyboard().Press("Enter"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	require.True(t, mustEvaluateBool(t, firstItem, "element => element === document.activeElement"), "Enter should focus the first menu item")
	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	require.True(t, mustEvaluateBool(t, trigger, "element => element === document.activeElement"), "Enter/Escape should restore trigger focus")

	require.NoError(t, trigger.Click())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	require.NoError(t, dismissOutside.Click())
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	require.True(t, mustEvaluateBool(t, dismissOutside, "element => element === document.activeElement"), "click outside should dismiss the popover and retain outside focus")
}

func assertNavbarVisualGeometry(t *testing.T, page playwright.Page, scrollable, scale2 bool) {
	t.Helper()
	require.LessOrEqual(t, mustEvaluateFloat(t, page, "() => document.documentElement.scrollWidth"), mustEvaluateFloat(t, page, "() => document.documentElement.clientWidth"))
	links := page.Locator("#navbar-secondary-row nav a")
	actions := page.Locator("#navbar-secondary-row [data-navbar-actions='true'] [data-popover-trigger] button")
	for _, locator := range []playwright.Locator{links, actions} {
		count := mustCount(t, locator)
		for index := 0; index < count; index++ {
			box := mustBox(t, locator.Nth(index))
			require.GreaterOrEqual(t, box.Width, 44.0)
			require.GreaterOrEqual(t, box.Height, 44.0)
		}
	}
	if scale2 {
		current := page.Locator("#navbar-secondary-row nav a[aria-current]")
		focusLinkWithoutScroll(t, current)
		assertFocusWithinVisualViewport(t, page, current)
		panVisualViewport(t, page, 96, 211)
		focusLinkWithoutScroll(t, current)
		assertFocusWithinVisualViewport(t, page, current)
		panVisualViewport(t, page, 195, 211)
	}
	assertNavbarActionPanelBounds(t, page, scale2)
	if scale2 {
		panVisualViewport(t, page, 0, 0)
	}
	if !scrollable {
		require.Equal(t, "0", mustEvaluateString(t, page, "() => String(document.querySelector('#navbar-secondary-row nav > div')?.scrollLeft || 0)"))
		return
	}
	scrollport := page.Locator("#navbar-secondary-row nav > div")
	first := links.First()
	last := links.Last()
	_, err := scrollport.Evaluate("element => { element.scrollLeft = 0; return element.scrollLeft; }", nil)
	require.NoError(t, err)
	focusKeyboardLink(t, page, first)
	assertFocusClearance(t, first)
	require.InDelta(t, 0, mustEvaluateFloat(t, scrollport, "element => element.scrollLeft"), 0.5)
	_, err = scrollport.Evaluate("element => { element.scrollLeft = element.scrollWidth - element.clientWidth; return element.scrollLeft; }", nil)
	require.NoError(t, err)
	focusKeyboardLink(t, page, last)
	assertFocusClearance(t, last)
	maxScroll := mustEvaluateFloat(t, scrollport, "element => element.scrollWidth - element.clientWidth")
	require.InDelta(t, maxScroll, mustEvaluateFloat(t, scrollport, "element => element.scrollLeft"), 0.5)
}

func assertNavbarContrastMatrix(t *testing.T, page playwright.Page, dark bool) {
	t.Helper()
	link := page.Locator("#navbar-secondary-row nav a:not([aria-current])").First()
	current := page.Locator("#navbar-secondary-row nav a[aria-current]")
	hoverOptions := playwright.LocatorHoverOptions{}
	if mustEvaluateFloat(t, page, "() => visualViewport.scale") > 1.01 {
		hoverOptions.Force = playwright.Bool(true)
	}
	require.GreaterOrEqual(t, measureRenderedContrast(t, link).Ratio, 4.5)
	require.GreaterOrEqual(t, measureRenderedContrast(t, current).Ratio, 4.5)
	require.True(t, mustEvaluateBool(t, link, `element => {
		const color = getComputedStyle(element).borderBottomColor;
		return color === 'transparent' || color === 'rgba(0, 0, 0, 0)';
	}`), "inactive links use a transparent bottom border")
	require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, current, "borderBottomColor").Ratio, 3.0)

	require.NoError(t, link.Hover(hoverOptions))
	page.WaitForTimeout(200)
	require.GreaterOrEqual(t, measureRenderedContrast(t, link).Ratio, 4.5)
	hoverBorder := measureRenderedPropertyContrast(t, link, "borderBottomColor")
	require.GreaterOrEqual(t, hoverBorder.Ratio, 3.0)
	require.NoError(t, current.Hover(hoverOptions))
	page.WaitForTimeout(200)
	require.GreaterOrEqual(t, measureRenderedContrast(t, current).Ratio, 4.5)
	require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, current, "borderBottomColor").Ratio, 3.0)

	focusKeyboardLink(t, page, link)
	assertNavbarFocusContrast(t, page, link, false)
	pressFocusedLocator(t, page, link, func() {
		assertNavbarFocusContrast(t, page, link, true)
	})
	focusKeyboardLink(t, page, current)
	assertNavbarFocusContrast(t, page, current, true)
	pressFocusedLocator(t, page, current, func() {
		assertNavbarFocusContrast(t, page, current, true)
	})
	require.NoError(t, current.Hover(hoverOptions))
	require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, current, "borderBottomColor").Ratio, 3.0)
	require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, current, "outlineColor").Ratio, 3.0)
	err := page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce})
	require.NoError(t, err)
	require.Equal(t, "none", mustEvaluateString(t, current, "element => getComputedStyle(element).transitionProperty"))
	_ = dark
}

func assertNavbarFocusContrast(t *testing.T, page playwright.Page, locator playwright.Locator, borderVisible bool) {
	t.Helper()
	require.True(t, mustEvaluateBool(t, locator, "element => element === document.activeElement"), "focused state must retain DOM focus")
	require.True(t, mustEvaluateBool(t, locator, "element => element.matches(':focus-visible')"), "focused state must use focus-visible")
	require.Equal(t, "solid", mustEvaluateString(t, locator, "element => getComputedStyle(element).outlineStyle"))
	require.Equal(t, "2px", mustEvaluateString(t, locator, "element => getComputedStyle(element).outlineWidth"))
	require.GreaterOrEqual(t, measureRenderedContrast(t, locator).Ratio, 4.5)
	require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, locator, "outlineColor").Ratio, 3.0)
	if borderVisible {
		require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, locator, "borderBottomColor").Ratio, 3.0)
	}
}

func pressFocusedLocator(t *testing.T, page playwright.Page, locator playwright.Locator, during func()) {
	t.Helper()
	require.NoError(t, locator.ScrollIntoViewIfNeeded())
	require.True(t, mustEvaluateBool(t, locator, "element => element === document.activeElement"), "pressed focus state must start with DOM focus")
	box, err := locator.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, box)
	require.NoError(t, page.Mouse().Move(box.X+box.Width/2, box.Y+box.Height/2))
	require.NoError(t, page.Mouse().Down())
	page.WaitForTimeout(200)
	during()
	require.NoError(t, page.Mouse().Move(1, 1))
	require.NoError(t, page.Mouse().Up())
}

func assertNavbarActionPanelBounds(t *testing.T, page playwright.Page, scale2 bool) {
	t.Helper()
	trigger := page.Locator("#navbar-secondary-row [data-navbar-actions='true'] button").First()
	assertElementWithinVisualViewport(t, page, trigger)
	clickOptions := playwright.LocatorClickOptions{}
	if scale2 {
		clickOptions.Force = playwright.Bool(true)
	}
	require.NoError(t, trigger.Click(clickOptions))
	panel := page.Locator("#navbar-secondary-row [role='menu']").First()
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	result := mustEvaluate(t, panel, `(element) => {
		const rect = element.getBoundingClientRect();
		const viewport = window.visualViewport || {offsetLeft: 0, offsetTop: 0, width: document.documentElement.clientWidth, height: innerHeight};
		return {left: rect.left, right: rect.right, top: rect.top, bottom: rect.bottom, viewportLeft: viewport.offsetLeft, viewportRight: viewport.offsetLeft + viewport.width, viewportTop: viewport.offsetTop, viewportBottom: viewport.offsetTop + viewport.height};
	}`)
	values := result.(map[string]any)
	require.GreaterOrEqual(t, numberAsFloat(t, values["left"]), numberAsFloat(t, values["viewportLeft"])-0.5)
	require.LessOrEqual(t, numberAsFloat(t, values["right"]), numberAsFloat(t, values["viewportRight"])+0.5)
	require.GreaterOrEqual(t, numberAsFloat(t, values["top"]), numberAsFloat(t, values["viewportTop"])-0.5)
	require.LessOrEqual(t, numberAsFloat(t, values["bottom"]), numberAsFloat(t, values["viewportBottom"])+0.5)
	if scale2 {
		require.InDelta(t, 2, mustEvaluateFloat(t, page, "() => visualViewport.scale"), 0.01)
	}
	require.NoError(t, page.Keyboard().Press("Escape"))
}

func assertElementWithinVisualViewport(t *testing.T, page playwright.Page, locator playwright.Locator) {
	t.Helper()
	result := mustEvaluate(t, locator, `element => {
		const rect = element.getBoundingClientRect();
		const viewport = window.visualViewport || {offsetLeft: 0, offsetTop: 0, width: document.documentElement.clientWidth, height: innerHeight};
		return {
			left: rect.left,
			right: rect.right,
			top: rect.top,
			bottom: rect.bottom,
			viewportLeft: viewport.offsetLeft,
			viewportRight: viewport.offsetLeft + viewport.width,
			viewportTop: viewport.offsetTop,
			viewportBottom: viewport.offsetTop + viewport.height,
		};
	}`)
	values := result.(map[string]any)
	require.GreaterOrEqual(t, numberAsFloat(t, values["left"]), numberAsFloat(t, values["viewportLeft"]))
	require.LessOrEqual(t, numberAsFloat(t, values["right"]), numberAsFloat(t, values["viewportRight"]))
	require.GreaterOrEqual(t, numberAsFloat(t, values["top"]), numberAsFloat(t, values["viewportTop"]))
	require.LessOrEqual(t, numberAsFloat(t, values["bottom"]), numberAsFloat(t, values["viewportBottom"]))
}

func assertFocusWithinVisualViewport(t *testing.T, page playwright.Page, locator playwright.Locator) {
	t.Helper()
	result := mustEvaluate(t, locator, `element => {
		const rect = element.getBoundingClientRect();
		const style = getComputedStyle(element);
		const extra = parseFloat(style.outlineWidth) + parseFloat(style.outlineOffset);
		const viewport = window.visualViewport || {offsetLeft: 0, offsetTop: 0, width: document.documentElement.clientWidth, height: innerHeight};
		return {
			left: rect.left - extra,
			right: rect.right + extra,
			top: rect.top - extra,
			bottom: rect.bottom + extra,
			viewportLeft: viewport.offsetLeft,
			viewportRight: viewport.offsetLeft + viewport.width,
			viewportTop: viewport.offsetTop,
			viewportBottom: viewport.offsetTop + viewport.height,
		};
	}`)
	values := result.(map[string]any)
	require.GreaterOrEqual(t, numberAsFloat(t, values["left"]), numberAsFloat(t, values["viewportLeft"]))
	require.LessOrEqual(t, numberAsFloat(t, values["right"]), numberAsFloat(t, values["viewportRight"]))
	require.GreaterOrEqual(t, numberAsFloat(t, values["top"]), numberAsFloat(t, values["viewportTop"]))
	require.LessOrEqual(t, numberAsFloat(t, values["bottom"]), numberAsFloat(t, values["viewportBottom"]))
}

func assertFocusClearance(t *testing.T, link playwright.Locator) {
	t.Helper()
	result := mustEvaluate(t, link, `link => {
		const scrollport = document.querySelector('#navbar-secondary-row nav > div');
		const s = scrollport.getBoundingClientRect();
		const f = link.getBoundingClientRect();
		const style = getComputedStyle(link);
		const extra = parseFloat(style.outlineWidth) + parseFloat(style.outlineOffset);
		return {left: f.left - extra, right: f.right + extra, top: f.top - extra, bottom: f.bottom + extra, scrollLeft: s.left, scrollRight: s.right, scrollTop: s.top, scrollBottom: s.bottom};
	}`)
	values := result.(map[string]any)
	require.GreaterOrEqual(t, numberAsFloat(t, values["left"]), numberAsFloat(t, values["scrollLeft"])+4-0.5)
	require.LessOrEqual(t, numberAsFloat(t, values["right"]), numberAsFloat(t, values["scrollRight"])-4+0.5)
	require.GreaterOrEqual(t, numberAsFloat(t, values["top"]), numberAsFloat(t, values["scrollTop"])+4-0.5)
	require.LessOrEqual(t, numberAsFloat(t, values["bottom"]), numberAsFloat(t, values["scrollBottom"])-4+0.5)
}

func navigateNavbarFixture(t *testing.T, page playwright.Page, fixture *navbarCurrentSourceFixture, path string) {
	t.Helper()
	_, err := page.Goto(fixture.server.URL+path, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#navbar-secondary-row").WaitFor())
}

func clickNavbarView(t *testing.T, page playwright.Page, view string) {
	t.Helper()
	href := navbarOverviewURL
	if view == "details" {
		href = navbarDetailsURL
	}
	link := page.Locator("#navbar-secondary-row a[href='" + href + "']")
	_, err := page.ExpectResponse("**"+navbarRowEndpoint+view, func() error { return link.Click() }, playwright.PageExpectResponseOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`href => window.location.pathname + window.location.search === href`, href)
	if err != nil {
		t.Fatalf("secondary-row URL did not settle: expected=%s actual=%s: %v", href, mustEvaluateString(t, page, "() => window.location.pathname + window.location.search"), err)
	}
	_, err = page.WaitForFunction(`label => {
		const row = document.querySelector('#navbar-secondary-row');
		const current = row?.querySelector('a[aria-current]');
		return current?.textContent?.trim() === label;
	}`, func() string {
		if view == "details" {
			return "Details"
		}
		return "Overview"
	}())
	if err != nil {
		state := mustEvaluateString(t, page, "() => document.querySelector('#navbar-secondary-row')?.outerHTML || ''")
		t.Fatalf("secondary-row state did not settle: url=%s state=%s: %v", mustEvaluateString(t, page, "() => location.href"), state, err)
	}
	_, err = page.WaitForFunction(`() => {
		const current = document.querySelector('#navbar-secondary-row a[aria-current]');
		return document.activeElement === current;
	}`, nil)
	if err != nil {
		active := mustEvaluateString(t, page, "() => document.activeElement?.outerHTML || ''")
		t.Fatalf("secondary-row focus restoration failed: active=%s: %v", active, err)
	}
}

func goBackNavbar(t *testing.T, page playwright.Page, path, label string) {
	t.Helper()
	_, err := page.GoBack(playwright.PageGoBackOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NoError(t, page.WaitForURL("**"+path))
	if strings.HasPrefix(path, "/components/navbar") {
		_, err = page.WaitForFunction(`label => document.querySelector('#navbar-secondary-row a[aria-current]')?.textContent?.trim() === label`, label)
		require.NoError(t, err)
		assertNavbarCurrentState(t, page, label)
		assertCurrentLinkFocused(t, page)
	}
}

func goForwardNavbar(t *testing.T, page playwright.Page, path, label string) {
	t.Helper()
	_, err := page.GoForward(playwright.PageGoForwardOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NoError(t, page.WaitForURL("**"+path))
	_, err = page.WaitForFunction(`label => document.querySelector('#navbar-secondary-row a[aria-current]')?.textContent?.trim() === label`, label)
	require.NoError(t, err)
	assertNavbarCurrentState(t, page, label)
	assertCurrentLinkFocused(t, page)
}

func assertCurrentLinkFocused(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.WaitForFunction(`() => {
		const current = document.querySelector('#navbar-secondary-row a[aria-current]');
		return current !== null && document.activeElement === current;
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2500)})
	require.NoError(t, err)
}

func clearNavbarHistoryEvents(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Evaluate("() => { window.__navbarHistoryEvents = []; }", nil)
	require.NoError(t, err)
}

func assertLastNavbarHistoryEvent(t *testing.T, page playwright.Page, path string, cacheMiss bool) {
	t.Helper()
	result := mustEvaluate(t, page, "() => window.__navbarHistoryEvents.at(-1)")
	values := result.(map[string]any)
	require.Equal(t, path, values["path"])
	require.Equal(t, cacheMiss, values["cacheMiss"])
}

func focusKeyboardLink(t *testing.T, page playwright.Page, link playwright.Locator) {
	t.Helper()
	_, err := link.Evaluate(`element => {
		const scrollport = element.closest('#navbar-secondary-row nav > div');
		if (scrollport) {
			const port = scrollport.getBoundingClientRect();
			const item = element.getBoundingClientRect();
			if (item.left < port.left) scrollport.scrollLeft -= port.left - item.left;
			if (item.right > port.right) scrollport.scrollLeft += item.right - port.right;
		}
		element.focus({preventScroll: true});
	}`, nil)
	require.NoError(t, err)
}

func focusLinkWithoutScroll(t *testing.T, link playwright.Locator) {
	t.Helper()
	_, err := link.Evaluate("element => element.focus({preventScroll: true})", nil)
	require.NoError(t, err)
}

func emulateNavbarScaleTwo(t *testing.T, page playwright.Page) {
	t.Helper()
	cdp, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	_, err = cdp.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 2})
	require.NoError(t, err)
	_, err = cdp.Send("Emulation.setTouchEmulationEnabled", map[string]any{"enabled": true, "configuration": "mobile"})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => Math.abs(visualViewport.scale - 2) < 0.01", nil)
	require.NoError(t, err)
	require.InDelta(t, 0, mustEvaluateFloat(t, page, "() => visualViewport.offsetLeft"), 1)
	require.InDelta(t, 0, mustEvaluateFloat(t, page, "() => visualViewport.offsetTop"), 1)
}

func panVisualViewport(t *testing.T, page playwright.Page, targetLeft, targetTop float64) {
	t.Helper()
	cdp, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	scale := mustEvaluateFloat(t, page, "() => visualViewport.scale")
	for attempt := 0; attempt < 5; attempt++ {
		current := mustEvaluate(t, page, `() => ({left: visualViewport.offsetLeft, top: visualViewport.offsetTop})`).(map[string]any)
		currentLeft := numberAsFloat(t, current["left"])
		currentTop := numberAsFloat(t, current["top"])
		if math.Abs(currentLeft-targetLeft) <= 1 && math.Abs(currentTop-targetTop) <= 1 {
			return
		}
		_, err = cdp.Send("Input.dispatchMouseEvent", map[string]any{
			"type":    "mouseWheel",
			"x":       10,
			"y":       200,
			"deltaX":  (targetLeft - currentLeft) * scale,
			"deltaY":  (targetTop - currentTop) * scale,
			"buttons": 0,
		})
		require.NoError(t, err)
		page.WaitForTimeout(40)
	}
	_, err = page.WaitForFunction(`([wantLeft, wantTop]) => {
		return Math.abs(visualViewport.offsetLeft - wantLeft) <= 1 &&
			Math.abs(visualViewport.offsetTop - wantTop) <= 1;
	}`, []float64{targetLeft, targetTop}, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2500)})
	require.NoErrorf(t, err, "visual viewport did not reach (%.0f, %.0f); observed (%.2f, %.2f)", targetLeft, targetTop, mustEvaluateFloat(t, page, "() => visualViewport.offsetLeft"), mustEvaluateFloat(t, page, "() => visualViewport.offsetTop"))
}

func mustEvaluate(t *testing.T, target any, expression string, args ...any) any {
	t.Helper()
	var result any
	var err error
	switch value := target.(type) {
	case playwright.Page:
		if len(args) == 0 {
			result, err = value.Evaluate(expression, nil)
		} else {
			result, err = value.Evaluate(expression, args[0])
		}
	case playwright.Locator:
		if len(args) == 0 {
			result, err = value.Evaluate(expression, nil)
		} else {
			result, err = value.Evaluate(expression, args[0])
		}
	default:
		t.Fatalf("unsupported Playwright evaluation target %T", target)
	}
	require.NoError(t, err)
	return result
}

func mustEvaluateString(t *testing.T, target any, expression string) string {
	t.Helper()
	value := mustEvaluate(t, target, expression)
	result, ok := value.(string)
	require.True(t, ok, "expected string result, got %T", value)
	return result
}

func mustEvaluateFloat(t *testing.T, target any, expression string) float64 {
	t.Helper()
	value := mustEvaluate(t, target, expression)
	switch result := value.(type) {
	case float64:
		return result
	case int:
		return float64(result)
	case int64:
		return float64(result)
	default:
		t.Fatalf("expected float result, got %T", value)
		return 0
	}
}

func mustEvaluateBool(t *testing.T, target any, expression string) bool {
	t.Helper()
	value := mustEvaluate(t, target, expression)
	result, ok := value.(bool)
	require.True(t, ok, "expected bool result, got %T", value)
	return result
}

type navbarBox struct{ Width, Height float64 }

func mustBox(t *testing.T, locator playwright.Locator) navbarBox {
	t.Helper()
	result := mustEvaluate(t, locator, "element => { const box = element.getBoundingClientRect(); return {width: box.width, height: box.height}; }")
	values := result.(map[string]any)
	return navbarBox{Width: numberAsFloat(t, values["width"]), Height: numberAsFloat(t, values["height"])}
}

func numberAsFloat(t *testing.T, value any) float64 {
	t.Helper()
	switch result := value.(type) {
	case float64:
		return result
	case int:
		return float64(result)
	case int64:
		return float64(result)
	default:
		t.Fatalf("expected numeric result, got %T", value)
		return 0
	}
}

func floatPtr(value float64) *float64 { return &value }

func boolFloat(condition bool, yes, no float64) float64 {
	if condition {
		return yes
	}
	return no
}
