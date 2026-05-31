package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// toInt coerces a Playwright Evaluate result (JS number) to int.
func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// gotoLogs opens the live log feed with a fast stream interval and waits for Alpine.
func gotoLogs(t *testing.T, page playwright.Page) {
	t.Helper()
	// Dismiss the first-run cookie banner so it can't intercept control clicks.
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: playwright.String("try{localStorage.setItem('cookieConsent','accepted')}catch(e){}"),
	}))
	_, err := page.Goto(baseURL + "/examples/logs?interval=50ms")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
}

// waitForRows waits until at least one log row has streamed in.
func waitForRows(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.WaitForFunction("() => !!document.querySelector('#log-feed .log-row')", nil)
	require.NoError(t, err)
}

// TestLogFeed_StreamsRows verifies rows stream in and carry a level badge.
func TestLogFeed_StreamsRows(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	waitForRows(t, page)
	// A streamed row contains a rendered level badge (span with a log level word).
	hasBadge, err := page.Evaluate("() => !!document.querySelector('#log-feed .log-row span')")
	require.NoError(t, err)
	require.Equal(t, true, hasBadge)
}

// TestLogFeed_FragmentNavNoErrors lands elsewhere, navigates to the feed via the
// sidebar (htmx fragment swap), and asserts the stream connects with no console/
// page errors — the regression guard for Alpine.data registering on fragment-nav.
func TestLogFeed_FragmentNavNoErrors(t *testing.T) {
	page := newIsolatedPage(t)

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
	require.NoError(t, page.Locator("a[href='/examples/logs']").First().Click())
	// Rows only appear if the logFeed Alpine component registered after the swap
	// and the SSE connection opened.
	_, err = page.WaitForFunction("() => !!document.querySelector('#log-feed .log-row')", nil)
	require.NoError(t, err)
	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav log feed: %v", jsErrors)
}

// TestLogFeed_PauseStopsAndResumes verifies Pause stops new rows (status→Paused)
// and Resume restarts the stream.
func TestLogFeed_PauseStopsAndResumes(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	waitForRows(t, page)

	pauseBtn := page.Locator("#logs-fragment button").Filter(playwright.LocatorFilterOptions{HasText: "Pause"})
	require.NoError(t, pauseBtn.First().Click())
	_, err := page.WaitForFunction(
		"() => Alpine.$data(document.getElementById('logs-fragment'))?.paused === true",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err)

	// Let any in-flight swap settle, then confirm the row count is stable across a
	// window in which an un-paused feed (50ms ticks) would have grown.
	page.WaitForTimeout(100)
	v1, err := page.Evaluate("() => document.querySelectorAll('#log-feed .log-row').length")
	require.NoError(t, err)
	page.WaitForTimeout(400)
	v2, err := page.Evaluate("() => document.querySelectorAll('#log-feed .log-row').length")
	require.NoError(t, err)
	require.Equal(t, toInt(v1), toInt(v2), "row count must not grow while paused")

	// Resume → rows grow again.
	before := toInt(v2)
	resumeBtn := page.Locator("#logs-fragment button").Filter(playwright.LocatorFilterOptions{HasText: "Resume"})
	require.NoError(t, resumeBtn.First().Click())
	_, err = page.WaitForFunction(
		fmt.Sprintf("() => document.querySelectorAll('#log-feed .log-row').length > %d", before), nil)
	require.NoError(t, err)
}

// TestLogFeed_FilterHidesLowerLevels sets the min level to error and asserts no
// debug/info rows remain visible (pure-CSS filter).
func TestLogFeed_FilterHidesLowerLevels(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	waitForRows(t, page)

	// Drive the filter via the Select's Alpine model directly (avoids dropdown choreography).
	_, err := page.Evaluate("() => { Alpine.$data(document.getElementById('logs-fragment')).minLevel = 'error'; }")
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		"() => document.querySelector('[x-ref=feedWrap]').classList.contains('flt-error')", nil)
	require.NoError(t, err)

	visibleLower, err := page.Evaluate(`() => [...document.querySelectorAll('#log-feed .log-level-debug, #log-feed .log-level-info')].filter(el => el.offsetParent !== null).length`)
	require.NoError(t, err)
	require.Equal(t, 0, toInt(visibleLower), "debug/info rows must be hidden under the error filter")
}

// TestLogFeed_ClearEmptiesFeed pauses (so the stream cannot refill mid-assert),
// then Clear empties the feed.
func TestLogFeed_ClearEmptiesFeed(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	waitForRows(t, page)

	pauseBtn := page.Locator("#logs-fragment button").Filter(playwright.LocatorFilterOptions{HasText: "Pause"})
	require.NoError(t, pauseBtn.First().Click())
	_, err := page.WaitForFunction(
		"() => Alpine.$data(document.getElementById('logs-fragment'))?.paused === true",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err)
	page.WaitForTimeout(100)

	require.NoError(t, page.Locator("button", playwright.PageLocatorOptions{HasText: "Clear"}).First().Click())
	_, err = page.WaitForFunction("() => document.querySelectorAll('#log-feed .log-row').length === 0", nil)
	require.NoError(t, err)
}
