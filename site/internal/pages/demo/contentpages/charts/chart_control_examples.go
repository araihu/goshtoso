package charts

import (
	"fmt"
	"maps"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	"github.com/araihu/goshtoso-charts/components/line"
)

const (
	staticChartControlTarget      = "static-chart-control-example"
	interactiveChartControlTarget = "interactive-chart-control-example"
	paletteChartControlTarget     = "palette-chart-control-example"
)

type staticChartControlState struct {
	Mode        chartcontrol.WrapperMode
	ModeValue   string
	StrokeWidth int
	Area        bool
	Errors      []string
}

type interactiveChartControlState struct {
	Orientation interactive.BarOrientation
	Scale       int
	ShowLabels  bool
	Errors      []string
}

// ChartControlExamples is the validated request state rendered by the canonical
// chart-controls guide. The fields remain private so callers cannot bypass the
// guide's closed request-value mapping.
type ChartControlExamples struct {
	static      staticChartControlState
	interactive interactiveChartControlState
	palette     paletteChartControlState
}

// ParseChartControlExamples maps query values to the guide's closed chart API.
func ParseChartControlExamples(values url.Values) ChartControlExamples {
	return ChartControlExamples{
		static:      parseStaticChartControlState(values),
		interactive: parseInteractiveChartControlState(values),
		palette:     parsePaletteChartControlState(values),
	}
}

func defaultChartControlExamples() ChartControlExamples {
	return ParseChartControlExamples(nil)
}

// ChartControlExampleForTarget returns one swap-safe example fragment for a
// recognized HTMX target. Full guide and navigation requests return false.
func ChartControlExampleForTarget(examples ChartControlExamples, target string) (templ.Component, bool) {
	switch target {
	case staticChartControlTarget:
		return staticChartControlExample(examples.static), true
	case interactiveChartControlTarget:
		return interactiveChartControlExample(examples.interactive), true
	case paletteChartControlTarget:
		return paletteChartControlExample(examples.palette), true
	default:
		return nil, false
	}
}

func parseStaticChartControlState(values url.Values) staticChartControlState {
	state := staticChartControlState{
		Mode:        chartcontrol.WrapperModeEnabled,
		ModeValue:   "enabled",
		StrokeWidth: 3,
		Area:        true,
	}
	if values.Get("static_present") != "1" {
		return state
	}

	state.Area = values.Has("static_area")
	switch mode := values.Get("static_mode"); mode {
	case "enabled":
		state.Mode = chartcontrol.WrapperModeEnabled
		state.ModeValue = mode
	case "disabled":
		state.Mode = chartcontrol.WrapperModeDisabled
		state.ModeValue = mode
	case "hidden":
		state.Mode = chartcontrol.WrapperModeHidden
		state.ModeValue = mode
	case "omitted":
		state.Mode = chartcontrol.WrapperModeOmitted
		state.ModeValue = mode
	default:
		state.Errors = append(state.Errors, "Wrapper mode must be enabled, disabled, hidden, or omitted; enabled was restored.")
	}

	strokeWidth, err := strconv.Atoi(values.Get("static_stroke"))
	if err != nil || strokeWidth < 1 || strokeWidth > 8 {
		state.Errors = append(state.Errors, "Stroke width must be a whole number from 1 through 8; 3 was restored.")
	} else {
		state.StrokeWidth = strokeWidth
	}
	return state
}

func parseInteractiveChartControlState(values url.Values) interactiveChartControlState {
	state := interactiveChartControlState{
		Orientation: interactive.BarOrientationVertical,
		Scale:       100,
		ShowLabels:  true,
	}
	if values.Get("interactive_present") != "1" {
		return state
	}

	state.ShowLabels = values.Has("interactive_labels")
	switch orientation := interactive.BarOrientation(values.Get("interactive_orientation")); orientation {
	case interactive.BarOrientationVertical, interactive.BarOrientationHorizontal:
		state.Orientation = orientation
	default:
		state.Errors = append(state.Errors, "Orientation must be vertical or horizontal; vertical was restored.")
	}

	scale, err := strconv.Atoi(values.Get("interactive_scale"))
	if err != nil || scale < 50 || scale > 150 || scale%10 != 0 {
		state.Errors = append(state.Errors, "Value scale must be a 10% step from 50% through 150%; 100% was restored.")
	} else {
		state.Scale = scale
	}
	return state
}

