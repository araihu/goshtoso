package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gotoThemePage navigates to /docs/theme, waits for Alpine, and resets all
// per-token overrides + theme/dark-mode state so each test starts from a
// known baseline. localStorage carries over across tests on the shared
// browser otherwise, masking regressions.
func gotoThemePage(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+"/docs/theme", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined' && Alpine.store('darkMode')`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err)
	// Wipe persisted overrides + theme + dark-mode so the test starts clean,
	// then reload so the wipe takes effect.
	_, err = page.Evaluate(`() => {
		localStorage.removeItem('themeOverrides');
		localStorage.removeItem('themeTitleFont');
		localStorage.removeItem('themeBodyFont');
		localStorage.removeItem('themeRadius');
		localStorage.removeItem('themeCssMode');
		localStorage.removeItem('themeCssFilter');
		localStorage.setItem('theme', 'minimal');
		localStorage.setItem('darkMode', 'false');
	}`, nil)
	require.NoError(t, err)
	_, err = page.Reload(playwright.PageReloadOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined' && Alpine.store('darkMode')`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err)
	// Wait for the themePage Alpine.data factory to finish init — the page's
	// x-model/x-data bindings come from this component, so interactions
	// before it's wired don't trigger watchers.
	_, err = page.WaitForFunction(`() => {
		const el = document.querySelector('[x-data="themePage"]');
		return el && Alpine && typeof Alpine.$data === 'function' && Alpine.$data(el) && typeof Alpine.$data(el).applyFont === 'function';
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

// inlineStyleVar reads an inline-style custom property set on <html>.
// getComputedStyle resolves through the cascade, so it can't tell us whether
// the page actually wrote an override or it's coming from a theme block —
// inline style does.
func inlineStyleVar(t *testing.T, page playwright.Page, name string) string {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => document.documentElement.style.getPropertyValue('%s')`, name), nil)
	require.NoError(t, err)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func TestThemePage_Loads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// All major section headings must appear.
	for _, heading := range []string{"Theme Settings", "Themes", "Typography", "Border", "Colors", "Theme Showcase", "Color Contrast Checker", "Get CSS Code"} {
		count, err := page.Locator(fmt.Sprintf("h1:has-text('%s'), h2:has-text('%s')", heading, heading)).Count()
		require.NoError(t, err)
		assert.Greater(t, count, 0, "missing section heading %q", heading)
	}
}

