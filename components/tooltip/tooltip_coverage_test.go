package tooltip

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestTooltipCoverageConfigHelpers(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		label    string
		options  []Option
		position string
		rich     bool
		trigger  string
	}{
		{
			name:     "defaults",
			id:       "default-tip",
			label:    "Default tooltip",
			position: "absolute bottom-full mb-2 left-1/2 -translate-x-1/2",
			trigger:  "Hover Me",
		},
		{
			name:  "bottom rich custom labels",
			id:    "help",
			label: "Help",
			options: []Option{
				WithDescription("More details"),
				WithPosition(PositionBottom),
				WithTriggerLabel("Details"),
			},
			position: "absolute top-full mt-2 left-1/2 -translate-x-1/2",
			rich:     true,
			trigger:  "Details",
		},
		{
			name:     "left",
			id:       "left-tip",
			label:    "Left",
			options:  []Option{WithPosition(PositionLeft)},
			position: "absolute right-full mr-2 top-1/2 -translate-y-1/2",
			trigger:  "Hover Me",
		},
		{
			name:     "right",
			id:       "right-tip",
			label:    "Right",
			options:  []Option{WithPosition(PositionRight)},
			position: "absolute left-full ml-2 top-1/2 -translate-y-1/2",
			trigger:  "Hover Me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(tt.id, tt.label, tt.options)
			if got := cfg.positionClasses(); got != tt.position {
				t.Fatalf("positionClasses() = %q, want %q", got, tt.position)
			}
			if got := cfg.isRich(); got != tt.rich {
				t.Fatalf("isRich() = %v, want %v", got, tt.rich)
			}
			if cfg.id != tt.id {
				t.Fatalf("id = %q, want %q", cfg.id, tt.id)
			}
			if cfg.triggerLabel != tt.trigger {
				t.Fatalf("triggerLabel = %q, want %q", cfg.triggerLabel, tt.trigger)
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
		name    string
		id      string
		label   string
		options []Option
		want    []string
	}{
		{
			name:    "default hover tooltip",
			id:      "plainTip",
			label:   "Helpful context",
			options: []Option{WithTriggerLabel("Hover help")},
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
			name:  "rich tooltip",
			id:    "richTip",
			label: "Storage",
			options: []Option{
				WithDescription("Backed up every hour"),
				WithPosition(PositionBottom),
				WithTriggerLabel("Show storage help"),
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
			name:  "click tooltip",
			id:    "clickTip",
			label: "Click details",
			options: []Option{
				WithPosition(PositionRight),
				WithActivation(ActivationClick),
				WithTriggerLabel("Toggle help"),
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
			name:  "custom trigger",
			id:    "customTip",
			label: "Custom trigger context",
			options: []Option{
				WithPosition(PositionLeft),
				WithTrigger(customTrigger),
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
			html := renderTooltip(t, tt.id, tt.label, tt.options...)
			for _, want := range tt.want {
				if !strings.Contains(html, want) {
					t.Fatalf("Tooltip render missing %q in %s", want, html)
				}
			}
		})
	}
}

func TestTooltipCoverageCustomTriggerRenderErrors(t *testing.T) {
	errTrigger := errors.New("trigger render failed")
	failingTrigger := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return errTrigger
	})

	tests := []struct {
		name       string
		activation Activation
		rich       bool
	}{
		{
			name: "default",
		},
		{
			name: "rich",
			rich: true,
		},
		{
			name:       "click",
			activation: ActivationClick,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			options := []Option{WithTrigger(failingTrigger)}
			if tt.rich {
				options = append(options, WithDescription("Trigger should fail before details render"))
			}
			if tt.activation != "" {
				options = append(options, WithActivation(tt.activation))
			}
			err := Tooltip(tt.name+"-tip", tt.name+" tooltip", options...).Render(context.Background(), &buf)
			if !errors.Is(err, errTrigger) {
				t.Fatalf("Tooltip render error = %v, want %v", err, errTrigger)
			}
		})
	}
}

func renderTooltip(t *testing.T, id, label string, options ...Option) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Tooltip(id, label, options...).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render tooltip: %v", err)
	}
	return buf.String()
}

func TestTooltipRequiresIDAndLabel(t *testing.T) {
	var buf bytes.Buffer
	err := Tooltip(
		"copy-url-tooltip",
		"Copies the URL",
		WithPosition(PositionBottom),
	).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render tooltip: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Copies the URL") {
		t.Fatalf("tooltip required label missing in %s", html)
	}
	if !strings.Contains(html, `id="copy-url-tooltip"`) {
		t.Fatalf("tooltip required ID missing in %s", html)
	}
	if !strings.Contains(html, "top-full mt-2") {
		t.Fatalf("tooltip position option missing in %s", html)
	}
}
