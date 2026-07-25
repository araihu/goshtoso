package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaFormDemoPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/schema-form", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	title, err := page.Locator("main h1").TextContent()
	require.NoError(t, err)
	assert.Contains(t, title, "Schema Form")

	form := page.Locator("#schema-form-generated")
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

	schemaVersion := page.Locator("#schema-form-schema-version")
	require.NoError(t, schemaVersion.WaitFor())
	schemaVersionText, err := schemaVersion.TextContent()
	require.NoError(t, err)
	assert.Contains(t, schemaVersionText, "JSON Schema object subset")
	assert.Contains(t, schemaVersionText, "$schema is optional and ignored")

	allowList := page.Locator("#schema-form-allow-list")
	require.NoError(t, allowList.WaitFor())
	allowListText, err := allowList.TextContent()
	require.NoError(t, err)
	assert.Contains(t, allowListText, "Editable")
	assert.Contains(t, allowListText, "Managed")
	assert.Contains(t, allowListText, "Disabled")

	submitPrune := page.Locator("#schema-form-submit-prune")
	require.NoError(t, submitPrune.WaitFor())
	submitPruneText, err := submitPrune.TextContent()
	require.NoError(t, err)
	assert.Contains(t, submitPruneText, "values.serviceType")
	assert.Contains(t, submitPruneText, "values.internalToken")

	apiLink := page.Locator("[data-go-api-link]")
	href, err := apiLink.GetAttribute("href")
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://pkg.go.dev/github.com/araihu/goshtoso@"+goshtosoDocsVersion+"/components/schemaform",
		href,
	)
}
