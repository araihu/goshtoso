//go:build e2e && (full || codeblock)

package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeBlockStandaloneRuntimeWithoutAlpineOrHTMX(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)
	require.NoError(t, page.SetContent(`<main id="mount">
<div data-code-block><button hidden data-code-block-copy data-code-block-target="source" aria-label="Copy example code"><span data-code-block-copy-status role="status" aria-live="polite">Copy</span></button><div id="source">hello standalone</div></div>
</main>`))

	initiallyHidden, err := page.Locator("[data-code-block-copy]").IsHidden()
	require.NoError(t, err)
	require.True(t, initiallyHidden, "copy control must remain hidden without runtime")

	source, err := os.ReadFile("../../../assets/js/code-block.js")
	require.NoError(t, err)
	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{Content: new(string(source))})
	require.NoError(t, err)

	runtimes, err := page.Evaluate(`() => ({ alpine: typeof Alpine, htmx: typeof htmx })`, nil)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"alpine": "undefined", "htmx": "undefined"}, runtimes)

	button := page.Locator("[data-code-block-copy]")
	require.NoError(t, button.WaitFor())
	_, err = page.Evaluate(`() => Object.defineProperty(navigator, "clipboard", { configurable: true, value: { writeText: text => { window.__copied = text; return Promise.resolve(); } } })`, nil)
	require.NoError(t, err)
	require.NoError(t, button.Click())
	status, err := button.Locator("[data-code-block-copy-status]").TextContent()
	require.NoError(t, err)
	require.Equal(t, "Copied!", status)
	copied, err := page.Evaluate(`() => window.__copied`, nil)
	require.NoError(t, err)
	require.Equal(t, "hello standalone", copied)

	_, err = page.Evaluate(`() => Object.defineProperty(navigator, "clipboard", { configurable: true, value: undefined })`, nil)
	require.NoError(t, err)
	require.NoError(t, button.Click())
	status, err = button.Locator("[data-code-block-copy-status]").TextContent()
	require.NoError(t, err)
	require.Equal(t, "Unable to copy", status)
	require.Equal(t, "Copy example code", mustAttribute(t, button, "aria-label"))

	_, err = page.Evaluate(`() => {
		const fragment = document.createElement("div");
		fragment.innerHTML = '<button hidden data-code-block-copy data-code-block-target="late"><span data-code-block-copy-status>Copy</span></button><div id="late">late fragment</div>';
		document.querySelector("#mount").append(fragment);
		fragment.dispatchEvent(new CustomEvent("htmx:afterSwap", { bubbles: true }));
	}`, nil)
	require.NoError(t, err)
	lateHidden, err := page.Locator("#mount > div:last-child [data-code-block-copy]").IsHidden()
	require.NoError(t, err)
	require.False(t, lateHidden, "fragment-inserted copy control should be enabled")
}

// TestCodeblockCoverageDemo loads the codeblock demo directly and confirms the
// page renders multiple highlighted blocks, that clicking the copy button (the
// delegated runtime path) writes the original source to the clipboard and
// flips the accessible status label, and that the page produces no uncaught JS
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
	// delegated CodeBlock handler runs end to end.
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

func TestCodeblockReducedMotionStopsCopyTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))
	_, err := page.Goto(baseURL+"/components/codeblock", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	_, err = page.WaitForFunction(`() => {
		const buttons = Array.from(document.querySelectorAll("[data-code-block] button"));
		return buttons.length > 0 && buttons.every(button =>
			getComputedStyle(button).transitionProperty === "none"
		);
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "reduced-motion CodeBlock copy transitions should be disabled")
}
