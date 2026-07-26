package badge

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderBadge(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Badge(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestBadge_LabelFieldRenders(t *testing.T) {
	html := renderBadge(t, Config{Label: "Live"})

	assert.Contains(t, html, "Live")
}

func TestBadge_SoftSemanticLabelsUseSurfaceContrast(t *testing.T) {
	tests := []struct {
		name string
		tone Tone
	}{
		{name: "primary", tone: TonePrimary},
		{name: "secondary", tone: ToneSecondary},
		{name: "info", tone: ToneInfo},
		{name: "success", tone: ToneSuccess},
		{name: "warning", tone: ToneWarning},
		{name: "danger", tone: ToneDanger},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			html := renderBadge(t, Config{
				Label:      strings.ToUpper(test.name),
				Tone:       test.tone,
				Appearance: AppearanceSoft,
			})

			assert.Contains(t, html, "text-on-surface-strong")
			assert.Contains(t, html, "dark:text-on-surface-dark-strong")
			assert.NotContains(t, html, fmt.Sprintf("text-%s ", test.name))
		})
	}
}
