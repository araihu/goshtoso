package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkComponentDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/link", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#link-fragment").WaitFor())

	defaultLink := page.Locator("#link-default a").First()
	href, err := defaultLink.GetAttribute("href")
	require.NoError(t, err)
	assert.Equal(t, "#", href)

	className, err := defaultLink.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, className, "text-primary")

	iconCount, err := page.Locator("#link-icon a svg[aria-hidden='true']").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, iconCount)

	buttonLink := page.Locator("#link-button a").First()
	role, err := buttonLink.GetAttribute("role")
	require.NoError(t, err)
	assert.Equal(t, "button", role)

	buttonClass, err := buttonLink.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, buttonClass, "bg-primary")
}
