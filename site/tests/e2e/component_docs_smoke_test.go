package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAllComponentDocsDirectLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	for _, entry := range catalog.ComponentPages() {
		entry := entry
		t.Run(entry.Active, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			failures := watchPageFailures(page)
			response, err := page.Goto(baseURL+entry.Path, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.Status())

			heading, err := page.Locator("main h1").TextContent()
			require.NoError(t, err)
			require.Equal(t, entry.Title, strings.TrimSpace(heading))

			descriptionCount, err := page.Locator("[data-component-description]").Count()
			require.NoError(t, err)
			require.Equal(t, 1, descriptionCount)

			previewCount, err := page.Locator("[data-demo-preview]").Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, previewCount, 1)

			codeCount, err := page.Locator("[data-demo-code]").Count()
			require.NoError(t, err)
			require.GreaterOrEqual(t, codeCount, 1)

			apiCount, err := page.Locator("[data-api-reference]").Count()
			require.NoError(t, err)
			require.Equal(t, 1, apiCount)
			failures.RequireEmpty(t)
		})
	}
}
