package radio

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderRadio(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Radio(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestRadio_HelperTextRenders(t *testing.T) {
	html := renderRadio(t, Config{
		ID:         "daily",
		Name:       "cadence",
		Value:      "daily",
		Label:      "Daily",
		HelperText: "Runs every day",
	})

	assert.Contains(t, html, "Runs every day")
}

func TestRadio_AnimationsProvideReducedMotionContracts(t *testing.T) {
	standard := renderRadio(t, Config{ID: "reduced-standard", Name: "reduced", Value: "standard", Label: "Standard"})
	assert.Contains(t, standard, "motion-reduce:before:transition-none")

	segmented := renderRadio(t, Config{ID: "reduced-segmented", Name: "reduced", Value: "segmented", Label: "Segmented", Segmented: true})
	assert.Contains(t, segmented, "transition-colors motion-reduce:transition-none")
}

func TestLabelLeadingRendersWithAccessibleText(t *testing.T) {
	for _, cfg := range []Config{
		{ID: "mode-dark", Name: "mode", Value: "dark", Label: "Dark", Segmented: true},
		{ID: "mode-standard", Name: "mode", Value: "system", Label: "System"},
		{ID: "mode-container", Name: "mode", Value: "light", Label: "Light", Container: true},
	} {
		cfg.LabelLeading = templ.Raw(`<svg aria-hidden="true" data-radio-leading></svg>`)
		html := renderRadio(t, cfg)
		assert.Contains(t, html, `data-radio-leading`)
		assert.Contains(t, html, cfg.Label)
		assert.Contains(t, html, `type="radio"`)
		assert.Contains(t, html, `for="`+cfg.ID+`"`)
	}
}
