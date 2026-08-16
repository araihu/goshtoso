//go:build e2e && (full || tooltipradius)

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tooltipRadiusCase struct {
	id        string
	variant   string
	placement string
}

func TestTooltipSemanticRadiusAcrossFoundationThemes(t *testing.T) {
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1440, Height: 1000},
	})
	failures := watchPageFailures(page)
	response, err := page.Goto(baseURL+"/components/tooltip", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())

	session, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 1})
		_ = session.Detach()
	})

	themes := []struct {
		name   string
		radius string
	}{
		{name: "araihu", radius: "4px"},
		{name: "goshtoso", radius: "8px"},
		{name: "minimal", radius: "0px"},
	}
	cases := []tooltipRadiusCase{
		{id: "demoTop", variant: "default", placement: "bottom-full"},
		{id: "demoBottom", variant: "default", placement: "top-full"},
		{id: "demoLeft", variant: "default", placement: "right-full"},
		{id: "demoRight", variant: "default", placement: "left-full"},
		{id: "richTop", variant: "rich", placement: "bottom-full"},
		{id: "richBottom", variant: "rich", placement: "top-full"},
		{id: "clickTop", variant: "click", placement: "bottom-full"},
		{id: "clickBottom", variant: "click", placement: "top-full"},
	}
	panelIDs := make([]string, 0, len(cases))
	for _, testCase := range cases {
		panelIDs = append(panelIDs, testCase.id)
	}

	cellCount := 0
	for _, theme := range themes {
		for _, dark := range []bool{false, true} {
			for _, zoom := range []float64{1, 2} {
				mode := fmt.Sprintf("%s/dark=%t/zoom=%.0f", theme.name, dark, zoom)
				t.Run(mode, func(t *testing.T) {
					_, err := session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": zoom})
					require.NoError(t, err)
					_, err = page.WaitForFunction(`zoom => Math.abs((window.visualViewport?.scale || 1) - zoom) < 0.05`, zoom)
					require.NoError(t, err)
					_, err = page.Evaluate(`([theme, dark]) => {
						document.documentElement.dataset.theme = theme;
						document.documentElement.classList.toggle('dark', dark);
					}`, []any{theme.name, dark})
					require.NoError(t, err)
					_, err = page.WaitForFunction(`args => args.ids.every(id => {
						const panel = document.getElementById(id);
						return panel && getComputedStyle(panel).borderTopLeftRadius === args.radius;
					})`, map[string]any{"ids": panelIDs, "radius": theme.radius}, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(1000)})
					require.NoErrorf(t, err, "%s Tooltip radius did not settle to %s", mode, theme.radius)

					for _, testCase := range cases {
						panel := page.Locator("#" + testCase.id)
						require.Equal(t, 1, tooltipRadiusCount(t, panel), "%s panel %s", mode, testCase.id)
						value, err := panel.Evaluate(`element => {
							const style = getComputedStyle(element);
							return {
								className: element.getAttribute('class') || '',
								topLeft: style.borderTopLeftRadius,
								topRight: style.borderTopRightRadius,
								bottomRight: style.borderBottomRightRadius,
								bottomLeft: style.borderBottomLeftRadius,
							};
						}`, nil)
						require.NoError(t, err)
						measurement := value.(map[string]any)
						className := measurement["className"].(string)
						assert.Containsf(t, strings.Fields(className), "rounded-radius", "%s %s/%s semantic class", mode, testCase.variant, testCase.placement)
						assert.NotContainsf(t, strings.Fields(className), "rounded-sm", "%s %s/%s fixed radius", mode, testCase.variant, testCase.placement)
						assert.Containsf(t, strings.Fields(className), testCase.placement, "%s %s placement", mode, testCase.variant)
						for _, corner := range []string{"topLeft", "topRight", "bottomRight", "bottomLeft"} {
							assert.Equalf(t, theme.radius, measurement[corner], "%s %s/%s %s radius", mode, testCase.variant, testCase.placement, corner)
						}
						cellCount++
					}
				})
			}
		}
	}

	require.Equal(t, 96, cellCount)
	_, err = session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 1})
	require.NoError(t, err)
	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
	t.Logf("Tooltip semantic-radius matrix passed %d cells", cellCount)
}

func tooltipRadiusCount(t *testing.T, locator playwright.Locator) int {
	t.Helper()
	count, err := locator.Count()
	require.NoError(t, err)
	return count
}
