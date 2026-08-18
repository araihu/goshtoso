package charts

import (
	"fmt"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactiveradar "github.com/araihu/goshtoso-charts/components/interactive/radar"
)

var interactiveRadarIndicators = []interactiveradar.Indicator{
	{Name: "AQI", Max: 300}, {Name: "PM2.5", Max: 250}, {Name: "PM10", Max: 300},
	{Name: "CO", Max: 5}, {Name: "NO2", Max: 200}, {Name: "SO2", Max: 100},
}

// The upstream generator stores a seventh day index beside six pollutant
// dimensions. Observation names preserve that index as Day 1 through Day 21;
// the aligned value vector retains only the six declared pollutant dimensions.
var interactiveRadarBeijing = [][]float64{
	{55, 9, 56, .46, 18, 6}, {25, 11, 21, .65, 34, 9}, {56, 7, 63, .3, 14, 5},
	{33, 7, 29, .33, 16, 6}, {42, 24, 44, .76, 40, 16}, {82, 58, 90, 1.77, 68, 33},
	{74, 49, 77, 1.46, 48, 27}, {78, 55, 80, 1.29, 59, 29}, {267, 216, 280, 4.8, 108, 64},
	{185, 127, 216, 2.52, 61, 27}, {39, 19, 38, .57, 31, 15}, {41, 11, 40, .43, 21, 7},
	{64, 38, 74, 1.04, 46, 22}, {108, 79, 120, 1.7, 75, 41}, {108, 63, 116, 1.48, 44, 26},
	{33, 6, 29, .34, 13, 5}, {94, 66, 110, 1.54, 62, 31}, {186, 142, 192, 3.88, 93, 79},
	{57, 31, 54, .96, 32, 14}, {22, 8, 17, .48, 23, 10}, {39, 15, 36, .61, 29, 13},
}

var interactiveRadarGuangzhou = [][]float64{
	{26, 37, 27, 1.163, 27, 13}, {85, 62, 71, 1.195, 60, 8}, {78, 38, 74, 1.363, 37, 7},
	{21, 21, 36, .634, 40, 9}, {41, 42, 46, .915, 81, 13}, {56, 52, 69, 1.067, 92, 16},
	{64, 30, 28, .924, 51, 2}, {55, 48, 74, 1.236, 75, 26}, {76, 85, 113, 1.237, 114, 27},
	{91, 81, 104, 1.041, 56, 40}, {84, 39, 60, .964, 25, 11}, {64, 51, 101, .862, 58, 23},
	{70, 69, 120, 1.198, 65, 36}, {77, 105, 178, 2.549, 64, 16}, {109, 68, 87, .996, 74, 29},
	{73, 68, 97, .905, 51, 34}, {54, 27, 47, .592, 53, 12}, {51, 61, 97, .811, 65, 19},
	{91, 71, 121, 1.374, 43, 18}, {73, 102, 182, 2.787, 44, 19}, {73, 50, 76, .717, 31, 20},
}

var interactiveRadarShanghai = [][]float64{
	{91, 45, 125, .82, 34, 23}, {65, 27, 78, .86, 45, 29}, {83, 60, 84, 1.09, 73, 27},
	{109, 81, 121, 1.28, 68, 51}, {106, 77, 114, 1.07, 55, 51}, {109, 81, 121, 1.28, 68, 51},
	{106, 77, 114, 1.07, 55, 51}, {89, 65, 78, .86, 51, 26}, {53, 33, 47, .64, 50, 17},
	{80, 55, 80, 1.01, 75, 24}, {117, 81, 124, 1.03, 45, 24}, {99, 71, 142, 1.1, 62, 42},
	{95, 69, 130, 1.28, 74, 50}, {116, 87, 131, 1.47, 84, 40}, {108, 80, 121, 1.3, 85, 37},
	{134, 83, 167, 1.16, 57, 43}, {79, 43, 107, 1.05, 59, 37}, {71, 46, 89, .86, 64, 25},
	{97, 71, 113, 1.17, 88, 31}, {84, 57, 91, .85, 55, 31}, {87, 63, 101, .9, 56, 41},
}

func interactiveRadarData(values [][]float64) []interactiveradar.Data {
	data := make([]interactiveradar.Data, len(values))
	for index, vector := range values {
		data[index] = interactiveradar.Data{Name: fmt.Sprintf("Day %d", index+1), Values: vector}
	}
	return data
}

func interactiveRadarOptions(title, filename string, mode chart.LegendSelectionMode) chart.ChartOptions {
	return chart.ChartOptions{
		Title:    &chart.TitleOptions{Text: title, Left: "center"},
		Legend:   &chart.LegendOptions{Show: new(true), Left: "center", Bottom: "0", SelectionMode: mode},
		Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "item"},
		Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: filename},
	}
}

