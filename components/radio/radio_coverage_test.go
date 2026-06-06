package radio

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// render is a small helper that renders any templ.Component to a string.
func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, c.Render(context.Background(), &buf))
	return buf.String()
}

// TestCoverageStandardRadio exercises the default branch of Radio: no helper
// text, no container, no segmented => standardRadio + radioInput.
func TestCoverageStandardRadio(t *testing.T) {
	html := render(t, Radio(Config{ID: "std", Name: "grp", Value: "v", Label: "Standard"}))

	assert.Contains(t, html, `type="radio"`)
	assert.Contains(t, html, `id="std"`)
	assert.Contains(t, html, `name="grp"`)
	assert.Contains(t, html, `value="v"`)
	assert.Contains(t, html, "Standard")
	// Default variant -> primary classes wired through InputClasses.
	assert.Contains(t, html, "checked:border-primary")
	assert.Contains(t, html, "size-4") // default SizeMD
}

// TestCoverageContainerRadio hits radioWithContainer and its bordered look.
func TestCoverageContainerRadio(t *testing.T) {
	html := render(t, Radio(Config{ID: "ct", Label: "Boxed", Container: true}))

	assert.Contains(t, html, "rounded-radius border border-outline")
	assert.Contains(t, html, "Boxed")
	// Container swaps the input background.
	assert.Contains(t, html, "bg-surface dark:bg-surface-dark")
}

// TestCoverageHelperTextWithID covers radioWithDescription including the
// aria-describedby + HelperTextID branch.
func TestCoverageHelperTextWithID(t *testing.T) {
	html := render(t, Radio(Config{
		ID:           "ht",
		Label:        "With help",
		HelperText:   "extra context",
		HelperTextID: "ht-desc",
	}))

	assert.Contains(t, html, "extra context")
	assert.Contains(t, html, `id="ht-desc"`)
	assert.Contains(t, html, `aria-describedby="ht-desc"`)
}

// TestCoverageHelperTextNoID covers radioWithDescription's else branch where
// the helper text has no id (no aria-describedby on the input).
func TestCoverageHelperTextNoID(t *testing.T) {
	html := render(t, Radio(Config{ID: "ht2", Label: "L", HelperText: "ctx only"}))

	assert.Contains(t, html, "ctx only")
	assert.NotContains(t, html, "aria-describedby")
}

// TestCoverageSegmented hits radioSegmented + segmentedInput (sr-only input).
func TestCoverageSegmented(t *testing.T) {
	html := render(t, Radio(Config{ID: "seg", Label: "Seg", Value: "seg", Segmented: true}))

	assert.Contains(t, html, "sr-only")
	assert.Contains(t, html, "Seg")
	// Segmented label styling rides has-checked:.
	assert.Contains(t, html, "has-checked:bg-primary")
}

// TestCoverageRadioBar renders the segmented bar wrapper.
func TestCoverageRadioBar(t *testing.T) {
	html := render(t, RadioBar())

	assert.Contains(t, html, "divide-x divide-outline")
	assert.Contains(t, html, "rounded-radius border border-outline")
}

// TestCoverageCheckedAndDisabled covers the checked + disabled attribute
// branches in radioInput.
func TestCoverageCheckedAndDisabled(t *testing.T) {
	html := render(t, Radio(Config{ID: "cd", Label: "L", Checked: true, Disabled: true}))

	assert.Contains(t, html, "checked")
	assert.Contains(t, html, "disabled")
}

// TestCoverageHTMXAttributes drives every HTMX branch in radioInput including
// the default hx-trigger="change".
func TestCoverageHTMXAttributes(t *testing.T) {
	html := render(t, Radio(Config{
		ID:    "hx",
		Label: "HX",
		HTMX: &HTMXConfig{
			Get:       "/get",
			Post:      "/post",
			Put:       "/put",
			Delete:    "/delete",
			Patch:     "/patch",
			Target:    "#tgt",
			Swap:      "outerHTML",
			Indicator: "#spin",
			PushURL:   true,
			Confirm:   "Sure?",
			Vals:      `{"a":1}`,
			Include:   "#form",
		},
	}))

	for _, want := range []string{
		`hx-get="/get"`, `hx-post="/post"`, `hx-put="/put"`,
		`hx-delete="/delete"`, `hx-patch="/patch"`, `hx-target="#tgt"`,
		`hx-swap="outerHTML"`, `hx-indicator="#spin"`, `hx-push-url="true"`,
		`hx-confirm="Sure?"`, `hx-include="#form"`,
		`hx-trigger="change"`, // defaulted because a verb is set, no Trigger given
	} {
		assert.Contains(t, html, want, "missing %s", want)
	}
	// hx-vals quotes are escaped by templ; assert on the attribute name.
	assert.Contains(t, html, "hx-vals=")
	// Regression: the trigger attribute must not leak a stray `else` text node.
	assert.NotContains(t, html, " else", "hx-trigger else-if must not leak a literal else")
}

