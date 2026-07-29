package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestJavaScriptBatch2_FirstPaintFactoriesAndDataIntegrity(t *testing.T) {
	page := newPage(t, sharedBrowser)
	dismissCookieBanner(t, page)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(message playwright.ConsoleMessage) {
		if message.Type() == "error" {
			jsErrors = append(jsErrors, message.Text())
		}
	})

	_, err := page.Goto(baseURL + "/components/select")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`async () => {
		const names = [
			'goshtosoParseData', 'goshtosoCarousel', 'goshtosoDropdown',
			'goshtosoPalette', 'goshtosoSelect', 'goshtosoSelectShell', 'goshtosoTabs',
		];
		if (!names.every(name => typeof window[name] === 'function')) return false;
		const providers = [
			'__actionGroupDemoRegistered', '__avatarShowcaseRegistered', '__logFeedRegistered',
			'__profileImagesRegistered', '__tickerPaneRegistered', '__themePageRegistered',
		];
		if (!window.Alpine || !providers.every(name => Alpine[name] === true) || window.__gtChatInit !== true) {
			return false;
		}

		const original = { message: 'Olá 世界 🚀', items: ['ação', '東京', '🙂'] };
		const bytes = new TextEncoder().encode(JSON.stringify(original));
		let binary = '';
		bytes.forEach(byte => { binary += String.fromCharCode(byte); });
		const decoded = goshtosoParseData(btoa(binary), null);
		if (JSON.stringify(decoded) !== JSON.stringify(original)) return false;

		const fallback = { fallback: true };
		if (goshtosoParseData('not valid base64%', fallback) !== fallback) return false;
		if (goshtosoParseData('/w==', fallback) !== fallback) return false;

		const root = document.querySelector('#os-trigger')?.closest('[data-select-config]');
		if (!root || !root._x_dataStack) return false;
		const config = goshtosoParseData(root.dataset.selectConfig, null);
		const option = config && config.options && config.options[0];
		if (!option || option.value !== 'mac' || option.label !== 'Mac') return false;
		if ('Value' in option || 'Label' in option || 'Selected' in option) return false;

		const scripts = Array.from(document.scripts);
		const componentBundle = scripts.find(script => script.src.includes('/assets/js/goshtoso.min.js'));
		const siteBundle = scripts.find(script => script.src.includes('/site-assets/js/goshtoso-demo.min.js'));
		const alpine = scripts.find(script => script.src.includes('/assets/js/runtime/alpinejs/'));
		if (!componentBundle || !siteBundle || !alpine || !componentBundle.defer || !alpine.defer || siteBundle.defer) {
			return false;
		}

		const [componentSource, siteSource] = await Promise.all([
			fetch(componentBundle.src).then(response => response.ok ? response.text() : ''),
			fetch(siteBundle.src).then(response => response.ok ? response.text() : ''),
		]);
		if (!componentSource || !siteSource) return false;
		if (['themePage', 'profileImages', 'tickerPane', '__gtChatInit'].some(name => componentSource.includes(name))) {
			return false;
		}
		if (['goshtosoCarousel', 'goshtosoDropdown', 'goshtosoSelect', 'goshtosoTabs'].some(name => siteSource.includes(name))) {
			return false;
		}
		return true;
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "split bundles, first-paint providers, Unicode data, fallback behavior, lowercase DTO, and runtime order should hold")
	require.Empty(t, jsErrors, "no JS console/page errors on first paint: %v", jsErrors)
}

func TestJavaScriptBatch2_LibraryFactoriesSurviveFragmentNavigation(t *testing.T) {
	page := newPage(t, sharedBrowser)
	dismissCookieBanner(t, page)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(message playwright.ConsoleMessage) {
		if message.Type() == "error" {
			jsErrors = append(jsErrors, message.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	navigate := func(path, readySelector string) {
		t.Helper()
		require.NoError(t, page.Locator("a[href='"+path+"']").First().Click())
		_, waitErr := page.WaitForFunction(
			"selector => { const el = document.querySelector(selector); return !!(el && el._x_dataStack); }",
			readySelector, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
		require.NoError(t, waitErr, "%s should initialize after HTMX fragment navigation", path)
	}

	navigate("/components/carousel", "#carousel-default-c")
	require.NoError(t, page.Locator("#carousel-default-c button[aria-label='next slide']").Click())
	require.NoError(t, waitForCarouselIndex(page, "#carousel-default-c", 2))

	navigate("/components/dropdown", "#dropdown-click")
	dropdownTrigger := page.Locator("#dropdown-click button").First()
	require.NoError(t, dropdownTrigger.Click())
	_, err = page.WaitForFunction(
		"() => document.querySelector('#dropdown-click button')?.getAttribute('aria-expanded') === 'true'", nil)
	require.NoError(t, err)

	navigate("/components/palette", "#demo-palette")
	require.NoError(t, page.Locator(`#demo-palette button[data-cls="blue-700"]`).Click())
	_, err = page.WaitForFunction(
		"() => Alpine.$data(document.querySelector('#demo-palette')).selectedHex !== '#000000'", nil)
	require.NoError(t, err)

	navigate("/components/select", "#select-default [data-select-config]")
	require.NoError(t, page.Locator("#os-trigger").Click())
	require.NoError(t, page.Locator("#os-option-1").Click())
	_, err = page.WaitForFunction("() => document.querySelector('input#os')?.value === 'windows'", nil)
	require.NoError(t, err)

	navigate("/components/tabs", "#tabs-default > [x-data]")
	require.NoError(t, page.Locator("#tabs-default").GetByRole(
		"tab", playwright.LocatorGetByRoleOptions{Name: "Likes"}).Click())
	require.NoError(t, waitForTabSelected(page, "#tabs-default [aria-controls='tabpaneldefaultlikes']", true))

	require.Empty(t, jsErrors, "no undefined globals or console/page errors across fragment navigation: %v", jsErrors)
}

