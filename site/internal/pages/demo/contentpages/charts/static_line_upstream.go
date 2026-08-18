package charts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/line"
)

var lineWeekLabels = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

func sampleBasicLine() line.Config {
	return line.Config{
		Label:   "Basic line chart with one missing Email observation",
		Caption: "Five series across Monday through Sunday; Email has no Thursday observation.",
		Title:   line.Title{Text: "Line", FontSize: 16},
		Labels:  append([]string(nil), lineWeekLabels...),
		Series: []line.Series{
			{Name: "Email", Points: []line.Point{{Value: 120}, {Value: 132}, {Value: 101}, {Missing: true}, {Value: 90}, {Value: 230}, {Value: 210}}},
			{Name: "Union Ads", Values: []float64{220, 182, 191, 234, 290, 330, 310}},
			{Name: "Video Ads", Values: []float64{150, 232, 201, 154, 190, 330, 410}},
			{Name: "Direct", Values: []float64{320, 332, 301, 334, 390, 330, 320}},
			{Name: "Search Engine", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}},
		},
		Legend: line.LegendOptions{Padding: line.Padding{Left: 100}},
		Symbol: line.Symbol{Shape: line.SymbolCircle}, StrokeWidth: 1.2,
		Width: 600, Height: 400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-line-chart"},
	}
}

func sampleSymbolLine() line.Config {
	values := [][]float64{
		{120, 132, 101, 96, 90, 230, 210},
		{220, 182, 191, 234, 290, 330, 310},
		{150, 232, 201, 154, 190, 330, 410},
		{320, 332, 301, 334, 390, 330, 320},
	}
	names := []string{"Email", "Union Ads", "Video Ads", "Direct"}
	shapes := []line.SymbolShape{line.SymbolCircle, line.SymbolDiamond, line.SymbolSquare, line.SymbolDot}
	series := make([]line.Series, len(names))
	for index := range series {
		series[index] = line.Series{Name: names[index], Values: values[index], Symbol: line.Symbol{Shape: shapes[index]}}
	}
	return line.Config{
		Label: "Line series with distinct point symbols", Caption: "Each series uses a distinct marker while exact values remain available below.",
		Labels: append([]string(nil), lineWeekLabels...), Series: series, StrokeWidth: 1.2,
		Width: 600, Height: 400, Controls: chartcontrol.Options{Fullscreen: true},
		Export: &chartcontrol.ExportOptions{Filename: "line-series-symbols"},
	}
}

func sampleSmoothLine() line.Config {
	cfg := sampleSymbolLine()
	cfg.Label = "Bold smoothed line series"
	cfg.Caption = "Four smoothed series without point markers or a legend."
	cfg.Legend = line.LegendOptions{Hidden: true}
	cfg.Symbol = line.Symbol{Shape: line.SymbolNone}
	cfg.StrokeWidth = 4
	cfg.SmoothingTension = .9
	for index := range cfg.Series {
		cfg.Series[index].Symbol = line.Symbol{}
	}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "smooth-line-series"}
	return cfg
}

func sampleMarkedLine() line.Config {
	references := line.References{Average: true, Maximum: true, Format: line.ValueFormatHumanized, Decimals: 1}
	return line.Config{
		Label:   "Line series with average and maximum references",
		Caption: "Each series shows an average reference line and maximum reference point.",
		Labels:  append([]string(nil), lineWeekLabels...),
		Series: []line.Series{
			{Name: "Email", Values: []float64{120, 132, 101, 95, 90, 230, 210}, References: references},
			{Name: "Direct", Values: []float64{320, 332, 301, 334, 390, 330, 320}, References: references},
			{Name: "Search Engine", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}, References: references},
		},
		Padding: line.Padding{Top: 20, Right: 48, Bottom: 20, Left: 20},
		Symbol:  line.Symbol{Shape: line.SymbolCircle}, StrokeWidth: 1.2,
		Width: 600, Height: 400, Controls: chartcontrol.Options{Fullscreen: true},
		Export: &chartcontrol.ExportOptions{Filename: "line-statistical-references"},
	}
}

