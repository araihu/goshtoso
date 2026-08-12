//go:build e2e && full

package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	demothemes "github.com/araihu/goshtoso/site/internal/themes"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type themeInventoryBrowserState struct {
	Path            string   `json:"path"`
	RootTheme       string   `json:"rootTheme"`
	AlpineTheme     string   `json:"alpineTheme"`
	SelectedKeys    []string `json:"selectedKeys"`
	Ownership       string   `json:"ownership"`
	Dark            bool     `json:"dark"`
	PrimaryColor    string   `json:"primaryColor"`
	SurfaceColor    string   `json:"surfaceColor"`
	Radius          string   `json:"radius"`
	SurfaceVariable string   `json:"surfaceVariable"`
}

// gotoThemePage navigates to /docs/theme, waits for Alpine, and resets all
// per-token overrides plus dark-mode state so each test starts from a known
// baseline. localStorage carries over across tests on the shared browser
// otherwise, masking regressions.
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
	// Wipe persisted overrides plus dark-mode so the test starts clean,
	// then reload so the wipe takes effect.
	_, err = page.Evaluate(`() => {
		localStorage.removeItem('themeOverrides');
		localStorage.removeItem('themeTitleFont');
		localStorage.removeItem('themeBodyFont');
		localStorage.removeItem('themeRadius');
		localStorage.removeItem('themeCssMode');
		localStorage.removeItem('themeCssFilter');
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
		const data = el && Alpine && typeof Alpine.$data === 'function' && Alpine.$data(el);
		return Alpine.__themePageRegistered === true && document.querySelector('#theme-page-data[type="application/json"]') &&
			data && typeof data.applyFont === 'function' && data.allTokens.length > 10;
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

func TestThemeInventory_BFullAllThemesAndDeepModes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)
	gotoThemePage(t, page)

	catalog := demothemes.All()
	require.Len(t, catalog, 16)
	require.Equal(t, 16, mustLocatorCount(t, page.Locator("button[data-theme-key][data-theme-ownership]")))

	for _, theme := range catalog {
		t.Run("Smoke/"+theme.Key, func(t *testing.T) {
			selectThemeInventoryCard(t, page, theme.Key)
			state := readThemeInventoryBrowserState(t, page)
			require.Equal(t, "/docs/theme", state.Path)
			require.Equal(t, theme.Key, state.RootTheme)
			require.Equal(t, theme.Key, state.AlpineTheme)
			require.Equal(t, []string{theme.Key}, state.SelectedKeys)
			require.Equal(t, string(theme.Ownership), state.Ownership)
			require.NotEmpty(t, state.SurfaceVariable)
		})
	}

	deepModes := []struct {
		name         string
		theme        string
		ownership    string
		dark         bool
		primaryColor string
		surfaceColor string
		radius       string
	}{
		{name: "AraiHuLight", theme: "araihu", ownership: "organization", primaryColor: "rgb(23, 59, 114)", surfaceColor: "rgb(243, 242, 233)", radius: "4px"},
		{name: "AraiHuDark", theme: "araihu", ownership: "organization", dark: true, primaryColor: "rgb(199, 255, 74)", surfaceColor: "rgb(7, 17, 31)", radius: "4px"},
		{name: "GoshtosoLight", theme: "goshtoso", ownership: "organization", primaryColor: "rgb(33, 114, 163)", surfaceColor: "rgb(255, 255, 255)", radius: "8px"},
		{name: "GoshtosoDark", theme: "goshtoso", ownership: "organization", dark: true, primaryColor: "rgb(59, 206, 247)", surfaceColor: "rgb(19, 25, 32)", radius: "8px"},
		{name: "MinimalLight", theme: "minimal", ownership: "generic", primaryColor: "rgb(0, 0, 0)", surfaceColor: "rgb(255, 255, 255)", radius: "0px"},
		{name: "MinimalDark", theme: "minimal", ownership: "generic", dark: true, primaryColor: "rgb(255, 255, 255)", surfaceColor: "oklch(0.145 0 none)", radius: "0px"},
	}
	for _, mode := range deepModes {
		t.Run("Deep/"+mode.name, func(t *testing.T) {
			selectThemeInventoryCard(t, page, mode.theme)
			setThemeInventoryDarkMode(t, page, mode.dark)
			state := readThemeInventoryBrowserState(t, page)
			require.Equal(t, "/docs/theme", state.Path)
			require.Equal(t, mode.theme, state.RootTheme)
			require.Equal(t, mode.theme, state.AlpineTheme)
			require.Equal(t, []string{mode.theme}, state.SelectedKeys)
			require.Equal(t, mode.ownership, state.Ownership)
			require.Equal(t, mode.dark, state.Dark)
			require.Equal(t, mode.primaryColor, state.PrimaryColor)
			require.Equal(t, mode.surfaceColor, state.SurfaceColor)
			require.Equal(t, mode.radius, state.Radius)
		})
	}
}

func selectThemeInventoryCard(t *testing.T, page playwright.Page, key string) {
	t.Helper()
	card := page.Locator(fmt.Sprintf("h2:has-text('Themes') ~ div button[data-theme-key='%s']", key)).First()
	require.NoError(t, card.Click())
	_, err := page.WaitForFunction(fmt.Sprintf(`() => {
		const root = document.querySelector('[x-data="themePage"]');
		const selected = document.querySelector('button[data-theme-key="%s"][aria-pressed="true"]');
		return document.documentElement.dataset.theme === %q && Alpine.$data(root).theme === %q && selected;
	}`, key, key, key), nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func setThemeInventoryDarkMode(t *testing.T, page playwright.Page, dark bool) {
	t.Helper()
	_, err := page.Evaluate(`dark => {
		const store = Alpine.store('darkMode');
		if (store.on !== dark) store.toggle();
	}`, dark)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`dark => Alpine.store('darkMode').on === dark && document.documentElement.classList.contains('dark') === dark`, dark,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func readThemeInventoryBrowserState(t *testing.T, page playwright.Page) themeInventoryBrowserState {
	t.Helper()
	raw, err := page.Evaluate(`() => {
		const root = document.querySelector('[x-data="themePage"]');
		const selected = [...document.querySelectorAll('button[data-theme-key][aria-pressed="true"]')];
		const probe = document.createElement('div');
		probe.className = 'bg-primary dark:bg-primary-dark rounded-radius';
		document.body.appendChild(probe);
		const probeStyle = getComputedStyle(probe);
		const rootStyle = getComputedStyle(document.documentElement);
		const state = {
			path: window.location.pathname,
			rootTheme: document.documentElement.dataset.theme || '',
			alpineTheme: Alpine.$data(root).theme,
			selectedKeys: selected.map(card => card.dataset.themeKey),
			ownership: selected[0]?.dataset.themeOwnership || '',
			dark: document.documentElement.classList.contains('dark') && Alpine.store('darkMode').on,
			primaryColor: probeStyle.backgroundColor,
			surfaceColor: '',
			radius: probeStyle.borderRadius,
			surfaceVariable: rootStyle.getPropertyValue('--color-surface').trim(),
		};
		const surfaceProbe = document.createElement('div');
		surfaceProbe.className = 'bg-surface dark:bg-surface-dark';
		document.body.appendChild(surfaceProbe);
		state.surfaceColor = getComputedStyle(surfaceProbe).backgroundColor;
		probe.remove();
		surfaceProbe.remove();
		return JSON.stringify(state);
	}`, nil)
	require.NoError(t, err)
	encoded, ok := raw.(string)
	require.True(t, ok, "unexpected browser state type %T", raw)
	var state themeInventoryBrowserState
	require.NoError(t, json.Unmarshal([]byte(encoded), &state))
	return state
}

func TestThemePage_FragmentNavBootstrapsData(t *testing.T) {
	page := newPage(t, sharedBrowser)
	dismissCookieBanner(t, page)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(message playwright.ConsoleMessage) {
		if message.Type() == "error" {
			jsErrors = append(jsErrors, message.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("a[href='/docs/theme']").First().Click())
	_, err = page.WaitForFunction(`() => {
		const root = document.querySelector('[x-data="themePage"]');
		const data = root && Alpine.$data(root);
		return root && root._x_dataStack && data && data.allThemes.length === 16 &&
			data.blocks.araihu && data.blocks.goshtoso && data.radiusMap['2xl'] === '1rem';
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "theme provider should parse inert data after fragment navigation")

	require.NoError(t, page.Locator("button[data-radius='2xl']").First().Click())
	_, err = page.WaitForFunction(
		"() => document.documentElement.style.getPropertyValue('--radius-radius') === '1rem'", nil)
	require.NoError(t, err, "fragment-loaded theme data should drive static behavior")
	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav theme page: %v", jsErrors)
}

