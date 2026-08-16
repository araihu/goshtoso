//go:build e2e && (full || avatar)

package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Selectors scoped to the test fixtures section of the avatar demo page.
const (
	avatarFixturesURL    = "/components/avatar"
	loadedAvatarSelector = "[data-testid='avatar-test-loaded']"
	errorAvatarSelector  = "[data-testid='avatar-test-error']"
)

// avatarIsImageVisible returns true if the <img> inside the avatar container is visible.
func avatarIsImageVisible(t *testing.T, page playwright.Page, containerSelector string) bool {
	t.Helper()
	img := page.Locator(containerSelector + " img")
	visible, err := img.IsVisible()
	require.NoError(t, err)
	return visible
}

// avatarIsSpinnerVisible returns true if the loading spinner inside the avatar container is visible.
func avatarIsSpinnerVisible(t *testing.T, page playwright.Page, containerSelector string) bool {
	t.Helper()
	spinner := page.Locator(containerSelector + " svg.animate-spin")
	visible, err := spinner.IsVisible()
	require.NoError(t, err)
	return visible
}

// avatarIsInitialsVisible returns true if the initials span inside the avatar container is visible.
// The initials span is the one with `font-bold tracking-wider` that's NOT the spinner.
func avatarIsInitialsVisible(t *testing.T, page playwright.Page, containerSelector string) bool {
	t.Helper()
	// The initials span has font-bold tracking-wider. It's inside the avatar root
	// (not nested inside spinner/image layers).
	initials := page.Locator(containerSelector + " span.font-bold.tracking-wider")
	visible, err := initials.IsVisible()
	require.NoError(t, err)
	return visible
}

// navigateToAvatarDemo loads the avatar component demo page and waits for network idle.
func navigateToAvatarDemo(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+avatarFixturesURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(5000),
	})
	require.NoError(t, err)
	// Make sure the fixtures section rendered
	require.NoError(t, page.Locator("[data-testid='avatar-e2e-fixtures']").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(3000),
		},
	))
}

func TestAvatarReducedMotionSpinnersStopAnimation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))
	navigateToAvatarDemo(t, page)

	_, err := page.WaitForFunction(`() => {
		const spinners = document.querySelectorAll("[data-testid='avatar-e2e-fixtures'] svg.animate-spin");
		return spinners.length > 0 && Array.from(spinners).every(spinner => getComputedStyle(spinner).animationName === 'none');
	}`, nil)
	require.NoError(t, err)
}

// TestAvatar_ImageLoaded_SpinnerGoneInitialsHidden verifies that when an image
// loads successfully, the loading spinner disappears AND the initials fallback
// is hidden (so transparent pixels don't show text bleeding through).
func TestAvatar_ImageLoaded_SpinnerGoneInitialsHidden(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToAvatarDemo(t, page)

	// Wait for the image to actually load in the browser.
	require.NoError(t, page.Locator(loadedAvatarSelector+" img").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		},
	))

	// After load: image visible, spinner hidden, initials hidden.
	assert.True(t, avatarIsImageVisible(t, page, loadedAvatarSelector),
		"image should be visible after successful load")
	assert.False(t, avatarIsSpinnerVisible(t, page, loadedAvatarSelector),
		"spinner should be hidden after image loads (not stuck)")
	assert.False(t, avatarIsInitialsVisible(t, page, loadedAvatarSelector),
		"initials should be hidden when image loads successfully (no text bleed-through)")
}

// TestAvatar_CachedReload verifies the cached-image race fix: when the page is
// reloaded and the browser has the image cached, the spinner must not get stuck.
// This is the bug that x-init="$el.complete && naturalWidth > 0" fixes.
func TestAvatar_CachedReload(t *testing.T) {
	page := newPage(t, sharedBrowser)

	// First visit: image is fetched from the server and cached by the browser.
	navigateToAvatarDemo(t, page)
	require.NoError(t, page.Locator(loadedAvatarSelector+" img").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		},
	))

	// Second visit: the image is cached. The browser may fire the load event
	// before Alpine attaches x-on:load. Without the x-init fix, imgLoaded would
	// stay false and the spinner would be stuck forever.
	_, err := page.Reload(playwright.PageReloadOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(5000),
	})
	require.NoError(t, err)

	// Give Alpine a moment to settle after reload.
	require.NoError(t, page.Locator(loadedAvatarSelector+" img").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		},
	))

	// After cached reload: image visible, spinner NOT stuck.
	assert.True(t, avatarIsImageVisible(t, page, loadedAvatarSelector),
		"image should be visible after cached reload")
	assert.False(t, avatarIsSpinnerVisible(t, page, loadedAvatarSelector),
		"spinner must not be stuck after cached reload (x-init catches already-complete images)")
	assert.False(t, avatarIsInitialsVisible(t, page, loadedAvatarSelector),
		"initials should still be hidden after cached reload")
}

