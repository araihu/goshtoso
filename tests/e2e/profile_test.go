package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// gotoProfile navigates to the profile page (with the given query, e.g.
// "?seed=0") and waits for Alpine.js + the #profile-fragment to be ready.
func gotoProfile(t *testing.T, page playwright.Page, query string) {
	t.Helper()
	// Suppress the first-run cookie/localStorage notice so its fixed banner can
	// never intercept clicks on the appearance controls.
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: playwright.String("try{localStorage.setItem('cookieConsent','accepted')}catch(e){}"),
	}))
	_, err := page.Goto(baseURL + "/examples/profile" + query)
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#profile-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))
}

// activateAppearanceTab clicks the "Appearance" tab button (the theme/dark
// controls live in that tab, which is NOT the default and is x-show-hidden on
// load) and waits for the theme Select trigger to become visible. The tabs
// component renders tab buttons as role="tab"; selecting by accessible text is
// stable across tab markup changes.
func activateAppearanceTab(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.Locator("button[role='tab']", playwright.PageLocatorOptions{
		HasText: "Appearance",
	}).Click())
	require.NoError(t, page.Locator("#profile-theme-trigger").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))
}

// onePxPNG is a valid 1x1 PNG used as a small uploadable image.
var onePxPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
	0x42, 0x60, 0x82,
}

// TestProfileFragmentNavNoConsoleErrors lands on the examples gallery, then
// navigates to the profile app via the gallery card (htmx fragment swap), and
// asserts there are zero console/page errors. This guards the examples mandate:
// the profileImages Alpine.data must register on fragment-nav (Alpine already
// running) and the IndexedDB/toast wiring must not error with no target.
func TestProfileFragmentNavNoConsoleErrors(t *testing.T) {
	page := newIsolatedPage(t)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: playwright.String("try{localStorage.setItem('cookieConsent','accepted')}catch(e){}"),
	}))
	// Land on the examples gallery first (seed=0 keeps state out of it).
	_, err := page.Goto(baseURL + "/examples?seed=0")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	// Click the Profile gallery card → htmx fragment swap.
	require.NoError(t, page.Locator("a[href='/examples/profile']").First().Click())
	_, err = page.WaitForFunction("() => !!document.querySelector('#profile-fragment')", nil)
	require.NoError(t, err)

	// Let Alpine init the swapped-in component (IndexedDB hydrate, etc.).
	// Wait for Alpine to have initialized the profileImages x-data on #profile-fragment
	// (_x_dataStack is set by Alpine v3 when it initialises an element).
	_, err = page.WaitForFunction(
		"() => { const el = document.querySelector('#profile-fragment'); return !!(el && el._x_dataStack); }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "profileImages Alpine component should initialise on fragment-nav")

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav profile page: %v", jsErrors)
}

// TestProfileIdentityPersists fills the name field, saves via the HTMX form,
// then reloads WITHOUT seed=0 (so the cookie is read) and asserts the name
// persisted. Proves the cookie-backed identity round-trip.
func TestProfileIdentityPersists(t *testing.T) {
	page := newIsolatedPage(t)
	gotoProfile(t, page, "?seed=0")

	require.NoError(t, page.Locator("input[name='name']").Fill("Katherine Johnson"))
	// The submit button is NOT itself swapped (only #profile-identity inner is),
	// so a normal click is fine.
	require.NoError(t, page.Locator("#profile-identity-section form button[type='submit']").Click())
	// Wait for the HTMX swap to re-render #profile-identity with the saved value.
	_, err := page.WaitForFunction(
		"() => { const el = document.querySelector(\"input[name='name']\"); return el && el.value === 'Katherine Johnson'; }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	// Reload reading the cookie (no seed=0) — same browser context.
	gotoProfile(t, page, "")

	val, err := page.Locator("input[name='name']").InputValue()
	require.NoError(t, err)
	require.Equal(t, "Katherine Johnson", val, "name should persist via cookie across reload")
}

// TestProfileDarkToggle clicks the dark-mode Toggle and asserts the <html> .dark
// class flips. Proves the Toggle's x-on:change → $store.darkMode.toggle().
func TestProfileDarkToggle(t *testing.T) {
	page := newIsolatedPage(t)
	gotoProfile(t, page, "?seed=0")
	activateAppearanceTab(t, page)

	before, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("label[for='profile-dark']").Click())
	// Wait for the <html> .dark class to actually flip rather than sleeping blindly.
	beforeBool, _ := before.(bool)
	_, err = page.WaitForFunction(
		fmt.Sprintf("() => document.documentElement.classList.contains('dark') === %t", !beforeBool),
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "dark mode class on <html> should flip after toggling")

	after, err := page.Evaluate("() => document.documentElement.classList.contains('dark')", nil)
	require.NoError(t, err)
	require.NotEqual(t, before, after, "dark mode class on <html> should flip after toggling")
}