func TestThemePage_ThemeGrid_SwitchesTheme(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	themes := []string{"arctic", "neo-brutalism", "halloween", "90s", "dracula"}
	for _, key := range themes {
		t.Run(key, func(t *testing.T) {
			// Theme picker cards live inside the grid section. Scope the
			// click so we don't fire the header dropdown's button instead.
			card := page.Locator(fmt.Sprintf("h2:has-text('Themes') ~ div button[data-theme-key='%s']", key)).First()
			require.NoError(t, card.Click())
			_, err := page.WaitForFunction(fmt.Sprintf(
				`() => document.documentElement.getAttribute('data-theme') === '%s'`, key), nil,
				playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
			require.NoError(t, err)

			stored, err := page.Evaluate(`() => localStorage.getItem('theme')`, nil)
			require.NoError(t, err)
			assert.Equal(t, key, stored)
		})
	}
}

func TestThemePage_ThemeGrid_OnlyOneActiveDot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// Pick a non-default theme so the regression that flipped every card
	// "active" on load can't pass by coincidence.
	require.NoError(t, page.Locator("h2:has-text('Themes') ~ div button[data-theme-key='dracula']").First().Click())
	_, err := page.WaitForFunction(`() => document.documentElement.getAttribute('data-theme') === 'dracula'`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	// Count theme-grid inner dots that are visibly rendered (no `hidden` class).
	visible, err := page.Evaluate(`() => {
		const dots = document.querySelectorAll('h2 + .grid button[data-theme-key] .size-2.rounded-full');
		let n = 0;
		dots.forEach(d => { if (!d.classList.contains('hidden')) n++; });
		return n;
	}`, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 1, visible, "exactly one theme card should show an active dot")
}

func TestThemePage_Typography_AppliesFontVar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// Set titleFont via the Alpine component directly. SelectOption flake-races
	// against the x-model wiring on cold loads in headless Chromium; setting
	// the reactive property is equivalent (Alpine.$data drives the same code
	// path that the change-event listener does) and avoids the race.
	setFont := func(prop, val string) {
		_, err := page.Evaluate(fmt.Sprintf(
			`() => { Alpine.$data(document.querySelector('[x-data="themePage"]')).%s = %q; }`, prop, val), nil)
		require.NoError(t, err)
	}

	setFont("titleFont", "Inter")
	_, err := page.WaitForFunction(`() => document.documentElement.style.getPropertyValue('--font-title').includes('Inter')`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
	assert.Contains(t, inlineStyleVar(t, page, "--font-title"), "Inter")

	setFont("bodyFont", "Roboto")
	_, err = page.WaitForFunction(`() => document.documentElement.style.getPropertyValue('--font-body').includes('Roboto')`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	// Resetting to the empty sentinel clears the inline override.
	setFont("titleFont", "")
	_, err = page.WaitForFunction(`() => document.documentElement.style.getPropertyValue('--font-title') === ''`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}

func TestThemePage_BorderRadius_AppliesVar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	cases := []struct {
		key  string
		want string
	}{
		{"none", "0"},
		{"sm", "0.25rem"},
		{"2xl", "1rem"},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			btn := page.Locator(fmt.Sprintf("button[data-radius='%s']", c.key)).First()
			require.NoError(t, btn.Click())
			_, err := page.WaitForFunction(fmt.Sprintf(
				`() => document.documentElement.style.getPropertyValue('--radius-radius') === '%s'`, c.want), nil,
				playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
			require.NoError(t, err)
		})
	}
}

func TestThemePage_ColorPalette_TogglesOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// Open the Primary color picker via its Select trigger.
	require.NoError(t, page.Locator("#color-primary-trigger").Click())
	// Primary palette should be rendered with swatches visible.
	tile := page.Locator(`#palette-primary button[data-cls="blue-700"]`).First()
	require.NoError(t, tile.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1500),
	}))

	// Open Secondary — clicking its trigger fires the Primary dropdown's
	// click-outside handler, so Primary should close (single-open invariant).
	// The open Primary dropdown (absolute, z-30) overlaps the Secondary row and
	// would intercept a real pointer click, so dispatch the click via JS: it
	// still bubbles, so both Secondary's @click and Primary's click.outside fire.
	_, err := page.Evaluate(`() => document.querySelector('#color-secondary-trigger').click()`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator(`#palette-secondary button[data-cls="blue-700"]`).First().
		WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1500),
		}))
	require.NoError(t, tile.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(1500),
	}))
}

func TestThemePage_ColorPalette_PickAppliesVar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	require.NoError(t, page.Locator("#color-primary-trigger").Click())
	require.NoError(t, page.Locator(`#palette-primary button[data-cls="fuchsia-500"]`).First().Click())

	// applyColors writes `var(--color-fuchsia-500)` inline on <html>.
	_, err := page.WaitForFunction(
		`() => document.documentElement.style.getPropertyValue('--color-primary').includes('--color-fuchsia-500')`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	// The trigger value (driven by x-text="classLabel('primary')") should now
	// read "Fuchsia-500".
	_, err = page.WaitForFunction(
		`() => document.querySelector('#color-primary-trigger').textContent.includes('Fuchsia-500')`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func TestThemePage_ColorPalette_BlackWhiteApply(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	require.NoError(t, page.Locator("#color-primary-trigger").Click())
	// The two preset chips in the palette are tagged data-cls="white" / "black".
	require.NoError(t, page.Locator(`#palette-primary button[data-cls="white"]`).First().Click())
	_, err := page.WaitForFunction(
		`() => document.documentElement.style.getPropertyValue('--color-primary').includes('--color-white')`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func TestThemePage_DarkModeGroup_Visibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	light := page.Locator("h3:has-text('Light Mode Colors')").First()
	dark := page.Locator("h3:has-text('Dark Mode Colors')").First()

	require.NoError(t, light.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1500),
	}))
	hidden, err := dark.IsHidden()
	require.NoError(t, err)
	assert.True(t, hidden, "Dark group should be hidden while in light mode")

	// Toggle dark mode and confirm the swap.
	require.NoError(t, page.Locator("#darkModeToggleBtn").Click())
	_, err = page.WaitForFunction(`() => Alpine.store('darkMode').on === true`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	require.NoError(t, dark.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1500),
	}))
	hidden, err = light.IsHidden()
	require.NoError(t, err)
	assert.True(t, hidden, "Light group should be hidden while in dark mode")
}

