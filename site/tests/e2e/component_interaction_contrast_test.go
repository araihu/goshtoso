//go:build e2e && full

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestCoreActionAndHelperContrastAcrossAcceptanceThemes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)

	cases := []struct {
		name     string
		path     string
		selector string
		hover    bool
		border   bool
		minimum  float64
	}{
		{name: "button", path: "/components/button", selector: "#button-fragment button", hover: true},
		{name: "danger button", path: "/components/button", selector: "#button-fragment button.bg-danger-action", hover: true},
		{name: "button-like link", path: "/components/link", selector: "#link-button a", hover: true},
		{name: "textarea helper", path: "/components/textarea", selector: "#textarea-required small"},
		{name: "text input feedback", path: "/components/text-input", selector: "#patternInput-feedback"},
		{name: "file input helper", path: "/components/fileinput", selector: "#fileinput-default small"},
		{name: "text input boundary", path: "/components/text-input", selector: "#textinput-default input", border: true, minimum: 3},
		{name: "required marker", path: "/components/form", selector: "#form-complete label > span"},
		{name: "alert status title", path: "/components/alert", selector: "#alert-default h3.text-success-text"},
	}

	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			for _, tc := range cases {
				t.Run(fmt.Sprintf("%s/%s/dark=%t", theme, tc.name, dark), func(t *testing.T) {
					_, err := page.Goto(baseURL+tc.path, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
					require.NoError(t, err)
					setThemeMode(t, page, theme, dark)
					locator := page.Locator(tc.selector).First()
					require.NoError(t, locator.WaitFor())

					base := measureRenderedContrast(t, locator)
					minimum := tc.minimum
					if minimum == 0 {
						minimum = 4.5
					}
					if tc.border {
						base = measureRenderedBorderContrast(t, locator)
					}
					require.GreaterOrEqualf(t, base.Ratio, minimum, "%s must meet its contrast target: %+v", tc.name, base)

					if tc.hover {
						require.NoError(t, locator.Hover())
						hovered := measureRenderedContrast(t, locator)
						require.Equal(t, "1", hovered.Opacity, "hover must not blend the whole control with its page surface")
						require.Truef(t, strings.Contains(hovered.Filter, "contrast"), "hover must preserve or strengthen contrast: %+v", hovered)
					}
				})
			}
		}
	}
}
