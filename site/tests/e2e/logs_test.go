package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// toInt coerces a Playwright Evaluate result (JS number) to int.
func toInt(v any) int {
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
		Content: new("try{document.cookie='gt_storage=allowed; Path=/; SameSite=Lax'}catch(e){}"),
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

// TestLogFeed_AutoScrollFollowsStream verifies the checked Auto-scroll control
// keeps the persistent feed pinned to the newest streamed rows.
func TestLogFeed_AutoScrollFollowsStream(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	_, err := page.Evaluate("() => { document.getElementById('log-feed').style.height = '64px'; }")
	require.NoError(t, err)
	waitForRows(t, page)

	_, err = page.WaitForFunction(
		`() => {
			const feed = document.getElementById('log-feed');
			return feed && feed.children.length >= 4 && feed.scrollHeight > feed.clientHeight;
		}`,
		nil)
	require.NoError(t, err)

	atBottom, err := page.Evaluate(
		`() => {
			const feed = document.getElementById('log-feed');
			return (feed.scrollHeight - feed.clientHeight - feed.scrollTop) <= 2;
		}`)
	require.NoError(t, err)
	require.Equal(t, true, atBottom, "auto-scroll should keep the log feed pinned to the newest rows")
}

// TestLogFeed_FragmentNavNoErrors lands elsewhere, navigates to the feed via the
// sidebar (htmx fragment swap), and asserts the stream connects with no console/
// page errors — the regression guard for Alpine.data registering on fragment-nav.
func TestLogFeed_FragmentNavNoErrors(t *testing.T) {
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
	require.NoError(t, page.Locator("a[href='/examples/logs']").First().Click())
	// Rows only appear if the logFeed Alpine component registered after the swap
	// and the SSE connection opened.
	_, err = page.WaitForFunction("() => !!document.querySelector('#log-feed .log-row')", nil)
	require.NoError(t, err)
	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav log feed: %v", jsErrors)
}

// TestLogFeed_PauseStopsAndResumes verifies Pause closes the SSE stream (the
// connector element is removed) and Resume reconnects and the feed grows again.
func TestLogFeed_PauseStopsAndResumes(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	waitForRows(t, page)

	pause := page.Locator("#logs-fragment button").Filter(playwright.LocatorFilterOptions{HasText: "Pause"})
	require.NoError(t, pause.Click())
	// Paused: Alpine flag set AND the SSE connector removed (stream closed) — no
	// more rows can arrive, asserted without any wall-clock wait.
	_, err := page.WaitForFunction(
		"() => { const d = Alpine.$data(document.getElementById('logs-fragment')); return d && d.paused === true && !document.querySelector('#logs-fragment [sse-connect]'); }",
		nil)
	require.NoError(t, err)

	// Capture the frozen count, then Resume and assert the feed grows past it.
	v, err := page.Evaluate("() => document.querySelectorAll('#log-feed .log-row').length")
	require.NoError(t, err)
	before := toInt(v)
	resume := page.Locator("#logs-fragment button").Filter(playwright.LocatorFilterOptions{HasText: "Resume"})
	require.NoError(t, resume.Click())
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

// TestLogFeed_ClearEmptiesFeed pauses (closing the stream so it cannot refill
// mid-assert), then Clear empties the feed.
func TestLogFeed_ClearEmptiesFeed(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	waitForRows(t, page)

	pause := page.Locator("#logs-fragment button").Filter(playwright.LocatorFilterOptions{HasText: "Pause"})
	require.NoError(t, pause.Click())
	_, err := page.WaitForFunction(
		"() => { const d = Alpine.$data(document.getElementById('logs-fragment')); return d && d.paused === true && !document.querySelector('#logs-fragment [sse-connect]'); }",
		nil)
	require.NoError(t, err)

	clear := page.Locator("#logs-fragment button").Filter(playwright.LocatorFilterOptions{HasText: "Clear"})
	require.NoError(t, clear.Click())
	_, err = page.WaitForFunction("() => document.querySelectorAll('#log-feed .log-row').length === 0", nil)
	require.NoError(t, err)
}

// TestLogFeed_SidebarScrollPreservedDuringStream guards a layout regression: the
// global htmx:afterSwap sidebar-scroll restore used to fire on EVERY htmx swap,
// so each SSE log append snapped the nav sidebar back to the top. The restore is
// now gated to #main-content (page-nav) swaps only; an in-page feed swap must
// leave the nav scroll position alone.
func TestLogFeed_SidebarScrollPreservedDuringStream(t *testing.T) {
	page := newIsolatedPage(t)
	// Force the nav list to overflow so it is actually scrollable (width >= lg so
	// the sidebar is visible; short height so the component list overflows).
	require.NoError(t, page.SetViewportSize(1100, 600))
	gotoLogs(t, page)
	waitForRows(t, page)

	// Scroll the nav sidebar down and confirm it moved.
	scrolled, err := page.Evaluate("() => { const sb = document.querySelector('.sidebar-scroll'); sb.scrollTop = 120; return sb.scrollTop; }")
	require.NoError(t, err)
	require.Greater(t, toInt(scrolled), 0, "nav sidebar must be scrollable for this test to be meaningful")
	start := toInt(scrolled)

	// Let several more SSE rows stream in — each previously reset the sidebar.
	v, err := page.Evaluate("() => document.querySelectorAll('#log-feed .log-row').length")
	require.NoError(t, err)
	base := toInt(v)
	_, err = page.WaitForFunction(
		fmt.Sprintf("() => document.querySelectorAll('#log-feed .log-row').length >= %d", base+3), nil)
	require.NoError(t, err)

	// Sidebar scroll position must be unchanged by the log appends.
	after, err := page.Evaluate("() => document.querySelector('.sidebar-scroll').scrollTop")
	require.NoError(t, err)
	require.Equal(t, start, toInt(after), "nav sidebar scroll must not reset while logs stream")
}

// TestLogFeed_StatusReachesConnected guards against the connection-status sticking
// on "Connecting": the htmx:sseOpen listener must sit on an ancestor of the
// sse-connect connector so the bubbled event flips `connected` to true once the
// stream opens (rows can stream while connected stays false if the listener is on
// a sibling element, never receiving the event).
func TestLogFeed_StatusReachesConnected(t *testing.T) {
	page := newIsolatedPage(t)
	gotoLogs(t, page)
	waitForRows(t, page)

	_, err := page.WaitForFunction(
		"() => { const d = Alpine.$data(document.getElementById('logs-fragment')); return d && d.connected === true; }", nil)
	require.NoError(t, err)
	// The visible status text reads "Connected", not "Connecting".
	_, err = page.WaitForFunction(
		"() => document.querySelector('#logs-fragment [x-text=\"statusText\"]').textContent.trim() === 'Connected'", nil)
	require.NoError(t, err)
}
