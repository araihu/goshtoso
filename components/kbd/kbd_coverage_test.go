package kbd

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKbdCoverageHelperBranches(t *testing.T) {
	tests := []struct {
		name      string
		size      Size
		sizeClass string
		iconClass string
	}{
		{
			name:      "extra small",
			size:      SizeXS,
			sizeClass: "min-h-5 min-w-5 px-1 py-0.5 text-[10px] leading-none",
			iconClass: "size-3",
		},
		{
			name:      "small",
			size:      SizeSM,
			sizeClass: "min-h-6 min-w-6 px-1.5 py-0.5 text-xs leading-none",
			iconClass: "size-3.5",
		},
		{
			name:      "medium default",
			size:      SizeMD,
			sizeClass: "min-h-7 min-w-7 px-2 py-1 text-sm leading-none",
			iconClass: "size-4",
		},
		{
			name:      "large",
			size:      SizeLG,
			sizeClass: "min-h-9 min-w-9 px-2.5 py-1 text-base leading-none",
			iconClass: "size-5",
		},
		{
			name:      "unknown size falls back to medium",
			size:      Size("wide"),
			sizeClass: "min-h-7 min-w-7 px-2 py-1 text-sm leading-none",
			iconClass: "size-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig("", []Option{WithSize(tt.size)})
			assert.Equal(t, tt.sizeClass, cfg.sizeClasses())
			assert.Equal(t, tt.iconClass, cfg.iconClasses())
			assert.Contains(t, cfg.rootClasses(), tt.sizeClass)
		})
	}

	cfg := newConfig("", []Option{WithRootClass("shortcut-key")})
	assert.Contains(t, cfg.rootClasses(), "shortcut-key")
}

func TestKbdCoverageAccessibleLabelBranches(t *testing.T) {
	assert.Empty(t, newConfig("Esc", []Option{WithLabel("Escape")}).accessibleLabel())
	assert.Equal(t, "Command", newConfig("", []Option{WithLabel("  Command  ")}).accessibleLabel())
	assert.Empty(t, newConfig("", []Option{WithLabel("   ")}).accessibleLabel())
}

func TestKbdCoverageRenderTextLabelAndAttributes(t *testing.T) {
	rendered := renderCoverageKbd(t, "<Esc>",
		WithLabel("Escape key"),
		WithSize(SizeLG),
		WithRootClass("shortcut-key"),
		WithAttrs(templ.Attributes{
			"id":            "escape-key",
			"data-shortcut": "escape",
			"aria-hidden":   "false",
		}),
	)

	assert.Contains(t, rendered, `<kbd`)
	assert.Contains(t, rendered, `id="escape-key"`)
	assert.Contains(t, rendered, `data-shortcut="escape"`)
	assert.Contains(t, rendered, `aria-label="Escape key"`)
	assert.Contains(t, rendered, `aria-hidden="false"`)
	assert.Contains(t, rendered, `shortcut-key`)
	assert.Contains(t, rendered, `min-h-9`)
	assert.Contains(t, rendered, `&lt;Esc&gt;`)
	assert.NotContains(t, rendered, `sr-only`)
}

func TestKbdCoverageRenderIconOnlyUsesTrimmedScreenReaderLabel(t *testing.T) {
	icon := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<svg data-testid="coverage-icon"></svg>`)
		return err
	})

	rendered := renderCoverageKbd(t, "",
		WithLabel("  Shift  "),
		WithIcon(icon),
		WithSize(SizeXS),
	)

	assert.Contains(t, rendered, `aria-label="  Shift  "`)
	assert.Contains(t, rendered, `aria-hidden="true"`)
	assert.Contains(t, rendered, `data-testid="coverage-icon"`)
	assert.Contains(t, rendered, `class="shrink-0 size-3"`)
	assert.Contains(t, rendered, `<span class="sr-only">Shift</span>`)
}

func renderCoverageKbd(t *testing.T, text string, options ...Option) string {
	t.Helper()

	var buf bytes.Buffer
	err := Kbd(text, options...).Render(context.Background(), &buf)
	require.NoError(t, err)
	return strings.TrimSpace(buf.String())
}
