//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextInput_ReadonlyAttribute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/text-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	input := page.Locator("#readonlyInput")
	require.NoError(t, input.WaitFor())

	// Should have the readonly attribute
	readonly, err := input.GetAttribute("readonly")
	require.NoError(t, err)
	assert.NotNil(t, readonly, "readonly input should have readonly attribute")

	// Should have the initial value
	val, err := input.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "Cannot change this", val)

	// Attempt to type — value should not change because of readonly
	err = input.Focus()
	require.NoError(t, err)
	err = page.Keyboard().Type("new text")
	require.NoError(t, err)
	val, err = input.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "Cannot change this", val, "readonly input should not accept typing")
}

func TestTextInput_PatternAttribute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/text-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	input := page.Locator("#patternInput")
	require.NoError(t, input.WaitFor())

	// Should have the pattern attribute
	pattern, err := input.GetAttribute("pattern")
	require.NoError(t, err)
	assert.Equal(t, "t[a-z0-9]{6}", pattern)
}

func TestTextInput_PatternValidationFeedback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/text-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	input := page.Locator("#patternInput")
	require.NoError(t, input.WaitFor())
	feedback := page.Locator("#patternInput-feedback")

	require.NoError(t, input.PressSequentially("fasds"))
	require.NoError(t, feedback.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	text, err := feedback.TextContent()
	require.NoError(t, err)
	assert.Contains(t, text, "Needs a leading t plus 6 lowercase letters or numbers")

	require.NoError(t, input.Fill(""))
	require.NoError(t, input.PressSequentially("tabc123"))
	text, err = feedback.TextContent()
	require.NoError(t, err)
	assert.Contains(t, text, "Looks good")
	valid, err := input.Evaluate("el => el.checkValidity()", nil)
	require.NoError(t, err)
	assert.Equal(t, true, valid)
}

func TestTextInput_MaxLengthAttribute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/text-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	input := page.Locator("#maxlengthInput")
	require.NoError(t, input.WaitFor())

	// Should have the maxlength attribute
	maxlength, err := input.GetAttribute("maxlength")
	require.NoError(t, err)
	assert.Equal(t, "7", maxlength)

	// Type more than 7 chars — browser should truncate
	err = input.Fill("")
	require.NoError(t, err)
	err = input.PressSequentially("abcdefghij")
	require.NoError(t, err)
	val, err := input.InputValue()
	require.NoError(t, err)
	assert.Len(t, val, 7, "maxlength should limit input to 7 characters")
}
