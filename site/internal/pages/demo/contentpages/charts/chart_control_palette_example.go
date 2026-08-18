package charts

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/heatmap"
	"github.com/araihu/goshtoso-charts/components/line"
	"github.com/araihu/goshtoso-charts/components/pie"
)

const (
	paletteChoiceTheme  = "theme"
	paletteChoiceCustom = "custom"
)

var defaultPaletteChartColors = [4]string{"#0e7490", "#2563eb", "#d97706", "#be123c"}

type paletteChartControlState struct {
	Palette      charttheme.Palette
	PaletteValue string
	Colors       [4]string
	Custom       bool
	Errors       []string
}

func parsePaletteChartControlState(values url.Values) paletteChartControlState {
	state := paletteChartControlState{
		PaletteValue: paletteChoiceTheme,
		Colors:       defaultPaletteChartColors,
	}
	if values.Get("palette_present") != "1" {
		return state
	}

	switch choice := values.Get("chart_palette"); choice {
	case paletteChoiceTheme:
		state.PaletteValue = choice
	case "araihu":
		state.Palette, state.PaletteValue = charttheme.PaletteAraiHu, choice
	case "bold":
		state.Palette, state.PaletteValue = charttheme.PaletteBold, choice
	case "neutral":
		state.Palette, state.PaletteValue = charttheme.PaletteNeutral, choice
	case "pastel":
		state.Palette, state.PaletteValue = charttheme.PalettePastel, choice
	case "status":
		state.Palette, state.PaletteValue = charttheme.PaletteStatus, choice
	case paletteChoiceCustom:
		state.PaletteValue, state.Custom = choice, true
		for index := range state.Colors {
			name := fmt.Sprintf("palette_color_%d", index+1)
			if !values.Has(name) {
				continue
			}
			color := strings.ToLower(strings.TrimSpace(values.Get(name)))
			if validHexColor(color) {
				state.Colors[index] = color
				continue
			}
			state.Errors = append(state.Errors, fmt.Sprintf("Color %d must use six-digit hexadecimal notation; %s was restored.", index+1, defaultPaletteChartColors[index]))
		}
	default:
		state.Errors = append(state.Errors, "Palette must be theme, Arai Hû, bold, neutral, pastel, status, or custom; theme was restored.")
	}
	return state
}

func validHexColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}
	for _, character := range value[1:] {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func paletteChartStyle(state paletteChartControlState) charttheme.Style {
	style := charttheme.Style{Palette: state.Palette}
	if state.Custom {
		style.Colors = append([]string(nil), state.Colors[:]...)
	}
	return style
}

func paletteLineConfig(state paletteChartControlState) line.Config {
	cfg := sampleAreaLine()
	cfg.Label = "Shared-palette line and area"
	cfg.Caption = "Color 1 controls the line stroke and its translucent area."
	cfg.Style = paletteChartStyle(state)
	cfg.Export.Filename = "shared-palette-line"
	return cfg
}

func paletteBarConfig(state paletteChartControlState) bar.Config {
	cfg := sampleBasicBar()
	cfg.Label = "Shared-palette grouped bars"
	cfg.Caption = "Colors 1 and 2 distinguish rainfall and evaporation."
	cfg.Labels = append([]string(nil), cfg.Labels[:6]...)
	for index := range cfg.Series {
		cfg.Series[index].Values = append([]float64(nil), cfg.Series[index].Values[:6]...)
	}
	cfg.Style = paletteChartStyle(state)
	cfg.Export.Filename = "shared-palette-bars"
	return cfg
}

func palettePieConfig(state paletteChartControlState) pie.Config {
	cfg := sampleBasicPie()
	cfg.Label = "Shared-palette channel shares"
	cfg.Caption = "Palette order maps to sectors; a custom four-color palette repeats Color 1 for the fifth sector."
	cfg.Style = paletteChartStyle(state)
	if state.Custom {
		cfg.Style.Colors = append(cfg.Style.Colors, state.Colors[0])
	}
	cfg.Export.Filename = "shared-palette-pie"
	return cfg
}

func paletteHeatMapConfig(state paletteChartControlState) heatmap.Config {
	cfg := sampleBasicHeatMap()
	cfg.Label = "Shared-palette heat map"
	cfg.Caption = "The same palette becomes an ordered cold-to-warm value gradient."
	cfg.Style = charttheme.Style{Palette: state.Palette}
	// The sample root attributes are cleared so repeated HTMX swaps cannot duplicate them.
	cfg.RootAttrs = nil
	if state.Custom {
		cfg.Gradient = heatmap.Gradient{Stops: []heatmap.GradientStop{
			{At: 0, Color: state.Colors[0], Class: "palette-low"},
			{At: 1.0 / 3.0, Color: state.Colors[1], Class: "palette-low-mid"},
			{At: 2.0 / 3.0, Color: state.Colors[2], Class: "palette-high-mid"},
			{At: 1, Color: state.Colors[3], Class: "palette-warm"},
		}}
	}
	cfg.Export.Filename = "shared-palette-heat-map"
	return cfg
}

func paletteColorLabel(index int) string {
	switch index {
	case 0:
		return "Color 1 · line / low"
	case 1:
		return "Color 2 · series / low-mid"
	case 2:
		return "Color 3 · sector / high-mid"
	default:
		return "Color 4 · sector / warm"
	}
}

func paletteAppliedText(state paletteChartControlState) string {
	if !state.Custom {
		return state.PaletteValue + " palette"
	}
	return fmt.Sprintf("custom palette %s, %s, %s, %s", state.Colors[0], state.Colors[1], state.Colors[2], state.Colors[3])
}

const paletteChartControlSource = `package chartsdemo

import (
  "fmt"
  "net/http"
  "strings"

  "github.com/a-h/templ"
  "github.com/araihu/goshtoso-charts/components/bar"
  "github.com/araihu/goshtoso-charts/components/chartcontrol"
  "github.com/araihu/goshtoso-charts/components/charttheme"
  "github.com/araihu/goshtoso-charts/components/heatmap"
  "github.com/araihu/goshtoso-charts/components/line"
  "github.com/araihu/goshtoso-charts/components/pie"
  "github.com/araihu/goshtoso/components/button"
  "github.com/araihu/goshtoso/components/form"
)

var customDefaults = [4]string{"#0e7490", "#2563eb", "#d97706", "#be123c"}

type paletteState struct {
  Palette charttheme.Palette
  Value string
  Colors [4]string
  Custom bool
  Errors []string
}

func paletteStateFromRequest(r *http.Request) paletteState {
  state := paletteState{Value: "theme", Colors: customDefaults}
  if r.URL.Query().Get("palette_present") != "1" { return state }
  switch value := r.URL.Query().Get("chart_palette"); value {
  case "theme": state.Value = value
  case "araihu": state.Palette, state.Value = charttheme.PaletteAraiHu, value
  case "bold": state.Palette, state.Value = charttheme.PaletteBold, value
  case "neutral": state.Palette, state.Value = charttheme.PaletteNeutral, value
  case "pastel": state.Palette, state.Value = charttheme.PalettePastel, value
  case "status": state.Palette, state.Value = charttheme.PaletteStatus, value
  case "custom":
    state.Value, state.Custom = value, true
    for index := range state.Colors {
      name := fmt.Sprintf("palette_color_%d", index+1)
      if !r.URL.Query().Has(name) { continue }
      color := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(name)))
      if validHex(color) { state.Colors[index] = color } else {
        state.Errors = append(state.Errors, fmt.Sprintf("Color %d must use six-digit hexadecimal notation; %s was restored.", index+1, customDefaults[index]))
      }
    }
  default: state.Errors = append(state.Errors, "Palette must be theme, Arai Hû, bold, neutral, pastel, status, or custom; theme was restored.")
  }
  return state
}

func validHex(value string) bool {
  if len(value) != 7 || value[0] != '#' { return false }
  for _, character := range value[1:] {
    if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) { return false }
  }
  return true
}

func style(state paletteState) charttheme.Style {
  result := charttheme.Style{Palette: state.Palette}
  if state.Custom { result.Colors = append([]string(nil), state.Colors[:]...) }
  return result
}

func lineConfig(state paletteState) line.Config {
  minimum, noGap := 0.0, false
  return line.Config{
    Label: "Shared-palette line and area", Caption: "Color 1 controls the line stroke and its translucent area.", Title: line.Title{Text: "Line"},
    Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
    Series: []line.Series{{Name: "Email", Values: []float64{120, 132, 101, 134, 90, 230, 210}}},
    Area: line.AreaOptions{Enabled: true, Opacity: 150.0 / 255.0},
    XAxis: line.CategoryAxisOptions{BoundaryGap: &noGap}, Legend: line.LegendOptions{Padding: line.Padding{Top: 5, Bottom: 10}}, YAxes: []line.Axis{{Min: &minimum}},
    Width: 600, Height: 400, Style: style(state), Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "shared-palette-line"},
  }
}

func barConfig(state paletteState) bar.Config {
  return bar.Config{
    Label: "Shared-palette grouped bars", Caption: "Colors 1 and 2 distinguish rainfall and evaporation.",
    Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
    Series: []bar.Series{
      {Name: "Rainfall", Values: []float64{2, 4.9, 7, 23.2, 25.6, 76.7}},
      {Name: "Evaporation", Values: []float64{2.6, 5.9, 9, 26.4, 28.7, 70.7}},
    },
    Title: "Bar Chart", Legend: bar.LegendOptions{Placement: bar.LegendPlacementEnd, Overlay: true},
    Width: 600, Height: 400, Style: style(state), Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "shared-palette-bars"},
  }
}

func pieConfig(state paletteState) pie.Config {
  colors := style(state)
  if state.Custom { colors.Colors = append(colors.Colors, state.Colors[0]) }
  return pie.Config{
    Label: "Shared-palette channel shares", Caption: "Palette order maps to sectors; Color 1 repeats for the fifth custom sector.",
    Title: pie.TitleOptions{Text: "Pie Chart", Subtitle: "(Fake Data)", Placement: pie.PlacementCenter, FontSize: 16, SubtitleFontSize: 10},
    Legend: pie.LegendOptions{Orientation: pie.LegendVertical, LeftPercent: 80, VerticalPlacement: pie.VerticalPlacementBottom, FontSize: 10},
    Padding: pie.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
    Slices: []pie.Slice{{Name: "Search Engine", Value: 1048}, {Name: "Direct", Value: 735}, {Name: "Email", Value: 580}, {Name: "Union Ads", Value: 484}, {Name: "Video Ads", Value: 300}},
    Width: 600, Height: 400, Style: colors, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "shared-palette-pie", Background: chartcontrol.ExportBackgroundTransparent},
  }
}

func heatMapConfig(state paletteState) heatmap.Config {
  config := heatmap.Config{
    Label: "Shared-palette heat map", Caption: "The same palette becomes an ordered cold-to-warm value gradient.", Title: "Heat Map Chart",
    XAxis: heatmap.Axis{Title: "X-Axis", Labels: []string{"0", "1", "2", "3", "4"}}, YAxis: heatmap.Axis{Title: "Y-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
    Rows: [][]float64{{4.4, 4.9, 7, 7.5, 4.3}, {2.6, 5.9, 9, 6.4, 2.3}, {3.3, 6.4, 7, 4.9, 3.2}, {1.9, 6, 9, 5.9, 2.6}, {4.4, 5.9, 7, 6.4, 4.6}},
    ValueRange: heatmap.ValueRange{Min: 1.9, Max: 9}, Width: 600, Height: 400,
    Style: charttheme.Style{Palette: state.Palette}, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "shared-palette-heat-map"},
  }
  if state.Custom { config.Gradient = heatmap.Gradient{Stops: []heatmap.GradientStop{
    {At: 0, Color: state.Colors[0]}, {At: 1.0 / 3.0, Color: state.Colors[1]},
    {At: 2.0 / 3.0, Color: state.Colors[2]}, {At: 1, Color: state.Colors[3]},
  }} }
  return config
}

func paletteChangeAttrs() templ.Attributes {
  return templ.Attributes{"hx-get": "/modules/charts/docs/chart-controls", "hx-target": "#palette-chart-control-example", "hx-swap": "outerHTML focus-scroll:false", "hx-include": "closest form", "hx-trigger": "change"}
}

func paletteFalsePointer() *bool { value := false; return &value }

func PaletteChartsHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  if err := PaletteCharts(paletteStateFromRequest(r)).Render(r.Context(), w); err != nil { http.Error(w, "render palette charts", http.StatusInternalServerError) }
}

templ PaletteCharts(state paletteState) {
  <div id="palette-chart-control-example">
    @form.Form(form.Config{ID: "palette-chart-control-form", Action: "/modules/charts/docs/chart-controls#palette-chart-control-example", Method: "get", PreventEnterSubmit: paletteFalsePointer(), HTMX: &form.HTMXConfig{Get: "/modules/charts/docs/chart-controls", Target: "#palette-chart-control-example", Swap: "outerHTML focus-scroll:false"}}) {
      <input type="hidden" name="palette_present" value="1"/>
      <label for="chart-palette">Chart palette</label>
      <select id="chart-palette" name="chart_palette" { paletteChangeAttrs()... }>
        <option value="theme" selected?={ state.Value == "theme" }>Theme tokens</option>
        <option value="araihu" selected?={ state.Value == "araihu" }>Arai Hû</option>
        <option value="bold" selected?={ state.Value == "bold" }>Bold</option>
        <option value="neutral" selected?={ state.Value == "neutral" }>Neutral</option>
        <option value="pastel" selected?={ state.Value == "pastel" }>Pastel</option>
        <option value="status" selected?={ state.Value == "status" }>Status</option>
        <option value="custom" selected?={ state.Value == "custom" }>Custom four colors</option>
      </select>
      for index, color := range state.Colors {
        <label for={ fmt.Sprintf("palette-color-%d", index+1) }>Color { fmt.Sprintf("%d", index+1) }</label>
        <input id={ fmt.Sprintf("palette-color-%d", index+1) } type="color" name={ fmt.Sprintf("palette_color_%d", index+1) } value={ color } disabled?={ !state.Custom } { paletteChangeAttrs()... }/>
      }
      @button.Button(button.WithType("submit"), button.WithID("palette-apply")) { Apply palette }
      if len(state.Errors) > 0 { <p role="alert">{ strings.Join(state.Errors, " ") }</p> }
	  <p aria-live="polite">Applied palette: { state.Value }.</p>
    }
    <div class="grid gap-5 lg:grid-cols-2">
      @line.Line(lineConfig(state))
      @bar.Bar(barConfig(state))
      @pie.Pie(pieConfig(state))
      @heatmap.HeatMap(heatMapConfig(state))
    </div>
  </div>
}`
