//go:build e2e && (full || combobox)

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComboboxCoverage_Demo asserts the demo page renders all three combobox
// variants and produces no console errors on first paint.
func TestComboboxCoverage_Demo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/combobox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	for _, id := range []string{"#industry-trigger", "#skills-trigger", "#users-trigger"} {
		vis, err := page.Locator(id).First().IsVisible()
		require.NoError(t, err)
		assert.True(t, vis, "%s trigger should be visible", id)
	}

	main, err := page.Locator("main").First().TextContent()
	require.NoError(t, err)
	assert.Contains(t, main, "Combobox")

	assert.Empty(t, consoleErrors, "no console errors on combobox demo (got: %v)", consoleErrors)
}

// TestComboboxCoverage_ServerLazySearchAndToggle exercises the server-mode
// (HTMX lazy) path: opening the users combobox, searching to load options over
// the wire, selecting one (server toggle round-trip), and verifying the OOB
// trigger-label and hidden input reflect the new selection.
func TestComboboxCoverage_ServerLazySearchAndToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		// htmx logs a benign warning when the search hx-include selector matches
		// no hidden inputs (i.e. before anything is selected). That is expected
		// component behavior, not a regression — ignore it.
		if msg.Type() == "error" && !strings.Contains(msg.Text(), "on hx-include returned no matches") {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/combobox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	trigger := page.Locator("#users-trigger").First()
	require.NoError(t, trigger.Click())

	// Dropdown opens (server mode uses the same Alpine shell).
	_, err = page.WaitForFunction(
		`() => document.querySelector('#users-trigger').getAttribute('aria-expanded') === 'true'`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	// Type into the search box; HTMX issues a GET that swaps the options list.
	search := page.Locator("#users [data-combobox-search]").First()
	require.NoError(t, search.Fill("al"))

	// Filtered options arrive over the wire: alice + albert, but not carol.
	alice := page.Locator(`#users [data-combobox-option][data-value="alice"]`).First()
	require.NoError(t, alice.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	albertCount, err := page.Locator(`#users [data-combobox-option][data-value="albert"]`).Count()
	require.NoError(t, err)
	assert.Equal(t, 1, albertCount, "search 'al' returns albert")
	carolCount, err := page.Locator(`#users [data-combobox-option][data-value="carol"]`).Count()
	require.NoError(t, err)
	assert.Equal(t, 0, carolCount, "search 'al' filters out carol")

	// Select alice — server toggle round-trip swaps the body and emits the OOB
	// trigger label. clickUntil retries past the HTMX rebind race.
	clickUntil(t, page, alice,
		`() => document.querySelectorAll('#users input[type=hidden][name="users"][value="alice"]').length === 1`)

	label, err := page.Locator("#users-trigger-label").TextContent()
	require.NoError(t, err)
	assert.Equal(t, "Alice", label, "single selection shows the option label via OOB swap")

	assert.Empty(t, consoleErrors, "no console errors during server-mode flow (got: %v)", consoleErrors)
}

func TestComboboxCoverage_CascadingProviderReloadsClusterOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/combobox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	trigger := page.Locator("#cluster-trigger").First()
	require.NoError(t, trigger.Click())
	search := page.Locator("#cluster [data-combobox-search]").First()
	require.NoError(t, search.Fill("prod"))

	awsCluster := page.Locator(`#cluster [data-combobox-option][data-value="prod-use1"]`).First()
	require.NoError(t, awsCluster.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	provider := page.Locator(`#combobox-cluster select[name="provider"]`).First()
	_, err = provider.SelectOption(playwright.SelectOptionValues{
		Values: &[]string{"gcp"},
	})
	require.NoError(t, err)

	gcpCluster := page.Locator(`#cluster [data-combobox-option][data-value="prod-us-central1"]`).First()
	require.NoError(t, gcpCluster.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	awsClusterCount, err := page.Locator(`#cluster [data-combobox-option][data-value="prod-use1"]`).Count()
	require.NoError(t, err)
	assert.Equal(t, 0, awsClusterCount, "provider dependency should filter out AWS clusters after selecting GCP")
}

// TestComboboxCoverage_KeyboardAndChevron covers the keyboard-open path
// (openedWithKeyboard) and the chevron rotate binding, then Escape-to-close.
func TestComboboxCoverage_KeyboardAndChevron(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/combobox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	trigger := page.Locator("#industry-trigger").First()
	require.NoError(t, trigger.Focus())

	// ArrowDown opens the dropdown via the keyboard handler.
	require.NoError(t, trigger.Press("ArrowDown"))
	_, err = page.WaitForFunction(
		`() => document.querySelector('#industry-trigger').getAttribute('aria-expanded') === 'true'`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	// Chevron rotates while open (x-bind:class adds rotate-180).
	_, err = page.WaitForFunction(
		`() => { const s = document.querySelector('#industry-trigger svg'); return s && s.classList.contains('rotate-180'); }`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#industry [data-combobox-body]").First().WaitFor(
		playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}),
		"dropdown body visible after keyboard open")

	// Escape closes and resets the rotate binding.
	require.NoError(t, page.Keyboard().Press("Escape"))
	_, err = page.WaitForFunction(
		`() => { const s = document.querySelector('#industry-trigger svg'); return s && !s.classList.contains('rotate-180'); }`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	expanded, err := page.Evaluate(`() => document.querySelector('#industry-trigger').getAttribute('aria-expanded')`, nil)
	require.NoError(t, err)
	assert.Equal(t, "false", expanded, "Escape clears expanded state")
}
