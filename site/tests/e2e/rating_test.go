//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRating_DefaultSelectionUpdates(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRatingDemo(t, page)

	assert.True(t, ratingChecked(t, page, "#rating-stars-3"))
	require.NoError(t, page.Locator("label[for='rating-stars-5']").Click())
	assert.True(t, ratingChecked(t, page, "#rating-stars-5"))
	assert.False(t, ratingChecked(t, page, "#rating-stars-3"))
}

func TestRating_EmojiSelectionUpdates(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRatingDemo(t, page)

	assert.True(t, ratingChecked(t, page, "#rating-sentiment-4"))
	require.NoError(t, page.Locator("label[for='rating-sentiment-1']").Click())
	assert.True(t, ratingChecked(t, page, "#rating-sentiment-1"))

	labelText, err := page.Locator("label[for='rating-sentiment-1'] span.sr-only").TextContent()
	require.NoError(t, err)
	assert.Equal(t, "very dissatisfied", labelText)

	activeEmojiCount, err := page.Locator("#rating-emoji [aria-hidden='true'].scale-110").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, activeEmojiCount, "emoji rating should activate exactly one sentiment")
}

func TestRating_DisabledAndDisplay(t *testing.T) {
	page := newPage(t, sharedBrowser)
	navigateToRatingDemo(t, page)

	_, err := page.Locator("#rating-disabled-control-5").Evaluate("el => el.click()", nil)
	require.NoError(t, err)
	assert.False(t, ratingChecked(t, page, "#rating-disabled-control-5"))
	assert.True(t, ratingChecked(t, page, "#rating-disabled-control-2"))

	displayRadios, err := page.Locator("#rating-readonly input[type='radio']").Count()
	require.NoError(t, err)
	assert.Equal(t, 0, displayRadios)

	label, err := page.Locator("#rating-readonly-summary").GetAttribute("aria-label")
	require.NoError(t, err)
	assert.Equal(t, "Average rating", label)
}
