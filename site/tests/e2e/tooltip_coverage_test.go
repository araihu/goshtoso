package e2e

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTooltipCoverageDemo complements the legacy tooltip E2E tests with a
// deterministic pass over the documented demo variants, Alpine click state, and
// console/page-error checks.
func TestTooltipCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/tooltip", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForTooltipAlpine(page))

	require.NoError(t, page.Locator("#tooltip-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	require.Contains(t, mainText, "Tooltip")

	t.Run("hover positions render with ARIA wiring", func(t *testing.T) {
		positions := map[string]string{
			"demoTop":    "bottom-full",
			"demoBottom": "top-full",
			"demoLeft":   "right-full",
			"demoRight":  "left-full",
		}
		for id, wantClass := range positions {
			trigger := page.Locator("[aria-describedby='" + id + "']").First()
			require.NoError(t, trigger.ScrollIntoViewIfNeeded())
			require.Equal(t, id, tooltipAttribute(t, trigger, "aria-describedby"))

			tip := page.Locator("#" + id)
			require.Equal(t, "tooltip", tooltipAttribute(t, tip, "role"))
			require.Contains(t, tooltipAttribute(t, tip, "class"), wantClass)
		}
	})

	t.Run("rich tooltip exposes heading and description text", func(t *testing.T) {
		rich := page.Locator("#richTop")
		require.NoError(t, rich.ScrollIntoViewIfNeeded())
		require.Contains(t, tooltipAttribute(t, rich, "class"), "flex w-64 flex-col")
		require.NoError(t, rich.Locator("span.text-sm.font-medium").WaitFor())
		require.NoError(t, rich.Locator("p.text-balance").WaitFor())

		text, err := rich.TextContent()
		require.NoError(t, err)
		require.Contains(t, text, "Tooltip top")
		require.Contains(t, text, "A rich tooltip that contains longer text")
	})

	t.Run("click trigger toggles live Alpine visibility", func(t *testing.T) {
		trigger := page.Locator("[aria-describedby='clickTop']").First()
		require.NoError(t, trigger.ScrollIntoViewIfNeeded())

		visible, err := page.Evaluate("() => getComputedStyle(document.querySelector('#clickTop')).display !== 'none'", nil)
		require.NoError(t, err)
		require.False(t, visible.(bool), "click tooltip should start hidden")

		require.NoError(t, trigger.Click())
		_, err = page.WaitForFunction(
			"() => getComputedStyle(document.querySelector('#clickTop')).display !== 'none'",
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)

		require.NoError(t, page.Locator("main").Click(playwright.LocatorClickOptions{
			Position: &playwright.Position{X: 10, Y: 10},
		}))
		_, err = page.WaitForFunction(
			"() => getComputedStyle(document.querySelector('#clickTop')).display === 'none'",
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
	})

	require.Empty(t, jsErrors, "no JS console/page errors on tooltip demo: %v", jsErrors)
}

func waitForTooltipAlpine(page playwright.Page) error {
	_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	return err
}

func tooltipAttribute(t *testing.T, loc playwright.Locator, name string) string {
	t.Helper()

	value, err := loc.GetAttribute(name)
	require.NoError(t, err)
	return value
}

func TestTooltipCustomButtonTriggerFocusWiring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/examples/logs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForTooltipAlpine(page))

	wrapper := page.Locator("#log-pause-tip").Locator("xpath=preceding-sibling::span[1]")
	trigger := wrapper.Locator("button").First()
	require.NoError(t, trigger.WaitFor())

	describedBy, err := trigger.GetAttribute("aria-describedby")
	require.NoError(t, err)
	assert.Equal(t, "log-pause-tip", describedBy,
		"the actual focusable custom trigger must describe the tooltip")
	wrapperTabIndex, err := wrapper.Evaluate("el => el.tabIndex", nil)
	require.NoError(t, err)
	assert.Equal(t, -1, wrapperTabIndex,
		"a focusable custom child must not leave a duplicate wrapper tab stop")

	require.NoError(t, trigger.Focus())
	assert.Eventually(t, func() bool {
		opacity, evalErr := page.Evaluate(
			"() => getComputedStyle(document.querySelector('#log-pause-tip')).opacity",
			nil,
		)
		return evalErr == nil && opacity == "1"
	}, 2*time.Second, 50*time.Millisecond,
		"keyboard focus on the custom button should reveal its tooltip")

	_, err = page.Evaluate(`() => {
		const content = document.createElement('div')
		content.id = 'fallback-tooltip-content'
		content.setAttribute('role', 'tooltip')
		document.body.appendChild(content)
		const root = document.createElement('span')
		root.id = 'fallback-tooltip-trigger'
		root.dataset.tooltipContentId = content.id
		root.textContent = 'More information'
		document.body.appendChild(root)
		window.goshtosoInitTooltipTrigger(root)
	}`, nil)
	require.NoError(t, err)

	fallback := page.Locator("#fallback-tooltip-trigger")
	require.Equal(t, "0", tooltipAttribute(t, fallback, "tabindex"))
	require.Equal(t, "button", tooltipAttribute(t, fallback, "role"))
	require.Equal(t, "fallback-tooltip-content", tooltipAttribute(t, fallback, "aria-describedby"))
}

func TestTooltipCustomTriggerInitializerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/examples/logs", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForTooltipAlpine(page))

	t.Run("repeated fallback initialization keeps one key listener", func(t *testing.T) {
		clicks, evalErr := page.Evaluate(`() => {
			const root = document.createElement('span')
			root.dataset.tooltipContentId = 'repeated-fallback-content'
			root.dataset.tooltipActivation = 'click'
			root.dataset.clicks = '0'
			root.addEventListener('click', () => {
				root.dataset.clicks = String(Number(root.dataset.clicks) + 1)
			})
			document.body.appendChild(root)
			window.goshtosoInitTooltipTrigger(root)
			window.goshtosoInitTooltipTrigger(root)
			root.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }))
			const afterEnter = root.dataset.clicks
			root.dispatchEvent(new KeyboardEvent('keydown', { key: ' ', bubbles: true }))
			return afterEnter + '|' + root.dataset.clicks
		}`, nil)
		require.NoError(t, evalErr)
		assert.Equal(t, "1|2", clicks,
			"reinitialization must not stack fallback keyboard listeners")
	})

	t.Run("focusable descendant replaces and cleans fallback", func(t *testing.T) {
		state, evalErr := page.Evaluate(`() => {
			const root = document.createElement('span')
			root.dataset.tooltipContentId = 'transition-content'
			root.dataset.tooltipActivation = 'click'
			root.dataset.clicks = '0'
			root.addEventListener('click', () => {
				root.dataset.clicks = String(Number(root.dataset.clicks) + 1)
			})
			document.body.appendChild(root)
			window.goshtosoInitTooltipTrigger(root)
			const button = document.createElement('button')
			button.id = 'transition-button'
			button.setAttribute('aria-describedby', 'existing-help')
			root.appendChild(button)
			window.goshtosoInitTooltipTrigger(root)
			root.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }))
			return [
				String(root.tabIndex),
				root.getAttribute('role') || '',
				root.getAttribute('aria-describedby') || '',
				button.getAttribute('aria-describedby') || '',
				root.dataset.clicks,
			].join('|')
		}`, nil)
		require.NoError(t, evalErr)
		assert.Equal(t, "-1|||existing-help transition-content|0", state,
			"transition must remove fallback semantics/listener and preserve descendant descriptions")
	})

	t.Run("disabled native child does not create wrapper button", func(t *testing.T) {
		state, evalErr := page.Evaluate(`() => {
			const root = document.createElement('span')
			root.dataset.tooltipContentId = 'disabled-native-content'
			const button = document.createElement('button')
			button.disabled = true
			root.appendChild(button)
			document.body.appendChild(root)
			window.goshtosoInitTooltipTrigger(root)
			return [
				String(root.tabIndex),
				root.getAttribute('role') || '',
				root.getAttribute('aria-describedby') || '',
				button.getAttribute('aria-describedby') || '',
			].join('|')
		}`, nil)
		require.NoError(t, evalErr)
		assert.Equal(t, "-1|||", state,
			"disabled native controls must remain non-activatable without wrapper button semantics")
	})
}
