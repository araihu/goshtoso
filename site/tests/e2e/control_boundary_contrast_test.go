//go:build e2e && (full || controlboundary)

package e2e

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/fileinput"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/palette"
	"github.com/araihu/goshtoso/components/radio"
	"github.com/araihu/goshtoso/components/schemaform"
	"github.com/araihu/goshtoso/components/search"
	selectcomponent "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/components/toggle"
)

const minimumBoundaryContrast = 3.0

type boundaryThemeMode struct {
	name  string
	theme string
	dark  bool
}

type boundaryCase struct {
	name            string
	family          string
	boundary        string
	focusTarget     string
	checkedTarget   string
	uncheckedTarget string
	ariaInvalid     bool
	disabled        bool
	expectedRoles   []string
}

type boundaryMeasurement struct {
	BorderColor       string  `json:"borderColor"`
	BorderWidth       float64 `json:"borderWidth"`
	BorderStyle       string  `json:"borderStyle"`
	AdjacentColor     string  `json:"adjacentColor"`
	RenderedColor     string  `json:"renderedColor"`
	Contrast          float64 `json:"contrast"`
	CumulativeOpacity float64 `json:"cumulativeOpacity"`
	OutlineColor      string  `json:"outlineColor"`
	OutlineWidth      string  `json:"outlineWidth"`
}

type boundaryFailure struct {
	caseName string
	family   string
	mode     string
	reason   string
	measure  boundaryMeasurement
}