func staticChartControlConfig(state staticChartControlState) line.Config {
	minimum := 0.0
	noGap := false
	area := line.AreaOptions{}
	if state.Area {
		area = line.AreaOptions{Enabled: true, Opacity: 150.0 / 255.0}
	}
	return line.Config{
		Label:       "Controlled weekly email line",
		Caption:     fmt.Sprintf("Seven exact values; %d px stroke; area fill %s; wrapper %s.", state.StrokeWidth, enabledText(state.Area), state.ModeValue),
		Title:       line.Title{Text: "Weekly email"},
		Labels:      []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Series:      []line.Series{{Name: "Email", Values: []float64{120, 132, 101, 134, 90, 230, 210}}},
		Area:        area,
		XAxis:       line.CategoryAxisOptions{BoundaryGap: &noGap},
		YAxes:       []line.Axis{{Min: &minimum}},
		StrokeWidth: float64(state.StrokeWidth),
		Width:       600,
		Height:      400,
		Controls: chartcontrol.Options{
			Mode:       state.Mode,
			Fullscreen: true,
		},
		Export: &chartcontrol.ExportOptions{Filename: "controlled-weekly-email"},
	}
}

func interactiveChartControlConfig(state interactiveChartControlState) interactive.BarConfig {
	cfg := sampleInteractiveBar()
	cfg.Label = "Controlled weekly category bar"
	cfg.Caption = fmt.Sprintf("Fourteen exact values at %d%% scale; %s orientation; value labels %s.", state.Scale, state.Orientation, enabledText(state.ShowLabels))
	cfg.Orientation = state.Orientation
	cfg.Options.Title = &interactive.TitleOptions{Text: "Weekly categories"}
	cfg.Options.Export = &chartcontrol.ExportOptions{Filename: "controlled-weekly-categories"}
	cfg.SeriesOptions.Label = &interactive.LabelOptions{Show: interactive.Bool(state.ShowLabels), Position: "top"}
	for seriesIndex := range cfg.Series {
		for valueIndex := range cfg.Series[seriesIndex].Data {
			scaled := cfg.Series[seriesIndex].Data[valueIndex].Value * float64(state.Scale) / 100
			cfg.Series[seriesIndex].Data[valueIndex].Value = math.Round(scaled*10) / 10
		}
	}
	return cfg
}

func enabledText(value bool) string {
	if value {
		return "on"
	}
	return "off"
}

func chartControlHTMXAttrs(target string) templ.Attributes {
	return templ.Attributes{
		"hx-get":     "/modules/charts/docs/chart-controls",
		"hx-target":  "#" + target,
		"hx-swap":    "outerHTML focus-scroll:false",
		"hx-include": "closest form",
		"hx-trigger": "change",
	}
}

func mergeAttributes(base, extra templ.Attributes) templ.Attributes {
	merged := make(templ.Attributes, len(base)+len(extra))
	maps.Copy(merged, base)
	maps.Copy(merged, extra)
	return merged
}

//go:fix inline
func boolPointer(value bool) *bool { return new(value) }

func selected(value, current string) bool { return value == current }

func joinedErrors(errors []string) string { return strings.Join(errors, " ") }