// TestCoverageHTMXExplicitTrigger covers the explicit Trigger branch (not the
// defaulted "change"). Regression guard for the templ else-if leak bug.
func TestCoverageHTMXExplicitTrigger(t *testing.T) {
	html := render(t, Radio(Config{
		ID:    "hxt",
		Label: "L",
		HTMX:  &HTMXConfig{Post: "/p", Trigger: "click"},
	}))

	assert.Contains(t, html, `hx-trigger="click"`)
	assert.NotContains(t, html, `hx-trigger="change"`)
	assert.NotContains(t, html, " else")
}

// TestCoverageAlpineAttributes drives every Alpine branch in radioInput.
func TestCoverageAlpineAttributes(t *testing.T) {
	html := render(t, Radio(Config{
		ID:    "al",
		Label: "L",
		Alpine: &AlpineConfig{
			Data:         "{open:false}",
			Model:        "selected",
			OnChange:     "doThing()",
			BindChecked:  "isOn",
			BindDisabled: "isLocked",
		},
	}))

	for _, want := range []string{"x-data", "x-model", "x-on:change", "x-bind:checked", "x-bind:disabled"} {
		assert.Contains(t, html, want, "missing %s", want)
	}
}

// TestCoverageSegmentedHTMXAndAlpine drives the HTMX + Alpine branches in
// segmentedInput (separate template from radioInput).
func TestCoverageSegmentedHTMXAndAlpine(t *testing.T) {
	html := render(t, Radio(Config{
		ID:        "seghx",
		Label:     "L",
		Segmented: true,
		HTMX:      &HTMXConfig{Post: "/p"},
		Alpine:    &AlpineConfig{Model: "v"},
	}))

	assert.Contains(t, html, `hx-post="/p"`)
	assert.Contains(t, html, `hx-trigger="change"`)
	assert.Contains(t, html, "x-model")
	assert.NotContains(t, html, " else")
}

// TestCoverageInputAttrsEscapeHatch covers the trailing cfg.InputAttrs spread.
func TestCoverageInputAttrsEscapeHatch(t *testing.T) {
	html := render(t, Radio(Config{
		ID:         "ia",
		Label:      "L",
		InputAttrs: templ.Attributes{"data-extra": "yes"},
	}))

	assert.Contains(t, html, `data-extra="yes"`)
}

// TestCoverageRadioGroup hits RadioGroup, groupRadioItem,
// groupRadioItemWithDescription and the BadgeColor branch.
func TestCoverageRadioGroup(t *testing.T) {
	html := render(t, RadioGroup(GroupConfig{
		Title: "Pick one",
		Items: []Config{
			{ID: "g1", Name: "g", Value: "1", Label: "Plain", Checked: true},
			{ID: "g2", Name: "g", Value: "2", Label: "Described", HelperText: "more"},
			{ID: "g3", Name: "g", Value: "3", Label: "Badged", HelperText: "h", BadgeColor: "success"},
		},
	}))

	assert.Contains(t, html, "Pick one")
	assert.Contains(t, html, "Plain")
	assert.Contains(t, html, "Described")
	assert.Contains(t, html, "more")
	assert.Contains(t, html, "Badged")
	// BadgeColor success -> BadgeClasses success palette.
	assert.Contains(t, html, "bg-success/10 text-success")
}

// TestCoverageRadioGroupNoTitle covers the empty-title branch of RadioGroup.
func TestCoverageRadioGroupNoTitle(t *testing.T) {
	html := render(t, RadioGroup(GroupConfig{
		Items: []Config{{ID: "n1", Label: "Only"}},
	}))

	assert.NotContains(t, html, "<h3")
	assert.Contains(t, html, "Only")
}

