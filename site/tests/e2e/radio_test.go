//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// radioIsChecked returns the live `checked` property of a radio input.
func radioIsChecked(t *testing.T, page playwright.Page, selector string) bool {
	t.Helper()
	v, err := page.Locator(selector).Evaluate("el => el.checked", nil)
	require.NoError(t, err)
	b, ok := v.(bool)
	require.True(t, ok, "expected bool checked value")
	return b
}

// TestRadio_DefaultChecks verifies clicking each radio in the default group
// flips `checked` exclusively per the shared name.
func TestRadio_DefaultChecks(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	// Initial state: Mac is checked, others are not.
	assert.True(t, radioIsChecked(t, page, "#r-d-mac"), "Mac should start checked")
	assert.False(t, radioIsChecked(t, page, "#r-d-win"), "Windows should start unchecked")

	require.NoError(t, page.Locator("#r-d-win").Click())
	assert.True(t, radioIsChecked(t, page, "#r-d-win"), "Windows should become checked after click")
	assert.False(t, radioIsChecked(t, page, "#r-d-mac"), "Mac must auto-uncheck (exclusive group)")
}

// TestRadio_Disabled verifies a disabled radio does not check on click.
func TestRadio_Disabled(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	// The first disabled radio starts unchecked. Force-click it (Playwright would
	// otherwise refuse to click a disabled element).
	_, err := page.Locator("#r-dis-1").Evaluate("el => el.click()", nil)
	require.NoError(t, err)
	assert.False(t, radioIsChecked(t, page, "#r-dis-1"), "disabled radio must remain unchecked")
}

// TestRadio_SegmentedContainer verifies the container variant renders with the
// expected border classes and toggles selection on click.
func TestRadio_SegmentedContainer(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	// Container variant's label has `rounded-radius border border-outline`.
	labelClass, err := page.Locator("label[for='r-c-mac']").GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, labelClass, "border-outline")
	assert.Contains(t, labelClass, "rounded-radius")

	require.NoError(t, page.Locator("#r-c-win").Click())
	assert.True(t, radioIsChecked(t, page, "#r-c-win"))
	assert.False(t, radioIsChecked(t, page, "#r-c-mac"))
}

// TestRadio_ColorVariants verifies each variant input renders the matching
// checked:bg-* token in its class string.
func TestRadio_ColorVariants(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	cases := []struct {
		id      string
		variant string
	}{
		{"r-v-primary", "primary"},
		{"r-v-secondary", "secondary"},
		{"r-v-info", "info"},
		{"r-v-success", "success"},
		{"r-v-warning", "warning"},
		{"r-v-danger", "danger"},
	}
	for _, tc := range cases {
		cls, err := page.Locator("#" + tc.id).GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, cls, "checked:before:bg-"+tc.variant,
			"%s radio missing checked:before:bg-%s", tc.id, tc.variant)
		assert.Contains(t, cls, "checked:border-"+tc.variant,
			"%s radio missing checked:border-%s", tc.id, tc.variant)
	}
}

// TestRadio_SizeSelector verifies the documented size radio selector controls
// the visible size preview without colliding with the demo radio groups.
func TestRadio_SizeSelector(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	cases := []struct {
		size      string
		inputID   string
		wantClass string
	}{
		{"sm", "r-size-sm", "size-3"},
		{"md", "r-size-md", "size-4"},
		{"lg", "r-size-lg", "size-5"},
		{"xl", "r-size-xl", "size-6"},
	}
	for _, tc := range cases {
		require.NoError(t, page.Locator("label[for='radio-size-selector-"+tc.size+"']").Click())
		require.NoError(t, page.Locator("[data-testid='radio-size-selected']").Filter(playwright.LocatorFilterOptions{
			HasText: tc.size,
		}).WaitFor())
		require.NoError(t, page.Locator("[data-testid='radio-size-preview-"+tc.size+"']").WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		cls, err := page.Locator("#" + tc.inputID).GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, cls, tc.wantClass, "%s radio missing %s", tc.inputID, tc.wantClass)
	}
}

// TestRadio_AlpinePrimitive verifies the Alpine-only showcase: clicking a radio
// updates the visible `selected` text without any server roundtrip.
func TestRadio_AlpinePrimitive(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	// Initial: selected = 'md'
	out := page.Locator("[data-testid='radio-alpine-out'] span")
	txt, err := out.TextContent()
	require.NoError(t, err)
	assert.Equal(t, "md", txt)

	require.NoError(t, page.Locator("label[for='r-a-lg']").Click())

	// Alpine reactivity: poll for change.
	_, err = page.WaitForFunction(
		`() => document.querySelector("[data-testid='radio-alpine-out'] span").textContent === 'lg'`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err, "Alpine should update selected to 'lg' after click")
	assert.True(t, radioIsChecked(t, page, "#r-a-lg"))

	// And BindChecked makes the previously-checked one unchecked.
	assert.False(t, radioIsChecked(t, page, "#r-a-md"))
}

// TestRadio_HTMXPrimitive verifies the HTMX-only showcase: clicking fires the
// /api/components/radio/echo endpoint and swaps the response into the target.
func TestRadio_HTMXPrimitive(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	require.NoError(t, page.Locator("label[for='r-h-lg']").Click())

	// Wait for the swap. The endpoint replies with text containing "Server: you picked".
	_, err := page.WaitForFunction(
		`() => {
			const el = document.querySelector("#radio-htmx-out");
			return el && el.textContent.includes("lg") && el.textContent.includes("Server: you picked");
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err, "HTMX swap should replace #radio-htmx-out with server response")
}

// TestRadio_HybridPrimitive verifies the hybrid showcase: a single click both
// updates Alpine local state and fires an HTMX swap.
func TestRadio_HybridPrimitive(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRadioDemo(t, page)

	require.NoError(t, page.Locator("label[for='r-hy-sm']").Click())

	// Alpine side
	_, err := page.WaitForFunction(
		`() => document.querySelector("[data-testid='radio-hybrid-alpine-out']").textContent === 'sm'`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
	)
	require.NoError(t, err, "hybrid Alpine state should flip to 'sm'")

	// HTMX side
	_, err = page.WaitForFunction(
		`() => {
			const el = document.querySelector("#radio-hybrid-out");
			return el && el.textContent.includes("sm") && el.textContent.includes("Server: you picked");
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err, "hybrid HTMX swap should land in #radio-hybrid-out")
}
