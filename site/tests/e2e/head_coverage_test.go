package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var headAssetRefRe = regexp.MustCompile(`(?:src|href)="(/assets/[^"]+)"`)
var headLoaderConfigRe = regexp.MustCompile(`data-goshtoso-dependencies="([^"]+)"`)

func matchAssetURLs(t *testing.T, output string) []string {
	t.Helper()
	seen := map[string]struct{}{}
	var urls []string
	add := func(url string) {
		if !strings.HasPrefix(url, "/assets/") {
			return
		}
		if _, ok := seen[url]; ok {
			return
		}
		seen[url] = struct{}{}
		urls = append(urls, url)
	}
	for _, match := range headAssetRefRe.FindAllStringSubmatch(output, -1) {
		add(match[1])
	}
	if match := headLoaderConfigRe.FindStringSubmatch(output); len(match) == 2 {
		var config struct {
			Dependencies []struct {
				PrimaryURL  string `json:"primary_url"`
				FallbackURL string `json:"fallback_url"`
			} `json:"dependencies"`
		}
		require.NoError(t, json.Unmarshal([]byte(html.UnescapeString(match[1])), &config))
		for _, dependency := range config.Dependencies {
			add(dependency.PrimaryURL)
			add(dependency.FallbackURL)
		}
	}
	sort.Strings(urls)
	return urls
}

