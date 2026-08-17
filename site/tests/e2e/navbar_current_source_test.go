//go:build e2e && full && goshtoso_current_source

package e2e

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/navbar"
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

func TestNavbar_CurrentSourceSecondaryRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	fixture := newNavbarCurrentSourceFixture(t)
	focusPage := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1280, Height: 800},
	})
	failures := watchPageFailures(focusPage)

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
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">%s<link rel="stylesheet" href="/assets/styles.css">%s</head>
<body class="min-h-[1200px] bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark">
<main class="mx-auto w-full max-w-[1280px]">
<button id="outside-focus" type="button" class="m-2 min-h-11 min-w-11 px-3 py-2" hx-get="/api/navbar/unrelated" hx-target="#unrelated-content" hx-swap="innerHTML" hx-push-url="/other">Outside focus</button>
<div id="navbar-secondary-host">%s</div>
<div id="unrelated-content" class="p-4">Initial unrelated content</div>
</main>
<script>
window.__navbarHistoryEvents = [];
const focusCurrent = () => requestAnimationFrame(() => {
  const row = document.getElementById("navbar-secondary-row");
  row?.querySelector('a[aria-current="page"], a[aria-current="location"]')?.focus();
});
document.body.addEventListener("htmx:afterSettle", (event) => {
  if (event.detail?.target?.id === "navbar-secondary-row") focusCurrent();
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
  window.__navbarHistoryEvents.push({
    path: event.detail?.path || "",
    cacheMiss: event.detail?.cacheMiss === true,
  });
  if (isSecondaryHistoryPath(event.detail?.path)) focusCurrent();
});
</script>
</body></html>`, header, dependencies, nav)
}

func navbarSecondaryConfig(view string, scrollable bool) navbar.SecondaryConfig {
	return navbar.SecondaryConfig{
		Links: []navbar.SecondaryLink{
			{Label: "Overview", Href: navbarOverviewURL, Current: currentForView(view, "overview"), LinkAttrs: navbarHTMXAttrs("overview")},
			{Label: "Details", Href: navbarDetailsURL, Current: currentForView(view, "details"), LinkAttrs: navbarHTMXAttrs("details")},
			{Label: "A long secondary navigation label for overflow", Href: "/components/navbar?view=long", LinkAttrs: navbarHTMXAttrs("overview")},
		},
		Actions: []templ.Component{dropdown.Dropdown(dropdown.Config{
			ID:        "navbar-secondary-actions",
			Label:     "Actions",
			MenuAlign: dropdown.AlignEnd,
			Trigger:   templ.Raw(`<button type="button" aria-haspopup="menu" class="inline-flex min-h-11 min-w-11 items-center gap-2 whitespace-nowrap px-3 py-2">Actions</button>`),
			Sections: []dropdown.Section{{Items: []dropdown.Item{
				{Label: "Open action", Href: "#open-action"},
			}}},
		})},
		AriaLabel:  "secondary navigation",
		Scrollable: scrollable,
		RootClass:  "overflow-x-hidden",
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
}

func assertNavbarHistoryCacheHit(t *testing.T, fixture *navbarCurrentSourceFixture) {
	t.Helper()
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 1280, Height: 800}})
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
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 1280, Height: 800}})
	navigateNavbarFixture(t, page, fixture, navbarOverviewURL)
	_, err := page.Evaluate("() => { htmx.config.historyCacheSize = 0; window.__navbarHistoryEvents = []; }", nil)
	require.NoError(t, err)
	clickNavbarView(t, page, "details")
	nativeBeforeBack := fixture.nativeRequests.Load()
	goBackNavbar(t, page, navbarOverviewURL, "Overview")
	require.Equal(t, nativeBeforeBack+1, fixture.nativeRequests.Load(), "cache-miss back must reload the native page")
	assertLastNavbarHistoryEvent(t, page, navbarOverviewURL, true)
	nativeBeforeForward := fixture.nativeRequests.Load()
	goForwardNavbar(t, page, navbarDetailsURL, "Details")
	require.Equal(t, nativeBeforeForward+1, fixture.nativeRequests.Load(), "cache-miss forward must reload the native page")
	assertLastNavbarHistoryEvent(t, page, navbarDetailsURL, true)
	_, err = page.Evaluate("() => { htmx.config.historyCacheSize = 10; }", nil)
	require.NoError(t, err)
}

func assertNavbarUnrelatedHistoryFocus(t *testing.T, fixture *navbarCurrentSourceFixture) {
	t.Helper()
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 1280, Height: 800}})
	navigateNavbarFixture(t, page, fixture, navbarOverviewURL)
	clearNavbarHistoryEvents(t, page)
	_, err := page.ExpectResponse("**/api/navbar/unrelated", func() error {
		return page.Locator("#outside-focus").Click()
	}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	require.NoError(t, page.WaitForURL("**/other"))
	require.NoError(t, page.Locator("#outside-focus").Focus())
	clickNavbarView(t, page, "details")
	goBackNavbar(t, page, "/other", "Overview")
	require.Equal(t, "outside-focus", mustEvaluateString(t, page, "() => document.activeElement?.id"), "unrelated restore must preserve outside focus")
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
			page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
				Viewport:          &playwright.Size{Width: viewport.width, Height: viewport.height},
				DeviceScaleFactor: floatPtr(boolFloat(viewport.scale2, 2, 1)),
			})
			if viewport.scale2 {
				emulateNavbarScaleTwo(t, page)
			}
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
						path := navbarOverviewURL + "&scrollable=" + strconv.FormatBool(scrollable)
						navigateNavbarFixture(t, page, fixture, path)
						if viewport.scale2 {
							emulateNavbarScaleTwo(t, page)
						}
						setThemeMode(t, page, theme.theme, theme.dark)
						assertNavbarCurrentState(t, page, "Overview")
						assertNavbarVisualGeometry(t, page, scrollable, viewport.scale2)
						assertNavbarContrastMatrix(t, page, theme.dark)
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

func assertNavbarActionSeparation(t *testing.T, page playwright.Page) {
	t.Helper()
	require.Equal(t, 1, mustCount(t, page.Locator("#navbar-secondary-row > nav")))
	require.Equal(t, 1, mustCount(t, page.Locator("#navbar-secondary-row > [data-navbar-actions='true']")))
	require.Equal(t, 0, mustCount(t, page.Locator("#navbar-secondary-row > nav button")))
	actions := page.Locator("#navbar-secondary-row > [data-navbar-actions='true']")
	require.GreaterOrEqual(t, mustCount(t, actions.Locator("button")), 1)
	require.Equal(t, "8px", mustEvaluateString(t, actions, "element => getComputedStyle(element).columnGap"))
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
	assertNavbarActionPanelBounds(t, page, scale2)
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
	link := page.Locator("#navbar-secondary-row nav a").Nth(1)
	current := page.Locator("#navbar-secondary-row nav a[aria-current]")
	for _, locator := range []playwright.Locator{link, current} {
		require.GreaterOrEqual(t, measureRenderedContrast(t, locator).Ratio, 4.5)
		require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, locator, "borderBottomColor").Ratio, 3.0)
	}
	require.NoError(t, link.Hover())
	require.GreaterOrEqual(t, measureRenderedContrast(t, link).Ratio, 4.5)
	require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, link, "borderBottomColor").Ratio, 3.0)
	focusKeyboardLink(t, page, current)
	require.GreaterOrEqual(t, measureRenderedContrast(t, current).Ratio, 4.5)
	require.GreaterOrEqual(t, measureRenderedPropertyContrast(t, current, "borderBottomColor").Ratio, 3.0)
	err := page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce})
	require.NoError(t, err)
	require.Equal(t, "none", mustEvaluateString(t, current, "element => getComputedStyle(element).transitionProperty"))
	_ = dark
}

func assertNavbarActionPanelBounds(t *testing.T, page playwright.Page, scale2 bool) {
	t.Helper()
	trigger := page.Locator("#navbar-secondary-row [data-navbar-actions='true'] button").First()
	require.NoError(t, trigger.ScrollIntoViewIfNeeded())
	require.NoError(t, trigger.Click())
	panel := page.Locator("#navbar-secondary-row [role='menu']").First()
	require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	result := mustEvaluate(t, panel, `(element) => {
		const rect = element.getBoundingClientRect();
		const viewport = window.visualViewport || {offsetLeft: 0, offsetTop: 0, width: document.documentElement.clientWidth, height: innerHeight};
		return {left: rect.left, right: rect.right, top: rect.top, bottom: rect.bottom, viewportLeft: viewport.offsetLeft, viewportRight: viewport.offsetLeft + viewport.width, viewportTop: viewport.offsetTop, viewportBottom: viewport.offsetTop + viewport.height};
	}`)
	values := result.(map[string]any)
	require.GreaterOrEqual(t, numberAsFloat(t, values["left"]), numberAsFloat(t, values["viewportLeft"]))
	require.LessOrEqual(t, numberAsFloat(t, values["right"]), numberAsFloat(t, values["viewportRight"]))
	require.GreaterOrEqual(t, numberAsFloat(t, values["top"]), numberAsFloat(t, values["viewportTop"]))
	require.LessOrEqual(t, numberAsFloat(t, values["bottom"]), numberAsFloat(t, values["viewportBottom"]))
	if scale2 {
		require.InDelta(t, 2, mustEvaluateFloat(t, page, "() => visualViewport.scale"), 0.01)
	}
	require.NoError(t, page.Keyboard().Press("Escape"))
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
	require.LessOrEqual(t, numberAsFloat(t, values["scrollLeft"])+4, numberAsFloat(t, values["left"]))
	require.LessOrEqual(t, numberAsFloat(t, values["right"]), numberAsFloat(t, values["scrollRight"])-4)
	require.LessOrEqual(t, numberAsFloat(t, values["scrollTop"])+4, numberAsFloat(t, values["top"]))
	require.LessOrEqual(t, numberAsFloat(t, values["bottom"]), numberAsFloat(t, values["scrollBottom"])-4)
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
	require.NoError(t, page.WaitForURL("**"+href))
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
	_, err = page.WaitForFunction(`label => document.querySelector('#navbar-secondary-row a[aria-current]')?.textContent?.trim() === label`, label)
	require.NoError(t, err)
	if strings.HasPrefix(path, "/components/navbar") {
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
	assertCurrentLinkFocused(t, page)
}

func assertCurrentLinkFocused(t *testing.T, page playwright.Page) {
	t.Helper()
	require.True(t, mustEvaluateBool(t, page, `() => {
		const current = document.querySelector('#navbar-secondary-row a[aria-current]');
		return current !== null && document.activeElement === current;
	}`))
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
	require.NoError(t, page.Keyboard().Press("Tab"))
	require.NoError(t, link.Focus())
}

func emulateNavbarScaleTwo(t *testing.T, page playwright.Page) {
	t.Helper()
	cdp, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	_, err = cdp.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 2})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => Math.abs(visualViewport.scale - 2) < 0.01", nil)
	require.NoError(t, err)
	require.InDelta(t, 0, mustEvaluateFloat(t, page, "() => visualViewport.offsetLeft"), 1)
	require.InDelta(t, 0, mustEvaluateFloat(t, page, "() => visualViewport.offsetTop"), 1)
	_, err = cdp.Send("Input.synthesizePinchGesture", map[string]any{
		"x":                 96,
		"y":                 211,
		"scaleFactor":       1,
		"relativeSpeed":     800,
		"gestureSourceType": "touch",
	})
	require.NoError(t, err)
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