func TestThemePage_ContrastMatrix_Renders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// At least one ratio cell should render once Alpine populates the matrix.
	_, err := page.WaitForFunction(
		`() => Array.from(document.querySelectorAll('table tbody td'))
			.some(td => td.textContent.trim().endsWith(':1'))`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	// Counts the ":1" suffix appearances in the matrix — one per non-base token.
	count, err := page.Evaluate(`() => Array.from(
		document.querySelectorAll('table tbody td')).filter(td => td.textContent.trim().endsWith(':1')).length`, nil)
	require.NoError(t, err)
	switch n := count.(type) {
	case int:
		assert.Greater(t, n, 10, "contrast matrix should report ratios for most tokens")
	case float64:
		assert.Greater(t, int(n), 10, "contrast matrix should report ratios for most tokens")
	}
}

func TestThemePage_ContrastMatrix_BaseColorSwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// Capture an initial ratio cell value, switch the base color, expect
	// the rendered ratios to change.
	before, err := page.Evaluate(`() => {
		// Row 3 is the ratio row (rows 1-2 are the swatch + inverted swatch).
		const cells = document.querySelectorAll('table tbody tr:nth-child(3) td');
		return Array.from(cells).map(c => c.textContent.trim()).join('|');
	}`, nil)
	require.NoError(t, err)

	_, err = page.Locator("h2:has-text('Color Contrast Checker') ~ div select").First().
		SelectOption(playwright.SelectOptionValues{Values: &[]string{"surface"}})
	require.NoError(t, err)

	_, err = page.WaitForFunction(fmt.Sprintf(`() => {
		// Row 3 is the ratio row (rows 1-2 are the swatch + inverted swatch).
		const cells = document.querySelectorAll('table tbody tr:nth-child(3) td');
		return Array.from(cells).map(c => c.textContent.trim()).join('|') !== %q;
	}`, before), nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func TestThemePage_CSSExport_FilterAndMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	output := page.Locator("#theme-css-output code")

	// Default state: single mode + active theme = minimal single block rendered.
	require.NoError(t, output.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1500),
	}))
	text, err := output.TextContent()
	require.NoError(t, err)
	assert.Contains(t, text, "[data-theme=minimal]")
	assert.NotContains(t, text, "@layer base")

	// Switch to All Themes + Multiple. cssFilter is the Goshtoso Select
	// component (custom dropdown), so open the trigger and click the option.
	require.NoError(t, page.Locator("#cssFilter-trigger").Click())
	require.NoError(t, page.GetByRole("option", playwright.PageGetByRoleOptions{
		Name: "All Themes", Exact: new(true),
	}).Click())
	require.NoError(t, page.Locator("button:has-text('Multiple Themes')").First().Click())

	_, err = page.WaitForFunction(`() => {
		const code = document.querySelector('#theme-css-output code');
		return code && code.textContent.includes('@layer base') && code.textContent.includes('[data-theme=dracula]');
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(1500)})
	require.NoError(t, err)

	// Switch back to a specific theme in single mode.
	require.NoError(t, page.Locator("button:has-text('Single Theme')").First().Click())
	require.NoError(t, page.Locator("#cssFilter-trigger").Click())
	require.NoError(t, page.GetByRole("option", playwright.PageGetByRoleOptions{
		Name: "Arctic", Exact: new(true),
	}).Click())
	_, err = page.WaitForFunction(`() => {
		const code = document.querySelector('#theme-css-output code');
		return code && code.textContent.includes('[data-theme=arctic]') && !code.textContent.includes('@layer base');
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(1500)})
	require.NoError(t, err)
}