func TestControlBoundaryContrastAcrossFoundationThemes(t *testing.T) {
	cases := controlBoundaryCases()
	require.Len(t, cases, 52, "the contract requires exactly 52 boundary states")

	server := httptest.NewServer(controlBoundaryHandler(t))
	t.Cleanup(server.Close)

	page := newPage(t, sharedBrowser)
	defer page.Close()

	modes := []boundaryThemeMode{
		{name: "araihu-light", theme: "araihu"},
		{name: "araihu-dark", theme: "araihu", dark: true},
		{name: "modern-light", theme: "modern"},
		{name: "modern-dark", theme: "modern", dark: true},
		{name: "goshtoso-light", theme: "goshtoso"},
		{name: "goshtoso-dark", theme: "goshtoso", dark: true},
	}

	failures := make([]boundaryFailure, 0)
	minimumByFamily := map[string]float64{}
	minimumByMode := map[string]float64{}
	overallMinimum := math.Inf(1)
	cells := 0

	for _, mode := range modes {
		url := fmt.Sprintf("%s/?theme=%s&dark=%t", server.URL, mode.theme, mode.dark)
		_, err := page.Goto(url)
		require.NoError(t, err)
		require.NoError(t, page.Locator("[data-boundary-ready]").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateAttached}))
		_, err = page.AddStyleTag(playwright.PageAddStyleTagOptions{Content: playwright.String("*,*::before,*::after{animation:none!important;transition:none!important}")})
		require.NoError(t, err)

		for _, tc := range cases {
			cells++
			boundary := page.Locator(tc.boundary)
			count, countErr := boundary.Count()
			require.NoError(t, countErr)
			require.Equalf(t, 1, count, "%s/%s boundary selector", mode.name, tc.name)

			if tc.ariaInvalid {
				_, err = boundary.Evaluate("element => element.setAttribute('aria-invalid', 'true')", nil)
				require.NoError(t, err)
			}
			if tc.uncheckedTarget != "" {
				target := page.Locator(tc.uncheckedTarget)
				count, countErr = target.Count()
				require.NoError(t, countErr)
				require.Equalf(t, 1, count, "%s/%s unchecked selector", mode.name, tc.name)
				_, err = target.Evaluate("element => { element.checked = false; element.dispatchEvent(new Event('input', { bubbles: true })); element.dispatchEvent(new Event('change', { bubbles: true })); }", nil)
				require.NoError(t, err)
			}
			if tc.checkedTarget != "" {
				target := page.Locator(tc.checkedTarget)
				count, countErr = target.Count()
				require.NoError(t, countErr)
				require.Equalf(t, 1, count, "%s/%s checked selector", mode.name, tc.name)
				_, err = target.Evaluate("element => { element.checked = true; element.dispatchEvent(new Event('input', { bubbles: true })); element.dispatchEvent(new Event('change', { bubbles: true })); }", nil)
				require.NoError(t, err)
			}
			if tc.focusTarget != "" {
				target := page.Locator(tc.focusTarget)
				count, countErr = target.Count()
				require.NoError(t, countErr)
				require.Equalf(t, 1, count, "%s/%s focus selector", mode.name, tc.name)
				require.NoError(t, target.Focus())
			}

			classes, err := boundary.GetAttribute("class")
			require.NoError(t, err)
			for _, role := range tc.expectedRoles {
				if !strings.Contains(classes, role) {
					failures = append(failures, boundaryFailure{caseName: tc.name, family: tc.family, mode: mode.name, reason: "missing role class " + role})
				}
			}

			measurement := measureBoundary(t, boundary)
			if prior, ok := minimumByFamily[tc.family]; !ok || measurement.Contrast < prior {
				minimumByFamily[tc.family] = measurement.Contrast
			}
			if prior, ok := minimumByMode[mode.name]; !ok || measurement.Contrast < prior {
				minimumByMode[mode.name] = measurement.Contrast
			}
			if measurement.Contrast < overallMinimum {
				overallMinimum = measurement.Contrast
			}

			if measurement.BorderWidth <= 0 || measurement.BorderStyle == "none" {
				failures = append(failures, boundaryFailure{caseName: tc.name, family: tc.family, mode: mode.name, reason: "no rendered boundary", measure: measurement})
			} else if measurement.Contrast+1e-9 < minimumBoundaryContrast {
				failures = append(failures, boundaryFailure{caseName: tc.name, family: tc.family, mode: mode.name, reason: fmt.Sprintf("contrast %.3f < %.1f", measurement.Contrast, minimumBoundaryContrast), measure: measurement})
			}
		}
		assertSchemaBooleanStateAppearance(t, page, mode.name)
	}

	require.Equal(t, 312, cells, "52 cases x 6 theme/modes")
	families := make([]string, 0, len(minimumByFamily))
	for family := range minimumByFamily {
		families = append(families, family)
	}
	sort.Strings(families)
	for _, family := range families {
		t.Logf("minimum family=%s contrast=%.3f", family, minimumByFamily[family])
	}
	failuresByMode := map[string]int{}
	failuresByFamily := map[string]int{}
	for _, failure := range failures {
		failuresByMode[failure.mode]++
		failuresByFamily[failure.family]++
	}
	for _, mode := range modes {
		t.Logf("minimum mode=%s contrast=%.3f failures=%d", mode.name, minimumByMode[mode.name], failuresByMode[mode.name])
	}
	for _, family := range families {
		t.Logf("failures family=%s count=%d", family, failuresByFamily[family])
	}
	t.Logf("minimum overall contrast=%.3f cells=%d failures=%d", overallMinimum, cells, len(failures))
	for _, failure := range failures {
		t.Errorf("boundary failure case=%s mode=%s reason=%s border=%s rendered=%s adjacent=%s opacity=%.3f borderWidth=%.2f borderStyle=%s outline=%s/%s",
			failure.caseName, failure.mode, failure.reason, failure.measure.BorderColor,
			failure.measure.RenderedColor, failure.measure.AdjacentColor, failure.measure.CumulativeOpacity,
			failure.measure.BorderWidth, failure.measure.BorderStyle, failure.measure.OutlineColor, failure.measure.OutlineWidth)
	}
}

