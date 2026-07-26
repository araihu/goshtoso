package selectfield

import (
	"html"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
)

// --- Pure helper coverage ---

func TestHelperTextClasses_AllStates(t *testing.T) {
	assert.Equal(t, "pl-0.5 text-danger", helperTextClasses(StateError))
	assert.Equal(t, "pl-0.5 text-success", helperTextClasses(StateSuccess))
	assert.Equal(t, "pl-0.5 text-on-surface-muted dark:text-on-surface-dark-muted", helperTextClasses(StateDefault))
}

func TestToOptions_MapsAndMarksSelected(t *testing.T) {
	type region struct{ Code, Name string }
	regions := []region{
		{Code: "us-east-1", Name: "US East"},
		{Code: "eu-west-1", Name: "EU West"},
	}

	opts := toOptions(regions,
		func(r region) string { return r.Code },
		func(r region) string { return r.Name },
		"eu-west-1",
	)

	assert.Len(t, opts, 2)
	assert.Equal(t, Option{Value: "us-east-1", Label: "US East", Selected: false}, opts[0])
	assert.Equal(t, Option{Value: "eu-west-1", Label: "EU West", Selected: true}, opts[1])
}

func TestToOptions_EmptyInput(t *testing.T) {
	opts := toOptions(nil,
		func(s string) string { return s },
		func(s string) string { return s },
		"",
	)
	assert.Empty(t, opts)
}

func TestSelectClasses_AllStates(t *testing.T) {
	errClasses := Config{State: StateError}.selectClasses()
	assert.Contains(t, errClasses, "border-danger")

	okClasses := Config{State: StateSuccess}.selectClasses()
	assert.Contains(t, okClasses, "border-success")

	defClasses := Config{}.selectClasses()
	assert.Contains(t, defClasses, "border-outline dark:border-outline-dark")
	// shared base survives across states
	assert.Contains(t, defClasses, "appearance-none")
}

func TestSelectedValue_ReturnsFirstSelected(t *testing.T) {
	cfg := Config{Options: []Option{
		{Value: "a", Label: "A"},
		{Value: "b", Label: "B", Selected: true},
		{Value: "c", Label: "C", Selected: true},
	}}
	assert.Equal(t, "b", cfg.selectedValue())
}

func TestSelectedValue_NoneSelected(t *testing.T) {
	cfg := Config{Options: []Option{{Value: "a", Label: "A"}}}
	assert.Equal(t, "", cfg.selectedValue())
	assert.Equal(t, "", Config{}.selectedValue())
}

func TestLabelClasses_AllStates(t *testing.T) {
	assert.Equal(t, "flex w-fit gap-1 pl-0.5 text-sm text-danger", Config{State: StateError}.labelClasses())
	assert.Equal(t, "flex w-fit gap-1 pl-0.5 text-sm text-success", Config{State: StateSuccess}.labelClasses())
	assert.Equal(t, "w-fit pl-0.5 text-sm", Config{}.labelClasses())
}

func TestTriggerClasses_StateBranches(t *testing.T) {
	errClasses := Config{State: StateError}.triggerClasses()
	assert.Contains(t, errClasses, "border-danger")
	assert.Contains(t, errClasses, "bg-surface")

	okClasses := Config{State: StateSuccess}.triggerClasses()
	assert.Contains(t, okClasses, "border-success")

	// Readonly is also "effectively disabled" → disabled vocabulary
	roClasses := Config{Readonly: true, State: StateError}.triggerClasses()
	assert.Contains(t, roClasses, "cursor-not-allowed")
	assert.Contains(t, roClasses, "bg-surface-alt")
}

func TestContainerClasses_RootClass(t *testing.T) {
	base := Config{}.containerClasses()
	assert.Contains(t, base, "relative flex w-full flex-col")
	assert.NotContains(t, base, "custom-root")

	withRoot := Config{RootClass: "custom-root"}.containerClasses()
	assert.Contains(t, withRoot, "custom-root")
	assert.True(t, strings.HasSuffix(withRoot, " custom-root"))
}

func TestGetPlaceholder_DefaultAndCustom(t *testing.T) {
	assert.Equal(t, "Please Select", Config{}.getPlaceholder())
	assert.Equal(t, "Pick one", Config{Placeholder: "Pick one"}.getPlaceholder())
}