// TestAvatar_FragmentNav_ShowcaseRegistered reproduces the first-render bug:
// the demo registered Alpine.data('avatarShowcase') only on the 'alpine:init'
// event, which does NOT re-fire on htmx fragment navigation. Arriving via the
// sidebar (fragment swap) left x-data="avatarShowcase" undefined, so reactive
// avatars never received a size class and the preview rendered broken until a
// full-page refresh. Direct page.Goto (every other avatar test) never exercised
// this path.
func TestAvatar_FragmentNav_ShowcaseRegistered(t *testing.T) {
	page := newIsolatedPage(t)

	// pageErrors captures uncaught JS exceptions. The avatar component used to
	// interpolate cfg.Src unescaped into the x-on:error single-quoted JS string;
	// a data-URI Src (with embedded single quotes) broke it with a SyntaxError.
	var pageErrors []string
	page.On("pageerror", func(err error) { pageErrors = append(pageErrors, err.Error()) })

	// consoleErrors captures every console.error with its source location. The
	// shared filter below recognizes only the intentional broken-image fixture.
	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, consoleFailureMessage(m))
		}
	})

	// Land elsewhere first, then navigate via the sidebar (fragment swap).
	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/components/avatar']").First().Click())

	// Wait for the avatar fragment to land.
	require.NoError(t, page.Locator("[data-testid='avatar-section-image']").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(3000),
		}))

	// avatarShowcase must be registered after fragment nav: the size selector
	// must drive the reactive avatars. Click 2xl and confirm a reactive avatar
	// root picks up size-32. If avatarShowcase is undefined, the binding never
	// applies and this times out.
	require.NoError(t, page.Locator("label[for='avatar-size-2xl']").Click())
	_, err = page.WaitForFunction(
		`() => {
			const roots = document.querySelectorAll("[data-testid='avatar-section-image'] .relative.inline-flex");
			return Array.from(roots).some(el => (el.getAttribute('class') || '').includes('size-32'));
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "avatarShowcase must be registered after fragment nav so the size selector drives reactive avatars")

	require.Empty(t, pageErrors, "no uncaught JS exceptions on fragment-nav avatar page: %v", pageErrors)
	require.Empty(t, filterIgnorable(consoleErrors), "no unexpected console errors on fragment-nav avatar page: %v", consoleErrors)
}

// TestAvatar_SizeSelector_RendersSixPills verifies the interactive size selector
// rendered six segmented radio pills (xs/sm/md/lg/xl/2xl).
func TestAvatar_SizeSelector_RendersSixPills(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToAvatarDemo(t, page)

	for _, val := range []string{"xs", "sm", "md", "lg", "xl", "2xl"} {
		count, err := page.Locator("#avatar-size-" + val).Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "expected size pill #avatar-size-%s to exist", val)
	}
	v, err := page.Locator("#avatar-size-md").Evaluate("el => el.checked", nil)
	require.NoError(t, err)
	assert.Equal(t, true, v, "md pill should start checked")
}

// TestAvatar_SizeSelector_DrivesReactiveAvatars verifies that clicking a size
// pill propagates the Tailwind size class to all reactive avatars on the page.
func TestAvatar_SizeSelector_DrivesReactiveAvatars(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToAvatarDemo(t, page)

	_, err := page.WaitForFunction(
		`() => typeof window.Alpine !== 'undefined'`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)

	require.NoError(t, page.Locator("label[for='avatar-size-2xl']").Click())

	_, err = page.WaitForFunction(
		`() => {
			const roots = document.querySelectorAll("[data-testid='avatar-section-image'] .relative.inline-flex");
			return Array.from(roots).some(el => {
				const c = el.getAttribute('class') || '';
				return c.includes('size-32') && c.includes('text-5xl');
			});
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err, "after clicking 2xl, at least one reactive avatar root should carry size-32 text-5xl")

	_, err = page.WaitForFunction(
		`() => {
			const dots = document.querySelectorAll("[data-testid='avatar-section-status'] span.absolute.rounded-full");
			return Array.from(dots).some(el => (el.getAttribute('class') || '').includes('size-7'));
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err, "after clicking 2xl, status indicator should carry size-7")

	require.NoError(t, page.Locator("label[for='avatar-size-xs']").Click())
	_, err = page.WaitForFunction(
		`() => {
			const roots = document.querySelectorAll("[data-testid='avatar-section-image'] .relative.inline-flex");
			return Array.from(roots).some(el => {
				const c = el.getAttribute('class') || '';
				return c.includes('size-8') && c.includes('text-xs');
			});
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err, "after clicking xs, reactive avatar root should carry size-8 text-xs")

	txt, err := page.Locator("[data-testid='avatar-size-selected']").TextContent()
	require.NoError(t, err)
	assert.Equal(t, "xs", txt)
}

// TestAvatar_RadiusSelector_DrivesSquareAvatar verifies the PenguinUI-style
// square avatar radius selector updates the reactive square preview.
func TestAvatar_RadiusSelector_DrivesSquareAvatar(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToAvatarDemo(t, page)

	require.NoError(t, page.Locator("label[for='avatar-radius-none']").Click())
	_, err := page.WaitForFunction(
		`() => {
			const root = document.querySelector("[data-testid='avatar-section-square'] .relative.inline-flex");
			return root && (root.getAttribute('class') || '').includes('rounded-none');
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err, "after clicking none, square avatar root should carry rounded-none")

	require.NoError(t, page.Locator("label[for='avatar-radius-lg']").Click())
	_, err = page.WaitForFunction(
		`() => {
			const root = document.querySelector("[data-testid='avatar-section-square'] .relative.inline-flex");
			return root && (root.getAttribute('class') || '').includes('rounded-lg');
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err, "after clicking lg, square avatar root should carry rounded-lg")
}

func TestAvatar_StackedAvatars_RenderOverlappedGroup(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToAvatarDemo(t, page)

	stack := page.Locator("[data-testid='avatar-section-stacked'] [role='list']")
	require.NoError(t, stack.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(3000),
	}))

	label, err := stack.GetAttribute("aria-label")
	require.NoError(t, err)
	assert.Equal(t, "Team members", label)

	count, err := stack.Locator("[role='listitem']").Count()
	require.NoError(t, err)
	assert.Equal(t, 4, count)

	cls, err := stack.Locator("[role='listitem']").Nth(1).Locator(".relative.inline-flex").GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, cls, "-ml-3")
	assert.Contains(t, cls, "ring-2")
}