func sampleStackedLine() line.Config {
	noGap := false
	labels := line.DataLabelOptions{Show: true, Format: line.ValueFormatHumanized, Decimals: 1, TrailingZeros: true}
	mark := line.References{Maximum: true, PointPrefix: "Max:", Format: line.ValueFormatHumanized, PointSize: 30}
	return line.Config{
		Label: "Stacked line contributions", Caption: "Three layers show both component values and their cumulative sum.",
		Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8"},
		Series: []line.Series{
			{Name: "A", Values: []float64{1.9, 23.2, 25.6, 102.6, 142.2, 32.6, 20, 2.3}, Labels: labels, References: mark},
			{Name: "B", Values: []float64{12, 26.4, 28.7, 144.6, 122.2, 48.7, 18.8, 13.3}, Labels: labels, References: mark},
			{Name: "C", Values: []float64{80, 40.4, 28.4, 48.8, 24.4, 24.2, 40.8, 80.8}, Labels: labels, References: mark},
		},
		Stacked: true, XAxis: line.CategoryAxisOptions{BoundaryGap: &noGap},
		YAxes: []line.Axis{{Title: "A+B+C Sum", TitleFontSize: 12, LabelFontSize: 8}}, Padding: line.Padding{Top: 10, Right: 40, Bottom: 10, Left: 10},
		Width: 600, Height: 400, Controls: chartcontrol.Options{Fullscreen: true},
		Export: &chartcontrol.ExportOptions{Filename: "stacked-line-contributions"},
	}
}

func sampleBoundaryGapLine(enabled bool) line.Config {
	values := [][]float64{
		{120, 132, 101, 90, 230, 210}, {220, 182, 191, 290, 330, 310},
		{150, 232, 201, 190, 330, 410}, {320, 332, 301, 390, 330, 320},
		{820, 932, 901, 1290, 1330, 1320},
	}
	names := []string{"Email", "Union Ads", "Video Ads", "Direct", "Search Engine"}
	series := make([]line.Series, len(names))
	for index := range series {
		series[index] = line.Series{Name: names[index], Values: values[index]}
	}
	title, filename := "Boundary Gap Disabled", "line-boundary-gap-disabled"
	if enabled {
		title, filename = "Boundary Gap", "line-boundary-gap"
	}
	return line.Config{
		Label: title, Caption: "Same six categories and five series for direct spacing comparison.",
		Title: line.Title{Text: title, FontSize: 16}, Labels: []string{"A", "B", "C", "D", "E", "F"}, Series: series,
		Legend: line.LegendOptions{Hidden: true}, XAxis: line.CategoryAxisOptions{BoundaryGap: &enabled},
		Padding: line.Padding{Top: 10, Right: 10, Bottom: 10, Left: 10}, Symbol: line.Symbol{Shape: line.SymbolCircle},
		Width: 600, Height: 400, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: filename},
	}
}

type lensDefinition struct{ name, encoded string }

var lensDefinitions = []lensDefinition{
	{name: "100-500mm f/4.5-7.1", encoded: "100f4.5,151f5,254f5.6,363f6.3,472f7.1,501f-"},
	{name: "70-200mm f/2.8", encoded: "70f2.8,201f-"},
	{name: "70-200mm f/2.8 + 1.4x TC", encoded: "98f4,281f-"},
	{name: "70-200mm f/2.8 + 2x TC", encoded: "140f5.6,401f-"},
}

func sampleCustomLensLine() line.Config {
	labels := make([]string, 451)
	for index := range labels {
		labels[index] = fmt.Sprintf("%dmm", index+60)
	}
	series := make([]line.Series, len(lensDefinitions))
	for index, definition := range lensDefinitions {
		series[index] = line.Series{Name: definition.name, Points: lensPoints(definition.encoded)}
	}
	minimum, maximum := 1.4, 8.0
	return line.Config{
		Label: "Canon RF zoom-lens aperture ranges", Caption: "Four lens configurations across 60–510mm; unavailable focal lengths break each line.",
		Title:  line.Title{Text: "Canon RF Zoom Lenses", Placement: line.PlacementCenter, FontSize: 16},
		Labels: labels, Series: series, Legend: line.LegendOptions{Hidden: true}, Symbol: line.Symbol{Shape: line.SymbolNone}, StrokeWidth: 1.5,
		XAxis:   line.CategoryAxisOptions{BoundaryGap: new(true), Unit: 40, LabelCount: 10, LabelRotation: 45, LabelFontSize: 6},
		YAxes:   []line.Axis{{Hidden: true, Min: &minimum, Max: &maximum, LabelCount: 4, SpineLine: true, LabelFontSize: 8}},
		Padding: line.Padding{Top: 20, Right: 20, Bottom: 10, Left: 20},
		Annotations: []line.TextAnnotation{
			{Text: lensDefinitions[0].name, X: 420, Y: 84, SeriesIndex: 0, FontSize: 12},
			{Text: lensDefinitions[1].name, X: 45, Y: 284, SeriesIndex: 1, FontSize: 12},
			{Text: lensDefinitions[2].name, X: 140, Y: 230, SeriesIndex: 2, FontSize: 12},
			{Text: lensDefinitions[3].name, X: 160, Y: 155, SeriesIndex: 3, FontSize: 12},
			{Text: "f/4.5", X: 42, Y: 220, SeriesIndex: 0, FontSize: 8}, {Text: "f/5.0", X: 105, Y: 196, SeriesIndex: 0, FontSize: 8},
			{Text: "f/6.3", X: 370, Y: 137, SeriesIndex: 0, FontSize: 8}, {Text: "f/7.1", X: 570, Y: 100, SeriesIndex: 0, FontSize: 8},
			{Text: "f/2.8", X: 5, Y: 298, SeriesIndex: 1, FontSize: 8}, {Text: "f/4.0", X: 40, Y: 244, SeriesIndex: 2, FontSize: 8},
			{Text: "f/5.6", X: 92, Y: 168, SeriesIndex: 3, FontSize: 8},
		},
		Width: 600, Height: 400, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "canon-rf-zoom-lenses"},
	}
}

