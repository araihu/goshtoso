//go:build e2e && (full || backdropelevation)

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

const backdropElevationKnownFontURL = "https://fonts.gstatic.com/s/montserrat/v31/JTU4jIg1_i6t8kCHKm4532VJOt5-QNFgpCtr6C41QYuEPgG_LbnLmMlbcDGy_rGKMQ.woff2"

func TestBackdropElevationRolesAcrossFoundationThemes(t *testing.T) {
	components := []struct {
		name  string
		route string
		check func(*testing.T, playwright.Page, bool)
	}{
		{name: "drawer", route: "/components/drawer", check: checkDrawerLayerRoles},
		{name: "modal", route: "/components/modal", check: checkModalLayerRoles},
		{name: "sidebar", route: "/components/sidebar", check: checkSidebarLayerRoles},
		{name: "search", route: "/components/search", check: checkSearchLayerRoles},
		{name: "kbd", route: "/components/kbd", check: checkKbdLayerRole},
	}
	themes := []string{"araihu", "goshtoso", "minimal"}

	for _, component := range components {
		component := component
		t.Run(component.name, func(t *testing.T) {
			page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
				Viewport: &playwright.Size{Width: 1440, Height: 1000},
			})
			failures := watchPageFailures(page)
			response, err := page.Goto(baseURL+component.route, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)
			require.Equal(t, 200, response.Status())
			if component.name != "kbd" {
				require.NoError(t, waitForAlpine(page))
			}

			session, err := page.Context().NewCDPSession(page)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, _ = session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 1})
				_ = session.Detach()
			})

			cellCount := 0
			for _, theme := range themes {
				for _, dark := range []bool{false, true} {
					for _, zoom := range []float64{1, 2} {
						name := fmt.Sprintf("%s/dark=%t/zoom=%.0f", theme, dark, zoom)
						t.Run(name, func(t *testing.T) {
							setLayerAppearance(t, page, session, theme, dark, zoom)
							component.check(t, page, dark)
							cellCount++
						})
					}
				}
			}
			require.Equal(t, 12, cellCount)
			requireBackdropElevationFailuresEmpty(t, failures)
		})
	}
}

