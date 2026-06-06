package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gotoSelectDemo opens the select demo on a fresh isolated page, wires up
// console/page-error capture, waits for Alpine, and returns the page plus
// accessors for any errors collected. filterIgnorable is defined in
// alert_coverage_test.go.
func gotoSelectDemo(t *testing.T) (playwright.Page, func() (pageErrs, consoleErrs []string)) {
	t.Helper()
	page := newIsolatedPage(t)

	var pageErrors, consoleErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/select", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	return page, func() ([]string, []string) { return pageErrors, consoleErrors }
}

// TestSelectOpenAndSelect exercises the live Alpine selection flow that static
// rendering tests can't reach: the aria-expanded toggle, revealing the option
// list, picking an option, and the hidden input mirroring the chosen value.
func TestSelectOpenAndSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page, errs := gotoSelectDemo(t)

	trigger := page.Locator("#os-trigger")
	require.NoError(t, trigger.WaitFor())

	// Live attribute starts false (Alpine binds isOpen || openedWithKeyboard).
	expanded, err := trigger.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "false", expanded)

	require.NoError(t, trigger.Click())
	require.NoError(t, page.Locator("#os-option-0").WaitFor())

	expanded, err = trigger.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "true", expanded)

	// Pick Linux (index 2); selectOption() closes the dropdown and updates state.
	require.NoError(t, page.Locator("#os-option-2").Click())
	_, err = page.WaitForFunction(
		"() => document.querySelector('#os-trigger span').textContent.trim() === 'Linux'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	value, err := page.Locator("input#os").InputValue()
	require.NoError(t, err)
	assert.Equal(t, "linux", value, "hidden input should mirror the selected value")

	expanded, err = trigger.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "false", expanded, "dropdown should close after selecting")

	pageErrs, consoleErrs := errs()
	assert.Empty(t, filterIgnorable(pageErrs), "unexpected page errors")
	assert.Empty(t, filterIgnorable(consoleErrs), "unexpected console errors")
}

// TestSelectClickOutsideCloses verifies the x-on:click.outside handler closes
// an open dropdown.
func TestSelectClickOutsideCloses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page, _ := gotoSelectDemo(t)

	require.NoError(t, page.Locator("#os-trigger").Click())
	require.NoError(t, page.Locator("#os-option-0").WaitFor())

	// Click the page heading, well outside the dropdown.
	require.NoError(t, page.Locator("main h1, main h2").First().Click())

	_, err := page.WaitForFunction(
		"() => document.querySelector('#os-trigger').getAttribute('aria-expanded') === 'false'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// TestSelectKeyboardOpenClose covers the keyboard-open path (openedWithKeyboard
// via ArrowDown) and the window-level Escape handler that closes it.
func TestSelectKeyboardOpenClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page, _ := gotoSelectDemo(t)

	trigger := page.Locator("#country-trigger")
	require.NoError(t, trigger.WaitFor())
	require.NoError(t, trigger.Focus())
	require.NoError(t, page.Keyboard().Press("ArrowDown"))

	_, err := page.WaitForFunction(
		"() => document.querySelector('#country-trigger').getAttribute('aria-expanded') === 'true'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	require.NoError(t, page.Keyboard().Press("Escape"))
	_, err = page.WaitForFunction(
		"() => document.querySelector('#country-trigger').getAttribute('aria-expanded') === 'false'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// TestSelectDependentBinding covers the Alpine Model + BindDisabled wiring: the
// second select stays disabled until the first has a value, then enables once a
// model is chosen.
func TestSelectDependentBinding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page, errs := gotoSelectDemo(t)

	require.NoError(t, page.Locator("#year-trigger").ScrollIntoViewIfNeeded())

	// BindDisabled "!firstValue" → disabled while firstValue is empty.
	_, err := page.WaitForFunction(
		"() => document.querySelector('#year-trigger').disabled === true",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	// Choose a model; the Alpine Model watcher sets firstValue.
	require.NoError(t, page.Locator("#modelName-trigger").Click())
	require.NoError(t, page.Locator("#modelName-option-0").WaitFor())
	require.NoError(t, page.Locator("#modelName-option-0").Click())

	_, err = page.WaitForFunction(
		"() => document.querySelector('#year-trigger').disabled === false",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "year select should enable once a model is chosen")

	pageErrs, consoleErrs := errs()
	assert.Empty(t, filterIgnorable(pageErrs), "unexpected page errors")
	assert.Empty(t, filterIgnorable(consoleErrs), "unexpected console errors")
}
