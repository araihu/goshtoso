package toggle

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderToggle(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Toggle(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render toggle: %v", err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultToggle(t *testing.T) {
	html := renderToggle(t, Config{ID: "t1", Label: "Enable"})

	for _, want := range []string{
		`for="t1"`,
		`id="t1"`,
		`type="checkbox"`,
		`role="switch"`,
		`class="peer sr-only"`,
		`aria-hidden="true"`,
		`Enable`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("default render missing %q in %s", want, html)
		}
	}

	// No Name set -> neither name attribute nor the always-off hidden input.
	if strings.Contains(html, `name=`) {
		t.Fatalf("expected no name attribute when Name unset: %s", html)
	}
	if strings.Contains(html, `type="hidden"`) {
		t.Fatalf("expected no hidden input when Name unset: %s", html)
	}
}

func TestCoverageCheckedAndDisabled(t *testing.T) {
	html := renderToggle(t, Config{ID: "t2", Label: "On", Checked: true, Disabled: true})

	if !strings.Contains(html, "checked") {
		t.Fatalf("expected checked attribute: %s", html)
	}
	if !strings.Contains(html, "disabled") {
		t.Fatalf("expected disabled attribute: %s", html)
	}
}

func TestCoverageNameOnlyEmitsHiddenInput(t *testing.T) {
	html := renderToggle(t, Config{ID: "t3", Label: "Notify", Name: "notify"})

	if !strings.Contains(html, `name="notify"`) {
		t.Fatalf("expected name attribute on checkbox: %s", html)
	}
	// Name set without Value: the always-off hidden input must be present.
	if !strings.Contains(html, `<input type="hidden" name="notify" value="off"`) {
		t.Fatalf("expected always-off hidden input when Name set and Value empty: %s", html)
	}
}

func TestCoverageNameAndValueOmitsHiddenInput(t *testing.T) {
	html := renderToggle(t, Config{ID: "t4", Label: "Accept", Name: "accept", Value: "yes"})

	if !strings.Contains(html, `name="accept"`) {
		t.Fatalf("expected name attribute: %s", html)
	}
	if !strings.Contains(html, `value="yes"`) {
		t.Fatalf("expected value attribute: %s", html)
	}
	// With Name + Value the hidden always-off input is omitted.
	if strings.Contains(html, `type="hidden"`) {
		t.Fatalf("expected no hidden input when Name and Value both set: %s", html)
	}
}

func TestCoverageInputAttrsPassthrough(t *testing.T) {
	html := renderToggle(t, Config{
		ID:         "t5",
		Label:      "Bound",
		InputAttrs: templ.Attributes{"x-on:change": "save()", "data-test": "toggle5"},
	})

	if !strings.Contains(html, `x-on:change="save()"`) {
		t.Fatalf("expected x-on:change passthrough: %s", html)
	}
	if !strings.Contains(html, `data-test="toggle5"`) {
		t.Fatalf("expected data-test passthrough: %s", html)
	}
}

func TestCoverageCheckedClassesAllVariants(t *testing.T) {
	cases := []struct {
		variant Variant
		want    string
	}{
		{Primary, "peer-checked:bg-primary"},
		{Secondary, "peer-checked:bg-secondary"},
		{Info, "peer-checked:bg-info"},
		{Success, "peer-checked:bg-success"},
		{Warning, "peer-checked:bg-warning"},
		{Danger, "peer-checked:bg-danger"},
		{Variant("unknown"), "peer-checked:bg-primary"}, // falls back to Primary
	}

	for _, c := range cases {
		got := Config{Variant: c.variant}.checkedClasses()
		if !strings.Contains(got, c.want) {
			t.Errorf("variant %q checkedClasses missing %q: %s", c.variant, c.want, got)
		}
	}
}

func TestCoverageToggleClassesStyleBranches(t *testing.T) {
	defaultClasses := Config{Style: StyleDefault}.ToggleClasses()
	if !strings.Contains(defaultClasses, "bg-surface-alt") {
		t.Errorf("default style should use bg-surface-alt: %s", defaultClasses)
	}

	containerClasses := Config{Style: StyleContainer}.ToggleClasses()
	if strings.Contains(containerClasses, "bg-surface-alt") {
		t.Errorf("container style should not use bg-surface-alt: %s", containerClasses)
	}
	if !strings.Contains(containerClasses, "bg-surface ") {
		t.Errorf("container style should use bg-surface: %s", containerClasses)
	}

	// ToggleClasses always appends variant (checked) classes.
	if !strings.Contains(defaultClasses, "peer-checked:bg-primary") {
		t.Errorf("ToggleClasses should append checked classes: %s", defaultClasses)
	}
}

func TestCoverageLabelClassesBranches(t *testing.T) {
	def := Config{Style: StyleDefault}.LabelClasses()
	if def != "inline-flex items-center gap-3" {
		t.Errorf("unexpected default LabelClasses: %s", def)
	}

	container := Config{Style: StyleContainer}.LabelClasses()
	if !strings.Contains(container, "min-w-52") || !strings.Contains(container, "rounded-radius") {
		t.Errorf("container LabelClasses missing expected styling: %s", container)
	}

	withRoot := Config{Style: StyleDefault, RootClass: "mb-4"}.LabelClasses()
	if !strings.HasSuffix(withRoot, " mb-4") {
		t.Errorf("RootClass should be appended: %s", withRoot)
	}

	containerWithRoot := Config{Style: StyleContainer, RootClass: "shadow"}.LabelClasses()
	if !strings.HasSuffix(containerWithRoot, " shadow") {
		t.Errorf("RootClass should append to container LabelClasses: %s", containerWithRoot)
	}
}

func TestCoverageVariantClassesRendered(t *testing.T) {
	html := renderToggle(t, Config{ID: "t6", Label: "Warn", Variant: Warning, Style: StyleContainer})

	if !strings.Contains(html, "peer-checked:bg-warning") {
		t.Fatalf("expected warning variant class in track: %s", html)
	}
	if !strings.Contains(html, "min-w-52") {
		t.Fatalf("expected container label styling: %s", html)
	}
}
