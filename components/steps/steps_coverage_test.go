package steps

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultSteps(t *testing.T) {
	html := render(t, Steps(Config{}))
	if !strings.Contains(html, "<ol") {
		t.Fatalf("expected ordered list, got %s", html)
	}
	if !strings.Contains(html, `aria-label="progress"`) {
		t.Fatalf("expected default aria-label, got %s", html)
	}
	// Default orientation is horizontal.
	if !strings.Contains(html, "items-center gap-2") {
		t.Fatalf("expected horizontal list classes, got %s", html)
	}
}

func TestStepsHorizontalFull(t *testing.T) {
	html := render(t, Steps(Config{
		ID:         "wizard",
		ShowLabels: true,
		AriaLabel:  "checkout flow",
		LiveRegion: true,
		RootClass:  "custom-root",
		RootAttrs:  templ.Attributes{"data-test": "root"},
		Steps: []Step{
			{Label: "Create account", Status: StatusCompleted},
			{Label: "Select plan", Status: StatusCurrent},
			{Label: "Checkout", Status: StatusUpcoming},
		},
	}))

	checks := []string{
		`id="wizard"`,
		`aria-label="checkout flow"`,
		`aria-live="polite"`,
		`aria-atomic="true"`,
		`custom-root`,
		`data-test="root"`,
		`aria-current="step"`, // current step
		`Create account`,
		`Select plan`,
		`Checkout`,
		`id="step-1"`,
		`id="step-2"`,
		`id="step-3"`,
		`completed`,             // sr-only completed text
		`m4.5 12.75 6 6 9-13.5`, // check icon path
		`outline outline-2`,     // current indicator outline
	}
	for _, want := range checks {
		if !strings.Contains(html, want) {
			t.Fatalf("horizontal render missing %q in %s", want, html)
		}
	}

	// First step has no connector; later steps do (bg-primary for completed/current).
	if strings.Count(html, "aria-hidden=\"true\"") < 2 {
		t.Fatalf("expected connectors on non-first steps, got %s", html)
	}
}

