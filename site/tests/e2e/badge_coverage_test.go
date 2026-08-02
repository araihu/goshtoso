//go:build e2e && (full || badge)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBadgeCoverageDemo complements TestBadgeComponentDemo with checks that the
// badge demo loads cleanly (no JS exceptions / console errors), that the
// notification and animating-dot helpers render their expected hooks, and that
// the badges survive a dark-mode toggle. filterIgnorable is defined in
// alert_coverage_test.go.
func TestBadgeCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)

	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/badge", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	t.Run("page heading renders", func(t *testing.T) {
		require.NoError(t, page.Locator("main").Filter(playwright.LocatorFilterOptions{
			HasText: "Badge",
		}).First().WaitFor())
	})

	t.Run("notification badge caps at 99 and exposes button labels", func(t *testing.T) {
		require.NoError(t, page.Locator("#badge-notification button[aria-label='notifications']").WaitFor())

		// NotificationBadge(count) caps display at "99" for count > 99.
		require.NoError(t, page.Locator("#badge-notification").GetByText("99", playwright.LocatorGetByTextOptions{
			Exact: new(true),
		}).WaitFor())

		// NotificationDot() renders a size-3 dot with no text.
		dots, err := page.Locator("#badge-notification span.size-3").Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, dots, 1)
	})

	t.Run("animating dots render aria-label and ping hook", func(t *testing.T) {
		dots := page.Locator("#badge-animating span[aria-label='notification']")
		count, err := dots.Count()
		require.NoError(t, err)
		assert.Equal(t, 6, count)

		// Each AnimatingDot emits an inner animate-ping span.
		ping, err := page.Locator("#badge-animating span.animate-ping").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, ping)
	})

	t.Run("indicator dots are aria-hidden decorations", func(t *testing.T) {
		count, err := page.Locator("#badge-indicators span[aria-hidden='true']").Count()
		require.NoError(t, err)
		assert.Equal(t, 6, count)
	})

	t.Run("badges survive a dark-mode toggle", func(t *testing.T) {
		before, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		beforeBool, _ := before.(bool)

		_, err = page.Evaluate(`() => { document.documentElement.classList.toggle('dark'); }`, nil)
		require.NoError(t, err)

		after, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
		require.NoError(t, err)
		afterBool, _ := after.(bool)
		assert.NotEqual(t, beforeBool, afterBool, "dark class should toggle")

		// Solid badges stay rendered after the theme switch.
		require.NoError(t, page.Locator("#badge-solid").GetByText("Primary", playwright.LocatorGetByTextOptions{
			Exact: new(true),
		}).First().WaitFor())
	})

	t.Run("soft semantic labels meet AA in the primary theme matrix", func(t *testing.T) {
		result, err := page.Evaluate(`() => {
			const root = document.documentElement
			const states = [
				{ name: 'goshtoso-light', theme: 'goshtoso', dark: false },
				{ name: 'goshtoso-dark', theme: 'goshtoso', dark: true },
				{ name: 'minimal-light', theme: 'minimal', dark: false },
				{ name: 'minimal-dark', theme: 'minimal', dark: true },
			]
			const canvas = document.createElement('canvas')
			canvas.width = 1
			canvas.height = 1
			const ctx = canvas.getContext('2d', { willReadFrequently: true })
			const pixel = colors => {
				ctx.clearRect(0, 0, 1, 1)
				for (const color of colors) {
					ctx.fillStyle = color
					ctx.fillRect(0, 0, 1, 1)
				}
				return [...ctx.getImageData(0, 0, 1, 1).data].slice(0, 3)
			}
			const luminance = rgb => {
				const linear = rgb.map(value => {
					const channel = value / 255
					return channel <= 0.03928 ? channel / 12.92 : Math.pow((channel + 0.055) / 1.055, 2.4)
				})
				return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2]
			}
			const ratio = (foreground, background) => {
				const a = luminance(foreground)
				const b = luminance(background)
				return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05)
			}
			return states.map(state => {
				root.dataset.theme = state.theme
				root.classList.toggle('dark', state.dark)
				const ratios = [...document.querySelectorAll('#badge-soft > div > span')].map(badge => {
					const inner = badge.firstElementChild
					const outerStyle = getComputedStyle(badge)
					const innerStyle = getComputedStyle(inner)
					const background = pixel([outerStyle.backgroundColor, innerStyle.backgroundColor])
					const foreground = pixel([innerStyle.color])
					return ratio(foreground, background)
				})
				return { name: state.name, minimum: Math.min(...ratios), ratios }
			})
		}`, nil)
		require.NoError(t, err)
		states, ok := result.([]any)
		require.True(t, ok, "unexpected contrast result: %T", result)
		for _, raw := range states {
			state, ok := raw.(map[string]any)
			require.True(t, ok, "unexpected contrast state: %T", raw)
			minimum, ok := state["minimum"].(float64)
			require.True(t, ok, "unexpected minimum ratio: %T", state["minimum"])
			require.GreaterOrEqual(t, minimum, 4.5, "%v soft badges: %v", state["name"], state["ratios"])
		}
	})

	require.Empty(t, pageErrors, "uncaught JS exceptions on badge demo")
	require.Empty(t, filterIgnorable(consoleErrors), "console errors on badge demo")
}
