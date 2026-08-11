//go:build e2e && full

package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

type foundationThemeExpectation struct {
	name         string
	font         string
	radius       string
	lightPrimary string
	darkPrimary  string
	lightSurface string
	darkSurface  string
}

func TestFoundationRoutesResolveAuthenticatedThemeCascade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	routes := []struct {
		path  string
		probe string
	}{
		{path: "/getting-started", probe: "#getting-started-preview"},
		{path: "/components/button", probe: "#button-fragment button"},
		{path: "/components/text-input", probe: "#textinput-default input"},
		{path: "/components/table", probe: "#table-default table"},
	}
	themes := []foundationThemeExpectation{
		{
			name: "araihu", font: "Lato", radius: "4px",
			lightPrimary: "rgb(23, 59, 114)", darkPrimary: "rgb(199, 255, 74)",
			lightSurface: "rgb(243, 242, 233)", darkSurface: "rgb(7, 17, 31)",
		},
		{
			name: "modern", font: "Lato", radius: "4px",
			lightPrimary: "rgb(0, 0, 0)", darkPrimary: "rgb(255, 255, 255)",
		},
		{
			name: "goshtoso", font: "Inter", radius: "8px",
			lightPrimary: "rgb(33, 114, 163)", darkPrimary: "rgb(59, 206, 247)",
		},
	}

	for _, route := range routes {
		t.Run(strings.TrimPrefix(route.path, "/"), func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			failures := watchPageFailures(page)
			response, err := page.Goto(baseURL+route.path, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.Status())

			require.Equal(t, 1, mustLocatorCount(t, page.Locator(`head link[rel="stylesheet"][href^="/assets/styles.css"]`)))
			require.Equal(t, 0, mustLocatorCount(t, page.Locator(`head link[href*="/componentdocshell/assets/araihu.css"]`)))
			require.NoError(t, page.Locator(route.probe).First().WaitFor())

			for _, theme := range themes {
				for _, dark := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/dark=%t", theme.name, dark), func(t *testing.T) {
						computed, err := page.Evaluate(`([theme, dark]) => {
							const root = document.documentElement;
							root.setAttribute('data-theme', theme);
							root.classList.toggle('dark', dark);
							let probe = document.getElementById('foundation-theme-probe');
							if (!probe) {
								probe = document.createElement('div');
								probe.id = 'foundation-theme-probe';
								probe.hidden = true;
								document.body.append(probe);
							}
							probe.style.fontFamily = 'var(--font-body)';
							probe.style.borderRadius = 'var(--radius-radius)';
							probe.style.backgroundColor = dark ? 'var(--color-primary-dark)' : 'var(--color-primary)';
							probe.style.color = dark ? 'var(--color-surface-dark)' : 'var(--color-surface)';
							const style = getComputedStyle(probe);
							return {
								font: style.fontFamily,
								radius: style.borderRadius,
								primary: style.backgroundColor,
								surface: style.color,
							};
						}`, []any{theme.name, dark})
						require.NoError(t, err)
						values, ok := computed.(map[string]any)
						require.True(t, ok, "unexpected computed style result %T", computed)
						require.Contains(t, values["font"], theme.font)
						require.Equal(t, theme.radius, values["radius"])
						wantPrimary := theme.lightPrimary
						if dark {
							wantPrimary = theme.darkPrimary
						}
						require.Equal(t, wantPrimary, values["primary"])
						if theme.name == "araihu" {
							wantSurface := theme.lightSurface
							if dark {
								wantSurface = theme.darkSurface
							}
							require.Equal(t, wantSurface, values["surface"])
						}
					})
				}
			}

			assertFoundationRouteStates(t, page, route.path)
			require.NoError(t, page.SetViewportSize(320, 720))
			noPageOverflow, err := page.Evaluate(`() => document.documentElement.scrollWidth <= document.documentElement.clientWidth`, nil)
			require.NoError(t, err)
			require.Equal(t, true, noPageOverflow, "%s must not overflow at 320px", route.path)
			waitForPageSettled(t, page)
			failures.RequireEmpty(t)
		})
	}
}

func assertFoundationRouteStates(t *testing.T, page playwright.Page, route string) {
	t.Helper()
	_, err := page.Evaluate(`() => {
		const root = document.documentElement;
		root.setAttribute('data-theme', 'araihu');
		root.classList.remove('dark');
	}`, nil)
	require.NoError(t, err)

	switch route {
	case "/getting-started":
		cta := page.Locator(`#getting-started-preview a[href="https://github.com/araihu/goshtoso-getting-started"]`)
		require.NoError(t, cta.Focus())
		require.True(t, mustEvaluateBool(t, cta, `element => element.matches(':focus-visible')`))
	case "/components/button":
		primary := page.Locator("#button-fragment button").First()
		require.NoError(t, primary.Hover())
		require.NoError(t, primary.Focus())
		require.True(t, mustEvaluateBool(t, primary, `element => element.matches(':focus-visible')`))
		disabled, err := page.Locator("#button-disabled button").First().IsDisabled()
		require.NoError(t, err)
		require.True(t, disabled)
	case "/components/text-input":
		input := page.Locator("#textinput-default input")
		require.NoError(t, input.Focus())
		require.True(t, mustEvaluateBool(t, input, `element => element.matches(':focus-visible')`))
		require.Equal(t, "true", mustAttribute(t, page.Locator("#inputError"), "aria-invalid"))
		disabled, err := page.Locator("#textinput-disabled input").IsDisabled()
		require.NoError(t, err)
		require.True(t, disabled)
	case "/components/table":
		require.Equal(t, 3, mustLocatorCount(t, page.Locator("#table-default tbody tr")))
		require.Equal(t, 4, mustLocatorCount(t, page.Locator("#table-default thead th")))
	}
}

func mustEvaluateBool(t *testing.T, locator playwright.Locator, expression string) bool {
	t.Helper()
	value, err := locator.Evaluate(expression, nil)
	require.NoError(t, err)
	result, ok := value.(bool)
	require.True(t, ok, "unexpected bool result %T", value)
	return result
}