func interactiveRadarConfig(label, caption, title, filename string) interactiveradar.Config {
	return interactiveradar.Config{
		Label: label, Caption: caption, Indicators: interactiveRadarIndicators,
		Width: "100%", Height: "520px", Options: interactiveRadarOptions(title, filename, chart.LegendSelectionDefault),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractiveRadarBase() interactiveradar.Config {
	config := interactiveRadarConfig("Daily Beijing air quality", "Twenty-one daily observations across six pollutant dimensions.", "Basic radar example", "daily-beijing-air-quality")
	config.Series = []interactiveradar.Series{{Name: "Beijing", Data: interactiveRadarData(interactiveRadarBeijing)}}
	config.Coordinate = interactiveradar.CoordinateOptions{SplitArea: new(true), SplitLine: &interactiveradar.SplitLineOptions{Show: new(true)}}
	return config
}

func sampleInteractiveRadarStyle() interactiveradar.Config {
	config := interactiveRadarConfig("Circular style", "The Beijing observations use a circular coordinate with five subtle rings.", "Style options", "circular-beijing-air-quality")
	config.Series = []interactiveradar.Series{{Name: "Beijing", Data: interactiveRadarData(interactiveRadarBeijing)}}
	config.Coordinate = interactiveradar.CoordinateOptions{Shape: interactiveradar.ShapeCircle, SplitNumber: 5, SplitLine: &interactiveradar.SplitLineOptions{Show: new(true), Style: &chart.LineStyle{Width: 1, Opacity: new(.1)}}}
	config.SeriesOptions = chart.SeriesOptions{LineStyle: &chart.LineStyle{Width: 1, Opacity: new(.5)}, AreaStyle: &chart.AreaStyle{Opacity: new(.1)}}
	return config
}

func interactiveRadarCitySeries() []interactiveradar.Series {
	return []interactiveradar.Series{
		{Name: "Beijing", Data: interactiveRadarData(interactiveRadarBeijing)},
		{Name: "Guangzhou", Data: interactiveRadarData(interactiveRadarGuangzhou)},
		{Name: "Shanghai", Data: interactiveRadarData(interactiveRadarShanghai)},
	}
}

func sampleInteractiveRadarLegendMulti() interactiveradar.Config {
	config := interactiveRadarConfig("Multiple-series legend", "Toggle any combination of Beijing, Guangzhou, and Shanghai.", "Multiple legend selection", "multiple-city-air-quality")
	config.Series = interactiveRadarCitySeries()
	config.Coordinate = interactiveradar.CoordinateOptions{Shape: interactiveradar.ShapeCircle, SplitNumber: 5, SplitLine: &interactiveradar.SplitLineOptions{Show: new(true), Style: &chart.LineStyle{Opacity: new(.1)}}}
	config.Options.Legend.SelectionMode = chart.LegendSelectionMultiple
	config.SeriesOptions = chart.SeriesOptions{LineStyle: &chart.LineStyle{Width: 1, Opacity: new(.5)}, AreaStyle: &chart.AreaStyle{Opacity: new(.1)}}
	return config
}

func sampleInteractiveRadarLegendSingle() interactiveradar.Config {
	config := interactiveRadarConfig("Single-series legend", "Select one city's twenty-one daily profiles at a time.", "Single legend selection", "single-city-air-quality")
	config.Series = interactiveRadarCitySeries()
	config.Coordinate = interactiveradar.CoordinateOptions{Shape: interactiveradar.ShapeCircle, SplitNumber: 5, SplitLine: &interactiveradar.SplitLineOptions{Show: new(true), Style: &chart.LineStyle{Opacity: new(.1)}}}
	config.Options.Legend.SelectionMode = chart.LegendSelectionSingle
	config.SeriesOptions = chart.SeriesOptions{LineStyle: &chart.LineStyle{Width: 1, Opacity: new(.5)}, AreaStyle: &chart.AreaStyle{Opacity: new(.5)}}
	return config
}

func interactiveRadarBaseCode() string {
	return `@interactiveradar.Radar(interactiveradar.Config{
  Label: "Daily Beijing air quality",
  Indicators: pollutantIndicators,
  Series: []interactiveradar.Series{{Name: "Beijing", Data: dailyBeijingValues}},
  Coordinate: interactiveradar.CoordinateOptions{
    SplitArea: chart.Bool(true),
    SplitLine: &interactiveradar.SplitLineOptions{Show: chart.Bool(true)},
  },
})`
}

func interactiveRadarStyleCode() string {
	return `Coordinate: interactiveradar.CoordinateOptions{
  Shape: interactiveradar.ShapeCircle,
  SplitNumber: 5,
  SplitLine: &interactiveradar.SplitLineOptions{
    Show: chart.Bool(true),
    Style: &chart.LineStyle{Width: 1, Opacity: chart.Float(.1)},
  },
},
SeriesOptions: chart.SeriesOptions{
  LineStyle: &chart.LineStyle{Width: 1, Opacity: chart.Float(.5)},
  AreaStyle: &chart.AreaStyle{Opacity: chart.Float(.1)},
}`
}

func interactiveRadarLegendCode(mode chart.LegendSelectionMode) string {
	return fmt.Sprintf(`Series: []interactiveradar.Series{beijing, guangzhou, shanghai},
Options: chart.ChartOptions{
  Legend: &chart.LegendOptions{SelectionMode: chart.LegendSelection%s},
}`, map[chart.LegendSelectionMode]string{chart.LegendSelectionMultiple: "Multiple", chart.LegendSelectionSingle: "Single"}[mode])
}