// These copyable .templ examples intentionally include request parsing and the
// native GET fallback. Tests keep input names, closed values, datasets, and
// component options aligned with the rendered guide examples.
const staticChartControlSource = `package chartsdemo

import (
  "fmt"
  "net/http"
  "strconv"
  "strings"

  "github.com/a-h/templ"
  "github.com/araihu/goshtoso-charts/components/chartcontrol"
  "github.com/araihu/goshtoso-charts/components/line"
  "github.com/araihu/goshtoso/components/button"
  "github.com/araihu/goshtoso/components/checkbox"
  "github.com/araihu/goshtoso/components/form"
  rangeinput "github.com/araihu/goshtoso/components/range"
)

type staticState struct {
  Mode chartcontrol.WrapperMode
  ModeValue string
  Stroke int
  Area bool
  Errors []string
}

func staticStateFromRequest(r *http.Request) staticState {
  state := staticState{Mode: chartcontrol.WrapperModeEnabled, ModeValue: "enabled", Stroke: 3, Area: true}
  if r.URL.Query().Get("static_present") != "1" { return state }
  state.Area = r.URL.Query().Has("static_area")
  switch r.URL.Query().Get("static_mode") {
  case "enabled": state.Mode = chartcontrol.WrapperModeEnabled
  case "disabled": state.Mode, state.ModeValue = chartcontrol.WrapperModeDisabled, "disabled"
  case "hidden": state.Mode, state.ModeValue = chartcontrol.WrapperModeHidden, "hidden"
  case "omitted": state.Mode, state.ModeValue = chartcontrol.WrapperModeOmitted, "omitted"
  default: state.Errors = append(state.Errors, "Wrapper mode must be enabled, disabled, hidden, or omitted; enabled was restored.")
  }
  if stroke, err := strconv.Atoi(r.URL.Query().Get("static_stroke")); err == nil && stroke >= 1 && stroke <= 8 {
    state.Stroke = stroke
  } else {
    state.Errors = append(state.Errors, "Stroke width must be a whole number from 1 through 8; 3 was restored.")
  }
  return state
}

func staticConfig(state staticState) line.Config {
  minimum, noGap := 0.0, false
  area := line.AreaOptions{}
  if state.Area { area = line.AreaOptions{Enabled: true, Opacity: 150.0 / 255.0} }
  return line.Config{
    Label: "Controlled weekly email line",
    Caption: fmt.Sprintf("Seven exact values; %d px stroke; area fill %s; wrapper %s.", state.Stroke, onOff(state.Area), state.ModeValue),
    Title: line.Title{Text: "Weekly email"},
    Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
    Series: []line.Series{{Name: "Email", Values: []float64{120, 132, 101, 134, 90, 230, 210}}},
    Area: area,
    XAxis: line.CategoryAxisOptions{BoundaryGap: &noGap},
    YAxes: []line.Axis{{Min: &minimum}},
    StrokeWidth: float64(state.Stroke), Width: 600, Height: 400,
    Controls: chartcontrol.Options{Mode: state.Mode, Fullscreen: true},
    Export: &chartcontrol.ExportOptions{Filename: "controlled-weekly-email"},
  }
}

func onOff(value bool) string { if value { return "on" }; return "off" }

func changeAttrs() templ.Attributes {
  return templ.Attributes{
    "hx-get": "/modules/charts/docs/chart-controls", "hx-target": "#static-chart-control-example",
    "hx-swap": "outerHTML focus-scroll:false", "hx-include": "closest form", "hx-trigger": "change",
  }
}

func falsePointer() *bool { value := false; return &value }

func StaticChartControlsHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  if err := StaticChartControls(staticStateFromRequest(r)).Render(r.Context(), w); err != nil {
    http.Error(w, "render static chart controls", http.StatusInternalServerError)
  }
}

templ StaticChartControls(state staticState) {
  <div id="static-chart-control-example">
    @form.Form(form.Config{Action: "/modules/charts/docs/chart-controls#static-chart-control-example", Method: "get", PreventEnterSubmit: falsePointer(), HTMX: &form.HTMXConfig{Get: "/modules/charts/docs/chart-controls", Target: "#static-chart-control-example", Swap: "outerHTML focus-scroll:false"}}) {
      <input type="hidden" name="static_present" value="1"/>
      <label for="static-mode">Wrapper mode</label>
      <select id="static-mode" name="static_mode" { changeAttrs()... }>
        <option value="enabled" selected?={ state.ModeValue == "enabled" }>Enabled</option>
        <option value="disabled" selected?={ state.ModeValue == "disabled" }>Disabled</option>
        <option value="hidden" selected?={ state.ModeValue == "hidden" }>Hidden</option>
        <option value="omitted" selected?={ state.ModeValue == "omitted" }>Omitted</option>
      </select>
      @rangeinput.Range(rangeinput.Config{ID: "static-stroke", Name: "static_stroke", Label: "Stroke width", Value: strconv.Itoa(state.Stroke), Min: "1", Max: "8", InputAttrs: changeAttrs()})
      @checkbox.Checkbox(checkbox.Config{ID: "static-area", Name: "static_area", Value: "on", Label: "Fill area below line", Checked: state.Area, InputAttrs: changeAttrs()})
      @button.Button(button.WithType("submit"), button.WithID("static-apply")) { Apply static controls }
      if len(state.Errors) > 0 { <p role="alert">{ strings.Join(state.Errors, " ") }</p> }
      <p aria-live="polite">Applied: { state.ModeValue } wrapper, { strconv.Itoa(state.Stroke) } px stroke, area { onOff(state.Area) }.</p>
    }
    @line.Line(staticConfig(state))
  </div>
}`

