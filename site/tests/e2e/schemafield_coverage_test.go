package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemafieldCoverageDemo exercises render paths the existing page test does
// not: the fallback section (object groups, array tagslist, managed nested
// input) and live enum selection / boolean defaults in the generated section.
// Deterministic — uses locator waits, no sleeps.
func TestSchemafieldCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/schema-field", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	title, err := page.Locator("main h1").TextContent()
	require.NoError(t, err)
	assert.Contains(t, title, "Schema Field")

	// --- Generated section: enum prefill + update, boolean default. ---
	generated := page.Locator("#schema-field-generated")
	require.NoError(t, generated.WaitFor())

	serviceType := generated.Locator("select[name='values.serviceType']")
	require.NoError(t, serviceType.WaitFor())
	// Current value "LoadBalancer" wins over the "ClusterIP" default.
	val, err := serviceType.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "LoadBalancer", val, "enum should prefill the current value")

	_, err = serviceType.SelectOption(playwright.SelectOptionValues{
		Values: &[]string{"NodePort"},
	})
	require.NoError(t, err)
	val, err = serviceType.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "NodePort", val, "enum selection should update")

	// Boolean default true renders a checked checkbox (tlsEnabled).
	tls := generated.Locator("input[name='values.tlsEnabled']")
	require.NoError(t, tls.WaitFor())
	checked, err := tls.IsChecked()
	require.NoError(t, err)
	assert.True(t, checked, "boolean default true should render checked")

	// --- Fallback section: object group, managed nested input, array tagslist. ---
	fallback := page.Locator("#schema-field-fallback")
	require.NoError(t, fallback.WaitFor())

	// resources is a 2-child object → fieldset/legend section.
	require.NoError(t, fallback.Locator("fieldset").First().WaitFor())

	// resources.cpu is managed → rendered disabled.
	cpu := fallback.Locator("input[name='values.resources.cpu']")
	require.NoError(t, cpu.WaitFor())
	disabled, err := cpu.IsDisabled()
	require.NoError(t, err)
	assert.True(t, disabled, "managed nested field should be disabled")

	// resources.memory is an editable nested string input.
	mem := fallback.Locator("input[name='values.resources.memory']")
	require.NoError(t, mem.WaitFor())
	memDisabled, err := mem.IsDisabled()
	require.NoError(t, err)
	assert.False(t, memDisabled, "non-managed nested field should be editable")

	// zones array → tagslist seeds the default elements.
	fallbackText, err := fallback.TextContent()
	require.NoError(t, err)
	assert.True(t, strings.Contains(fallbackText, "us-east-1a"), "tagslist should seed array defaults")
	assert.True(t, strings.Contains(fallbackText, "us-east-1b"), "tagslist should seed array defaults")

	require.Empty(t, consoleErrors, "no console errors on schema-field demo: %v", consoleErrors)
}
