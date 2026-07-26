package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestCompositionComponentDemosRenderPublicContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tests := []struct {
		path     string
		selector string
		text     string
	}{
		{path: "/components/app-shell", selector: "#app-shell-default main", text: "Operations"},
		{path: "/components/page-header", selector: "#page-header-default h1", text: "Operations"},
		{path: "/components/toolbar", selector: "#toolbar-default [role='toolbar']", text: "Create incident"},
		{path: "/components/empty-state", selector: "#empty-state-default section", text: "No incidents found"},
		{path: "/components/skeleton", selector: "#skeleton-default [role='status']", text: "Loading incidents"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			_, err := page.Goto(baseURL+test.path, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)

			preview := page.Locator(test.selector)
			require.NoError(t, preview.WaitFor())
			if test.path == "/components/app-shell" {
				requireAppShellSkipLinkStartsClipped(t, page)
			}
			if test.path == "/components/skeleton" {
				require.Equal(t, test.text, mustAttribute(t, preview, "aria-label"))
			} else {
				require.NoError(t, preview.GetByText(test.text).WaitFor())
			}
		})
	}
}

func requireAppShellSkipLinkStartsClipped(t *testing.T, page playwright.Page) {
	t.Helper()

	clipped, err := page.Locator("#app-shell-default").Evaluate(`frame => {
		const link = frame.querySelector('a[href="#app-shell-default-main"]')
		if (!link) return false
		const frameBounds = frame.getBoundingClientRect()
		const linkBounds = link.getBoundingClientRect()
		return linkBounds.bottom <= frameBounds.top
	}`, nil)
	require.NoError(t, err)
	require.Equal(t, true, clipped)
	require.Equal(t, "-1", mustAttribute(t, page.Locator("#app-shell-default main"), "tabindex"))
}

func TestCardBodyDemoRendersBetweenDescriptionAndFooter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/card", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	body := page.Locator("#card-body [data-card-body]")
	require.NoError(t, body.WaitFor())
	require.NoError(t, body.GetByText("Server-rendered activity").WaitFor())
	require.NoError(t, page.Locator("#card-body [data-card-footer]").GetByText("View all activity").WaitFor())

	order, err := page.Locator("#card-body article").Evaluate(`element => {
		const description = element.querySelector("p")
		const body = element.querySelector("[data-card-body]")
		const footer = element.querySelector("[data-card-footer]")
		return Boolean(
			description &&
			body &&
			footer &&
			(description.compareDocumentPosition(body) & Node.DOCUMENT_POSITION_FOLLOWING) &&
			(body.compareDocumentPosition(footer) & Node.DOCUMENT_POSITION_FOLLOWING)
		)
	}`, nil)
	require.NoError(t, err)
	require.Equal(t, true, order)
}
