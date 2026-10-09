//go:build e2e && (full || select)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func selectEventPage(t *testing.T) playwright.Page {
	t.Helper()
	page := newIsolatedPage(t)
	require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`
		window.selectEvents = [];
		window.selectRequests = [];
		for (const type of ['input', 'change']) {
			document.addEventListener(type, event => {
				if (event.target.matches('input[hidden]')) {
					selectEvents.push({id: event.target.id, type, value: event.target.value});
				}
			});
		}
		const originalFetch = window.fetch;
		window.fetch = function(input, init) {
			const url = input instanceof Request ? input.url : String(input);
			if (url.includes('/api/components/select/filter')) selectRequests.push(url);
			return originalFetch.call(this, input, init);
		};
	`)}))
	_, err := page.Goto(baseURL + "/components/select")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	return page
}

// Wait through reactive updates and event handlers before checking silence.
func settleSelectEvents(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Evaluate(`() => new Promise(resolve => Alpine.nextTick(() => setTimeout(resolve, 100)))`)
	require.NoError(t, err)
}

func TestSelectUserChangeTriggersHTMXFilter(t *testing.T) {
	for _, mode := range []string{"light", "dark", "minimal-light", "minimal-dark"} {
		t.Run(mode, func(t *testing.T) {
			page := selectEventPage(t)
			_, err := page.Evaluate(`mode => {
				document.documentElement.classList.toggle('dark', mode.includes('dark'));
				if (mode.startsWith('minimal')) document.documentElement.dataset.theme = 'minimal';
				const sibling = document.createElement('input');
				sibling.type = 'hidden'; sibling.name = 'scope'; sibling.value = 'demo';
				document.querySelector('#select-htmx-filter').append(sibling);
			}`, mode)
			require.NoError(t, err)
			settleSelectEvents(t, page)
			initial, err := page.Evaluate(`() => [selectEvents.length, selectRequests.length]`)
			require.NoError(t, err)
			require.Equal(t, []any{0, 0}, initial, "initial defaults must be silent")

			trigger := page.Locator("#allocation-type-trigger")
			require.NoError(t, trigger.Click())
			require.NoError(t, page.Locator("#allocation-type-option-2").Click())
			_, err = page.WaitForFunction(`() => document.querySelector('#allocation-result').textContent.trim() === 'Showing reserved items'`, nil)
			require.NoError(t, err)
			settleSelectEvents(t, page)
			state, err := page.Evaluate(`() => ({
				events: selectEvents, requests: selectRequests.length,
				value: new URL(selectRequests[0], location.href).searchParams.get('type'),
				scope: new URL(selectRequests[0], location.href).searchParams.get('scope')
			})`)
			require.NoError(t, err)
			require.Equal(t, map[string]any{
				"events":   []any{map[string]any{"id": "allocation-type", "type": "change", "value": "reserved"}},
				"requests": 1, "value": "reserved", "scope": "demo",
			}, state)

			// Re-selecting the same option must not send a second request.
			require.NoError(t, trigger.Click())
			require.NoError(t, page.Locator("#allocation-type-option-2").Click())
			settleSelectEvents(t, page)
			counts, err := page.Evaluate(`() => [selectEvents.length, selectRequests.length]`)
			require.NoError(t, err)
			require.Equal(t, []any{1, 1}, counts)

			// ArrowDown wraps from Reserved to the empty All types option.
			require.NoError(t, trigger.Press("ArrowDown"))
			_, err = page.WaitForFunction(`() => document.activeElement.id === 'allocation-type-option-0'`, nil)
			require.NoError(t, err)
			require.NoError(t, page.Locator("#allocation-type-option-0").Press("Enter"))
			_, err = page.WaitForFunction(`() => document.querySelector('#allocation-result').textContent.trim() === 'Showing all types'`, nil)
			require.NoError(t, err)
			settleSelectEvents(t, page)
			state, err = page.Evaluate(`() => [selectEvents.length, selectRequests.length, selectEvents[1].value, new URL(selectRequests[1], location.href).searchParams.get('type')]`)
			require.NoError(t, err)
			require.Equal(t, []any{2, 2, "", ""}, state)

			// Space chooses Allocated using the same event contract.
			require.NoError(t, trigger.Press("ArrowDown"))
			_, err = page.WaitForFunction(`() => document.activeElement.id === 'allocation-type-option-1'`, nil)
			require.NoError(t, err)
			require.NoError(t, page.Locator("#allocation-type-option-1").Press("Space"))
			_, err = page.WaitForFunction(`() => document.querySelector('#allocation-result').textContent.trim() === 'Showing allocated items'`, nil)
			require.NoError(t, err)
			settleSelectEvents(t, page)
			counts, err = page.Evaluate(`() => [selectEvents.length, selectRequests.length, selectEvents[2].value]`)
			require.NoError(t, err)
			require.Equal(t, []any{3, 3, "allocated"}, counts)
		})
	}
}

func TestSelectSynchronizationDoesNotEchoEvents(t *testing.T) {
	page := selectEventPage(t)
	for _, eventType := range []string{"input", "change"} {
		_, err := page.Evaluate(`type => {
			const input = document.querySelector('#allocation-type');
			input.value = type === 'input' ? 'reserved' : 'allocated';
			input.dispatchEvent(new Event(type, {bubbles: true}));
		}`, eventType)
		require.NoError(t, err)
		settleSelectEvents(t, page)
	}
	_, err := page.WaitForFunction(`() => document.querySelector('#allocation-result').textContent.trim() === 'Showing allocated items'`, nil)
	require.NoError(t, err)
	state, err := page.Evaluate(`() => [selectEvents.length, selectRequests.length, document.querySelector('#allocation-type-trigger span').textContent.trim()]`)
	require.NoError(t, err)
	require.Equal(t, []any{2, 1, "Allocated"}, state, "only the caller's change event triggers HTMX")

	// Updating the public Model expression is silent; user picks still notify.
	_, err = page.Evaluate(`() => Alpine.evaluate(document.querySelector('#modelName-trigger'), "firstValue = 'tacoma'")`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => document.querySelector('#modelName').value === 'tacoma'`, nil)
	require.NoError(t, err)
	settleSelectEvents(t, page)
	count, err := page.Evaluate(`() => selectEvents.length`)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// Capture Model in the same event turn, before any subsequent reactive tick.
	_, err = page.Evaluate(`() => document.querySelector('#modelName').addEventListener('change', event => {
		window.modelAtChange = Alpine.evaluate(event.target, 'firstValue');
	})`)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#modelName-trigger").Click())
	require.NoError(t, page.Locator("#modelName-option-3").Click())
	settleSelectEvents(t, page)
	state, err = page.Evaluate(`() => [selectEvents.length, selectEvents[2].value, modelAtChange]`)
	require.NoError(t, err)
	require.Equal(t, []any{3, "rav4", "rav4"}, state)

	// Custom shell children own their events and do not get an automatic change.
	require.NoError(t, page.Locator("#access-shell-trigger").Click())
	require.NoError(t, page.Locator("#access-shell-listbox button").First().Click())
	settleSelectEvents(t, page)
	count, err = page.Evaluate(`() => selectEvents.length`)
	require.NoError(t, err)
	require.Equal(t, 3, count)
}