func dependencyFallbackFixture(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	broken := func(name string) string { return "/cdn-unavailable/" + name + ".js" }
	var dependencyHead strings.Builder
	ctx := templ.WithNonce(context.Background(), "fallback-nonce")
	require.NoError(t, head.Dependencies(
		head.WithDependencyCDNURL(head.DependencyAlpineCollapse, broken("alpine-collapse")),
		head.WithDependencyCDNURL(head.DependencyAlpineFocus, broken("alpine-focus")),
		head.WithDependencyCDNURL(head.DependencyAlpineMask, broken("alpine-mask")),
		head.WithDependencyCDNURL(head.DependencyAlpineJS, broken("alpine")),
		head.WithDependencyCDNURL(head.DependencyHTMX, broken("htmx")),
	).Render(ctx, &dependencyHead))

	var failedRequests atomic.Int32
	mux := http.NewServeMux()
	mux.Handle("/assets/", assets.Handler())
	mux.HandleFunc("/cdn-unavailable/", func(w http.ResponseWriter, _ *http.Request) {
		failedRequests.Add(1)
		http.Error(w, "simulated CDN outage", http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/fragment", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<strong id="htmx-result">HTMX fallback works</strong>`)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html><head><meta charset="utf-8">%s</head><body>
  <div id="alpine-fixture" x-data="{ open: false }">
    <button id="alpine-toggle" x-on:click="open = !open">Toggle</button>
    <div id="collapse-panel" x-show="open" x-collapse>Alpine collapse works</div>
  </div>
  <div x-data>
    <label for="mask-input">Masked value</label>
    <input id="mask-input" x-mask="999-999">
  </div>
  <div x-data="{ trapped: false }">
    <button id="trap-toggle" x-on:click="trapped = true">Trap focus</button>
    <div x-show="trapped" x-trap="trapped">
      <input id="trapped-input">
      <button id="trap-close" x-on:click="trapped = false">Close trap</button>
    </div>
  </div>
  <button id="htmx-trigger" hx-get="/fragment" hx-target="#htmx-target">Load fragment</button>
  <div id="htmx-target"></div>
  <div id="combobox" data-combobox tabindex="0">
    <button id="combobox-first" role="option" tabindex="-1">First</button>
    <button id="combobox-second" role="option" tabindex="-1">Second</button>
  </div>
  <div id="action-group" data-goshtoso-action-group data-action-group-overflow-counts="1">
    <div data-action-group-primary><button>Primary</button></div>
    <div data-action-group-secondary><button>Secondary</button></div>
    <div data-action-group-overflow>
      <button aria-haspopup="true" aria-expanded="false">More</button>
      <div role="menu"><button role="menuitem">Secondary</button></div>
    </div>
  </div>
</body></html>`, dependencyHead.String())
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, &failedRequests
}

// TestHeadAssetContractServed renders head.Dependencies and DependenciesMinimal,
// then asserts every asset URL they reference is actually served (HTTP 200) by
// the running server's mounted asset handler. This is the component's core
// documented contract: the tags 404 (unstyled page / missing runtime) unless the
// handler is mounted. The existing TestDependenciesDemoPage only checks the demo
// code samples, not that the referenced assets resolve.
func TestHeadAssetContractServed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	var full, minimal strings.Builder
	require.NoError(t, head.Dependencies().Render(context.Background(), &full))
	require.NoError(t, head.DependenciesMinimal().Render(context.Background(), &minimal))

	fullURLs := matchAssetURLs(t, full.String())
	minimalURLs := matchAssetURLs(t, minimal.String())

	require.NotEmpty(t, fullURLs, "Dependencies() should reference asset URLs")
	require.NotEmpty(t, minimalURLs, "DependenciesMinimal() should reference asset URLs")

	// Minimal omits the collapse/focus plugins the full set ships.
	assert.Greater(t, len(fullURLs), len(minimalURLs),
		"Dependencies() should reference more assets than DependenciesMinimal()")

	for _, url := range fullURLs {
		resp, err := http.Get(baseURL + url)
		require.NoErrorf(t, err, "GET %s", url)
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "asset %s must be served", url)
		require.NoError(t, resp.Body.Close())
	}
}

// TestDependenciesPageNoConsoleErrors loads the dependencies demo page in a real
// browser and asserts it boots its runtime without console errors.
func TestDependenciesPageNoConsoleErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/dependencies", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	// Runtime referenced from the live <head> must initialise.
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined' && typeof window.htmx !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("main h1").WaitFor())
	title, err := page.Locator("main h1").TextContent()
	require.NoError(t, err)
	assert.Contains(t, title, "Dependencies")
	assert.Empty(t, consoleErrors, "page should load without console errors")
}

// TestDependenciesCDNFailureLoadsOrderedLocalFallback is the browser-level
// contract for the default mode. Every configured CDN request fails, yet all
// runtime-backed behaviors initialise from the exact embedded versions.
func TestDependenciesCDNFailureLoadsOrderedLocalFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	server, failedRequests := dependencyFallbackFixture(t)
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new(`window.__goshtosoFallbacks = [];
window.addEventListener("goshtoso:dependency-fallback", event => {
  window.__goshtosoFallbacks.push(event.detail.dependency);
});`),
	}))

	var pageErrors []string
	page.On("pageerror", func(err string) { pageErrors = append(pageErrors, err) })
	_, err := page.Goto(server.URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.Evaluate(`async () => await window.goshtosoDependencies.ready`, nil)
	require.NoError(t, err)
	require.Equal(t, int32(5), failedRequests.Load(), "every third-party CDN primary should be attempted")

	sourcesValue, err := page.Evaluate(`() => window.goshtosoDependencies.sources`, nil)
	require.NoError(t, err)
	sources, ok := sourcesValue.(map[string]any)
	require.True(t, ok, "dependency source ledger should be an object: %#v", sourcesValue)
	for _, name := range []string{"alpine-collapse", "alpine-focus", "alpine-mask", "alpine", "htmx"} {
		assert.Equal(t, "fallback", sources[name], "%s should use its local fallback", name)
	}
	assert.Equal(t, "primary", sources["first-party"], "the local first-party bundle should load without a fallback")
	nonces, err := page.Evaluate(`() => Array.from(document.querySelectorAll('script[data-goshtoso-dependency]'), script => script.nonce)`, nil)
	require.NoError(t, err)
	for _, nonce := range nonces.([]any) {
		assert.Equal(t, "fallback-nonce", nonce, "loader must propagate its nonce to every child script")
	}

	require.NoError(t, page.Locator("#alpine-toggle").Click())
	require.NoError(t, page.Locator("#collapse-panel").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, page.Locator("#mask-input").PressSequentially("123456"))
	assert.Equal(t, "123-456", requireInputValue(t, page, "#mask-input"))

	require.NoError(t, page.Locator("#trap-toggle").Click())
	_, err = page.WaitForFunction(`() => document.activeElement && document.activeElement.id === "trapped-input"`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#trap-close").Click())

	require.NoError(t, page.Locator("#htmx-trigger").Click())
	require.NoError(t, page.Locator("#htmx-result").WaitFor())

	require.NoError(t, page.Locator("#combobox").Focus())
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	activeID, err := page.Evaluate(`() => document.activeElement && document.activeElement.id`, nil)
	require.NoError(t, err)
	assert.Equal(t, "combobox-first", activeID)
	_, err = page.WaitForFunction(`() => document.querySelector("#action-group").dataset.actionGroupInitialized === "true"`, nil)
	require.NoError(t, err)

	fallbacks, err := page.Evaluate(`() => window.__goshtosoFallbacks`, nil)
	require.NoError(t, err)
	assert.Equal(t, []any{"alpine-collapse", "alpine-focus", "alpine-mask", "alpine", "htmx"}, fallbacks)
	assert.Empty(t, pageErrors, "fallback must not cause uncaught JavaScript errors")
}

func TestDependenciesLocalRuntimeBootsCombinedFirstPartyBundle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	var dependencyHead strings.Builder
	require.NoError(t, head.Dependencies(head.WithLocalRuntime()).Render(context.Background(), &dependencyHead))
	mux := http.NewServeMux()
	mux.Handle("/assets/", assets.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `<!doctype html><html><head>%s</head><body>
<div id="local-combobox" data-combobox tabindex="0"><button id="local-option" role="option">Option</button></div>
<div id="local-action-group" data-goshtoso-action-group data-action-group-overflow-counts="1">
<div data-action-group-primary><button>Primary</button></div><div data-action-group-secondary><button>Secondary</button></div>
<div data-action-group-overflow><button aria-haspopup="true" aria-expanded="false">More</button><div role="menu"><button role="menuitem">Secondary</button></div></div>
</div></body></html>`, dependencyHead.String())
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(server.URL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== "undefined" && typeof htmx !== "undefined" && document.querySelector("#local-action-group").dataset.actionGroupInitialized === "true"`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#local-combobox").Focus())
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	activeID, err := page.Evaluate(`() => document.activeElement.id`, nil)
	require.NoError(t, err)
	assert.Equal(t, "local-option", activeID)
}

func requireInputValue(t *testing.T, page playwright.Page, selector string) string {
	t.Helper()
	value, err := page.Locator(selector).InputValue()
	require.NoError(t, err)
	return value
}
