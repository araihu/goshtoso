package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCodeBlock_ChromaHighlighting verifies the codeblock component renders
// Chroma-highlighted HTML server-side (no Prism CDN) and the copy button
// still writes the original source text to the clipboard.
func TestCodeBlock_ChromaHighlighting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/button", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("renders chroma spans", func(t *testing.T) {
		// Wrapper div carries the .codeblock class.
		wrapper := page.Locator(".codeblock").First()
		count, err := wrapper.Count()
		require.NoError(t, err)
		require.Greater(t, count, 0, "expected at least one .codeblock wrapper")

		// Chroma class-mode output uses the "ch-" prefix.
		pre := page.Locator(".codeblock pre.ch-chroma").First()
		preVisible, err := pre.IsVisible()
		require.NoError(t, err)
		require.True(t, preVisible, "expected pre.ch-chroma to be visible")

		// Should contain at least one syntax-highlighted token span.
		anyToken := page.Locator(`.codeblock pre.ch-chroma span[class^="ch-"]`)
		tokenCount, err := anyToken.Count()
		require.NoError(t, err)
		require.Greater(t, tokenCount, 0, "expected at least one ch-* token span")
	})

	t.Run("no prism scripts or styles", func(t *testing.T) {
		html, err := page.Content()
		require.NoError(t, err)

		require.NotContains(t, strings.ToLower(html), "prism",
			"page HTML should not reference Prism")
		require.NotContains(t, html, "cdnjs.cloudflare.com",
			"page HTML should not load Prism from cdnjs")
		require.NotContains(t, html, `class="language-`,
			"page HTML should not contain Prism language-* classes")
	})

	t.Run("copy button preserves original source", func(t *testing.T) {
		// The textContent of the highlighted block (= the value the Alpine
		// handler copies) must match the original source character-for-char.
		got, err := page.Evaluate(`() => {
			const el = document.querySelector('.codeblock');
			return el ? el.textContent : null;
		}`, nil)
		require.NoError(t, err)
		text, ok := got.(string)
		require.True(t, ok, "expected string textContent")

		require.Contains(t, text, "button.Button",
			"expected source to contain the templ component call")
		require.Contains(t, text, "button.WithTone(button.ToneSecondary)",
			"expected source to contain the current tone option")

		// Stub navigator.clipboard.writeText. navigator.clipboard is a
		// read-only property in Chromium, so defineProperty is required —
		// plain assignment silently fails.
		_, err = page.Evaluate(`() => {
			window.__copied = null;
			Object.defineProperty(navigator, 'clipboard', {
				configurable: true,
				value: {
					writeText: (t) => { window.__copied = t; return Promise.resolve(); }
				}
			});
		}`, nil)
		require.NoError(t, err)

		// The copy button sits inside the outer x-data wrapper that also
		// contains the .codeblock div. Walk up one level.
		copyBtn := page.Locator(".codeblock").First().
			Locator("xpath=..").Locator("button").First()
		require.NoError(t, copyBtn.Click())

		copied, err := page.Evaluate(`() => window.__copied`, nil)
		require.NoError(t, err)
		require.Equal(t, text, copied,
			"clipboard payload should equal the rendered code textContent")
	})
}

func TestCodeBlock_DirectDemoPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/codeblock", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	for _, label := range []string{"main.go", "html", "css", "Install", "long.go"} {
		require.NoError(t, page.GetByText(label, playwright.PageGetByTextOptions{
			Exact: new(true),
		}).First().WaitFor())
	}

	for _, source := range []string{
		`fmt.Println("hello, world")`,
		`x-data="{ open: false }"`,
		`color: var(--color-primary)`,
		`go get github.com/a-h/templ`,
		`http.ListenAndServe(":8080", nil)`,
	} {
		require.NoError(t, page.Locator(".codeblock").Filter(playwright.LocatorFilterOptions{
			HasText: source,
		}).First().WaitFor())
	}

	scrollable := page.Locator(".codeblock").Filter(playwright.LocatorFilterOptions{
		HasText: `http.ListenAndServe(":8080", nil)`,
	}).First()
	maxHeight, err := scrollable.Evaluate("el => getComputedStyle(el).maxHeight", nil)
	require.NoError(t, err)
	assert.Equal(t, "180px", maxHeight)

	_, err = page.Evaluate(`() => {
		window.__copied = null;
		Object.defineProperty(navigator, 'clipboard', {
			configurable: true,
			value: {
				writeText: (t) => { window.__copied = t; return Promise.resolve(); }
			}
		});
	}`, nil)
	require.NoError(t, err)

	firstBlock := page.Locator(".codeblock").First()
	expected, err := firstBlock.TextContent()
	require.NoError(t, err)
	copyButton := firstBlock.Locator("xpath=..").Locator("button[aria-label='Copy main.go code']").First()
	require.NoError(t, copyButton.WaitFor())

	_, err = firstBlock.Evaluate(`el => Alpine.$data(el.parentElement).copyCode()`, nil)
	require.NoError(t, err)

	copied, err := page.Evaluate("() => window.__copied", nil)
	require.NoError(t, err)
	assert.Equal(t, expected, copied)
}
