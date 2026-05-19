package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
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
