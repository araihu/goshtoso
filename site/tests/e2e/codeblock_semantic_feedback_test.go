//go:build e2e && (full || codeblock)

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

type codeBlockFeedbackContrast struct {
	Color      string
	Background string
	Ratio      float64
}

func setCodeBlockThemeMode(t *testing.T, page playwright.Page, theme string, dark bool) {
	t.Helper()
	_, err := page.Evaluate(`([theme, dark]) => {
		const html = document.documentElement;
		html.setAttribute('data-theme', theme);
		html.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		`([theme, dark]) => document.documentElement.dataset.theme === theme && document.documentElement.classList.contains('dark') === dark`,
		[]any{theme, dark},
	)
	require.NoError(t, err)
}

func measureCodeBlockFeedbackContrast(t *testing.T, locator playwright.Locator) codeBlockFeedbackContrast {
	t.Helper()
	result, err := locator.Evaluate(`el => {
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
		const layers = [];
		for (let current = el; current; current = current.parentElement) {
			layers.push(rgba(getComputedStyle(current).backgroundColor));
		}
		const background = layers.reverse().reduce(
			(composited, layer) => composite(layer, composited),
			[255, 255, 255, 1],
		);
		const color = getComputedStyle(el).color;
		const foreground = composite(rgba(color), background);
		const foregroundLuminance = luminance(foreground);
		const backgroundLuminance = luminance(background);
		return {
			color,
			background: 'rgb(' + background.slice(0, 3).map(Math.round).join(', ') + ')',
			ratio: (Math.max(foregroundLuminance, backgroundLuminance) + 0.05) /
				(Math.min(foregroundLuminance, backgroundLuminance) + 0.05),
		};
	}`, nil)
	require.NoError(t, err)

	values, ok := result.(map[string]any)
	require.True(t, ok, "unexpected contrast result %T: %v", result, result)
	ratio, ok := values["ratio"].(float64)
	require.True(t, ok, "unexpected contrast ratio %T: %v", values["ratio"], values["ratio"])
	return codeBlockFeedbackContrast{
		Color:      values["color"].(string),
		Background: values["background"].(string),
		Ratio:      ratio,
	}
}

func codeBlockCopied(t *testing.T, wrapper playwright.Locator) bool {
	t.Helper()
	value, err := wrapper.Evaluate(`el => Alpine.$data(el).copied`, nil)
	require.NoError(t, err)
	copied, ok := value.(bool)
	require.True(t, ok, "unexpected copied state %T: %v", value, value)
	return copied
}

func TestCodeBlockCopiedFeedbackSemanticContrastAndClipboardStates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	for _, theme := range []string{"araihu", "modern", "goshtoso"} {
		for _, dark := range []bool{false, true} {
			name := fmt.Sprintf("%s/dark=%t", theme, dark)
			t.Run(name, func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				response, err := page.Goto(baseURL+"/components/codeblock", playwright.PageGotoOptions{
					WaitUntil: playwright.WaitUntilStateDomcontentloaded,
				})
				require.NoError(t, err)
				require.Equal(t, 200, response.Status())
				require.NoError(t, waitForAlpine(page))

				wrapper := page.Locator("[data-code-block]").First()
				require.NoError(t, wrapper.WaitFor())
				code := wrapper.Locator(".codeblock").First()
				expectedClipboardText, err := code.TextContent()
				require.NoError(t, err)
				copyButton := wrapper.Locator("button[aria-label='Copy main.go code']").First()
				require.NoError(t, copyButton.WaitFor())
				copiedIcon := wrapper.Locator(`svg[x-show="copied"]`).First()
				require.NoError(t, copiedIcon.WaitFor(playwright.LocatorWaitForOptions{
					State: playwright.WaitForSelectorStateAttached,
				}))

				setCodeBlockThemeMode(t, page, theme, dark)

				require.False(t, codeBlockCopied(t, wrapper), "default state must not report copied")
				require.Equal(t, "Copy", strings.TrimSpace(mustText(t, copyButton)))
				visible, err := copiedIcon.IsVisible()
				require.NoError(t, err)
				require.False(t, visible, "copied icon must be hidden by default")

				require.NoError(t, copyButton.Hover())
				_, err = page.WaitForFunction(
					`() => document.querySelector("[data-code-block] button[aria-label='Copy main.go code']")?.matches(':hover') === true`,
					nil,
				)
				require.NoError(t, err)
				hovered, err := copyButton.Evaluate(`el => el.matches(':hover')`, nil)
				require.NoError(t, err)
				require.Equal(t, true, hovered, "copy control must expose hover state")

				require.NoError(t, copyButton.Focus())
				focused, err := copyButton.Evaluate(`el => document.activeElement === el`, nil)
				require.NoError(t, err)
				require.Equal(t, true, focused, "copy control must retain keyboard focus")

				_, err = page.Evaluate(`() => {
					window.__codeBlockClipboardText = null;
					Object.defineProperty(navigator, 'clipboard', {
						configurable: true,
						value: {
							writeText: text => {
								window.__codeBlockClipboardText = text;
								return Promise.resolve();
							},
						},
					});
				}`)
				require.NoError(t, err)
				require.NoError(t, copyButton.Click())
				_, err = page.WaitForFunction(
					`() => Alpine.$data(document.querySelector('[data-code-block]')).copied === true`,
					nil,
				)
				require.NoError(t, err)
				require.Equal(t, expectedClipboardText, mustPageValue(t, page, "() => window.__codeBlockClipboardText"))
				require.Equal(t, "Copied!", strings.TrimSpace(mustText(t, copyButton)))
				require.True(t, codeBlockCopied(t, wrapper), "successful clipboard write must report copied")
				require.NoError(t, copiedIcon.WaitFor(playwright.LocatorWaitForOptions{
					State: playwright.WaitForSelectorStateVisible,
				}))
				visible, err = copiedIcon.IsVisible()
				require.NoError(t, err)
				require.True(t, visible, "copied icon must be visible after clipboard success")

				className, err := copiedIcon.GetAttribute("class")
				require.NoError(t, err)
				require.NotContains(t, className, "text-green-500")
				require.Contains(t, className, "text-success-text")
				require.Contains(t, className, "dark:text-success-text-dark")

				contrast := measureCodeBlockFeedbackContrast(t, copiedIcon)
				t.Logf("copied feedback contrast theme=%s dark=%t color=%s background=%s ratio=%.3f", theme, dark, contrast.Color, contrast.Background, contrast.Ratio)
				require.GreaterOrEqualf(t, contrast.Ratio, 4.5, "copied feedback must meet WCAG AA on actual header background: %+v", contrast)

				_, err = page.WaitForFunction(
					`() => Alpine.$data(document.querySelector('[data-code-block]')).copied === false`,
					nil,
					playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
				)
				require.NoError(t, err)
				require.Equal(t, "Copy", strings.TrimSpace(mustText(t, copyButton)))
				require.NoError(t, copiedIcon.WaitFor(playwright.LocatorWaitForOptions{
					State: playwright.WaitForSelectorStateHidden,
				}))
				visible, err = copiedIcon.IsVisible()
				require.NoError(t, err)
				require.False(t, visible, "copied icon must hide after reset")

				_, err = page.Evaluate(`() => {
					window.__codeBlockClipboardDenied = false;
					window.addEventListener('unhandledrejection', event => {
						if (event.reason && event.reason.name === 'NotAllowedError') {
							window.__codeBlockClipboardDenied = true;
							event.preventDefault();
						}
					}, { once: true });
					Object.defineProperty(navigator, 'clipboard', {
						configurable: true,
						value: {
							writeText: () => Promise.reject(new DOMException('Clipboard denied', 'NotAllowedError')),
						},
					});
				}`)
				require.NoError(t, err)
				require.NoError(t, copyButton.Click())
				_, err = page.WaitForFunction(`() => window.__codeBlockClipboardDenied === true`, nil)
				require.NoError(t, err)
				require.False(t, codeBlockCopied(t, wrapper), "clipboard denial must not report copied")
				require.Equal(t, "Copy", strings.TrimSpace(mustText(t, copyButton)))
				require.NoError(t, copiedIcon.WaitFor(playwright.LocatorWaitForOptions{
					State: playwright.WaitForSelectorStateHidden,
				}))
				visible, err = copiedIcon.IsVisible()
				require.NoError(t, err)
				require.False(t, visible, "copied icon must stay hidden after clipboard denial")
			})
		}
	}
}

func mustPageValue(t *testing.T, page playwright.Page, expression string) any {
	t.Helper()
	value, err := page.Evaluate(expression)
	require.NoError(t, err)
	return value
}
