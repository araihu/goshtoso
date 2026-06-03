package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCombobox_ClientMode_NoHTTPOnToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	// Record all requests to the combobox demo endpoints; start BEFORE goto.
	var demoReqs []string
	page.On("request", func(req playwright.Request) {
		url := req.URL()
		if strings.Contains(url, "/combobox/industry/") || strings.Contains(url, "/combobox/skills/") {
			demoReqs = append(demoReqs, url)
		}
	})

	_, err := page.Goto(baseURL+"/components/combobox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	// Wait for Alpine.
	_, err = page.Evaluate(`() => new Promise(resolve => {
		if (typeof Alpine !== 'undefined') return resolve(true);
		document.addEventListener('alpine:init', () => resolve(true), { once: true });
	})`, nil)
	require.NoError(t, err)

	// Baseline: discard any stray requests before the toggle.
	demoReqs = demoReqs[:0]

	trigger := page.Locator("#industry-trigger").First()
	require.NoError(t, trigger.Click())
	time.Sleep(300 * time.Millisecond)

	option := page.Locator(`#industry [data-combobox-option][data-value="tech"]`).First()
	require.NoError(t, option.Click())
	time.Sleep(300 * time.Millisecond)

	// 1. Option marked selected client-side.
	aria, err := page.Evaluate(`() => document.querySelector('#industry [data-combobox-option][data-value="tech"]').getAttribute('aria-selected')`, nil)
	require.NoError(t, err)
	assert.Equal(t, "true", aria, "option marked selected client-side")

	// 2. Hidden input created client-side.
	hidden, err := page.Locator(`#industry input[type=hidden][name="industry"][value="tech"]`).Count()
	require.NoError(t, err)
	assert.Equal(t, 1, hidden, "hidden input created client-side")

	// 3. Zero HTTP hits for client-mode toggle.
	assert.Empty(t, demoReqs, "client-mode toggle must not fire any HTTP request (got: %v)", demoReqs)

	// 4. Outer trigger label updated.
	label, err := page.Locator(`#industry-trigger-label`).TextContent()
	require.NoError(t, err)
	assert.Equal(t, "Technology", label)

	// 5. Single-select closes the dropdown after choosing an option.
	expanded, err := trigger.GetAttribute("aria-expanded")
	require.NoError(t, err)
	assert.Equal(t, "false", expanded)

	visible, err := page.Locator(`#industry [data-combobox-body]`).First().IsVisible()
	require.NoError(t, err)
	assert.False(t, visible, "single-select dropdown should close after selection")
}

func TestCombobox_ClientMode_MultiToggleAndClear(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	var skillsReqs []string
	page.On("request", func(req playwright.Request) {
		url := req.URL()
		if strings.Contains(url, "/combobox/skills/") {
			skillsReqs = append(skillsReqs, url)
		}
	})

	_, err := page.Goto(baseURL+"/components/combobox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	_, err = page.Evaluate(`() => new Promise(resolve => {
		if (typeof Alpine !== 'undefined') return resolve(true);
		document.addEventListener('alpine:init', () => resolve(true), { once: true });
	})`, nil)
	require.NoError(t, err)
	skillsReqs = skillsReqs[:0]

	trigger := page.Locator("#skills-trigger").First()
	require.NoError(t, trigger.Click())
	time.Sleep(300 * time.Millisecond)

	// Toggle two options by row click, then one by clicking directly on the visible checkbox square.
	require.NoError(t, page.Locator(`#skills [data-combobox-option][data-value="go"]`).First().Click())
	time.Sleep(100 * time.Millisecond)

	expanded, err := trigger.GetAttribute("aria-expanded")
	require.NoError(t, err)
	assert.Equal(t, "true", expanded, "multi-select should stay open after selecting an option")

	require.NoError(t, page.Locator(`#skills [data-combobox-option][data-value="rust"]`).First().Click())
	time.Sleep(100 * time.Millisecond)
	kotlinCheckbox := page.Locator(`#skills [data-combobox-option][data-value="kt"] input[type=checkbox]`).First()
	kotlinBox, err := kotlinCheckbox.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, kotlinBox)
	require.NoError(t, page.Mouse().Click(kotlinBox.X+kotlinBox.Width/2, kotlinBox.Y+kotlinBox.Height/2))
	time.Sleep(100 * time.Millisecond)

	// Trigger label shows "3 selected".
	label, err := page.Locator(`#skills-trigger-label`).TextContent()
	require.NoError(t, err)
	assert.Equal(t, "3 selected", label)

	// Three hidden inputs present.
	hiddenCount, err := page.Locator(`#skills input[type=hidden][name="skills"]`).Count()
	require.NoError(t, err)
	assert.Equal(t, 3, hiddenCount)

	kotlinChecked, err := kotlinCheckbox.IsChecked()
	require.NoError(t, err)
	assert.True(t, kotlinChecked, "direct checkbox click should keep selected option visually checked")

	// Deselect "go" by clicking it again (toggle off).
	require.NoError(t, page.Locator(`#skills [data-combobox-option][data-value="go"]`).First().Click())
	time.Sleep(100 * time.Millisecond)

	// Label still shows multiple selections; two hidden inputs remain.
	label, err = page.Locator(`#skills-trigger-label`).TextContent()
	require.NoError(t, err)
	assert.Equal(t, "2 selected", label)

	hiddenCount, err = page.Locator(`#skills input[type=hidden][name="skills"]`).Count()
	require.NoError(t, err)
	assert.Equal(t, 2, hiddenCount)

	// Re-select "go" so we have 2 selected again, then use clear-all.
	require.NoError(t, page.Locator(`#skills [data-combobox-option][data-value="go"]`).First().Click())
	time.Sleep(100 * time.Millisecond)

	// Clear-all button visible now.
	clearBtn := page.Locator(`#skills [data-combobox-clear]`).First()
	require.NoError(t, clearBtn.Click())
	time.Sleep(100 * time.Millisecond)

	label, err = page.Locator(`#skills-trigger-label`).TextContent()
	require.NoError(t, err)
	assert.Equal(t, "Pick some skills", label)

	hiddenCount, err = page.Locator(`#skills input[type=hidden][name="skills"]`).Count()
	require.NoError(t, err)
	assert.Equal(t, 0, hiddenCount)

	// Zero HTTP throughout.
	assert.Empty(t, skillsReqs, "client-mode multi-toggle + clear must not fire any HTTP (got: %v)", skillsReqs)
}