func TestJavaScriptBatch2_ProviderLifecycleCleansOwnedResources(t *testing.T) {
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)
	require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`(() => {
		const stats = window.__batch2Lifecycle = {
			themeActive: 0, themeDisconnects: 0, themeCallbacks: 0,
			tickerAdds: 0, tickerRemoves: 0, chatAdds: 0,
			urlsCreated: 0, urlsRevoked: 0,
		};

		const NativeObserver = window.MutationObserver;
		window.MutationObserver = class extends NativeObserver {
			constructor(callback) {
				let instance;
				super((records, observer) => {
					if (instance && instance.__batch2ThemeObserver) stats.themeCallbacks += 1;
					callback(records, observer);
				});
				instance = this;
				this.__batch2ThemeObserver = false;
			}
			observe(target, options) {
				const filter = options && options.attributeFilter;
				if (!this.__batch2ThemeObserver && target === document.documentElement &&
					Array.isArray(filter) && filter.includes('class') && filter.includes('data-theme')) {
					this.__batch2ThemeObserver = true;
					stats.themeActive += 1;
				}
				return super.observe(target, options);
			}
			disconnect() {
				if (this.__batch2ThemeObserver) {
					this.__batch2ThemeObserver = false;
					stats.themeActive -= 1;
					stats.themeDisconnects += 1;
				}
				return super.disconnect();
			}
		};

		const nativeAdd = EventTarget.prototype.addEventListener;
		const nativeRemove = EventTarget.prototype.removeEventListener;
		EventTarget.prototype.addEventListener = function (type, listener, options) {
			if ((type === 'htmx:sseBeforeMessage' || type === 'htmx:sseError') &&
				this instanceof Element && this.closest('#ticker-fragment')) {
				this.__batch2TickerProvider = true;
				stats.tickerAdds += 1;
			}
			if ((type === 'htmx:wsAfterSend' || type === 'htmx:oobAfterSwap') && this === document) {
				stats.chatAdds += 1;
			}
			return nativeAdd.call(this, type, listener, options);
		};
		EventTarget.prototype.removeEventListener = function (type, listener, options) {
			if ((type === 'htmx:sseBeforeMessage' || type === 'htmx:sseError') && this.__batch2TickerProvider) {
				stats.tickerRemoves += 1;
			}
			return nativeRemove.call(this, type, listener, options);
		};

		const nativeCreateObjectURL = URL.createObjectURL.bind(URL);
		const nativeRevokeObjectURL = URL.revokeObjectURL.bind(URL);
		URL.createObjectURL = function (value) {
			stats.urlsCreated += 1;
			return nativeCreateObjectURL(value);
		};
		URL.revokeObjectURL = function (value) {
			stats.urlsRevoked += 1;
			return nativeRevokeObjectURL(value);
		};
	})()`)}))

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction(
		"() => window.__gtChatInit === true && window.__batch2Lifecycle.chatAdds === 2", nil)
	require.NoError(t, err, "chat document hooks should install exactly once")
	_, err = page.Evaluate(`() => localStorage.setItem('themeOverrides', JSON.stringify({
		primary: 'blue-700', danger: 17, unknown: 'red-500'
	}))`, nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/docs/theme']").First().Click())
	_, err = page.WaitForFunction(
		`() => {
			const root = document.querySelector('[x-data=themePage]');
			if (!root?._x_dataStack || window.__batch2Lifecycle.themeActive !== 1) return false;
			const overrides = Alpine.$data(root).overrides;
			return overrides.primary === 'blue-700' && !('danger' in overrides) && !('unknown' in overrides);
		}`, nil)
	require.NoError(t, err, "theme provider should own one active observer")

	require.NoError(t, page.Locator("a[href='/examples/ticker']").First().Click())
	_, err = page.WaitForFunction(`() => {
		const stats = window.__batch2Lifecycle;
		return document.querySelector('#ticker-fragment')?._x_dataStack &&
			stats.themeActive === 0 && stats.themeDisconnects >= 1 && stats.tickerAdds >= 2;
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "theme observer should disconnect and ticker listeners should attach")

	require.NoError(t, page.Locator("a[href='/examples/profile']").First().Click())
	_, err = page.WaitForFunction(`() => {
		const stats = window.__batch2Lifecycle;
		return document.querySelector('#profile-fragment')?._x_dataStack && stats.tickerRemoves >= 2;
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "ticker provider should remove both owned listeners")

	require.NoError(t, page.Locator("#profile-avatar-input").SetInputFiles(playwright.InputFile{
		Name: "avatar.png", MimeType: "image/png", Buffer: onePxPNG,
	}))
	_, err = page.WaitForFunction("() => window.__batch2Lifecycle.urlsCreated >= 1", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/getting-started']").First().Click())
	_, err = page.WaitForFunction(
		"() => !document.querySelector('#profile-fragment') && window.__batch2Lifecycle.urlsRevoked >= 1", nil)
	require.NoError(t, err, "profile provider should revoke owned object URLs on teardown")

	chatAdds, err := page.Evaluate("() => window.__batch2Lifecycle.chatAdds", nil)
	require.NoError(t, err)
	require.Equal(t, 2, chatAdds, "chat hooks should remain document-singleton across fragment navigation")
}
