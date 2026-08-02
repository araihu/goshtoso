//go:build e2e && full

package e2e

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

const componentDocsHeadingSelector = "#main-content h1[data-toc-heading]"

func TestAllComponentDocsDirectLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	for _, entry := range catalog.ComponentPages() {
		t.Run(entry.Active, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			failures := watchPageFailures(page)
			response, err := page.Goto(baseURL+entry.Path, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.Status())

			heading, err := page.Locator(componentDocsHeadingSelector).TextContent()
			require.NoError(t, err)
			require.Equal(t, entry.Title, strings.TrimSpace(heading))

			descriptionCount, err := page.Locator("[data-component-description]").Count()
			require.NoError(t, err)
			require.Equal(t, 1, descriptionCount)

			previewCount, err := page.Locator("[data-component-preview]").Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, previewCount, 1)

			codeCount, err := page.Locator("[data-component-code]").Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, codeCount, 1)

			apiCount, err := page.Locator("[data-api-reference]").Count()
			require.NoError(t, err)
			require.Equal(t, 1, apiCount)
			requireComponentGoAPILink(t, page, entry)
			waitForPageSettled(t, page)
			failures.RequireEmpty(t)
		})
	}
}

func TestComponentDocsPreserveApplicationHeadAndRuntimeContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	response, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Status())

	for selector, want := range map[string]int{
		"head title": 1,
		`head link[rel="manifest"][href="/site.webmanifest"]`: 1,
		`head link[rel="apple-touch-icon"]`:                   1,
		`head link[rel="shortcut icon"][href="/favicon.ico"]`: 1,
		`head script[src*="htmx-ext-ws"]`:                     1,
		`head script[src*="htmx-ext-sse"]`:                    1,
	} {
		count, countErr := page.Locator(selector).Count()
		require.NoError(t, countErr)
		require.Equal(t, want, count, selector)
	}
	require.Equal(t, 1, mustLocatorCount(t, page.Locator("#docs-search")))
	require.Equal(t, 1, mustLocatorCount(t, page.Locator("#docs-search-dialog")))
}

func TestAllComponentDocsFragmentNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	entries := slices.DeleteFunc(catalog.ComponentPages(), func(entry catalog.Entry) bool {
		return entry.Active == "app-shell"
	})
	require.Len(t, entries, 49)

	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	response, err := page.Goto(baseURL+entries[0].Path, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusOK, response.Status())

	requireComponentDocsDestination(t, page, failures, entries[0])
	sentinel := installComponentDocsHTMXProof(t, page)
	mainContentSwapCount := 0
	for _, entry := range entries[1:] {
		link := page.Locator(fmt.Sprintf(
			`#componentdocshell-sidebar-content a[href=%q]`,
			entry.Path,
		))
		require.NoError(t, link.ScrollIntoViewIfNeeded())
		clickUntil(t, page, link, componentDocsDestinationReady(entry))
		requireComponentDocsDestination(t, page, failures, entry)
		mainContentSwapCount = requireComponentDocsHTMXProof(
			t,
			page,
			sentinel,
			mainContentSwapCount,
		)
	}
}

func installComponentDocsHTMXProof(t *testing.T, page playwright.Page) string {
	t.Helper()

	sentinel := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
	_, err := page.Evaluate(
		`sentinel => {
			const proof = { sentinel, mainContentSwaps: 0 };
			window.__componentDocsHTMXProof = proof;
			document.addEventListener("htmx:afterSwap", event => {
				if (event.detail?.target?.id === "main-content") {
					proof.mainContentSwaps += 1;
				}
			});
		}`,
		sentinel,
	)
	require.NoError(t, err)
	return sentinel
}

func requireComponentDocsHTMXProof(
	t *testing.T,
	page playwright.Page,
	sentinel string,
	previousMainContentSwapCount int,
) int {
	t.Helper()

	sentinelSurvived, err := componentDocsHTMXSentinelSurvived(page, sentinel)
	require.NoError(t, err)
	require.True(
		t,
		sentinelSurvived,
		"component docs navigation replaced the document instead of preserving the HTMX sentinel",
	)

	actualSwapCountValue, err := page.Evaluate(
		"() => window.__componentDocsHTMXProof?.mainContentSwaps",
	)
	require.NoError(t, err)
	actualSwapCount := playwrightInt(t, actualSwapCountValue)
	require.Greater(
		t,
		actualSwapCount,
		previousMainContentSwapCount,
		"component docs navigation did not increment the #main-content htmx:afterSwap count for this click",
	)
	return actualSwapCount
}