func setLayerAppearance(t *testing.T, page playwright.Page, session playwright.CDPSession, theme string, dark bool, zoom float64) {
	t.Helper()
	_, err := session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": zoom})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`zoom => Math.abs((window.visualViewport?.scale || 1) - zoom) < 0.05`, zoom)
	require.NoError(t, err)
	_, err = page.Evaluate(`([theme, dark]) => {
		document.documentElement.dataset.theme = theme;
		document.documentElement.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
}

func checkDrawerLayerRoles(t *testing.T, page playwright.Page, dark bool) {
	trigger := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Open details"})
	overlay := page.Locator("#drawer-default div[aria-hidden='true']")
	panel := page.Locator("#drawer-default [role='dialog']")
	requireHidden(t, panel)
	require.NoError(t, trigger.Focus())
	requireFocused(t, trigger)
	require.NoError(t, page.Keyboard().Press("Enter"))
	requireVisible(t, overlay)
	requireVisible(t, panel)
	activeClass, alpha := "bg-backdrop/40", 40
	if dark {
		activeClass, alpha = "dark:bg-backdrop/60", 60
	}
	requireBackdropRole(t, overlay, []string{"bg-backdrop/40", "dark:bg-backdrop/60"}, activeClass, "--color-backdrop", alpha, panel)
	requireShadowRole(t, panel, []string{"shadow-elevation-raised"}, "shadow-elevation-raised", "--shadow-elevation-raised")
	require.NoError(t, page.Keyboard().Press("Escape"))
	requireHidden(t, panel)
}

func checkModalLayerRoles(t *testing.T, page playwright.Page, _ bool) {
	trigger := page.Locator("#modal-default button").Filter(playwright.LocatorFilterOptions{HasText: "Open Modal"})
	overlay := page.Locator("#modal-default [role='dialog']")
	panel := page.Locator("#modal-default [role='dialog'] > div")
	requireHidden(t, overlay)
	require.NoError(t, trigger.Focus())
	requireFocused(t, trigger)
	require.NoError(t, page.Keyboard().Press("Enter"))
	requireVisible(t, overlay)
	requireBackdropRole(t, overlay, []string{"bg-backdrop/20"}, "bg-backdrop/20", "--color-backdrop", 20, panel)
	require.NoError(t, page.Keyboard().Press("Escape"))
	requireHidden(t, overlay)
}

func checkSidebarLayerRoles(t *testing.T, page playwright.Page, _ bool) {
	root := page.Locator("#sidebar-overlay")
	trigger := root.Locator("button[aria-label='Open sidebar']")
	backdrop := root.Locator("div[aria-hidden='true']")
	panel := page.Locator("#sidebar-overlay-demo-panel")
	requireHidden(t, panel)
	require.NoError(t, trigger.Focus())
	requireFocused(t, trigger)
	require.NoError(t, page.Keyboard().Press("Enter"))
	requireVisible(t, backdrop)
	requireVisible(t, panel)
	requireBackdropRole(t, backdrop, []string{"bg-backdrop/50"}, "bg-backdrop/50", "--color-backdrop", 50, panel)
	require.NoError(t, page.Keyboard().Press("Escape"))
	requireHidden(t, panel)
	requireFocused(t, trigger)
}

func checkSearchLayerRoles(t *testing.T, page playwright.Page, dark bool) {
	trigger := page.Locator("#component-search button[aria-haspopup='dialog']")
	overlay := page.Locator("#component-search-dialog")
	panel := page.Locator("#component-search-dialog > div")
	input := page.Locator("#component-search-input")
	requireHidden(t, overlay)
	require.NoError(t, trigger.Focus())
	requireFocused(t, trigger)
	require.NoError(t, page.Keyboard().Press("Enter"))
	requireVisible(t, overlay)
	requireVisible(t, input)
	requireFocused(t, input)
	activeClass, activeRole, alpha := "bg-backdrop-surface/55", "--color-backdrop-surface", 55
	if dark {
		activeClass, activeRole, alpha = "dark:bg-backdrop/60", "--color-backdrop", 60
	}
	requireBackdropRole(t, overlay, []string{"bg-backdrop-surface/55", "dark:bg-backdrop/60"}, activeClass, activeRole, alpha, panel)
	requireShadowRole(t, panel, []string{"shadow-elevation-overlay"}, "shadow-elevation-overlay", "--shadow-elevation-overlay")
	require.NoError(t, page.Keyboard().Press("Escape"))
	requireHidden(t, overlay)
}

func checkKbdLayerRole(t *testing.T, page playwright.Page, dark bool) {
	key := page.Locator("#kbd-frequently-used kbd").First()
	requireVisible(t, key)
	activeClass, activeRole := "shadow-elevation-control", "--shadow-elevation-control"
	if dark {
		activeClass, activeRole = "dark:shadow-elevation-control-dark", "--shadow-elevation-control-dark"
	}
	requireShadowRole(t, key, []string{"shadow-elevation-control", "dark:shadow-elevation-control-dark"}, activeClass, activeRole)
}

func requireBackdropRole(t *testing.T, backdrop playwright.Locator, requiredClasses []string, activeClass, role string, alpha int, panel playwright.Locator) {
	t.Helper()
	measurement := measureLayer(t, backdrop, role, alpha)
	classes := strings.Fields(measurement.className)
	for _, className := range requiredClasses {
		require.Contains(t, classes, className)
	}
	require.Contains(t, classes, activeClass)
	for _, raw := range []string{"bg-black/20", "bg-black/40", "bg-black/50", "dark:bg-black/60", "bg-surface-dark/55"} {
		require.NotContains(t, classes, raw)
	}
	require.NotEmpty(t, measurement.roleValue)
	require.NotEmpty(t, measurement.expectedBackground)
	require.NotEqual(t, "transparent", measurement.background)
	require.NotEqual(t, "rgba(0, 0, 0, 0)", measurement.background)
	require.Equal(t, measurement.expectedBackground, measurement.background,
		"active class %s must compute from %s at %d%%", activeClass, role, alpha)
	panelBackground, err := panel.Evaluate(`element => getComputedStyle(element).backgroundColor`, nil)
	require.NoError(t, err)
	require.NotEqual(t, measurement.background, panelBackground)
}

func requireShadowRole(t *testing.T, element playwright.Locator, requiredClasses []string, activeClass, role string) {
	t.Helper()
	measurement := measureLayer(t, element, role, 0)
	classes := strings.Fields(measurement.className)
	for _, className := range requiredClasses {
		require.Contains(t, classes, className)
	}
	require.Contains(t, classes, activeClass)
	for _, raw := range []string{"shadow-sm", "shadow-xl", "shadow-2xl", "shadow-outline/30", "shadow-black/20", "dark:shadow-black/20"} {
		require.NotContains(t, classes, raw)
	}
	require.NotEmpty(t, measurement.roleValue)
	require.NotEmpty(t, measurement.twShadow)
	require.Equal(t, strings.Join(strings.Fields(measurement.roleValue), " "), strings.Join(strings.Fields(measurement.twShadow), " "),
		"active class %s must bind --tw-shadow to %s", activeClass, role)
	require.NotEmpty(t, measurement.shadow)
	require.NotEqual(t, "none", measurement.shadow)
}

type layerMeasurement struct {
	className          string
	background         string
	expectedBackground string
	shadow             string
	twShadow           string
	roleValue          string
}

func measureLayer(t *testing.T, element playwright.Locator, role string, alpha int) layerMeasurement {
	t.Helper()
	value, err := element.Evaluate(`(element, expectation) => {
		const style = getComputedStyle(element);
		const root = getComputedStyle(document.documentElement);
		const reference = document.createElement('div');
		if (expectation.alpha > 0) {
			reference.style.backgroundColor = 'color-mix(in oklab, var(' + expectation.role + ') ' + expectation.alpha + '%, transparent)';
			document.body.appendChild(reference);
		}
		const expectedBackground = expectation.alpha > 0 ? getComputedStyle(reference).backgroundColor : '';
		reference.remove();
		return {
			className: element.getAttribute('class') || '',
			background: style.backgroundColor,
			expectedBackground,
			shadow: style.boxShadow,
			twShadow: style.getPropertyValue('--tw-shadow').trim(),
			roleValue: root.getPropertyValue(expectation.role).trim(),
		};
	}`, map[string]any{"role": role, "alpha": alpha})
	require.NoError(t, err)
	values := value.(map[string]any)
	return layerMeasurement{
		className:          values["className"].(string),
		background:         values["background"].(string),
		expectedBackground: values["expectedBackground"].(string),
		shadow:             values["shadow"].(string),
		twShadow:           values["twShadow"].(string),
		roleValue:          values["roleValue"].(string),
	}
}

func requireVisible(t *testing.T, locator playwright.Locator) {
	t.Helper()
	require.NoError(t, locator.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
}

func requireHidden(t *testing.T, locator playwright.Locator) {
	t.Helper()
	require.NoError(t, locator.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
}

func requireFocused(t *testing.T, locator playwright.Locator) {
	t.Helper()
	expect := playwright.NewPlaywrightAssertions(2000)
	require.NoError(t, expect.Locator(locator).ToBeFocused())
}

func requireBackdropElevationFailuresEmpty(t *testing.T, failures *pageFailures) {
	t.Helper()
	failures.mu.Lock()
	messages := append([]string(nil), failures.messages...)
	failures.mu.Unlock()

	kept := make([]string, 0, len(messages))
	for _, message := range filterIgnorable(messages) {
		knownHTTP404 := message == "HTTP response: 404 : "+backdropElevationKnownFontURL ||
			message == "HTTP response: 404 Not Found: "+backdropElevationKnownFontURL
		knownConsole404 := strings.HasPrefix(message,
			"console error: Failed to load resource: the server responded with a status of 404 () [url="+backdropElevationKnownFontURL+" ")
		knownRequestAbort := strings.HasPrefix(message,
			"request failed: GET "+backdropElevationKnownFontURL+": net::ERR_ABORTED")
		if knownHTTP404 || knownConsole404 || knownRequestAbort {
			continue
		}
		kept = append(kept, message)
	}
	require.Empty(t, kept, "unexpected page failures: %s", strings.Join(kept, "; "))
}
