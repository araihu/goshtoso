//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationPatternsPageAndResponsiveRecipes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1440, Height: 900},
	})

	_, err := page.Goto(baseURL+"/docs/application-patterns", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	title, err := page.Title()
	require.NoError(t, err)
	assert.Equal(t, "Application Patterns for Goshtoso", title)

	patterns := page.Locator("[data-application-pattern]")
	count, err := patterns.Count()
	require.NoError(t, err)
	assert.Equal(t, 4, count)

	for _, marker := range []string{
		"[data-pattern-preview]",
		"[data-pattern-problem]",
		"[data-pattern-components]",
		"[data-pattern-states]",
		"[data-pattern-390]",
		"[data-pattern-1440]",
		"[data-pattern-accessibility]",
		"[data-pattern-app-specific]",
		"[data-pattern-source-map]",
		"[data-pattern-done]",
	} {
		count, err = page.Locator(marker).Count()
		require.NoError(t, err)
		assert.Equal(t, 4, count, marker)
	}

	proofs := page.Locator("[data-field-proven-pattern]")
	count, err = proofs.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	for _, title := range []string{"Decision Queue", "Interruption-safe Workflow", "Content-first Review"} {
		require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: title}).WaitFor())
	}

	require.NoError(t, page.Locator(`a[href="/modules/app-shells"]`).Last().WaitFor())

	require.NoError(t, page.Locator("a[href='/docs/application-patterns'][aria-current='page']").WaitFor())
	require.NoError(t, page.Locator("#operations-pattern-table").WaitFor())
	require.NoError(t, page.Locator("#detail-workspace-preview [role='tablist']").WaitFor())
	require.NoError(t, page.Locator("#workflow-steps-desktop").WaitFor())

	desktopTableFits, err := page.Locator("#operations-list-preview .overflow-x-auto").Evaluate(
		"el => el.scrollWidth <= el.clientWidth",
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, true, desktopTableFits, "operations table actions should remain visible at 1440 px")

	for _, target := range []string{"#surface-brief-title", "#field-proven-title"} {
		require.NoError(t, page.Locator("#toc-list a[href='"+target+"']").Click())
		_, err = page.WaitForFunction("() => { const el = document.querySelector('"+target+"'); const header = document.querySelector('.component-doc-shell__header'); return Boolean(el && header) && el.getBoundingClientRect().top >= header.getBoundingClientRect().bottom; }", nil, playwright.PageWaitForFunctionOptions{
			Timeout: playwright.Float(2000),
		})
		require.NoError(t, err, "%s should remain visible below the sticky header", target)
	}

	visible, err := page.Locator("#app-shell-desktop-sidebar").IsVisible()
	require.NoError(t, err)
	assert.True(t, visible, "desktop app-shell navigation should be visible at 1440 px")
	visible, err = page.Locator("#workflow-steps-desktop").IsVisible()
	require.NoError(t, err)
	assert.True(t, visible, "desktop workflow progress should be visible at 1440 px")
	visible, err = page.Locator("#workflow-steps-mobile").IsVisible()
	require.NoError(t, err)
	assert.False(t, visible, "mobile workflow progress should be hidden at 1440 px")

	require.NoError(t, page.SetViewportSize(390, 844))
	documentFitsViewport, err := page.Evaluate("document.body.scrollWidth === document.body.clientWidth")
	require.NoError(t, err)
	assert.Equal(t, true, documentFitsViewport, "the shared documentation layout should not overflow at 390 px")

	visible, err = page.Locator("#app-shell-desktop-sidebar").IsVisible()
	require.NoError(t, err)
	assert.False(t, visible, "persistent app-shell navigation should collapse at 390 px")
	visible, err = page.Locator("#workflow-steps-mobile").IsVisible()
	require.NoError(t, err)
	assert.True(t, visible, "mobile workflow progress should be visible at 390 px")
	visible, err = page.Locator("#workflow-steps-desktop").IsVisible()
	require.NoError(t, err)
	assert.False(t, visible, "desktop workflow progress should be hidden at 390 px")

	scrollsHorizontally, err := page.Locator("#operations-list-preview .overflow-x-auto").Evaluate(
		"el => el.scrollWidth > el.clientWidth",
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, true, scrollsHorizontally, "operations table should preserve horizontal access at 390 px")
}

func TestApplicationPatternsSearchNavigatesToRecipePage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	_, err := page.Goto(baseURL+"/getting-started", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Keyboard().Press("Meta+K"))
	input := page.Locator("#docs-search-input")
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, input.Fill("application patterns"))

	result := page.Locator("#search-application-patterns:visible")
	require.NoError(t, result.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, result.Click())

	require.NoError(t, page.WaitForURL("**/docs/application-patterns"))
	require.NoError(t, page.Locator("#application-patterns-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
}
