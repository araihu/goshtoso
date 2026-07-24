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
	require.NotContains(t, html, `type="radio"`)
}

func renderStructuralRating(t *testing.T, component templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buf))
	return buf.String()
}