func assertSchemaBooleanStateAppearance(t *testing.T, page playwright.Page, mode string) {
	t.Helper()
	checked := page.Locator("[id='values.schema-boolean']")
	unchecked := page.Locator("[id='values.schema-boolean-focus']")
	disabled := page.Locator("[id='values.schema-boolean-disabled']")

	isChecked, err := checked.IsChecked()
	require.NoError(t, err)
	require.Truef(t, isChecked, "%s checked boolean state", mode)
	isUnchecked, err := unchecked.IsChecked()
	require.NoError(t, err)
	require.Falsef(t, isUnchecked, "%s unchecked boolean state", mode)
	isDisabled, err := disabled.IsDisabled()
	require.NoError(t, err)
	require.Truef(t, isDisabled, "%s managed boolean must remain disabled", mode)

	for _, id := range []string{"values.schema-boolean", "values.schema-boolean-focus", "values.schema-boolean-disabled"} {
		input := page.Locator(fmt.Sprintf("[id='%s']", id))
		inputType, attrErr := input.GetAttribute("type")
		require.NoError(t, attrErr)
		require.Equalf(t, "checkbox", inputType, "%s/%s native checkbox semantics", mode, id)
		name, attrErr := input.GetAttribute("name")
		require.NoError(t, attrErr)
		require.Equalf(t, id, name, "%s/%s stable name hook", mode, id)
		labelCount, countErr := page.Locator(fmt.Sprintf("label[for='%s']", id)).Count()
		require.NoError(t, countErr)
		require.Equalf(t, 1, labelCount, "%s/%s label association", mode, id)
	}

	activeID, err := page.Evaluate("() => document.activeElement && document.activeElement.id", nil)
	require.NoError(t, err)
	require.Equalf(t, "values.schema-boolean-focus", activeID, "%s boolean focus hook", mode)

	value, err := page.Evaluate(`() => {
		const canvas = document.createElement('canvas');
		canvas.width = 1;
		canvas.height = 1;
		const context = canvas.getContext('2d', { willReadFrequently: true });
		const parse = value => {
			context.clearRect(0, 0, 1, 1);
			context.fillStyle = 'rgba(0, 0, 0, 0)';
			context.fillStyle = String(value);
			context.fillRect(0, 0, 1, 1);
			const pixel = context.getImageData(0, 0, 1, 1).data;
			return [pixel[0], pixel[1], pixel[2]];
		};
		const linear = value => {
			value /= 255;
			return value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
		};
		const luminance = color => 0.2126 * linear(color[0]) + 0.7152 * linear(color[1]) + 0.0722 * linear(color[2]);
		const checked = getComputedStyle(document.getElementById('values.schema-boolean')).backgroundColor;
		const unchecked = getComputedStyle(document.getElementById('values.schema-boolean-focus')).backgroundColor;
		const checkedLuminance = luminance(parse(checked));
		const uncheckedLuminance = luminance(parse(unchecked));
		return {
			checked,
			unchecked,
			contrast: (Math.max(checkedLuminance, uncheckedLuminance) + 0.05) / (Math.min(checkedLuminance, uncheckedLuminance) + 0.05),
		};
	}`, nil)
	require.NoError(t, err)
	measurement, ok := value.(map[string]any)
	require.True(t, ok, "unexpected schema boolean state measurement %T", value)
	ratio := browserNumber(measurement["contrast"])
	require.GreaterOrEqualf(t, ratio+1e-9, minimumBoundaryContrast,
		"%s checked fill must be perceptibly distinct from unchecked fill: checked=%s unchecked=%s contrast=%.3f",
		mode, measurement["checked"], measurement["unchecked"], ratio)
}