func lensPoints(encoded string) []line.Point {
	parts := strings.Split(encoded, ",")
	points := make([]line.Point, 451)
	current := line.Point{Missing: true}
	next := current
	nextMM := 60
	partIndex := 0
	for index := range points {
		mm := index + 60
		if mm == nextMM {
			current = next
			if partIndex < len(parts) {
				next, nextMM = parseLensPart(parts[partIndex])
				partIndex++
			} else {
				next = line.Point{Missing: true}
			}
		}
		points[index] = current
	}
	return points
}

func parseLensPart(encoded string) (line.Point, int) {
	parts := strings.Split(encoded, "f")
	if len(parts) != 2 {
		panic("invalid pinned lens definition: " + encoded)
	}
	mm, err := strconv.Atoi(parts[0])
	if err != nil {
		panic("invalid pinned lens focal length: " + encoded)
	}
	if parts[1] == "-" {
		return line.Point{Missing: true}, mm
	}
	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		panic("invalid pinned lens aperture: " + encoded)
	}
	return line.Point{Value: value}, mm
}

func sampleGradientLabelLine() line.Config {
	return line.Config{
		Label:   "Sales performance with cold-to-warm value labels",
		Caption: "Theme-aware labels progress from the lowest value through the midpoint to the highest value.",
		Title:   line.Title{Text: "Sales Performance with Gradient Label Colors", Subtext: "Cold = Low Values, Warm = High Values"},
		Labels:  []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct"},
		Series:  []line.Series{{Name: "Sales Performance", Values: []float64{20, 15, 35, 40, 10, 55, 25, 45, 30, 50}, Labels: line.DataLabelOptions{Show: true, ColorScale: line.LabelColorScaleColdToWarm}}},
		Legend:  line.LegendOptions{Hidden: true}, Padding: line.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
		Width: 800, Height: 500, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "sales-performance-gradient-labels"},
	}
}

//go:fix inline
func boolRef(value bool) *bool { return new(value) }

func basicLineCode() string {
	return `@line.Line(line.Config{
  Label: "Basic line chart with one missing Email observation",
  Title: line.Title{Text: "Line", FontSize: 16},
  Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Series: []line.Series{
    {Name: "Email", Points: []line.Point{{Value: 120}, {Value: 132}, {Value: 101}, {Missing: true}, {Value: 90}, {Value: 230}, {Value: 210}}},
    {Name: "Union Ads", Values: []float64{220, 182, 191, 234, 290, 330, 310}},
    {Name: "Video Ads", Values: []float64{150, 232, 201, 154, 190, 330, 410}},
    {Name: "Direct", Values: []float64{320, 332, 301, 334, 390, 330, 320}},
    {Name: "Search Engine", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}},
  },
  Legend: line.LegendOptions{Padding: line.Padding{Left: 100}},
  Symbol: line.Symbol{Shape: line.SymbolCircle}, StrokeWidth: 1.2,
  Width: 600, Height: 400,
})`
}

func symbolLineCode() string {
	return `@line.Line(line.Config{
  Label: "Line series with distinct point symbols",
  Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Series: []line.Series{
    {Name: "Email", Values: email, Symbol: line.Symbol{Shape: line.SymbolCircle}},
    {Name: "Union Ads", Values: unionAds, Symbol: line.Symbol{Shape: line.SymbolDiamond}},
    {Name: "Video Ads", Values: videoAds, Symbol: line.Symbol{Shape: line.SymbolSquare}},
    {Name: "Direct", Values: direct, Symbol: line.Symbol{Shape: line.SymbolDot}},
  },
  StrokeWidth: 1.2,
})`
}