func TestThemePage_FragmentNavigationPreservesActiveTheme(t *testing.T) {
	for _, activeTheme := range []string{"araihu", "minimal"} {
		t.Run(activeTheme, func(t *testing.T) {
			page := newIsolatedPage(t)
			dismissCookieBanner(t, page)

			_, err := page.Goto(baseURL + "/getting-started")
			require.NoError(t, err)
			require.NoError(t, waitForAlpine(page))
			_, err = page.Locator("html").Evaluate("(root, theme) => root.setAttribute('data-theme', theme)", activeTheme)
			require.NoError(t, err)

			require.NoError(t, page.Locator("a[href='/docs/theme']").First().Click())
			_, err = page.WaitForFunction(`() => {
				const root = document.querySelector('[x-data="themePage"]');
				const data = root && Alpine.$data(root);
				return window.location.pathname === '/docs/theme' && data && data.allThemes.length === 16;
			}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
			require.NoError(t, err)

			appliedTheme, err := page.Locator("html").GetAttribute("data-theme")
			require.NoError(t, err)
			require.Equal(t, activeTheme, appliedTheme, "opening the theme workbench must preserve the active theme")

			pressed, err := page.Locator(fmt.Sprintf("h2:has-text('Themes') ~ div button[data-theme-key='%s']", activeTheme)).First().GetAttribute("aria-pressed")
			require.NoError(t, err)
			require.Equal(t, "true", pressed, "theme workbench selection must match the applied theme")
		})
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
		})
	}
}

func TestThemePage_ThemeGrid_OnlyOneSelectedCard(t *testing.T) {
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

	selected, err := page.Evaluate(`() => {
		const buttons = [...document.querySelectorAll('h2 + .grid button[data-theme-key]')];
		const pressed = buttons.filter(button => button.getAttribute('aria-pressed') === 'true');
		const visibleChecks = buttons.flatMap(button => [...button.querySelectorAll('[data-theme-selected-icon]')])
			.filter(icon => getComputedStyle(icon).display !== 'none');
		return { pressed: pressed.length, checks: visibleChecks.length, key: pressed[0]?.dataset.themeKey || '' };
	}`, nil)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"pressed": 1, "checks": 1, "key": "dracula"}, selected,
		"exactly one theme card should expose selected state and check icon")
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

	// Default state: single mode + active Arai Hû theme rendered.
	require.NoError(t, output.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1500),
	}))
	text, err := output.TextContent()
	require.NoError(t, err)
	assert.Contains(t, text, "[data-theme=araihu]")
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
