//go:build e2e && full

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// Inline markup can accidentally join words even when a template compiles.
// Check prose in the rendered DOM, excluding previews and structured widgets.
func TestDocsProseInlineSpacing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	page := newPage(t, sharedBrowser)
	for _, route := range []string{
		"/docs/iconpack", "/docs/component-model", "/docs/application-patterns",
		"/docs/internationalization", "/docs/internationalization/files",
		"/docs/internationalization/examples", "/docs/agents", "/docs/icon-catalog",
	} {
		t.Run(route, func(t *testing.T) {
			_, err := page.Goto(baseURL+route, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			require.NoError(t, err)
			issues, err := page.Evaluate(`() => {
                const issues = [];
                for (const paragraph of document.querySelectorAll('#main-content p')) {
                    if (paragraph.closest('pre, [data-component-preview], [data-pattern-preview], #expression-schema')) continue;
                    const walker = document.createTreeWalker(paragraph, NodeFilter.SHOW_TEXT);
                    let previous = null;
                    let node;
                    while ((node = walker.nextNode())) {
                        if (!node.textContent) continue;
                        if (previous && /[\p{L}\p{N}]$/u.test(previous.textContent) && /^[\p{L}\p{N}]/u.test(node.textContent) &&
                            (previous.parentElement.closest('code, a') || node.parentElement.closest('code, a'))) {
                            issues.push(paragraph.textContent);
                        }
                        previous = node;
                    }
                }
                return issues;
            }`)
			require.NoError(t, err)
			require.Empty(t, issues, "words joined across inline code or links")
		})
	}
}
