package checkbox

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderCheckbox(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Checkbox(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestCheckbox_HelperTextRenders(t *testing.T) {
	html := renderCheckbox(t, Config{
		ID:         "terms",
		Label:      "Terms",
		HelperText: "Required before continuing",
	})

	assert.Contains(t, html, "Required before continuing")
}

func TestCheckbox_AnimationsProvideReducedMotionContracts(t *testing.T) {
	for _, animation := range []Animation{AnimationSlideUp, AnimationScaleUp, AnimationSlideDown} {
		html := renderCheckbox(t, Config{ID: "reduced-" + string(animation), Animation: animation})
		assert.Contains(t, html, "motion-reduce:transition-none", "SVG animation %q", animation)
		if animation != AnimationSlideUp {
			assert.Contains(t, html, "motion-reduce:before:transition-none", "pseudo-element animation %q", animation)
		}
	}
}