func TestStepsVerticalFull(t *testing.T) {
	html := render(t, Steps(Config{
		Orientation: OrientationVertical,
		ShowLabels:  true,
		Steps: []Step{
			{Label: "One", Status: StatusCompleted},
			{Label: "Two", Status: StatusCurrent},
			{Label: "Three"}, // empty status -> upcoming default
		},
	}))

	if !strings.Contains(html, "flex-col gap-14") {
		t.Fatalf("expected vertical list classes, got %s", html)
	}
	// Vertical connector classes.
	if !strings.Contains(html, "absolute bottom-8 left-3") {
		t.Fatalf("expected vertical connector, got %s", html)
	}
	for _, want := range []string{"One", "Two", "Three", `aria-current="step"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("vertical render missing %q in %s", want, html)
		}
	}
}

func TestStepsLabelsHiddenWhenDisabled(t *testing.T) {
	html := render(t, Steps(Config{
		ShowLabels: false,
		Steps:      []Step{{Label: "Hidden", Status: StatusCurrent}},
	}))
	// The visible label span (labelClasses base "inline w-max") must be absent;
	// the text may still appear in the item's aria-label via resolvedItemLabel.
	if strings.Contains(html, "inline w-max") {
		t.Fatalf("visible label span should be hidden when ShowLabels is false, got %s", html)
	}
	// Number shown for non-completed.
	if !strings.Contains(html, ">1<") {
		t.Fatalf("expected default number 1, got %s", html)
	}
}

func TestStepsAriaLabelOverrideAndNumber(t *testing.T) {
	html := render(t, Steps(Config{
		Steps: []Step{
			{Label: "Visible", AriaLabel: "explicit aria", Number: 9, Status: StatusUpcoming},
		},
	}))
	if !strings.Contains(html, `aria-label="explicit aria"`) {
		t.Fatalf("expected step aria-label override, got %s", html)
	}
	if !strings.Contains(html, ">9<") {
		t.Fatalf("expected custom number 9, got %s", html)
	}
}

func TestStepsItemLabelFallsBackToLabel(t *testing.T) {
	html := render(t, Steps(Config{
		Steps: []Step{{Label: "FallbackLabel", Status: StatusUpcoming}},
	}))
	if !strings.Contains(html, `aria-label="FallbackLabel"`) {
		t.Fatalf("expected item aria-label to fall back to Label, got %s", html)
	}
}

func TestNormalizedOrientation(t *testing.T) {
	if got := (Config{}).normalizedOrientation(); got != OrientationHorizontal {
		t.Fatalf("default orientation = %q, want horizontal", got)
	}
	if got := (Config{Orientation: OrientationVertical}).normalizedOrientation(); got != OrientationVertical {
		t.Fatalf("vertical orientation = %q, want vertical", got)
	}
	// Unknown value normalizes to horizontal.
	if got := (Config{Orientation: Orientation("weird")}).normalizedOrientation(); got != OrientationHorizontal {
		t.Fatalf("unknown orientation = %q, want horizontal", got)
	}
}

func TestListClasses(t *testing.T) {
	h := Config{}.listClasses()
	if !strings.Contains(h, "w-full items-center") {
		t.Fatalf("horizontal listClasses = %q", h)
	}
	v := Config{Orientation: OrientationVertical, RootClass: "extra"}.listClasses()
	if !strings.Contains(v, "flex-col") || !strings.Contains(v, "extra") {
		t.Fatalf("vertical listClasses = %q", v)
	}
}

func TestResolvedAriaLabel(t *testing.T) {
	if got := (Config{}).resolvedAriaLabel(); got != "progress" {
		t.Fatalf("default aria label = %q", got)
	}
	if got := (Config{AriaLabel: "custom"}).resolvedAriaLabel(); got != "custom" {
		t.Fatalf("custom aria label = %q", got)
	}
}

func TestResolvedStepDefaults(t *testing.T) {
	got := resolvedStep(Step{}, 2)
	if got.ID != "step-3" {
		t.Fatalf("default ID = %q, want step-3", got.ID)
	}
	if got.Number != 3 {
		t.Fatalf("default Number = %d, want 3", got.Number)
	}
	if got.Status != StatusUpcoming {
		t.Fatalf("default Status = %q, want upcoming", got.Status)
	}

	preset := resolvedStep(Step{ID: "x", Number: 7, Status: StatusCurrent}, 0)
	if preset.ID != "x" || preset.Number != 7 || preset.Status != StatusCurrent {
		t.Fatalf("preset step mutated: %+v", preset)
	}
}

func TestConnectorClasses(t *testing.T) {
	completedH := connectorClasses(StatusCompleted, OrientationHorizontal)
	if !strings.Contains(completedH, "bg-primary") || !strings.Contains(completedH, "w-full") {
		t.Fatalf("completed horizontal connector = %q", completedH)
	}
	currentV := connectorClasses(StatusCurrent, OrientationVertical)
	if !strings.Contains(currentV, "bg-primary") || !strings.Contains(currentV, "absolute") {
		t.Fatalf("current vertical connector = %q", currentV)
	}
	upcomingH := connectorClasses(StatusUpcoming, OrientationHorizontal)
	if !strings.Contains(upcomingH, "bg-outline") {
		t.Fatalf("upcoming connector = %q", upcomingH)
	}
}

func TestIndicatorClasses(t *testing.T) {
	if !strings.Contains(indicatorClasses(StatusCompleted), "bg-primary text-on-primary") {
		t.Fatalf("completed indicator wrong")
	}
	if !strings.Contains(indicatorClasses(StatusCurrent), "outline-offset-2") {
		t.Fatalf("current indicator wrong")
	}
	if !strings.Contains(indicatorClasses(StatusUpcoming), "bg-surface-alt") {
		t.Fatalf("upcoming indicator wrong")
	}
}

func TestLabelClasses(t *testing.T) {
	if !strings.Contains(labelClasses(StatusCompleted), "text-primary") {
		t.Fatalf("completed label wrong")
	}
	if !strings.Contains(labelClasses(StatusCurrent), "font-bold text-primary") {
		t.Fatalf("current label wrong")
	}
	if !strings.Contains(labelClasses(StatusUpcoming), "text-on-surface") {
		t.Fatalf("upcoming label wrong")
	}
}

func TestHorizontalItemClasses(t *testing.T) {
	if horizontalItemClasses(0) != "" {
		t.Fatalf("first item should have empty classes")
	}
	if !strings.Contains(horizontalItemClasses(1), "w-full") {
		t.Fatalf("non-first item classes wrong")
	}
}

func TestResolvedItemLabel(t *testing.T) {
	if got := resolvedItemLabel(Step{AriaLabel: "a", Label: "b"}); got != "a" {
		t.Fatalf("aria override = %q, want a", got)
	}
	if got := resolvedItemLabel(Step{Label: "b"}); got != "b" {
		t.Fatalf("label fallback = %q, want b", got)
	}
	if got := resolvedItemLabel(Step{}); got != "" {
		t.Fatalf("empty label = %q, want empty", got)
	}
}
