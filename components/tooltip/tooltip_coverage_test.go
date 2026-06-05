package tooltip

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestTooltipCoverageConfigHelpers(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		position string
		rich     bool
		id       string
		label    string
	}{
		{
			name:     "defaults",
			cfg:      Config{},
			position: "absolute bottom-full mb-2 left-1/2 -translate-x-1/2",
			id:       "tooltipExample",
			label:    "Hover Me",
		},
		{
			name: "bottom rich custom labels",
			cfg: Config{
				ID:           "help",
				Description:  "More details",
				Position:     Bottom,
				TriggerLabel: "Details",
			},
			position: "absolute top-full mt-2 left-1/2 -translate-x-1/2",
			rich:     true,
			id:       "help",
			label:    "Details",
		},
		{
			name:     "left",
			cfg:      Config{Position: Left},
			position: "absolute right-full mr-2 top-1/2 -translate-y-1/2",
			id:       "tooltipExample",
			label:    "Hover Me",
		},
		{
			name:     "right",
			cfg:      Config{Position: Right},
			position: "absolute left-full ml-2 top-1/2 -translate-y-1/2",
			id:       "tooltipExample",
			label:    "Hover Me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.positionClasses(); got != tt.position {
				t.Fatalf("positionClasses() = %q, want %q", got, tt.position)
			}
			if got := tt.cfg.isRich(); got != tt.rich {
				t.Fatalf("isRich() = %v, want %v", got, tt.rich)
			}
			if got := tt.cfg.tooltipID(); got != tt.id {
				t.Fatalf("tooltipID() = %q, want %q", got, tt.id)
			}
			if got := tt.cfg.triggerLabel(); got != tt.label {
				t.Fatalf("triggerLabel() = %q, want %q", got, tt.label)
			}
		})
	}
}

func TestTooltipCoverageRenderDefaultRichClickAndCustomTrigger(t *testing.T) {
	customTrigger := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(`<span data-testid="custom-trigger">?</span>`))
		return err
	})

	tests := []struct {
		name string
		cfg  Config
		want []string
	}{
		{
			name: "default hover tooltip",
			cfg: Config{
				ID:           "plainTip",
				Label:        "Helpful context",
				Position:     Top,
				TriggerLabel: "Hover help",
			},
			want: []string{
				`type="button"`,
				`aria-describedby="plainTip"`,
				`id="plainTip"`,
				`role="tooltip"`,
				"Hover help",
				"Helpful context",
				"peer-hover:opacity-100",
				"bottom-full mb-2",
			},
		},
		{
			name: "rich tooltip",
			cfg: Config{
				ID:           "richTip",
				Label:        "Storage",
				Description:  "Backed up every hour",
				Position:     Bottom,
				TriggerLabel: "Show storage help",
			},
			want: []string{
				`aria-describedby="richTip"`,
				`id="richTip"`,
				`role="tooltip"`,
				"Storage",
				"Backed up every hour",
				"flex w-64 flex-col",
				"text-balance",
				"top-full mt-2",
			},
		},
		{
			name: "click tooltip",
			cfg: Config{
				ID:           "clickTip",
				Label:        "Click details",
				Position:     Right,
				TriggerMode:  Click,
				TriggerLabel: "Toggle help",
			},
			want: []string{
				`x-data="{ showTooltip: false }"`,
				`x-on:click="showTooltip = !showTooltip"`,
				`x-show="showTooltip"`,
				`x-on:click.outside="showTooltip = false"`,
				`aria-describedby="clickTip"`,
				`id="clickTip"`,
				"left-full ml-2",
				"Click details",
			},
		},
		{
			name: "custom trigger",
			cfg: Config{
				ID:       "customTip",
				Label:    "Custom trigger context",
				Position: Left,
				Trigger:  customTrigger,
			},
			want: []string{
				`class="peer"`,
				`aria-describedby="customTip"`,
				`data-testid="custom-trigger"`,
				"Custom trigger context",
				"right-full mr-2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderTooltip(t, tt.cfg)
			for _, want := range tt.want {
				if !strings.Contains(html, want) {
					t.Fatalf("Tooltip render missing %q in %s", want, html)
				}
			}
		})
	}
}

func renderTooltip(t *testing.T, cfg Config) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Tooltip(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render tooltip: %v", err)
	}
	return buf.String()
}