const interactiveChartControlSource = `package chartsdemo

import (
  "fmt"
  "math"
  "math/rand"
  "net/http"
  "strconv"
  "strings"

  "github.com/a-h/templ"
  "github.com/araihu/goshtoso-charts/components/chartcontrol"
  interactive "github.com/araihu/goshtoso-charts/components/interactive"
  "github.com/araihu/goshtoso/components/button"
  "github.com/araihu/goshtoso/components/checkbox"
  "github.com/araihu/goshtoso/components/form"
  rangeinput "github.com/araihu/goshtoso/components/range"
)

type interactiveState struct {
  Orientation interactive.BarOrientation
  Scale int
  Labels bool
  Errors []string
}

func interactiveStateFromRequest(r *http.Request) interactiveState {
  state := interactiveState{Orientation: interactive.BarOrientationVertical, Scale: 100, Labels: true}
  if r.URL.Query().Get("interactive_present") != "1" { return state }
  state.Labels = r.URL.Query().Has("interactive_labels")
  switch r.URL.Query().Get("interactive_orientation") {
  case "vertical":
  case "horizontal": state.Orientation = interactive.BarOrientationHorizontal
  default: state.Errors = append(state.Errors, "Orientation must be vertical or horizontal; vertical was restored.")
  }
  if scale, err := strconv.Atoi(r.URL.Query().Get("interactive_scale")); err == nil && scale >= 50 && scale <= 150 && scale%10 == 0 {
    state.Scale = scale
  } else {
    state.Errors = append(state.Errors, "Value scale must be a 10% step from 50% through 150%; 100% was restored.")
  }
  return state
}

func fixedData(seed int64, scale int) []interactive.BarData {
  source := rand.New(rand.NewSource(seed))
  data := make([]interactive.BarData, 7)
  for index := range data {
    scaled := float64(source.Intn(300)) * float64(scale) / 100
    data[index].Value = math.Round(scaled * 10) / 10
  }
  return data
}

func interactiveConfig(state interactiveState) interactive.BarConfig {
  return interactive.BarConfig{
    Label: "Controlled weekly category bar",
    Caption: fmt.Sprintf("Fourteen exact values at %d%% scale; %s orientation; value labels %s.", state.Scale, state.Orientation, interactiveOnOff(state.Labels)),
    XAxis: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
    Series: []interactive.BarSeries{{Name: "Category A", Data: fixedData(11, state.Scale)}, {Name: "Category B", Data: fixedData(12, state.Scale)}},
    Orientation: state.Orientation,
    SeriesOptions: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(state.Labels), Position: "top"}},
    Options: interactive.ChartOptions{
      Title: &interactive.TitleOptions{Text: "Weekly categories"},
      Legend: &interactive.LegendOptions{Bottom: "0"},
      Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "axis"},
      Controls: chartcontrol.Options{Fullscreen: true},
      Export: &chartcontrol.ExportOptions{Filename: "controlled-weekly-categories"},
    },
  }
}

func interactiveOnOff(value bool) string { if value { return "on" }; return "off" }

func interactiveChangeAttrs() templ.Attributes {
  return templ.Attributes{
    "hx-get": "/modules/charts/docs/chart-controls", "hx-target": "#interactive-chart-control-example",
    "hx-swap": "outerHTML focus-scroll:false", "hx-include": "closest form", "hx-trigger": "change",
  }
}

func interactiveFalsePointer() *bool { value := false; return &value }

func InteractiveChartControlsHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  if err := InteractiveChartControls(interactiveStateFromRequest(r)).Render(r.Context(), w); err != nil {
    http.Error(w, "render interactive chart controls", http.StatusInternalServerError)
  }
}

templ InteractiveChartControls(state interactiveState) {
  <div id="interactive-chart-control-example">
    @form.Form(form.Config{Action: "/modules/charts/docs/chart-controls#interactive-chart-control-example", Method: "get", PreventEnterSubmit: interactiveFalsePointer(), HTMX: &form.HTMXConfig{Get: "/modules/charts/docs/chart-controls", Target: "#interactive-chart-control-example", Swap: "outerHTML focus-scroll:false"}}) {
      <input type="hidden" name="interactive_present" value="1"/>
      <label for="interactive-orientation">Orientation</label>
      <select id="interactive-orientation" name="interactive_orientation" { interactiveChangeAttrs()... }>
        <option value="vertical" selected?={ state.Orientation == interactive.BarOrientationVertical }>Vertical</option>
        <option value="horizontal" selected?={ state.Orientation == interactive.BarOrientationHorizontal }>Horizontal</option>
      </select>
      @rangeinput.Range(rangeinput.Config{ID: "interactive-scale", Name: "interactive_scale", Label: "Value scale (%)", Value: strconv.Itoa(state.Scale), Min: "50", Max: "150", Step: "10", InputAttrs: interactiveChangeAttrs()})
      @checkbox.Checkbox(checkbox.Config{ID: "interactive-labels", Name: "interactive_labels", Value: "on", Label: "Show value labels", Checked: state.Labels, InputAttrs: interactiveChangeAttrs()})
      @button.Button(button.WithType("submit"), button.WithID("interactive-apply")) { Apply interactive controls }
      if len(state.Errors) > 0 { <p role="alert">{ strings.Join(state.Errors, " ") }</p> }
      <p aria-live="polite">Applied: { string(state.Orientation) }, { strconv.Itoa(state.Scale) }% scale, labels { interactiveOnOff(state.Labels) }.</p>
    }
    @interactive.Bar(interactiveConfig(state))
  </div>
}`