func smoothLineCode() string {
	return `@line.Line(line.Config{
  Label: "Bold smoothed line series",
  Labels: weekLabels,
  Series: series,
  Legend: line.LegendOptions{Hidden: true},
  Symbol: line.Symbol{Shape: line.SymbolNone},
  StrokeWidth: 4,
  SmoothingTension: 0.9,
})`
}

func markedLineCode() string {
	return `references := line.References{
  Average: true, Maximum: true,
  Format: line.ValueFormatHumanized, Decimals: 1,
}
@line.Line(line.Config{
  Label: "Line series with average and maximum references",
  Labels: weekLabels,
  Series: []line.Series{
    {Name: "Email", Values: email, References: references},
    {Name: "Direct", Values: direct, References: references},
    {Name: "Search Engine", Values: searchEngine, References: references},
  },
  Padding: line.Padding{Top: 20, Right: 48, Bottom: 20, Left: 20},
})`
}

func stackedLineCode() string {
	return `noGap := false
labels := line.DataLabelOptions{Show: true, Format: line.ValueFormatHumanized, Decimals: 1, TrailingZeros: true}
maximum := line.References{Maximum: true, PointPrefix: "Max:", PointSize: 30}
@line.Line(line.Config{
  Label: "Stacked line contributions",
  Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8"},
  Series: []line.Series{
    {Name: "A", Values: []float64{1.9, 23.2, 25.6, 102.6, 142.2, 32.6, 20, 2.3}, Labels: labels, References: maximum},
    {Name: "B", Values: []float64{12, 26.4, 28.7, 144.6, 122.2, 48.7, 18.8, 13.3}, Labels: labels, References: maximum},
    {Name: "C", Values: []float64{80, 40.4, 28.4, 48.8, 24.4, 24.2, 40.8, 80.8}, Labels: labels, References: maximum},
  },
  Stacked: true, XAxis: line.CategoryAxisOptions{BoundaryGap: &noGap},
  YAxes: []line.Axis{{Title: "A+B+C Sum", TitleFontSize: 12, LabelFontSize: 8}},
})`
}

func boundaryGapLineCode() string {
	return `withGapValue := true
withoutGapValue := false
withGap := line.Config{
  Label: "Boundary Gap", Title: line.Title{Text: "Boundary Gap"},
  Labels: []string{"A", "B", "C", "D", "E", "F"}, Series: series,
  XAxis: line.CategoryAxisOptions{BoundaryGap: &withGapValue},
}
withoutGap := withGap
withoutGap.Label = "Boundary Gap Disabled"
withoutGap.Title.Text = "Boundary Gap Disabled"
withoutGap.XAxis.BoundaryGap = &withoutGapValue

@line.Line(withGap)
@line.Line(withoutGap)`
}

func customLensLineCode() string {
	return `boundaryGap := true
minimum, maximum := 1.4, 8.0
@line.Line(line.Config{
  Label: "Canon RF zoom-lens aperture ranges",
  Title: line.Title{Text: "Canon RF Zoom Lenses", Placement: line.PlacementCenter, FontSize: 16},
  Labels: focalLengthLabels,
  Series: lensSeries, // Each aligned []line.Point uses Missing for unavailable focal lengths.
  Legend: line.LegendOptions{Hidden: true},
  Symbol: line.Symbol{Shape: line.SymbolNone}, StrokeWidth: 1.5,
  XAxis: line.CategoryAxisOptions{BoundaryGap: &boundaryGap, Unit: 40, LabelCount: 10, LabelRotation: 45, LabelFontSize: 6},
  YAxes: []line.Axis{{Hidden: true, Min: &minimum, Max: &maximum, LabelCount: 4, SpineLine: true, LabelFontSize: 8}},
  Annotations: []line.TextAnnotation{
    {Text: "100-500mm f/4.5-7.1", X: 420, Y: 84, SeriesIndex: 0, FontSize: 12},
    {Text: "f/4.5", X: 42, Y: 220, SeriesIndex: 0, FontSize: 8},
  },
  Width: 600, Height: 400,
})`
}

func gradientLabelLineCode() string {
	return `@line.Line(line.Config{
  Label: "Sales performance with cold-to-warm value labels",
  Title: line.Title{Text: "Sales Performance with Gradient Label Colors", Subtext: "Cold = Low Values, Warm = High Values"},
  Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct"},
  Series: []line.Series{{
    Name: "Sales Performance",
    Values: []float64{20, 15, 35, 40, 10, 55, 25, 45, 30, 50},
    Labels: line.DataLabelOptions{Show: true, ColorScale: line.LabelColorScaleColdToWarm},
  }},
  Legend: line.LegendOptions{Hidden: true},
  Width: 800, Height: 500,
})`
}