func controlBoundaryCases() []boundaryCase {
	roles := []string{"border-control-outline", "dark:border-control-outline-dark"}
	return []boundaryCase{
		{name: "checkbox/default", family: "checkbox", boundary: "#cb-default", expectedRoles: roles},
		{name: "checkbox/disabled", family: "checkbox", boundary: "#cb-disabled", disabled: true, expectedRoles: roles},
		{name: "checkbox/checked", family: "checkbox", boundary: "#cb-checked", checkedTarget: "#cb-checked", expectedRoles: appendRoles(roles, "checked:border-primary", "dark:checked:border-primary-dark")},
		{name: "checkbox/focus-visible", family: "checkbox", boundary: "#cb-focus", focusTarget: "#cb-focus", expectedRoles: roles},

		{name: "radio/default", family: "radio", boundary: "#radio-default", expectedRoles: roles},
		{name: "radio/disabled", family: "radio", boundary: "#radio-disabled", disabled: true, expectedRoles: roles},
		{name: "radio/checked", family: "radio", boundary: "#radio-checked", checkedTarget: "#radio-checked", expectedRoles: appendRoles(roles, "checked:border-primary", "dark:checked:border-primary-dark")},
		{name: "radio/focus-visible", family: "radio", boundary: "#radio-focus", focusTarget: "#radio-focus", expectedRoles: roles},

		{name: "toggle/default", family: "toggle", boundary: "label[for='toggle-default'] > div", expectedRoles: roles},
		{name: "toggle/disabled", family: "toggle", boundary: "label[for='toggle-disabled'] > div", disabled: true, expectedRoles: roles},
		{name: "toggle/checked", family: "toggle", boundary: "label[for='toggle-checked'] > div", checkedTarget: "#toggle-checked", expectedRoles: roles},
		{name: "toggle/focus-visible", family: "toggle", boundary: "label[for='toggle-focus'] > div", focusTarget: "#toggle-focus", expectedRoles: roles},

		{name: "file-upload/default", family: "file-input", boundary: "label[for='file-default'][data-fileinput-variant='upload']", expectedRoles: roles},
		{name: "file-upload/disabled", family: "file-input", boundary: "label[for='file-disabled'][data-fileinput-variant='upload']", disabled: true, expectedRoles: roles},
		{name: "file-upload/focus-visible", family: "file-input", boundary: "label[for='file-focus'][data-fileinput-variant='upload']", focusTarget: "#file-focus", expectedRoles: roles},
		{name: "file-dropzone/default", family: "file-input", boundary: ".drop-default [data-fileinput-variant='dropzone']", expectedRoles: roles},
		{name: "file-dropzone/disabled", family: "file-input", boundary: ".drop-disabled [data-fileinput-variant='dropzone']", disabled: true, expectedRoles: roles},

		{name: "select/default", family: "select", boundary: "#select-default-trigger", expectedRoles: roles},
		{name: "select/disabled", family: "select", boundary: "#select-disabled-trigger", disabled: true, expectedRoles: roles},
		{name: "select/invalid", family: "select", boundary: "#select-invalid-trigger", ariaInvalid: true, expectedRoles: []string{"border-danger"}},
		{name: "select/focus-visible", family: "select", boundary: "#select-focus-trigger", focusTarget: "#select-focus-trigger", expectedRoles: roles},

		{name: "structured/text-default", family: "structured-input", boundary: "#structured-default input[type='text']", expectedRoles: roles},
		{name: "structured/text-disabled", family: "structured-input", boundary: "#structured-disabled input[type='text']", disabled: true, expectedRoles: roles},
		{name: "structured/text-focus-visible", family: "structured-input", boundary: "#structured-focus input[type='text']", focusTarget: "#structured-focus input[type='text']", expectedRoles: roles},
		{name: "structured/select-default", family: "structured-input", boundary: "#structured-default select", expectedRoles: roles},
		{name: "structured/select-disabled", family: "structured-input", boundary: "#structured-disabled select", disabled: true, expectedRoles: roles},
		{name: "structured/add-default", family: "structured-input", boundary: "#structured-default button[data-add-row]", expectedRoles: roles},
		{name: "structured/add-focus-visible", family: "structured-input", boundary: "#structured-focus button[data-add-row]", focusTarget: "#structured-focus button[data-add-row]", expectedRoles: roles},

		{name: "palette/swatch-default", family: "palette", boundary: "#palette-default button[data-cls='white']", expectedRoles: roles},
		{name: "palette/swatch-focus-visible", family: "palette", boundary: "#palette-focus button[data-cls='blue-500']", focusTarget: "#palette-focus button[data-cls='blue-500']", expectedRoles: roles},
		{name: "palette/preview", family: "palette", boundary: "#palette-default [data-selected-preview]", expectedRoles: roles},
		{name: "palette/hex-default", family: "palette", boundary: "#palette-default input[type='text']", expectedRoles: roles},
		{name: "palette/hex-invalid", family: "palette", boundary: "#palette-invalid input[type='text']", ariaInvalid: true, expectedRoles: appendRoles(roles, "aria-[invalid=true]:border-danger")},
		{name: "palette/hex-focus-visible", family: "palette", boundary: "#palette-focus input[type='text']", focusTarget: "#palette-focus input[type='text']", expectedRoles: roles},
		{name: "palette/black-swatch", family: "palette", boundary: "#palette-default button[data-cls='black']", expectedRoles: roles},
		{name: "palette/hue-swatch", family: "palette", boundary: "#palette-default button[data-cls='blue-500']", expectedRoles: roles},

		{name: "sidebar/search-default", family: "sidebar", boundary: ".sidebar-boundary input[type='search']", expectedRoles: roles},
		{name: "sidebar/search-focus-visible", family: "sidebar", boundary: ".sidebar-boundary input[type='search']", focusTarget: ".sidebar-boundary input[type='search']", expectedRoles: roles},

		{name: "search/trigger-default", family: "search", boundary: "#boundary-search button", expectedRoles: roles},
		{name: "search/trigger-focus-visible", family: "search", boundary: "#boundary-search button", focusTarget: "#boundary-search button", expectedRoles: roles},

		{name: "table/selection-default", family: "table", boundary: "#checkAll", expectedRoles: roles},
		{name: "table/selection-checked", family: "table", boundary: "#checkAll", checkedTarget: "#checkAll", expectedRoles: appendRoles(roles, "checked:border-primary", "dark:checked:border-primary-dark")},
		{name: "table/selection-focus-visible", family: "table", boundary: "#checkAll", focusTarget: "#checkAll", uncheckedTarget: "#checkAll", expectedRoles: roles},
		{name: "table/filter-search", family: "table", boundary: "#table-boundary-filters-query", expectedRoles: roles},
		{name: "table/filter-search-focus-visible", family: "table", boundary: "#table-boundary-filters-query", focusTarget: "#table-boundary-filters-query", expectedRoles: roles},
		{name: "table/filter-select", family: "table", boundary: "#table-boundary-filters-kind", expectedRoles: roles},

		{name: "schema/default", family: "schema-form", boundary: "[id='values.schema-default']", expectedRoles: roles},
		{name: "schema/disabled", family: "schema-form", boundary: "[id='values.schema-disabled']", disabled: true, expectedRoles: roles},
		{name: "schema/invalid", family: "schema-form", boundary: "[id='values.schema-invalid']", ariaInvalid: true, expectedRoles: roles},
		{name: "schema/boolean-checked", family: "schema-form", boundary: "[id='values.schema-boolean']", checkedTarget: "[id='values.schema-boolean']", expectedRoles: roles},
		{name: "schema/focus-visible", family: "schema-form", boundary: "[id='values.schema-focus']", focusTarget: "[id='values.schema-focus']", expectedRoles: roles},
		{name: "schema/boolean-focus-visible", family: "schema-form", boundary: "[id='values.schema-boolean-focus']", focusTarget: "[id='values.schema-boolean-focus']", expectedRoles: roles},
	}
}

