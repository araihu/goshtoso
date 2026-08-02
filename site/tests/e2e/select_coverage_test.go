//go:build e2e

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

// TestSelectKeyboardMovesActiveOption proves both trigger directions and
// in-list movement use the rendered option set instead of depending on focus
// plugin traversal order.
func TestSelectKeyboardMovesActiveOption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page, errs := gotoSelectDemo(t)
	trigger := page.Locator("#os-trigger")
	require.NoError(t, trigger.Focus())

	require.NoError(t, trigger.Press("ArrowUp"))
	_, err := page.WaitForFunction(
		"() => document.activeElement?.id === 'os-option-2'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "ArrowUp should open on the last option when no value is selected")

	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	_, err = page.WaitForFunction(
		"() => document.activeElement?.id === 'os-option-0'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "ArrowDown should wrap to the first option")

	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	_, err = page.WaitForFunction(
		"() => document.activeElement?.id === 'os-option-1'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "ArrowDown should move to the next option")

	require.NoError(t, page.Keyboard().Press("ArrowUp"))
	_, err = page.WaitForFunction(
		"() => document.activeElement?.id === 'os-option-0'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "ArrowUp should move to the previous option")

	pageErrs, consoleErrs := errs()
	assert.Empty(t, filterIgnorable(pageErrs), "unexpected page errors")
	assert.Empty(t, filterIgnorable(consoleErrs), "unexpected console errors")
}

// TestSelectKeyboardAdvancesFromSelectedValue guards the consumer workflow
// contract: ArrowDown on a closed Select advances from the current value,
// Enter commits it, and focus returns to the trigger.
func TestSelectKeyboardAdvancesFromSelectedValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page, errs := gotoSelectDemo(t)
	trigger := page.Locator("#os-success-trigger")
	require.NoError(t, trigger.Focus())
	require.NoError(t, trigger.Press("ArrowDown"))

	_, err := page.WaitForFunction(
		"() => document.activeElement?.id === 'os-success-option-1'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "ArrowDown should advance from selected Mac to Windows")

	require.NoError(t, page.Keyboard().Press("Enter"))
	_, err = page.WaitForFunction(
		"() => document.querySelector('#os-success').value === 'windows' && document.activeElement?.id === 'os-success-trigger'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	if err != nil {
		state, stateErr := page.Evaluate(`() => ({
			value: document.querySelector('#os-success')?.value,
			focus: document.activeElement?.id,
			label: document.querySelector('#os-success-trigger span')?.textContent?.trim(),
			expanded: document.querySelector('#os-success-trigger')?.getAttribute('aria-expanded'),
		})`, nil)
		require.NoError(t, stateErr)
		require.NoErrorf(t, err, "Enter should commit the active option and return focus to the trigger: state=%v", state)
	}

	pageErrs, consoleErrs := errs()
	assert.Empty(t, filterIgnorable(pageErrs), "unexpected page errors")
	assert.Empty(t, filterIgnorable(consoleErrs), "unexpected console errors")
}

// TestSelectExternalDraftRestoration exercises the public DOM synchronization
// contract: external code sets the hidden submission input and dispatches a
// bubbling standard event. The visible value and live ARIA selection follow.
func TestSelectExternalDraftRestoration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page, errs := gotoSelectDemo(t)
	trigger := page.Locator("#os-trigger")

	controls, err := trigger.GetAttribute("aria-controls")
	require.NoError(t, err)
	assert.Equal(t, "os-listbox", controls)

	_, err = page.Evaluate(`() => {
		const input = document.querySelector('#os');
		input.value = 'windows';
		input.dispatchEvent(new Event('change', { bubbles: true }));
	}`, nil)
	require.NoError(t, err)

	_, err = page.WaitForFunction(
		"() => document.querySelector('#os-trigger span').textContent.trim() === 'Windows'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	value, err := page.Locator("#os").InputValue()
	require.NoError(t, err)
	assert.Equal(t, "windows", value)

	_, err = page.Evaluate(`() => {
		const input = document.querySelector('#os');
		input.value = 'linux';
		input.dispatchEvent(new Event('input', { bubbles: true }));
	}`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		"() => document.querySelector('#os-trigger span').textContent.trim() === 'Linux'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	require.NoError(t, trigger.Click())
	require.NoError(t, page.Locator("#os-option-2").WaitFor())
	selected, err := page.Locator("#os-option-2").GetAttribute("aria-selected")
	require.NoError(t, err)
	assert.Equal(t, "true", selected)
	notSelected, err := page.Locator("#os-option-1").GetAttribute("aria-selected")
	require.NoError(t, err)
	assert.Equal(t, "false", notSelected)

	restoreButton := page.Locator("#restore-linux-draft")
	require.NoError(t, restoreButton.ScrollIntoViewIfNeeded())
	require.NoError(t, restoreButton.Click())
	_, err = page.WaitForFunction(
		"() => document.querySelector('#draft-os-trigger span').textContent.trim() === 'Linux'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "the documented draft restoration control should synchronize Select")
	draftValue, err := page.Locator("#draft-os").InputValue()
	require.NoError(t, err)
	assert.Equal(t, "linux", draftValue)

	pageErrs, consoleErrs := errs()
	assert.Empty(t, filterIgnorable(pageErrs), "unexpected page errors")
	assert.Empty(t, filterIgnorable(consoleErrs), "unexpected console errors")
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
