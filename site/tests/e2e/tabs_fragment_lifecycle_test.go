//go:build e2e && (full || tabs)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestTabs_HTMXFragmentLifecycle proves that a real sidebar HTMX navigation
// initializes Tabs through Alpine's MutationObserver. It intentionally does
// not call Alpine.initTree or install consumer-side tab listeners.
func TestTabs_HTMXFragmentLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	failures := watchPageFailures(page)
	require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`
		window.__tabsFragmentLifecycleProbe = {
			duplicateClickBindings: 0,
			retryTimers: 0,
		};
	`)}))

	_, err := page.Goto(baseURL+"/components/tabs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	assertTabsState(t, page, "#tabs-default", "Groups")
	require.NoError(t, page.Locator("#tabs-default").GetByRole("tab", playwright.LocatorGetByRoleOptions{Name: "Likes"}).Click())
	assertTabsState(t, page, "#tabs-default", "Likes")

	installTabsFragmentLifecycleProbe(t, page)
	for range 3 {
		navigateTabsFragment(t, page, "/components/accordion", "#accordion-fragment")
		navigateTabsFragment(t, page, "/components/tabs", "#tabs-fragment")
		assertTabsState(t, page, "#tabs-default", "Groups")

		likes := page.Locator("#tabs-default").GetByRole("tab", playwright.LocatorGetByRoleOptions{Name: "Likes"})
		require.NoError(t, likes.Click())
		assertTabsState(t, page, "#tabs-default", "Likes")
		require.NoError(t, likes.Press("ArrowRight"))
		assertTabsState(t, page, "#tabs-default", "Likes")
		focused, err := page.Evaluate(`() => document.activeElement?.textContent?.trim()`, nil)
		require.NoError(t, err)
		require.Contains(t, focused, "Comments")
	}

	probe, err := page.Evaluate(`() => {
		const probe = window.__tabsFragmentLifecycleProbe;
		return [probe.duplicateClickBindings, probe.retryTimers];
	}`, nil)
	require.NoError(t, err)
	metrics, ok := probe.([]any)
	require.True(t, ok, "fragment lifecycle probe should return metrics")
	require.Equal(t, 0, metrics[0])
	require.Equal(t, 0, metrics[1])
	failures.RequireEmpty(t)
}

func navigateTabsFragment(t *testing.T, page playwright.Page, href, readySelector string) {
	t.Helper()
	link := page.Locator(`a[hx-get="` + href + `"]`).First()
	_, err := page.ExpectResponse("**"+href, func() error {
		return link.Click()
	}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err)
	require.NoError(t, page.Locator(readySelector).WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}))
}

func assertTabsState(t *testing.T, page playwright.Page, root, selected string) {
	t.Helper()
	_, err := page.WaitForFunction(`([root, selected]) => {
		const tabs = Array.from(document.querySelectorAll(root + ' [role="tab"]'));
		const panels = Array.from(document.querySelectorAll(root + ' [role="tabpanel"]'));
		const activeTabs = tabs.filter((tab) => tab.getAttribute('aria-selected') === 'true');
		const activePanels = panels.filter((panel) => getComputedStyle(panel).display !== 'none');
		return activeTabs.length === 1 &&
			activeTabs[0].textContent.trim().startsWith(selected) &&
			activeTabs[0].getAttribute('tabindex') === '0' &&
			tabs.filter((tab) => tab !== activeTabs[0]).every((tab) => tab.getAttribute('aria-selected') === 'false' && tab.getAttribute('tabindex') === '-1') &&
			activePanels.length === 1 && activePanels[0].getAttribute('aria-label') === selected;
	}`, []any{root, selected}, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err)
}

func installTabsFragmentLifecycleProbe(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Evaluate(`() => {
		const probe = window.__tabsFragmentLifecycleProbe;
		const addEventListener = HTMLElement.prototype.addEventListener;
		HTMLElement.prototype.addEventListener = function (type, listener, options) {
			if (type === 'click' && this.getAttribute('role') === 'tab') {
				this.__tabsClickBindings = (this.__tabsClickBindings || 0) + 1;
				if (this.__tabsClickBindings > 1) probe.duplicateClickBindings += 1;
			}
			return addEventListener.call(this, type, listener, options);
		};
		const setTimeout = window.setTimeout;
		window.setTimeout = function (handler, delay, ...args) {
			if (delay === 25) probe.retryTimers += 1;
			return setTimeout.call(this, handler, delay, ...args);
		};
	}`)
	require.NoError(t, err)
}