// TestCoverageVariantClasses asserts each Variant flows through the three
// checked-color helpers via InputClasses.
func TestCoverageVariantClasses(t *testing.T) {
	cases := []struct {
		variant Variant
		border  string
	}{
		{Primary, "checked:border-primary"},
		{Secondary, "checked:border-secondary"},
		{Info, "checked:border-info"},
		{Success, "checked:border-success"},
		{Warning, "checked:border-warning"},
		{Danger, "checked:border-danger"},
		{Variant("bogus"), "checked:border-primary"}, // default fallthrough
	}
	for _, tc := range cases {
		cls := Config{Variant: tc.variant}.InputClasses()
		assert.Contains(t, cls, tc.border, "variant %q border", tc.variant)
	}
}

// TestCoverageSizeClasses asserts every Size maps to a box size.
func TestCoverageSizeClasses(t *testing.T) {
	cases := map[Size]string{
		SizeSM:        "size-3",
		SizeMD:        "size-4",
		SizeLG:        "size-5",
		SizeXL:        "size-6",
		Size("bogus"): "size-4", // default
	}
	for size, want := range cases {
		assert.Contains(t, Config{Size: size}.InputClasses(), want, "size %q", size)
	}
}

// TestCoverageBadgeClasses covers every color branch including the empty
// default.
func TestCoverageBadgeClasses(t *testing.T) {
	cases := map[string]string{
		"success":   "bg-success/10 text-success",
		"danger":    "bg-danger/10 text-danger",
		"warning":   "bg-warning/10 text-warning",
		"info":      "bg-info/10 text-info",
		"primary":   "bg-primary/10",
		"secondary": "bg-secondary/10 text-secondary",
		"neutral":   "bg-on-surface/10",
	}
	for color, want := range cases {
		got := BadgeClasses(color)
		assert.Contains(t, got, want, "color %q", color)
		assert.True(t, strings.HasPrefix(got, "w-fit rounded-radius"), "color %q base prefix", color)
	}
	assert.Equal(t, "", BadgeClasses("unknown"), "unknown color yields empty string")
}

// TestCoverageSegmentedLabelClasses covers segmentedCheckedClasses for every
// variant.
func TestCoverageSegmentedLabelClasses(t *testing.T) {
	cases := []struct {
		variant Variant
		want    string
	}{
		{Primary, "has-checked:bg-primary"},
		{Secondary, "has-checked:bg-secondary"},
		{Info, "has-checked:bg-info"},
		{Success, "has-checked:bg-success"},
		{Warning, "has-checked:bg-warning"},
		{Danger, "has-checked:bg-danger"},
		{Variant("bogus"), "has-checked:bg-primary"},
	}
	for _, tc := range cases {
		cls := Config{Variant: tc.variant}.SegmentedLabelClasses()
		assert.Contains(t, cls, tc.want, "variant %q", tc.variant)
		assert.Contains(t, cls, "has-disabled:cursor-not-allowed")
	}
}

// TestCoveragePredicates covers HasAlpine / HasHTMX / HasHxVerb and
// EffectiveTrigger including the nil receiver paths.
func TestCoveragePredicates(t *testing.T) {
	assert.False(t, Config{}.HasAlpine())
	assert.False(t, Config{}.HasHTMX())
	assert.True(t, Config{Alpine: &AlpineConfig{}}.HasAlpine())
	assert.True(t, Config{HTMX: &HTMXConfig{}}.HasHTMX())

	var nilHTMX *HTMXConfig
	assert.False(t, nilHTMX.HasHxVerb(), "nil HTMXConfig has no verb")
	assert.Equal(t, "", nilHTMX.EffectiveTrigger(), "nil HTMXConfig has no trigger")
	assert.False(t, (&HTMXConfig{}).HasHxVerb())
	assert.True(t, (&HTMXConfig{Get: "/x"}).HasHxVerb())
	assert.True(t, (&HTMXConfig{Patch: "/x"}).HasHxVerb())

	// EffectiveTrigger precedence: explicit > defaulted > empty.
	assert.Equal(t, "click", (&HTMXConfig{Post: "/p", Trigger: "click"}).EffectiveTrigger())
	assert.Equal(t, "change", (&HTMXConfig{Post: "/p"}).EffectiveTrigger())
	assert.Equal(t, "", (&HTMXConfig{}).EffectiveTrigger())
}
