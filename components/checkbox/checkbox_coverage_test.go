package checkbox

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderCheckboxGroup(t *testing.T, cfg GroupConfig) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, CheckboxGroup(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestCoverageCheckboxToneClassesAndIcons(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantBorder  string
		wantBg      string
		wantFocus   string
		wantSVGText string
		wantPath    string
	}{
		{
			name:        "primary defaults",
			cfg:         Config{},
			wantBorder:  "checked:border-primary dark:checked:border-primary-dark",
			wantBg:      "checked:before:bg-primary dark:checked:before:bg-primary-dark",
			wantFocus:   "checked:focus:outline-primary dark:checked:focus:outline-primary-dark",
			wantSVGText: "text-on-primary dark:text-on-primary-dark",
			wantPath:    "M4.5 12.75l6 6 9-13.5",
		},
		{
			name:        "secondary xmark",
			cfg:         Config{Tone: ToneSecondary, Icon: IconXmark},
			wantBorder:  "checked:border-secondary dark:checked:border-secondary-dark",
			wantBg:      "checked:before:bg-secondary dark:checked:before:bg-secondary-dark",
			wantFocus:   "checked:focus:outline-secondary dark:checked:focus:outline-secondary-dark",
			wantSVGText: "text-on-secondary dark:text-on-secondary-dark",
			wantPath:    "M6 18L18 6M6 6l12 12",
		},
		{
			name:        "info minus",
			cfg:         Config{Tone: ToneInfo, Icon: IconMinus},
			wantBorder:  "checked:border-info dark:checked:border-info",
			wantBg:      "checked:before:bg-info dark:checked:before:bg-info",
			wantFocus:   "checked:focus:outline-info dark:checked:focus:outline-info",
			wantSVGText: "text-on-info dark:text-on-info-dark",
			wantPath:    "M18 12H6",
		},
		{
			name:        "success plus",
			cfg:         Config{Tone: ToneSuccess, Icon: IconPlus},
			wantBorder:  "checked:border-success dark:checked:border-success",
			wantBg:      "checked:before:bg-success dark:checked:before:bg-success",
			wantFocus:   "checked:focus:outline-success dark:checked:focus:outline-success",
			wantSVGText: "text-on-success dark:text-on-success-dark",
			wantPath:    "M12 4.5v15m7.5-7.5h-15",
		},
		{
			name:        "warning",
			cfg:         Config{Tone: ToneWarning},
			wantBorder:  "checked:border-warning dark:checked:border-warning",
			wantBg:      "checked:before:bg-warning dark:checked:before:bg-warning",
			wantFocus:   "checked:focus:outline-warning dark:checked:focus:outline-warning",
			wantSVGText: "text-on-warning dark:text-on-warning-dark",
			wantPath:    "M4.5 12.75l6 6 9-13.5",
		},
		{
			name:        "danger",
			cfg:         Config{Tone: ToneDanger},
			wantBorder:  "checked:border-danger dark:checked:border-danger",
			wantBg:      "checked:before:bg-danger dark:checked:before:bg-danger",
			wantFocus:   "checked:focus:outline-danger dark:checked:focus:outline-danger",
			wantSVGText: "text-on-danger dark:text-on-danger-dark",
			wantPath:    "M4.5 12.75l6 6 9-13.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantBorder, tt.cfg.checkedBorderClass())
			assert.Equal(t, tt.wantBg, tt.cfg.checkedBgClass())
			assert.Equal(t, tt.wantFocus, tt.cfg.focusCheckedClass())
			assert.Equal(t, tt.wantSVGText, tt.cfg.svgTextClass())
			assert.Equal(t, tt.wantPath, tt.cfg.iconPath())
		})
	}
}

func TestCoverageCheckboxAnimationClasses(t *testing.T) {
	scaleInput := Config{Animation: AnimationScaleUp}.inputClasses()
	assert.Contains(t, scaleInput, "before:scale-0")
	assert.Contains(t, scaleInput, "checked:before:scale-125")

	slideDownInput := Config{Animation: AnimationSlideDown}.inputClasses()
	assert.Contains(t, slideDownInput, "before:-translate-y-4")
	assert.Contains(t, slideDownInput, "checked:before:translate-y-0")

	slideUpSVG := Config{Animation: AnimationSlideUp}.svgClasses()
	assert.Contains(t, slideUpSVG, "-translate-y-1/4")
	assert.Contains(t, slideUpSVG, "peer-checked:-translate-y-1/2")

	scaleSVG := Config{Animation: AnimationScaleUp}.svgClasses()
	assert.Contains(t, scaleSVG, "scale-0")
	assert.Contains(t, scaleSVG, "peer-checked:scale-100")

	slideDownSVG := Config{Animation: AnimationSlideDown}.svgClasses()
	assert.Contains(t, slideDownSVG, "opacity-0")
	assert.Contains(t, slideDownSVG, "peer-checked:opacity-100")
}

func TestCoverageRenderStandardContainerAndHelperBranches(t *testing.T) {
	standard := renderCheckbox(t, Config{
		ID:    "terms",
		Name:  "accepted_terms",
		Value: "yes",
		Label: "Accept terms",
	})
	assert.Contains(t, standard, `for="terms"`)
	assert.Contains(t, standard, `id="terms"`)
	assert.Contains(t, standard, `name="accepted_terms"`)
	assert.Contains(t, standard, `value="yes"`)
	assert.Contains(t, standard, "Accept terms")

	container := renderCheckbox(t, Config{
		ID:        "delivery",
		Label:     "Delivery updates",
		Checked:   true,
		Container: true,
	})
	assert.Contains(t, container, "justify-between")
	assert.Contains(t, container, "bg-surface")
	assert.Contains(t, container, "Delivery updates")
	assert.Contains(t, container, "checked")

	helperDisabled := renderCheckbox(t, Config{
		ID:           "digest",
		Label:        "Weekly digest",
		HelperText:   "Sent every Friday",
		HelperTextID: "digest-help",
		Disabled:     true,
	})
	assert.Contains(t, helperDisabled, `aria-describedby="digest-help"`)
	assert.Contains(t, helperDisabled, "Sent every Friday")
	assert.Contains(t, helperDisabled, "disabled")
}

func TestCoverageRenderCheckboxGroup(t *testing.T) {
	html := renderCheckboxGroup(t, GroupConfig{
		Title: "Notification channels",
		Items: []Config{
			{ID: "email", Name: "channels", Value: "email", Label: "Email", Checked: true},
			{ID: "sms", Name: "channels", Value: "sms", Label: "SMS"},
			{ID: "push", Name: "channels", Value: "push", Label: "Push", Disabled: true},
		},
	})

	assert.Contains(t, html, "Notification channels")
	assert.Contains(t, html, "<ul")
	for _, want := range []string{
		`for="email"`,
		`id="email"`,
		`value="email"`,
		"Email",
		`for="sms"`,
		`value="sms"`,
		"SMS",
		`for="push"`,
		`value="push"`,
		"Push",
		"disabled",
	} {
		assert.Contains(t, html, want)
	}
}