func playwrightInt(t *testing.T, value any) int {
	t.Helper()

	switch value := value.(type) {
	case int:
		return value
	case float64:
		integer := int(value)
		require.Equal(t, value, float64(integer), "Playwright number is not an integer")
		return integer
	default:
		require.Failf(t, "unexpected Playwright number", "got %T, want int or float64", value)
		return 0
	}
}

func componentDocsHTMXSentinelSurvived(page playwright.Page, sentinel string) (bool, error) {
	value, err := page.Evaluate(
		"expected => window.__componentDocsHTMXProof?.sentinel === expected",
		sentinel,
	)
	if err != nil {
		return false, err
	}
	survived, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("HTMX sentinel predicate returned %T, want bool", value)
	}
	return survived, nil
}

func componentDocsDestinationReady(entry catalog.Entry) string {
	return fmt.Sprintf(
		`() => window.location.pathname === %q &&
			document.querySelector(%q)?.textContent.trim() === %q`,
		entry.Path,
		componentDocsHeadingSelector,
		entry.Title,
	)
}

func requireComponentDocsDestination(
	t *testing.T,
	page playwright.Page,
	failures *pageFailures,
	entry catalog.Entry,
) {
	t.Helper()

	activeSelector := fmt.Sprintf(
		`#componentdocshell-sidebar-content a[href=%q][aria-current="page"]`,
		entry.Path,
	)
	_, err := page.WaitForFunction(
		fmt.Sprintf(
			`() => window.location.pathname === %q &&
				document.querySelector(%q)?.textContent.trim() === %q &&
				document.querySelector(%q) !== null`,
			entry.Path,
			componentDocsHeadingSelector,
			entry.Title,
			activeSelector,
		),
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)

	path, err := page.Evaluate("() => window.location.pathname")
	require.NoError(t, err)
	require.Equal(t, entry.Path, path)

	heading, err := page.Locator(componentDocsHeadingSelector).TextContent()
	require.NoError(t, err)
	require.Equal(t, entry.Title, strings.TrimSpace(heading))

	apiCount, err := page.Locator("[data-api-reference]").Count()
	require.NoError(t, err)
	require.Equal(t, 1, apiCount)
	requireComponentGoAPILink(t, page, entry)

	previewCount, err := page.Locator("[data-component-preview]").Count()
	require.NoError(t, err)
	require.GreaterOrEqual(t, previewCount, 1)

	activeLink := page.Locator(fmt.Sprintf(
		`#componentdocshell-sidebar-content a[href=%q]`,
		entry.Path,
	))
	current, err := activeLink.GetAttribute("aria-current")
	require.NoError(t, err)
	require.Equal(t, "page", current)

	runtimesReady, err := page.Evaluate(
		"() => typeof window.Alpine !== 'undefined' && typeof window.htmx !== 'undefined'",
	)
	require.NoError(t, err)
	require.Equal(t, true, runtimesReady)

	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
}

func TestComponentDocsHTMXProofRejectsFullPageReload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	entries := slices.DeleteFunc(catalog.ComponentPages(), func(entry catalog.Entry) bool {
		return entry.Active == "app-shell"
	})
	require.GreaterOrEqual(t, len(entries), 2)

	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	response, err := page.Goto(baseURL+entries[0].Path, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusOK, response.Status())
	requireComponentDocsDestination(t, page, failures, entries[0])

	sentinel := installComponentDocsHTMXProof(t, page)
	destination := entries[1]
	linkSelector := fmt.Sprintf(`#componentdocshell-sidebar-content a[href=%q]`, destination.Path)
	replaceComponentDocsLinkWithOrdinaryHref(t, page, destination.Path)

	require.NoError(t, page.Locator(linkSelector).Click())
	require.NoError(t, page.WaitForURL("**"+destination.Path))
	requireComponentDocsDestination(t, page, failures, destination)

	sentinelSurvived, err := componentDocsHTMXSentinelSurvived(page, sentinel)
	require.NoError(t, err)
	require.False(
		t,
		sentinelSurvived,
		"ordinary href navigation must replace the document and invalidate the HTMX sentinel",
	)
}