func TestIsEffectivelyDisabled(t *testing.T) {
	assert.False(t, Config{}.isEffectivelyDisabled())
	assert.True(t, Config{Disabled: true}.isEffectivelyDisabled())
	assert.True(t, Config{Readonly: true}.isEffectivelyDisabled())
}

func TestOptionsToJS_EmptyAndPopulated(t *testing.T) {
	assert.Equal(t, "[]", optionsToJS(nil))
	assert.Equal(t, "[]", optionsToJS([]Option{}))
	assert.Equal(t,
		"[{value:'a',label:'A'},{value:'b',label:'B'}]",
		optionsToJS([]Option{{Value: "a", Label: "A"}, {Value: "b", Label: "B"}}),
	)
}

// --- Render branch coverage ---

func TestRenderSelect_LabelWithErrorIconAndHelperText(t *testing.T) {
	out := renderSelect(t, Config{
		ID:         "country",
		Label:      "Country",
		State:      StateError,
		HelperText: "Required",
		Options:    []Option{{Value: "us", Label: "US"}},
	}, nil)

	assert.Contains(t, out, `<label for="country-trigger"`)
	assert.Contains(t, out, "Country")
	// error icon path fragment
	assert.Contains(t, out, "M5.28 4.22a.75.75")
	assert.Contains(t, out, "Required")
	assert.Contains(t, out, "text-danger")
}

func TestRenderSelect_LabelWithSuccessIcon(t *testing.T) {
	out := renderSelect(t, Config{
		ID:      "country",
		Label:   "Country",
		State:   StateSuccess,
		Options: []Option{{Value: "us", Label: "US"}},
	}, nil)

	// success icon path fragment
	assert.Contains(t, out, "M12.416 3.376a.75.75")
}

func TestRenderSelect_AutocompleteAttribute(t *testing.T) {
	out := renderSelect(t, Config{
		ID:           "country",
		Name:         "country",
		Autocomplete: "country-name",
		Options:      []Option{{Value: "us", Label: "US"}},
	}, nil)

	assert.Contains(t, out, `autocomplete="country-name"`)
}

func TestRenderSelect_AlpineBindDisabled(t *testing.T) {
	out := renderSelect(t, Config{
		ID:      "country",
		Alpine:  &AlpineConfig{BindDisabled: "isLoading"},
		Options: []Option{{Value: "us", Label: "US"}},
	}, nil)

	browserHTML := html.UnescapeString(out)
	assert.Contains(t, browserHTML, `x-bind:disabled="isLoading"`)
}

func TestRenderSelect_AlpineModelInitWatcher(t *testing.T) {
	out := renderSelect(t, Config{
		ID:      "country",
		Alpine:  &AlpineConfig{Model: "form.country"},
		Options: []Option{{Value: "us", Label: "US", Selected: true}},
	}, nil)

	browserHTML := html.UnescapeString(out)
	assert.Contains(t, browserHTML, "this.$watch('selectedOption'")
	assert.Contains(t, browserHTML, "form.country")
}

func TestRenderSelect_ShellModeWithTriggerLabel(t *testing.T) {
	out := renderSelect(t, Config{
		ID:           "role",
		Shell:        true,
		TriggerLabel: "Role",
		ValueExpr:    "roleLabel()",
		Label:        "Access",
	}, nil)

	browserHTML := html.UnescapeString(out)
	assert.Contains(t, browserHTML, "Role")
	assert.Contains(t, browserHTML, `x-text="roleLabel()"`)
	// shell mode omits the hidden form input + option list
	assert.NotContains(t, out, `type="text"`)
	assert.NotContains(t, out, "allOptions")
}

func TestRenderSelect_ShellModeWithTriggerLeading(t *testing.T) {
	out := renderSelect(t, Config{
		ID:             "swatch",
		Shell:          true,
		ValueExpr:      "swatchLabel()",
		TriggerLeading: templ.Raw(`<span data-test-swatch aria-hidden="true"></span>`),
	}, templ.Raw(`<div data-test-body></div>`))

	assert.Contains(t, out, "data-test-swatch")
	assert.Contains(t, out, "data-test-body")
}

func TestRenderSelect_DisabledTrigger(t *testing.T) {
	out := renderSelect(t, Config{
		ID:       "country",
		Disabled: true,
		Options:  []Option{{Value: "us", Label: "US"}},
	}, nil)

	assert.Contains(t, out, "disabled")
	assert.Contains(t, out, "cursor-not-allowed")
}
