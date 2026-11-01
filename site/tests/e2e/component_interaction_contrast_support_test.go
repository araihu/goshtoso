//go:build e2e && full

package e2e

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"testing"
)

type renderedContrast struct {
	Color      string  `json:"color"`
	Background string  `json:"background"`
	Ratio      float64 `json:"ratio"`
	Opacity    string  `json:"opacity"`
	Filter     string  `json:"filter"`
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