func TestThemePage_Showcase_UpdateBanner_Dismiss(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	banner := page.Locator("text=Update Available").Locator("xpath=ancestor::div[1]")
	require.NoError(t, banner.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1500),
	}))

	// "Later" closes the banner without firing the primary action.
	require.NoError(t, page.Locator("button:has-text('Later')").First().Click())
	require.NoError(t, page.Locator("text=Update Available").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(1500),
	}))
}

func TestThemePage_Showcase_Toggles_Flip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// Each toggle starts on. Clicking flips the position of the inner knob.
	firstToggle := page.Locator("h5:has-text('Toggles')").Locator("xpath=ancestor::div[1]").Locator("button[aria-label='Toggle']").First()
	before, err := firstToggle.GetAttribute("class")
	require.NoError(t, err)
	require.NoError(t, firstToggle.Click())
	_, err = page.WaitForFunction(fmt.Sprintf(`() => {
		const t = document.querySelector("button[aria-label='Toggle']");
		return t && t.getAttribute('class') !== %q;
	}`, before), nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func TestThemePage_ResetAll_ClearsOverrides(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// Apply a couple of overrides first.
	require.NoError(t, page.Locator("#color-primary-trigger").Click())
	require.NoError(t, page.Locator(`#palette-primary button[data-cls="emerald-500"]`).First().Click())
	_, err := page.WaitForFunction(
		`() => document.documentElement.style.getPropertyValue('--color-primary').includes('--color-emerald-500')`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	// Border Radius buttons are SVG-only — target by data-radius.
	require.NoError(t, page.Locator("button[data-radius='xl']").First().Click())
	_, err = page.WaitForFunction(
		`() => document.documentElement.style.getPropertyValue('--radius-radius') === '0.75rem'`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	// Reset is a two-step confirm: the link reveals an inline prompt, then
	// "Yes, reset" performs the wipe.
	require.NoError(t, page.Locator("button:has-text('Reset all customizations')").First().Click())
	require.NoError(t, page.Locator("button:has-text('Yes, reset')").First().Click())

	_, err = page.WaitForFunction(`() => {
		const s = document.documentElement.style;
		return s.getPropertyValue('--color-primary') === '' && s.getPropertyValue('--radius-radius') === '';
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func TestThemePage_OverridesSurviveReload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	require.NoError(t, page.Locator("#color-primary-trigger").Click())
	require.NoError(t, page.Locator(`#palette-primary button[data-cls="rose-600"]`).First().Click())
	_, err := page.WaitForFunction(
		`() => document.documentElement.style.getPropertyValue('--color-primary').includes('--color-rose-600')`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)

	_, err = page.Reload(playwright.PageReloadOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		`() => document.documentElement.style.getPropertyValue('--color-primary').includes('--color-rose-600')`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "override should survive a reload via localStorage")
}

func TestThemePage_PaletteVariablesExist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	gotoThemePage(t, page)

	// Regression: when Tailwind v4's JIT didn't see `bg-{hue}-{shade}` in
	// source it omitted the matching `--color-{hue}-{shade}` variable. The
	// palette popover then rendered the affected tiles as blank cells.
	missing, err := page.Evaluate(`() => {
		const hues = ['red','orange','amber','yellow','lime','green','emerald','teal','cyan','sky','blue','indigo','violet','purple','fuchsia','pink','rose','slate','gray','zinc','neutral','stone'];
		const shades = ['50','100','200','300','400','500','600','700','800','900','950'];
		const cs = getComputedStyle(document.documentElement);
		const missing = [];
		hues.forEach(h => shades.forEach(s => {
			if (!cs.getPropertyValue('--color-' + h + '-' + s).trim()) missing.push(h + '-' + s);
		}));
		return missing;
	}`, nil)
	require.NoError(t, err)
	if arr, ok := missing.([]any); ok {
		assert.Empty(t, arr, "every Tailwind palette CSS variable should resolve")
	}
}
