package rangeinput

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderCov(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Range(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render range: %v", err)
	}
	return buf.String()
}

// Icons exercise iconClasses (0% baseline), the leading/trailing icon
// branches in rangeInner, the icon-aware controlClasses path, and the
// label-with-decorations branch.
func TestCoverageWithIcons(t *testing.T) {
	html := renderCov(t, Config{
		ID:           "brightness",
		Label:        "Brightness",
		LeadingIcon:  templ.Raw(`<svg id="lead"></svg>`),
		TrailingIcon: templ.Raw(`<svg id="trail"></svg>`),
	})

	for _, want := range []string{
		`<svg id="lead">`,
		`<svg id="trail">`,
		`aria-hidden="true"`,
		`shrink-0`,                       // iconClasses
		`flex w-full items-center gap-3`, // controlClasses with decorations
		`Brightness`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in:\n%s", want, html)
		}
	}

	// With decorations present the label moves inside the control row, so the
	// standalone top label branch must NOT render twice.
	if got := strings.Count(html, `>Brightness</label>`); got != 1 {
		t.Fatalf("expected exactly one label, got %d in:\n%s", got, html)
	}
}

// LeadingIcon alone still takes the decorated controlClasses branch.
func TestCoverageLeadingIconOnly(t *testing.T) {
	html := renderCov(t, Config{ID: "x", LeadingIcon: templ.Raw(`<svg id="only"></svg>`)})

	if !strings.Contains(html, `<svg id="only">`) {
		t.Fatalf("expected leading icon in:\n%s", html)
	}
	if !strings.Contains(html, `flex w-full items-center gap-3`) {
		t.Fatalf("expected decorated control row in:\n%s", html)
	}
}

// RootClass and InputClass exercise the non-empty branches of rootClasses and
// inputClasses.
func TestCoverageExtraClasses(t *testing.T) {
	html := renderCov(t, Config{
		ID:         "x",
		RootClass:  "my-root-extra",
		InputClass: "my-input-extra",
	})

	for _, want := range []string{"my-root-extra", "my-input-extra"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in:\n%s", want, html)
		}
	}
}

// InputAttrs spreads arbitrary hx-*/x-* hooks onto the native input.
func TestCoverageInputAttrs(t *testing.T) {
	html := renderCov(t, Config{
		ID:   "x",
		Name: "x",
		InputAttrs: templ.Attributes{
			"hx-post":   "/api/range",
			"data-test": "range-hook",
		},
	})

	for _, want := range []string{`hx-post="/api/range"`, `data-test="range-hook"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in:\n%s", want, html)
		}
	}
}

// alpineData must fall back to 0 for non-numeric values rather than emitting
// invalid Alpine state.
func TestCoverageAlpineDataInvalidValue(t *testing.T) {
	cfg := Config{ID: "x", Value: "not-a-number", ShowValue: true}
	if got := cfg.alpineData(); got != "{ currentVal: 0 }" {
		t.Fatalf("expected fallback alpine data, got %q", got)
	}

	html := renderCov(t, cfg)
	if !strings.Contains(html, `x-data="{ currentVal: 0 }"`) {
		t.Fatalf("expected sanitized x-data in:\n%s", html)
	}
}

// alpineData should preserve fractional values exactly.
func TestCoverageAlpineDataFractional(t *testing.T) {
	cfg := Config{Value: "12.5"}
	if got := cfg.alpineData(); got != "{ currentVal: 12.5 }" {
		t.Fatalf("expected fractional alpine data, got %q", got)
	}
}

// Custom Ticks take the override branch of tickLabels and render verbatim.
func TestCoverageCustomTicks(t *testing.T) {
	html := renderCov(t, Config{
		ID:        "x",
		ShowTicks: true,
		Ticks: []Tick{
			{Label: "Low"},
			{Label: "Mid", HideOnMobile: true},
			{Label: "High"},
		},
	})

	for _, want := range []string{`>Low</span>`, `>Mid</span>`, `>High</span>`, `hidden sm:inline`} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in:\n%s", want, html)
		}
	}
	// Exactly the three custom ticks, not the default 21-mark scale.
	if got := strings.Count(html, `aria-hidden="true"`); got != 1 {
		t.Fatalf("expected single tick row wrapper, got %d in:\n%s", got, html)
	}
}

// tickClasses returns the non-hidden base classes when HideOnMobile is false.
func TestCoverageTickClassesVisible(t *testing.T) {
	cfg := Config{}
	visible := cfg.tickClasses(Tick{Label: "0"})
	if strings.Contains(visible, "hidden") {
		t.Fatalf("expected visible tick classes, got %q", visible)
	}
	hidden := cfg.tickClasses(Tick{Label: "|", HideOnMobile: true})
	if !strings.Contains(hidden, "hidden sm:inline") {
		t.Fatalf("expected hidden tick classes, got %q", hidden)
	}
}

// stepOrDefault defaults to "1" and honors an explicit step.
func TestCoverageStepOrDefault(t *testing.T) {
	if got := (Config{}).stepOrDefault(); got != "1" {
		t.Fatalf("expected default step 1, got %q", got)
	}
	if got := (Config{Step: "5"}).stepOrDefault(); got != "5" {
		t.Fatalf("expected step 5, got %q", got)
	}
}

// ShowValue with no label and no icons still routes through the decorated
// control row (ShowValue forces it) and renders the live badge.
func TestCoverageShowValueNoLabel(t *testing.T) {
	html := renderCov(t, Config{ID: "x", Value: "7", ShowValue: true})

	if !strings.Contains(html, `flex w-full items-center gap-3`) {
		t.Fatalf("expected decorated control row in:\n%s", html)
	}
	if !strings.Contains(html, `x-text="currentVal"`) || !strings.Contains(html, `>7</span>`) {
		t.Fatalf("expected live value badge in:\n%s", html)
	}
}
