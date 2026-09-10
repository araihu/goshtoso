//go:build e2e && full

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"

	"github.com/stretchr/testify/require"
)

// The integration must preserve reactive state during morphs and initialize
// replacements exactly once, including components with Alpine-generated IDs.
func TestHTMX4AlpineSwapLifecycle(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/components/button")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => window.Alpine && window.htmx && Alpine.version === '3.17.2' && htmx.version === '4.0.0'`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => {
  window.swapProbe = {init: 0, destroy: 0};
  Alpine.data('swapProbeComponent', () => ({
   count: 0,
   init() { window.swapProbe.init++; },
   destroy() { window.swapProbe.destroy++; }
  }));
  const host = document.createElement('div');
  host.id = 'swap-probe-host';
  document.body.appendChild(host);
  window.swapProbeHTML = label => '<section id="swap-probe" x-data="swapProbeComponent" x-id="[\'probe-input\']"><label :for="$id(\'probe-input\')">' + label + '</label><input :id="$id(\'probe-input\')"><button @click="count++" x-text="count"></button></section>';
  host.innerHTML = window.swapProbeHTML('Before');
 }`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => window.swapProbe.init === 1`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#swap-probe button").Click())
	_, err = page.Evaluate(`async () => {
  const node = document.getElementById('swap-probe');
  window.swapProbeOldNode = node;
  window.swapProbeOldInput = node.querySelector('input');
  await htmx.swap({sourceElement: node, target: node, text: window.swapProbeHTML('After'), swap: 'outerMorph'});
  await Alpine.nextTick();
 }`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => {
  const node = document.getElementById('swap-probe');
  return node === window.swapProbeOldNode && node.querySelector('input') === window.swapProbeOldInput &&
   node.querySelector('label').textContent === 'After' && node.querySelector('button').textContent === '1' &&
   node.querySelector('label').htmlFor === node.querySelector('input').id && window.swapProbe.init === 1;
 }`, nil)
	require.NoError(t, err, "morph must update HTML while preserving Alpine state and reactive ID identity")
	_, err = page.Evaluate(`async () => {
  const node = document.getElementById('swap-probe');
  await htmx.swap({sourceElement: node, target: node, text: window.swapProbeHTML('Replacement'), swap: 'outerHTML'});
  await Alpine.nextTick();
 }`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => window.swapProbe.init === 2 && window.swapProbe.destroy === 1 && document.querySelector('#swap-probe button').textContent === '0'`, nil)
	require.NoError(t, err, "replacement must clean up old state and initialize the new component once")
}

// An active stream must close without leaking its reader or an unhandled
// rejection when navigation removes its connector.
func TestHTMX4SSECleanup(t *testing.T) {
	for _, path := range []string{"ticker", "logs"} {
		t.Run(path, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`
window.sseConnections = [];
window.sseRejections = [];
document.addEventListener('htmx:sse:after:connection', event => sseConnections.push(event.detail.connection));
window.addEventListener('unhandledrejection', event => sseRejections.push(String(event.reason)));
`)}))
			_, err := page.Goto(baseURL + "/examples/" + path)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => sseConnections.length > 0 && sseConnections.every(c => c.reader)`, nil)
			require.NoError(t, err)
			_, err = page.Evaluate(`async () => {
await htmx.ajax('GET', '/components/button', {target: '#main-content', swap: 'innerHTML'});
// Give unhandled-rejection reporting its own event-loop turn.
await new Promise(resolve => setTimeout(resolve, 50));
}`)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => sseConnections.every(c => c.abortController.signal.aborted && c.reader === null)`, nil)
			require.NoError(t, err)
			rejections, err := page.Evaluate(`() => sseRejections`)
			require.NoError(t, err)
			require.Empty(t, rejections)
		})
	}
}

func TestHTMX4PartialEnhancesCodeBlock(t *testing.T) {
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/components/button")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => window.htmx`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`async () => {
const main = document.createElement('div');
main.id = 'swap-main';
main.textContent = 'Keep main';
const secondary = document.createElement('div');
secondary.id = 'swap-secondary';
document.body.append(main, secondary);
await htmx.swap({sourceElement: main, target: main, text: '<hx-partial hx-target="#swap-secondary" hx-swap="innerHTML"><button hidden data-code-block-copy data-code-block-target="partial-code"><span data-code-block-copy-status>Copy</span></button><pre id="partial-code">Partial content</pre></hx-partial>', swap: 'innerHTML'});
}`)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#swap-secondary [data-code-block-copy]").WaitFor())
	mainText, err := page.Locator("#swap-main").TextContent()
	require.NoError(t, err)
	require.Equal(t, "Keep main", mainText, "partial-only response must preserve the main target")
}
