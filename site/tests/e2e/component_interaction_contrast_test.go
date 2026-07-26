package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

type renderedContrast struct {
	Color      string  `json:"color"`
	Background string  `json:"background"`
	Ratio      float64 `json:"ratio"`
	Opacity    string  `json:"opacity"`
	Filter     string  `json:"filter"`
}

func setThemeMode(t *testing.T, page playwright.Page, theme string, dark bool) {
	t.Helper()
	_, err := page.Evaluate(`([theme, dark]) => {
		const html = document.documentElement;
		let style = document.getElementById('contrast-test-no-transitions');
		if (!style) {
			style = document.createElement('style');
			style.id = 'contrast-test-no-transitions';
			style.textContent = '* { transition: none !important; }';
			document.head.append(style);
		}
		localStorage.setItem('theme', theme);
		html.setAttribute('data-theme', theme);
		html.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
	_, err = page.WaitForFunction("theme => document.documentElement.dataset.theme === theme", theme)
	require.NoError(t, err)
}

func measureRenderedContrast(t *testing.T, locator playwright.Locator) renderedContrast {
	return measureRenderedPropertyContrast(t, locator, "color")
}

func measureRenderedBorderContrast(t *testing.T, locator playwright.Locator) renderedContrast {
	return measureRenderedPropertyContrast(t, locator, "borderTopColor")
}

func measureRenderedPropertyContrast(t *testing.T, locator playwright.Locator, property string) renderedContrast {
	t.Helper()
	result, err := locator.Evaluate(`(el, property) => {
		const rgba = value => {
			const canvas = document.createElement('canvas');
			canvas.width = 1;
			canvas.height = 1;
			const context = canvas.getContext('2d', { willReadFrequently: true });
			context.clearRect(0, 0, 1, 1);
			context.fillStyle = '#000000';
			context.fillStyle = value;
			context.fillRect(0, 0, 1, 1);
			const [r, g, b, a] = context.getImageData(0, 0, 1, 1).data;
			return [r, g, b, a / 255];
		};
		const composite = (foreground, background) => {
			const alpha = foreground[3] + background[3] * (1 - foreground[3]);
			if (alpha === 0) return [0, 0, 0, 0];
			return [
				(foreground[0] * foreground[3] + background[0] * background[3] * (1 - foreground[3])) / alpha,
				(foreground[1] * foreground[3] + background[1] * background[3] * (1 - foreground[3])) / alpha,
				(foreground[2] * foreground[3] + background[2] * background[3] * (1 - foreground[3])) / alpha,
				alpha,
			];
		};
		const linear = value => {
			value /= 255;
			return value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
		};
		const luminance = value => {
			const [r, g, b] = value.slice(0, 3).map(linear);
			return 0.2126 * r + 0.7152 * g + 0.0722 * b;
		};
		const background = node => {
			const layers = [];
			for (let current = node; current; current = current.parentElement) {
				layers.push(rgba(getComputedStyle(current).backgroundColor));
			}
			return layers.reverse().reduce(
				(result, layer) => composite(layer, result),
				[255, 255, 255, 1],
			);
		};
		const style = getComputedStyle(el);
		const bg = background(el);
		const fg = composite(rgba(style[property]), bg);
		const foregroundLuminance = luminance(fg);
		const backgroundLuminance = luminance(bg);
		return {
			color: style[property],
			background: 'rgb(' + bg.slice(0, 3).map(Math.round).join(', ') + ')',
			ratio: (Math.max(foregroundLuminance, backgroundLuminance) + 0.05) /
				(Math.min(foregroundLuminance, backgroundLuminance) + 0.05),
			opacity: style.opacity,
			filter: style.filter,
		};
	}`, property)
	require.NoError(t, err)

	values, ok := result.(map[string]any)
	require.True(t, ok, "unexpected contrast result %T: %v", result, result)
	ratio := 0.0
	switch value := values["ratio"].(type) {
	case float64:
		ratio = value
	case int:
		ratio = float64(value)
	default:
		t.Fatalf("unexpected ratio %T: %v", values["ratio"], values["ratio"])
	}
	return renderedContrast{
		Color:      values["color"].(string),
		Background: values["background"].(string),
		Ratio:      ratio,
		Opacity:    values["opacity"].(string),
		Filter:     values["filter"].(string),
	}
}

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
