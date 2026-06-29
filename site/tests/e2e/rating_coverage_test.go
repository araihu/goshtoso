package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// iconClassForLabel reads the live (Alpine-bound) class of the visual icon span
// inside the label for the given input id.
func iconClassForLabel(t *testing.T, page playwright.Page, inputID string) string {
	t.Helper()
	loc := page.Locator("label[for='" + inputID + "'] span[aria-hidden='true']")
	v, err := loc.Evaluate("el => el.getAttribute('class')", nil)
	require.NoError(t, err)
	class, ok := v.(string)
	require.True(t, ok)
	return class
}

func TestRating_LiveStarBindingHighlightsCumulative(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRatingDemo(t, page)

	// Default value is 3: stars 1..3 active, 4..5 inactive.
	assert.Contains(t, iconClassForLabel(t, page, "rating-stars-3"), "text-warning")
	assert.NotContains(t, iconClassForLabel(t, page, "rating-stars-5"), "text-warning")

	require.NoError(t, page.Locator("label[for='rating-stars-5']").Click())

	// After selecting 5, all stars become active (cumulative binding).
	for _, id := range []string{"rating-stars-1", "rating-stars-3", "rating-stars-5"} {
		assert.Contains(t, iconClassForLabel(t, page, id), "text-warning",
			"star %s should be active after selecting 5", id)
	}
}

func TestRating_KeyboardArrowSelects(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRatingDemo(t, page)

	require.NoError(t, page.Locator("#rating-stars-3").Focus())
	require.NoError(t, page.Keyboard().Press("ArrowRight"))

	assert.True(t, ratingChecked(t, page, "#rating-stars-4"),
		"arrow-right should move radio selection to the next star")
	// Live binding should follow the keyboard-driven selection.
	assert.Contains(t, iconClassForLabel(t, page, "rating-stars-4"), "text-warning")
}

func TestRating_SizesRenderDistinctIconScales(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRatingDemo(t, page)

	cases := []struct {
		size      string
		inputID   string
		wantClass string
	}{
		{"sm", "rating-size-preview-sm-1", "size-5"},
		{"lg", "rating-size-preview-lg-1", "size-8"},
		{"xl", "rating-size-preview-xl-1", "size-10"},
	}
	for _, tc := range cases {
		require.NoError(t, page.Locator("label[for='rating-size-"+tc.size+"']").Click())
		require.NoError(t, page.Locator("[data-testid='rating-size-selected']").Filter(playwright.LocatorFilterOptions{
			HasText: tc.size,
		}).WaitFor())
		require.NoError(t, page.Locator("[data-testid='rating-size-preview-"+tc.size+"']").WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		assert.Contains(t, iconClassForLabel(t, page, tc.inputID), tc.wantClass,
			"icon for %s should carry %s", tc.inputID, tc.wantClass)
	}
}

func TestRating_DemoLoadsWithoutConsoleErrors(t *testing.T) {
	page := newPage(t, sharedBrowser)

	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	navigateToRatingDemo(t, page)

	require.NoError(t, page.Locator("label[for='rating-stars-4']").Click())
	require.NoError(t, page.Locator("label[for='rating-sentiment-2']").Click())

	assert.Empty(t, consoleErrors,
		"rating demo should not log console errors: %s", strings.Join(consoleErrors, "; "))
}
