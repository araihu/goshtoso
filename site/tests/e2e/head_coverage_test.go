package e2e

import (
	"context"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/head"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var headAssetRefRe = regexp.MustCompile(`(?:src|href)="(/assets/[^"]+)"`)

func matchAssetURLs(html string) []string {
	seen := map[string]struct{}{}
	var urls []string
	for _, m := range headAssetRefRe.FindAllStringSubmatch(html, -1) {
		if _, ok := seen[m[1]]; ok {
			continue
		}
		seen[m[1]] = struct{}{}
		urls = append(urls, m[1])
	}
	sort.Strings(urls)
	return urls
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

	fullURLs := matchAssetURLs(full.String())
	minimalURLs := matchAssetURLs(minimal.String())

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
