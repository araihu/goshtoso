package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaFieldDemoPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/schema-field", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	title, err := page.Locator("main h1").TextContent()
	require.NoError(t, err)
	assert.Contains(t, title, "Schema Field")

	form := page.Locator("#schema-field-generated")
	require.NoError(t, form.GetByLabel("Replica count").WaitFor())
	require.NoError(t, form.GetByLabel("Image tag").WaitFor())
	require.NoError(t, form.GetByLabel("TLS enabled").WaitFor())
	require.NoError(t, form.GetByLabel("Service type").WaitFor())

	managed := form.GetByLabel("Team owner")
	require.NoError(t, managed.WaitFor())
	disabled, err := managed.IsDisabled()
	require.NoError(t, err)
	assert.True(t, disabled, "managed field should render disabled")

	hiddenCount, err := form.Locator("[name='values.internalToken']").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, hiddenCount, "disabled allow-list fields should be hidden")
}
