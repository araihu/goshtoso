package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCodeblockCoverageDemo loads the codeblock demo directly and confirms the
// page renders multiple highlighted blocks, that clicking the copy button (the
// @click="copyCode()" path) writes the original source to the clipboard and
// flips the Alpine "copied" label, and that the page produces no uncaught JS
// exceptions or console errors.
func TestCodeblockCoverageDemo(t *testing.T) {
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

	_, err := page.Goto(baseURL+"/components/codeblock", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	assert.True(t, strings.Contains(mainText, "Code Block"), "main should mention Code Block")

	// Several live codeblock instances should render with highlighted markup.
	blocks := page.Locator(".codeblock")
	count, err := blocks.Count()
	require.NoError(t, err)
	require.Greater(t, count, 1, "expected multiple .codeblock instances")

	tokens := page.Locator(`.codeblock pre.ch-chroma span[class^="ch-"]`)
	tokenCount, err := tokens.Count()
	require.NoError(t, err)
	require.Greater(t, tokenCount, 0, "expected highlighted token spans")

	// Stub the clipboard: navigator.clipboard is read-only in Chromium, so
	// defineProperty is required.
	_, err = page.Evaluate(`() => {
		window.__copied = null;
		Object.defineProperty(navigator, 'clipboard', {
			configurable: true,
			value: { writeText: (t) => { window.__copied = t; return Promise.resolve(); } }
		});
	}`, nil)
	require.NoError(t, err)

	// Target the labeled "main.go" block and click its real copy button so the
	// @click="copyCode()" handler runs end to end.
	codeEl := page.Locator(".codeblock").Filter(playwright.LocatorFilterOptions{
		HasText: `fmt.Println("hello, world")`,
	}).First()
	require.NoError(t, codeEl.WaitFor())

	expected, err := codeEl.TextContent()
	require.NoError(t, err)

	wrapper := codeEl.Locator("xpath=..")
	copyBtn := wrapper.Locator("button[aria-label='Copy main.go code']").First()
	require.NoError(t, copyBtn.WaitFor())
	require.NoError(t, copyBtn.Click())

	// Clipboard receives the original (un-highlighted) source text.
	copied, err := page.Evaluate("() => window.__copied", nil)
	require.NoError(t, err)
	assert.Equal(t, expected, copied, "clipboard payload should equal rendered code text")

	// The Alpine label flips to "Copied!" after the click.
	require.NoError(t, wrapper.GetByText("Copied!", playwright.LocatorGetByTextOptions{
		Exact: new(true),
	}).First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	assert.Empty(t, pageErrors, "uncaught JS exceptions on codeblock demo")
	assert.Empty(t, consoleErrors, "console errors on codeblock demo")
}
