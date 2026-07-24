package rating

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestRatingDisplayHasNoFormInputs(t *testing.T) {
	html := renderStructuralRating(t, RatingDisplay(DisplayConfig{
		Value: 4,
		Label: "Four out of five",
	}))

	require.Contains(t, html, `role="img"`)
	require.Contains(t, html, `aria-label="Four out of five"`)
	require.NotContains(t, html, `type="radio"`)
}

func TestRatingDisplayFallsBackToValueLabel(t *testing.T) {
	html := renderStructuralRating(t, RatingDisplay(DisplayConfig{
		Value: 4,
	}))

	require.Contains(t, html, `aria-label="4 stars"`)
}

func renderStructuralRating(t *testing.T, component templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buf))
	return buf.String()
}