// TestProfileThemePicker opens the theme Select and picks "Dracula", asserting
// the root data-theme attribute updates. Proves the Select's AlpineModel
// ("theme") two-way binds to the app's root theme engine.
func TestProfileThemePicker(t *testing.T) {
	page := newIsolatedPage(t)
	gotoProfile(t, page, "?seed=0")
	activateAppearanceTab(t, page)

	// Open the Select dropdown.
	require.NoError(t, page.Locator("#profile-theme-trigger").Click())
	// The options listbox is rendered via x-for; wait for the Dracula option.
	dracula := page.Locator("#profile-appearance li[role='option']", playwright.PageLocatorOptions{HasText: "Dracula"})
	require.NoError(t, dracula.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(2000),
	}))
	require.NoError(t, dracula.Click())

	_, err := page.WaitForFunction(
		"() => document.documentElement.getAttribute('data-theme') === 'dracula'",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "selecting Dracula should set data-theme='dracula' on <html>")
}

// TestProfileAvatarUploadPersists uploads a small PNG via the hidden avatar file
// input, asserts the avatar <img> gets a blob: src, then reloads (same context)
// and asserts the avatar is rehydrated from IndexedDB with a blob: src again.
func TestProfileAvatarUploadPersists(t *testing.T) {
	page := newIsolatedPage(t)
	gotoProfile(t, page, "?seed=0")

	require.NoError(t, page.Locator("#profile-avatar-input").SetInputFiles(playwright.InputFile{
		Name:     "avatar.png",
		MimeType: "image/png",
		Buffer:   onePxPNG,
	}))

	_, err := page.WaitForFunction(
		"() => { const img = document.querySelector('#profile-avatar img'); return img && img.src.startsWith('blob:'); }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "avatar img should get a blob: src after upload")

	// Reload (same browser context → same IndexedDB) and verify rehydration.
	gotoProfile(t, page, "")
	_, err = page.WaitForFunction(
		"() => { const img = document.querySelector('#profile-avatar img'); return img && img.src.startsWith('blob:'); }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "avatar should rehydrate from IndexedDB with a blob: src on reload")
}

// TestProfileOversizeRejected uploads a >1MB image and asserts it is rejected
// client-side (no blob: preview on the avatar img).
func TestProfileOversizeRejected(t *testing.T) {
	page := newIsolatedPage(t)
	gotoProfile(t, page, "?seed=0")

	require.NoError(t, page.Locator("#profile-avatar-input").SetInputFiles(playwright.InputFile{
		Name:     "huge.png",
		MimeType: "image/png",
		Buffer:   make([]byte, 1024*1024+10),
	}))
	// The rejection path in onFile is synchronous: size check fires before any
	// async call, clears the input value, and dispatches a 'notify' event.
	// Wait for the file input to be cleared (ev.target.value='') as a positive
	// signal that the handler ran and rejected the file.
	_, err := page.WaitForFunction(
		"() => { const el = document.querySelector('#profile-avatar-input'); return el && el.value === ''; }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "oversize file input should be cleared by rejection handler")

	isBlob, err := page.Evaluate(
		"() => { const img = document.querySelector('#profile-avatar img'); return !!(img && img.src.startsWith('blob:')); }", nil)
	require.NoError(t, err)
	require.Equal(t, false, isBlob, "oversize image must be rejected client-side (no blob: preview)")
}

// TestProfileRemoveModal uploads an avatar (so the Remove modal trigger becomes
// visible), opens the confirm modal, clicks Remove, and asserts the avatar img
// no longer has a blob: src. Proves the Modal PrimaryAction.OnClick
// ("remove('avatar'); remove('banner')") resolves against the ancestor
// profileImages x-data and clears the image.
func TestProfileRemoveModal(t *testing.T) {
	page := newIsolatedPage(t)
	gotoProfile(t, page, "?seed=0")

	// Upload a small PNG so an image exists → the Remove modal trigger shows.
	require.NoError(t, page.Locator("#profile-avatar-input").SetInputFiles(playwright.InputFile{
		Name:     "avatar.png",
		MimeType: "image/png",
		Buffer:   onePxPNG,
	}))
	_, err := page.WaitForFunction(
		"() => { const img = document.querySelector('#profile-avatar img'); return img && img.src.startsWith('blob:'); }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "avatar img should get a blob: src after upload")

	// Open the confirm modal via its trigger button.
	trigger := page.Locator("#profile-photos button", playwright.PageLocatorOptions{HasText: "Remove photos"})
	require.NoError(t, trigger.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))
	require.NoError(t, trigger.Click())

	// The modal dialog (role="dialog", aria-labelledby="profile-remove-modalTitle").
	dialog := page.Locator("div[role='dialog'][aria-labelledby='profileRemoveModalTitle']")
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))

	// Click the modal's primary "Remove" button.
	confirm := dialog.Locator("button", playwright.LocatorLocatorOptions{HasText: "Remove"})
	require.NoError(t, confirm.Click())

	// The avatar img should no longer have a blob: src (removed).
	_, err = page.WaitForFunction(
		"() => { const img = document.querySelector('#profile-avatar img'); return !img || !img.src.startsWith('blob:'); }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "avatar img blob: src should be cleared after confirming remove in modal")
}
