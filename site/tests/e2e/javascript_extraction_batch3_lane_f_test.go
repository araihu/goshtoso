//go:build e2e && full

package e2e

import (
	"os"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestJavaScriptBatch3LaneF_SourceProvidersRunBeforeBundleIntegration(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/site.webmanifest")
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<main>
		<div id="page-scroll"><div id="main-content">
			<h2 id="first" data-toc-heading>First</h2>
			<h2 id="second" data-toc-heading>Second</h2>
		</div></div>
		<aside id="toc-rail"><nav id="toc-list"></nav></aside>
		<input id="draft-os">
	</main>`))

	_, err = page.Evaluate(`() => {
		document.cookie = 'gt_storage=allowed; Path=/; SameSite=Lax';
		localStorage.setItem('theme', 'minimal');
		localStorage.setItem('darkMode', 'true');
		document.documentElement.setAttribute('data-demo-storage-policy', 'strict');
		document.documentElement.setAttribute('data-demo-theme-bootstrap', '');
	}`)
	require.NoError(t, err)
	addLaneFSource(t, page, "../../assets/js/src/site-bootstrap.js")

	bootstrapReady, err := page.Evaluate(`() =>
		window.goshtosoStorageConsent.allowed() &&
		document.documentElement.dataset.theme === 'minimal' &&
		document.documentElement.classList.contains('dark')`)
	require.NoError(t, err)
	require.Equal(t, true, bootstrapReady)

	_, err = page.Evaluate(`() => {
		window.__laneFProviders = {};
		window.__laneFStores = {};
		window.__laneFObserverCreates = 0;
		window.__laneFObserverDisconnects = 0;
		window.IntersectionObserver = class {
			constructor() { window.__laneFObserverCreates += 1; }
			observe() {}
			disconnect() { window.__laneFObserverDisconnects += 1; }
		};
		window.Alpine = {
			data(name, factory) { window.__laneFProviders[name] = factory; },
			store(name, value) {
				if (arguments.length === 2) window.__laneFStores[name] = value;
				return window.__laneFStores[name];
			},
			initTree() {},
		};
		window.htmx = { process() {} };
	}`)
	require.NoError(t, err)

	for _, source := range []string{
		"../../assets/js/src/demo-layout.js",
		"../../assets/js/src/tab-view.js",
		"../../assets/js/src/select-demo.js",
	} {
		addLaneFSource(t, page, source)
	}

	providerReady, err := page.Evaluate(`() => {
		if (!['demoLayout', 'demoStorageConsent', 'demoTabView'].every(
			name => typeof window.__laneFProviders[name] === 'function')) return false;
		window.buildTOC();
		if (document.querySelectorAll('#toc-list [data-toc-link]').length !== 2) return false;
		if (window.__laneFObserverCreates < 2 || window.__laneFObserverDisconnects < 1) return false;

		let changed = false;
		document.getElementById('draft-os').addEventListener('change', () => { changed = true; });
		if (!window.goshtosoRestoreSelectDraft('draft-os', 'linux') || !changed ||
			document.getElementById('draft-os').value !== 'linux') return false;

		let stopped = false;
		const layout = window.__laneFProviders.demoLayout();
		layout.$watch = () => () => { stopped = true; };
		layout.init();
		layout.destroy();
		if (!stopped || layout._stopThemeWatch !== null) return false;

		const tabs = window.__laneFProviders.demoTabView();
		tabs._copyTimer = window.setTimeout(() => {}, 1000);
		tabs.destroy();
		if (!tabs._destroyed || tabs._copyTimer !== 0) return false;

		const consent = window.__laneFProviders.demoStorageConsent();
		consent.deny();
		return consent.show === false && localStorage.getItem('theme') === null &&
			document.cookie.includes('gt_storage=denied');
	}`)
	require.NoError(t, err)
	require.Equal(t, true, providerReady)
}

func addLaneFSource(t *testing.T, page playwright.Page, path string) {
	t.Helper()
	source, err := os.ReadFile(path)
	require.NoError(t, err)
	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{Content: new(string(source))})
	require.NoError(t, err)
}
