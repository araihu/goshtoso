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
		cfg       Config
		sizeClass string
		iconClass string
	}{
		{
			name:      "extra small",
			cfg:       Config{Size: SizeXS},
			sizeClass: "min-h-5 min-w-5 px-1 py-0.5 text-[10px] leading-none",
			iconClass: "size-3",
		},
		{
			name:      "small",
			cfg:       Config{Size: SizeSM},
			sizeClass: "min-h-6 min-w-6 px-1.5 py-0.5 text-xs leading-none",
			iconClass: "size-3.5",
		},
		{
			name:      "medium default",
			cfg:       Config{Size: SizeMD},
			sizeClass: "min-h-7 min-w-7 px-2 py-1 text-sm leading-none",
			iconClass: "size-4",
		},
		{
			name:      "large",
			cfg:       Config{Size: SizeLG},
			sizeClass: "min-h-9 min-w-9 px-2.5 py-1 text-base leading-none",
			iconClass: "size-5",
		},
		{
			name:      "unknown size falls back to medium",
			cfg:       Config{Size: Size("wide")},
			sizeClass: "min-h-7 min-w-7 px-2 py-1 text-sm leading-none",
			iconClass: "size-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.sizeClass, tt.cfg.SizeClasses())
			assert.Equal(t, tt.iconClass, tt.cfg.IconClasses())
			assert.Contains(t, tt.cfg.RootClasses(), tt.sizeClass)
		})
	}

	assert.Contains(t, Config{Class: "shortcut-key"}.RootClasses(), "shortcut-key")
}

func TestKbdCoverageAccessibleLabelBranches(t *testing.T) {
	assert.Empty(t, Config{Text: "Esc", Label: "Escape"}.AccessibleLabel())
	assert.Equal(t, "Command", Config{Label: "  Command  "}.AccessibleLabel())
	assert.Empty(t, Config{Label: "   "}.AccessibleLabel())
}

func TestKbdCoverageRenderTextLabelAndAttributes(t *testing.T) {
	rendered := renderCoverageKbd(t, Config{
		Text:  "<Esc>",
		Label: "Escape key",
		Size:  SizeLG,
		Class: "shortcut-key",
		Attrs: templ.Attributes{
			"id":            "escape-key",
			"data-shortcut": "escape",
			"aria-hidden":   "false",
		},
	})

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

	rendered := renderCoverageKbd(t, Config{
		Label: "  Shift  ",
		Icon:  icon,
		Size:  SizeXS,
	})

	assert.Contains(t, rendered, `aria-label="  Shift  "`)
	assert.Contains(t, rendered, `aria-hidden="true"`)
	assert.Contains(t, rendered, `data-testid="coverage-icon"`)
	assert.Contains(t, rendered, `class="shrink-0 size-3"`)
	assert.Contains(t, rendered, `<span class="sr-only">Shift</span>`)
}

func renderCoverageKbd(t *testing.T, cfg Config) string {
	t.Helper()

	var buf bytes.Buffer
	err := Kbd(cfg).Render(context.Background(), &buf)
	require.NoError(t, err)
	return strings.TrimSpace(buf.String())
}
