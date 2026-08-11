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

// inputTag returns just the opening <input> tag (the checkbox) so attribute
// assertions are not confused by peer-checked / peer-disabled utility classes
// that appear elsewhere in the rendered markup.
func inputTag(t *testing.T, html string) string {
	t.Helper()
	start := strings.Index(html, "<input")
	if start < 0 {
		t.Fatalf("no <input> in render: %s", html)
	}
	end := strings.Index(html[start:], ">")
	if end < 0 {
		t.Fatalf("unterminated <input> in render: %s", html)
	}
	return html[start : start+end+1]
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
	// Assert on the <input> tag only: the track div carries peer-checked /
	// peer-disabled utility classes, so a whole-document substring search would
	// pass even when the boolean attributes are absent.
	on := inputTag(t, renderToggle(t, Config{ID: "t2", Label: "On", Checked: true, Disabled: true}))
	if !strings.Contains(on, "checked") {
		t.Fatalf("expected checked attribute on input: %s", on)
	}
	if !strings.Contains(on, "disabled") {
		t.Fatalf("expected disabled attribute on input: %s", on)
	}

	// Guard against a false positive: an unconfigured toggle must NOT render the
	// boolean attributes on its input.
	off := inputTag(t, renderToggle(t, Config{ID: "t2", Label: "Off"}))
	if strings.Contains(off, "checked") {
		t.Fatalf("unexpected checked attribute on default input: %s", off)
	}
	if strings.Contains(off, "disabled") {
		t.Fatalf("unexpected disabled attribute on default input: %s", off)
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
		variant Tone
		want    string
	}{
		{TonePrimary, "peer-checked:bg-primary"},
		{ToneSecondary, "peer-checked:bg-secondary"},
		{ToneInfo, "peer-checked:bg-info"},
		{ToneSuccess, "peer-checked:bg-success"},
		{ToneWarning, "peer-checked:bg-warning"},
		{ToneDanger, "peer-checked:bg-danger"},
		{Tone("unknown"), "peer-checked:bg-primary"}, // falls back to TonePrimary
	}

	for _, c := range cases {
		got := Config{Tone: c.variant}.checkedClasses()
		if !strings.Contains(got, c.want) {
			t.Errorf("variant %q checkedClasses missing %q: %s", c.variant, c.want, got)
		}
	}
}

func TestCoverageToggleClassesStyleBranches(t *testing.T) {
	defaultClasses := Config{Appearance: AppearanceDefault}.toggleClasses()
	if !strings.Contains(defaultClasses, "border-control-outline") || !strings.Contains(defaultClasses, "dark:border-control-outline-dark") {
		t.Errorf("default track should use control boundary roles: %s", defaultClasses)
	}
	if !strings.Contains(defaultClasses, "bg-surface-alt") {
		t.Errorf("default style should use bg-surface-alt: %s", defaultClasses)
	}

	containerClasses := Config{Appearance: AppearanceContainer}.toggleClasses()
	if strings.Contains(containerClasses, "bg-surface-alt") {
		t.Errorf("container style should not use bg-surface-alt: %s", containerClasses)
	}
	if !strings.Contains(containerClasses, "bg-surface ") {
		t.Errorf("container style should use bg-surface: %s", containerClasses)
	}

	// toggleClasses always appends variant (checked) classes.
	if !strings.Contains(defaultClasses, "peer-checked:bg-primary") {
		t.Errorf("toggleClasses should append checked classes: %s", defaultClasses)
	}
}

func TestDisabledToggleDimsLabelWithoutFadingTrackBoundary(t *testing.T) {
	track := Config{Disabled: true}.toggleClasses()
	if strings.Contains(track, "peer-disabled:opacity-70") {
		t.Fatalf("disabled opacity must not fade the track boundary: %s", track)
	}
	html := renderToggle(t, Config{ID: "disabled-boundary", Label: "Disabled", Disabled: true})
	if !strings.Contains(html, "peer-disabled:opacity-70") {
		t.Fatalf("disabled label must retain content-level opacity semantics: %s", html)
	}
}

func TestCoverageLabelClassesBranches(t *testing.T) {
	def := Config{Appearance: AppearanceDefault}.labelClasses()
	if def != "inline-flex items-center gap-3" {
		t.Errorf("unexpected default labelClasses: %s", def)
	}

	container := Config{Appearance: AppearanceContainer}.labelClasses()
	if !strings.Contains(container, "min-w-52") || !strings.Contains(container, "rounded-radius") {
		t.Errorf("container labelClasses missing expected styling: %s", container)
	}

	withRoot := Config{Appearance: AppearanceDefault, RootClass: "mb-4"}.labelClasses()
	if !strings.HasSuffix(withRoot, " mb-4") {
		t.Errorf("RootClass should be appended: %s", withRoot)
	}

	containerWithRoot := Config{Appearance: AppearanceContainer, RootClass: "shadow"}.labelClasses()
	if !strings.HasSuffix(containerWithRoot, " shadow") {
		t.Errorf("RootClass should append to container labelClasses: %s", containerWithRoot)
	}
}

func TestCoverageToneClassesRendered(t *testing.T) {
	html := renderToggle(t, Config{ID: "t6", Label: "Warn", Tone: ToneWarning, Appearance: AppearanceContainer})

	if !strings.Contains(html, "peer-checked:bg-warning") {
		t.Fatalf("expected warning variant class in track: %s", html)
	}
	if !strings.Contains(html, "min-w-52") {
		t.Fatalf("expected container label styling: %s", html)
	}
}