func appendRoles(base []string, extra ...string) []string {
	result := append([]string{}, base...)
	return append(result, extra...)
}

func measureBoundary(t *testing.T, locator playwright.Locator) boundaryMeasurement {
	t.Helper()
	value, err := locator.Evaluate(`element => {
		const canvas = document.createElement('canvas');
		canvas.width = 1;
		canvas.height = 1;
		const context = canvas.getContext('2d', { willReadFrequently: true });
		const parse = value => {
			context.clearRect(0, 0, 1, 1);
			context.fillStyle = 'rgba(0, 0, 0, 0)';
			context.fillStyle = String(value);
			context.fillRect(0, 0, 1, 1);
			const pixel = context.getImageData(0, 0, 1, 1).data;
			return [pixel[0], pixel[1], pixel[2], pixel[3] / 255];
		};
		const composite = (front, back) => {
			const alpha = front[3] + back[3] * (1 - front[3]);
			if (alpha === 0) return [0, 0, 0, 0];
			return [
				(front[0] * front[3] + back[0] * back[3] * (1 - front[3])) / alpha,
				(front[1] * front[3] + back[1] * back[3] * (1 - front[3])) / alpha,
				(front[2] * front[3] + back[2] * back[3] * (1 - front[3])) / alpha,
				alpha,
			];
		};
		const exterior = node => {
			let color = [255, 255, 255, 1];
			const layers = [];
			for (let current = node; current; current = current.parentElement) {
				layers.push(parse(getComputedStyle(current).backgroundColor));
			}
			for (let index = layers.length - 1; index >= 0; index--) color = composite(layers[index], color);
			return color;
		};
		const linear = value => {
			value /= 255;
			return value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
		};
		const luminance = color => 0.2126 * linear(color[0]) + 0.7152 * linear(color[1]) + 0.0722 * linear(color[2]);
		const ratio = (a, b) => {
			const light = Math.max(luminance(a), luminance(b));
			const dark = Math.min(luminance(a), luminance(b));
			return (light + 0.05) / (dark + 0.05);
		};
		const css = getComputedStyle(element);
		const adjacent = exterior(element.parentElement);
		let opacity = 1;
		for (let current = element; current; current = current.parentElement) opacity *= Number(getComputedStyle(current).opacity || 1);
		const border = parse(css.borderTopColor);
		border[3] *= opacity;
		const rendered = composite(border, adjacent);
		const format = color => "rgba(" + color[0].toFixed(3) + ", " + color[1].toFixed(3) + ", " + color[2].toFixed(3) + ", " + color[3].toFixed(3) + ")";
		return {
			borderColor: css.borderTopColor,
			borderWidth: Number.parseFloat(css.borderTopWidth) || 0,
			borderStyle: css.borderTopStyle,
			adjacentColor: format(adjacent),
			renderedColor: format(rendered),
			contrast: ratio(rendered, adjacent),
			cumulativeOpacity: opacity,
			outlineColor: css.outlineColor,
			outlineWidth: css.outlineWidth,
		};
	}`, nil)
	require.NoError(t, err)

	raw, ok := value.(map[string]any)
	require.True(t, ok, "unexpected browser measurement type %T", value)
	return boundaryMeasurement{
		BorderColor:       raw["borderColor"].(string),
		BorderWidth:       browserNumber(raw["borderWidth"]),
		BorderStyle:       raw["borderStyle"].(string),
		AdjacentColor:     raw["adjacentColor"].(string),
		RenderedColor:     raw["renderedColor"].(string),
		Contrast:          browserNumber(raw["contrast"]),
		CumulativeOpacity: browserNumber(raw["cumulativeOpacity"]),
		OutlineColor:      raw["outlineColor"].(string),
		OutlineWidth:      raw["outlineWidth"].(string),
	}
}

