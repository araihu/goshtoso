package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKbdCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/kbd", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("#kbd-fragment").WaitFor())
	require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
		HasText: "KBD",
	}).First().WaitFor())

	frequentlyUsed := page.Locator("#kbd-frequently-used")
	for _, key := range []string{"Option", "Backspace", "Delete", "Caps Lock"} {
		require.NoError(t, frequentlyUsed.Locator("kbd").Filter(playwright.LocatorFilterOptions{
			HasText: key,
		}).WaitFor())
	}

	iconKey := page.Locator("#kbd-icons kbd[aria-label='Command']")
	require.NoError(t, iconKey.WaitFor())
	ariaLabel, err := iconKey.GetAttribute("aria-label")
	require.NoError(t, err)
	assert.Equal(t, "Command", ariaLabel)
	require.NoError(t, iconKey.Locator("span[aria-hidden='true'] svg").WaitFor())
	require.NoError(t, iconKey.Locator(".sr-only").Filter(playwright.LocatorFilterOptions{
		HasText: "Command",
	}).WaitFor())

	inlineClass, err := page.Locator("#kbd-inline kbd").First().GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, inlineClass, "min-h-6")
	assert.Contains(t, inlineClass, "text-xs")

	sizeClasses := map[string]string{
		"xs": "min-h-5",
		"sm": "min-h-6",
		"md": "min-h-7",
		"lg": "min-h-9",
	}
	for label, className := range sizeClasses {
		require.NoError(t, page.Locator("label[for='kbd-size-"+label+"']").Click())
		require.NoError(t, page.Locator("[data-testid='kbd-size-selected']").Filter(playwright.LocatorFilterOptions{
			HasText: label,
		}).WaitFor())
		require.NoError(t, page.Locator("[data-testid='kbd-size-preview-"+label+"']").WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		count, err := page.Locator("[data-testid='kbd-size-preview-" + label + "'] kbd." + className).Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "expected one visible %s size key", label)
	}

	functionCount, err := page.Locator("#kbd-functions kbd").Count()
	require.NoError(t, err)
	assert.Equal(t, 11, functionCount)
	require.NoError(t, page.Locator("#kbd-functions kbd").Filter(playwright.LocatorFilterOptions{
		HasText: "F12",
	}).WaitFor())

	require.NoError(t, page.Locator("table").Filter(playwright.LocatorFilterOptions{
		HasText: "Accessible label",
	}).WaitFor())

	before, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => document.documentElement.classList.toggle('dark')`, nil)
	require.NoError(t, err)
	after, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)
	assert.NotEqual(t, before, after, "dark class should toggle while kbd demo remains mounted")
	require.NoError(t, page.Locator("#kbd-fragment kbd").First().WaitFor())

	require.Empty(t, pageErrors, "uncaught JS exceptions on kbd demo: %v", pageErrors)
	require.Empty(t, filterIgnorable(consoleErrors), "console errors on kbd demo: %s", strings.Join(consoleErrors, "; "))
}

func TestKbdCoverageFragmentNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/getting-started", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	link := page.Locator("a[href='/components/kbd']").First()
	require.NoError(t, link.ScrollIntoViewIfNeeded())
	require.NoError(t, link.Click())
	require.NoError(t, page.Locator("#kbd-fragment").WaitFor())
	require.NoError(t, page.Locator("#kbd-icons kbd[aria-label='Shift']").WaitFor())
	require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
		HasText: "Keys With Icons",
	}).First().WaitFor())

	require.Empty(t, filterIgnorable(consoleErrors), "console errors after kbd fragment navigation: %s", strings.Join(consoleErrors, "; "))
}