// TestAvatar_SizeSelector_FixturesUnchanged verifies that the E2E test fixtures
// section stays at fixed SizeMD even when the selector changes.
func TestAvatar_SizeSelector_FixturesUnchanged(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToAvatarDemo(t, page)

	require.NoError(t, page.Locator("label[for='avatar-size-2xl']").Click())

	cls, err := page.Locator("[data-testid='avatar-test-loaded'] .relative.inline-flex").GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, cls, "size-14", "E2E fixture must not be size-driven by the selector")
	assert.NotContains(t, cls, "size-32", "E2E fixture must NOT receive 2xl bindings")
}

// TestAvatar_ImageError_FallsBackToInitials verifies that when an image fails
// to load (404, network error, etc.), the initials fallback becomes visible
// and the spinner disappears.
func TestAvatar_ImageError_FallsBackToInitials(t *testing.T) {
	page := newPage(t, sharedBrowser)

	// Listen for the 404 response so we can confirm the browser attempted the
	// fetch and the failure reached Alpine's x-on:error.
	errResp := make(chan playwright.Response, 1)
	page.On("response", func(resp playwright.Response) {
		if resp.Status() == 404 && strings.Contains(resp.URL(), "does-not-exist-404.png") {
			select {
			case errResp <- resp:
			default:
			}
		}
	})

	navigateToAvatarDemo(t, page)

	// Confirm the 404 actually returned. If this times out, the server route
	// changed and the test fixture needs updating.
	select {
	case <-errResp:
	case <-time.After(5 * time.Second):
		t.Fatal("404 for /assets/images/does-not-exist-404.png never returned")
	}

	// Poll Alpine state directly: imgError must flip to true. WaitFor on DOM
	// visibility races with Alpine's reactive re-render after the error event.
	_, err := page.WaitForFunction(
		`() => {
			const el = document.querySelector("[data-testid='avatar-test-error'] [x-data]");
			if (!el || !window.Alpine) return false;
			const data = window.Alpine.$data(el);
			return data.imgError === true;
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)},
	)
	require.NoError(t, err, "Alpine imgError should flip to true after 404")

	assert.True(t, avatarIsInitialsVisible(t, page, errorAvatarSelector),
		"initials should be visible as fallback after image error")
	assert.False(t, avatarIsSpinnerVisible(t, page, errorAvatarSelector),
		"spinner should be hidden after image error (not stuck loading)")
	assert.False(t, avatarIsImageVisible(t, page, errorAvatarSelector),
		"image should be hidden after load error")
}