func browserNumber(value any) float64 {
	switch number := value.(type) {
	case float64:
		return number
	case int:
		return float64(number)
	default:
		panic(fmt.Sprintf("unexpected browser number type %T", value))
	}
}

func controlBoundaryHandler(t *testing.T) http.Handler {
	t.Helper()
	page := controlBoundaryPage(t)
	mux := http.NewServeMux()
	mux.Handle("/assets/", assets.Handler())
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		themeName := request.URL.Query().Get("theme")
		darkClass := ""
		if request.URL.Query().Get("dark") == "true" {
			darkClass = " dark"
		}
		document := fmt.Sprintf(`<!doctype html><html data-theme=%q class=%q><head><meta charset="utf-8">%s</head><body class="bg-white text-neutral-950 dark:bg-neutral-950 dark:text-white"><main class="space-y-8 p-8">%s</main></body></html>`,
			themeName, darkClass, renderComponent(t, head.Dependencies(head.WithLocalRuntime())), page)
		_, _ = writer.Write([]byte(document))
	})
	return mux
}

func controlBoundaryPage(t *testing.T) string {
	t.Helper()
	components := []templ.Component{
		checkbox.Checkbox(checkbox.Config{ID: "cb-default", Label: "Default"}),
		checkbox.Checkbox(checkbox.Config{ID: "cb-disabled", Label: "Disabled", Disabled: true}),
		checkbox.Checkbox(checkbox.Config{ID: "cb-checked", Label: "Checked"}),
		checkbox.Checkbox(checkbox.Config{ID: "cb-focus", Label: "Focus"}),
		radio.Radio(radio.Config{ID: "radio-default", Name: "radio-default-name", Label: "Default"}),
		radio.Radio(radio.Config{ID: "radio-disabled", Name: "radio-disabled-name", Label: "Disabled", Disabled: true}),
		radio.Radio(radio.Config{ID: "radio-checked", Name: "radio-checked-name", Label: "Checked"}),
		radio.Radio(radio.Config{ID: "radio-focus", Name: "radio-focus-name", Label: "Focus"}),
		toggle.Toggle(toggle.Config{ID: "toggle-default", Label: "Default"}),
		toggle.Toggle(toggle.Config{ID: "toggle-disabled", Label: "Disabled", Disabled: true}),
		toggle.Toggle(toggle.Config{ID: "toggle-checked", Label: "Checked"}),
		toggle.Toggle(toggle.Config{ID: "toggle-focus", Label: "Focus"}),
		fileinput.FileInput(fileinput.Config{ID: "file-default", Label: "Default", Appearance: fileinput.AppearanceUpload}),
		fileinput.FileInput(fileinput.Config{ID: "file-disabled", Label: "Disabled", Appearance: fileinput.AppearanceUpload, Disabled: true}),
		fileinput.FileInput(fileinput.Config{ID: "file-focus", Label: "Focus", Appearance: fileinput.AppearanceUpload}),
		fileinput.FileInput(fileinput.Config{ID: "drop-default", Label: "Dropzone", Appearance: fileinput.AppearanceDropZone, RootClass: "drop-default"}),
		fileinput.FileInput(fileinput.Config{ID: "drop-disabled", Label: "Disabled dropzone", Appearance: fileinput.AppearanceDropZone, Disabled: true, RootClass: "drop-disabled"}),
		selectcomponent.Select(selectcomponent.Config{ID: "select-default", Label: "Default", Options: []selectcomponent.Option{{Value: "one", Label: "One"}}}),
		selectcomponent.Select(selectcomponent.Config{ID: "select-disabled", Label: "Disabled", Disabled: true, Options: []selectcomponent.Option{{Value: "one", Label: "One"}}}),
		selectcomponent.Select(selectcomponent.Config{ID: "select-invalid", Label: "Invalid", State: selectcomponent.StateError, Options: []selectcomponent.Option{{Value: "one", Label: "One"}}}),
		selectcomponent.Select(selectcomponent.Config{ID: "select-focus", Label: "Focus", Options: []selectcomponent.Option{{Value: "one", Label: "One"}}}),
		structuredinput.StructuredInput(structuredinput.Config{ID: "structured-default", Columns: controlStructuredColumns(), Entries: []structuredinput.Entry{{"name": "One", "kind": "a"}}}),
		structuredinput.StructuredInput(structuredinput.Config{ID: "structured-disabled", Disabled: true, Columns: controlStructuredColumns(), Entries: []structuredinput.Entry{{"name": "One", "kind": "a"}}}),
		structuredinput.StructuredInput(structuredinput.Config{ID: "structured-focus", Columns: controlStructuredColumns(), Entries: []structuredinput.Entry{{"name": "One", "kind": "a"}}}),
		palette.Palette(palette.Config{ID: "palette-default", Hues: []string{"blue"}, Shades: []string{"500"}, ShowHex: true}),
		palette.Palette(palette.Config{ID: "palette-invalid", Hues: []string{"blue"}, Shades: []string{"500"}, ShowHex: true}),
		palette.Palette(palette.Config{ID: "palette-focus", Hues: []string{"blue"}, Shades: []string{"500"}, ShowHex: true}),
		sidebar.Sidebar(sidebar.Config{RootClass: "sidebar-boundary", ShowSearch: true}),
		search.Search(search.Config{ID: "boundary-search"}),
		table.Table(table.Config{ID: "table-boundary", ShowCheckbox: true, Columns: []table.Column{{Key: "name", Label: "Name"}}, Rows: []table.Row{{ID: "one", Cells: map[string]table.Cell{"name": {Text: "One"}}}}, Filters: &table.FilterConfig{Appearance: table.FilterAppearanceInline, Filters: []table.Filter{{Key: "query", Label: "Query", Type: table.FilterSearch}, {Key: "kind", Label: "Kind", Type: table.FilterSelect, Options: []table.FilterOption{{Value: "one", Label: "One"}}}}}}),
		schemaform.Fields(schemaform.FieldsConfig{Fields: controlSchemaFields()}),
	}
	var builder strings.Builder
	for _, component := range components {
		builder.WriteString(renderComponent(t, component))
	}
	builder.WriteString(`<div data-boundary-ready></div>`)
	return builder.String()
}

