//go:build e2e && (full || pageheader)

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageHeaderTitleHooksStayOnTheHeading(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/page-header", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	heading := page.Locator(`#page-header-custom-title h1[data-heading-voice="editorial"]`)
	require.NoError(t, heading.WaitFor())
	classes, err := heading.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, classes, "font-mono")
	assert.Contains(t, classes, "tracking-tight")
	assert.Equal(t, 0, mustCount(t, page.Locator(`#page-header-custom-title header[data-heading-voice]`)))
}