func TestComponentDocsComponentPageSpacing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	response, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Status())

	primary := page.Locator("[data-component-page] [data-component-example-body] > *").Nth(1)
	primaryGap, err := primary.Evaluate("element => parseFloat(getComputedStyle(element).marginTop)", nil)
	require.NoError(t, err)
	require.EqualValues(t, 16, primaryGap)

	variant := page.Locator("section[data-component-example]").First()
	variantMargin, err := variant.Evaluate("element => parseFloat(getComputedStyle(element).marginTop)", nil)
	require.NoError(t, err)
	require.EqualValues(t, 40, variantMargin)
}

func replaceComponentDocsLinkWithOrdinaryHref(
	t *testing.T,
	page playwright.Page,
	path string,
) {
	t.Helper()

	link := page.Locator(fmt.Sprintf(`#componentdocshell-sidebar-content a[href=%q]`, path))
	_, err := link.Evaluate(
		`link => {
			const ordinaryLink = link.cloneNode(true);
			for (const attribute of ["hx-get", "hx-target", "hx-swap", "hx-push-url"]) {
				ordinaryLink.removeAttribute(attribute);
			}
			link.replaceWith(ordinaryLink);
		}`,
		nil,
	)
	require.NoError(t, err)
}

func TestComponentDocsThemeMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	pages := []string{
		"/components/button",
		"/components/form",
		"/components/modal",
		"/components/table",
		"/components/toast",
	}
	states := []struct {
		name        string
		storedTheme string
		dark        bool
	}{
		{name: "goshtoso-light", storedTheme: "goshtoso", dark: false},
		{name: "goshtoso-dark", storedTheme: "goshtoso", dark: true},
		{name: "minimal-light", storedTheme: "minimal", dark: false},
		{name: "minimal-dark", storedTheme: "minimal", dark: true},
	}

	for _, path := range pages {
		for _, state := range states {
			t.Run(strings.TrimPrefix(path, "/components/")+"/"+state.name, func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				failures := watchPageFailures(page)
				script := fmt.Sprintf(`
    document.cookie = "gt_storage=allowed; Path=/; SameSite=Lax";
    localStorage.setItem("theme", %q);
    localStorage.setItem("darkMode", %q);
`, state.storedTheme, strconv.FormatBool(state.dark))
				require.NoError(t, page.AddInitScript(playwright.Script{
					Content: &script,
				}))

				response, err := page.Goto(baseURL+path, playwright.PageGotoOptions{
					WaitUntil: playwright.WaitUntilStateDomcontentloaded,
				})
				require.NoError(t, err)
				require.NotNil(t, response)
				require.Equal(t, http.StatusOK, response.Status())

				html := page.Locator("html")
				theme, err := html.GetAttribute("data-theme")
				require.NoError(t, err)
				require.Equal(t, "araihu", theme, "component docs should ignore legacy stored themes")

				className, err := html.GetAttribute("class")
				require.NoError(t, err)
				require.Equal(t, state.dark, hasClass(className, "dark"))

				preview := page.Locator("[data-component-preview]").First()
				visible, err := preview.IsVisible()
				require.NoError(t, err)
				require.True(t, visible)

				box, err := preview.BoundingBox()
				require.NoError(t, err)
				require.NotNil(t, box)
				require.Greater(t, box.Width, float64(0))
				require.Greater(t, box.Height, float64(0))

				value, err := page.Evaluate(`() => getComputedStyle(document.documentElement)
    .getPropertyValue("--color-surface").trim()`)
				require.NoError(t, err)
				surfaceToken, ok := value.(string)
				require.True(t, ok)
				require.NotEmpty(t, surfaceToken)

				waitForPageSettled(t, page)
				failures.RequireEmpty(t)
			})
		}
	}
}

func hasClass(className, target string) bool {
	return slices.Contains(strings.Fields(className), target)
}