func controlStructuredColumns() []structuredinput.Column {
	return []structuredinput.Column{
		{Key: "name", Label: "Name", Type: structuredinput.ColumnText},
		{Key: "kind", Label: "Kind", Type: structuredinput.ColumnSelect, Options: []structuredinput.Option{{Value: "a", Label: "A"}}},
	}
}

func controlSchemaFields() []schemaform.Field {
	return []schemaform.Field{
		{Path: "schema-default", Name: "schema-default", Label: "Default", Kind: schemaform.KindString},
		{Path: "schema-disabled", Name: "schema-disabled", Label: "Disabled", Kind: schemaform.KindString, Managed: true},
		{Path: "schema-invalid", Name: "schema-invalid", Label: "Invalid", Kind: schemaform.KindString, Errors: []string{"Invalid"}},
		{Path: "schema-boolean", Name: "schema-boolean", Label: "Boolean", Kind: schemaform.KindBoolean},
		{Path: "schema-focus", Name: "schema-focus", Label: "Focus", Kind: schemaform.KindString},
		{Path: "schema-boolean-focus", Name: "schema-boolean-focus", Label: "Boolean focus", Kind: schemaform.KindBoolean},
		{Path: "schema-boolean-disabled", Name: "schema-boolean-disabled", Label: "Boolean disabled", Kind: schemaform.KindBoolean, Managed: true},
	}
}

func renderComponent(t *testing.T, component templ.Component) string {
	t.Helper()
	var builder strings.Builder
	require.NoError(t, component.Render(context.Background(), &builder))
	return builder.String()
}
