package charts

import (
	"math/rand"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

func interactiveBarCategories() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
}

func fixedInteractiveBarData(seed int64) []interactivebar.Data {
	source := rand.New(rand.NewSource(seed))
	data := make([]interactivebar.Data, len(interactiveBarCategories()))
	for index := range data {
		data[index].Value = float64(source.Intn(300))
	}
	return data
}

func interactiveBarSeries(name string, seed int64) interactivebar.Series {
	return interactivebar.Series{Name: name, Data: fixedInteractiveBarData(seed)}
}

func controlledInteractiveBarOptions(title, filename string) chart.ChartOptions {
	return chart.ChartOptions{
		Title:    &chart.TitleOptions{Text: title},
		Legend:   &chart.LegendOptions{Bottom: "0"},
		Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "axis"},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: filename},
	}
}

func sampleInteractiveBar() interactivebar.Config {
	options := controlledInteractiveBarOptions("basic bar example", "basic-bar-example")
	options.Title.Subtitle = "This is the subtitle."
	return interactivebar.Config{
		Label: "Basic bar example", Caption: "Two deterministic seven-day series preserve the upstream categorical shape.",
		XAxis: interactiveBarCategories(),
		Series: []interactivebar.Series{
			interactiveBarSeries("Category A", 11),
			interactiveBarSeries("Category B", 12),
		},
		Options: options,
	}
}

func sampleInteractiveBarLabels() interactivebar.Config {
	return interactivebar.Config{
		Label: "Visible value labels", Caption: "Every bar exposes its exact value above the mark and in the adjacent table.",
		XAxis: interactiveBarCategories(),
		Series: []interactivebar.Series{
			interactiveBarSeries("Category A", 21),
			interactiveBarSeries("Category B", 22),
		},
		SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true), Position: "top"}},
		Options:       controlledInteractiveBarOptions("label options", "visible-bar-labels"),
	}
}

func sampleInteractiveBarAxes() interactivebar.Config {
	options := controlledInteractiveBarOptions("axis names, units, and split lines", "bar-axis-options")
	options.XAxis = &chart.AxisOptions{Name: "XAxisName", LabelSuffix: " x-unit", ShowSplitLine: new(true)}
	options.YAxis = &chart.AxisOptions{Name: "YAxisName", LabelSuffix: " y-unit", ShowSplitLine: new(true)}
	return interactivebar.Config{
		Label: "Named axes with literal units", Caption: "Axis names, unit suffixes, and split lines clarify how categories and values are read.",
		XAxis:   interactiveBarCategories(),
		Series:  []interactivebar.Series{interactiveBarSeries("Category A", 31), interactiveBarSeries("Category B", 32)},
		Options: options,
	}
}

func sampleInteractiveBarColors() interactivebar.Config {
	return interactivebar.Config{
		Label: "Explicit series colors", Caption: "Caller-selected colors override the theme palette for both series.",
		XAxis:   interactiveBarCategories(),
		Series:  []interactivebar.Series{interactiveBarSeries("Category A", 41), interactiveBarSeries("Category B", 42)},
		Style:   charttheme.Style{Colors: []string{"#2563eb", "#db2777"}},
		Options: controlledInteractiveBarOptions("user-defined colors", "explicit-bar-colors"),
	}
}

func sampleInteractiveBarWidthsAndGap() interactivebar.Config {
	return interactivebar.Config{
		Label: "Bar widths and gap", Caption: "One absolute width, one percentage width, and a 150% inter-series gap preserve the upstream size treatments.",
		XAxis: interactiveBarCategories(),
		Series: []interactivebar.Series{
			{Name: "Category A", Data: fixedInteractiveBarData(51), Options: chart.SeriesOptions{BarWidth: "35"}},
			{Name: "Category B", Data: fixedInteractiveBarData(52), Options: chart.SeriesOptions{BarWidth: "15%"}},
		},
		SeriesOptions: chart.SeriesOptions{BarGap: "150%"},
		Options:       controlledInteractiveBarOptions("bar width and gap", "bar-width-and-gap"),
	}
}

func sampleInteractiveBarHorizontal() interactivebar.Config {
	return interactivebar.Config{
		Label: "Horizontal bar orientation", Caption: "Categories move to the vertical axis for easier comparison of long labels.",
		XAxis: interactiveBarCategories(), Orientation: interactivebar.OrientationHorizontal,
		Series:  []interactivebar.Series{interactiveBarSeries("Category A", 61), interactiveBarSeries("Category B", 62)},
		Options: controlledInteractiveBarOptions("reverse category and value axes", "horizontal-bar"),
	}
}

func sampleInteractiveBarStacked() interactivebar.Config {
	return interactivebar.Config{
		Label: "Stacked bar series", Caption: "Both series share one stack so each category shows a combined total.",
		XAxis:         interactiveBarCategories(),
		Series:        []interactivebar.Series{interactiveBarSeries("Category A", 71), interactiveBarSeries("Category B", 72)},
		SeriesOptions: chart.SeriesOptions{Stack: "stackA"},
		Options:       controlledInteractiveBarOptions("stack style", "stacked-bar"),
	}
}

func sampleInteractiveBarZoom(mode interactivebar.ZoomMode) interactivebar.Config {
	label, title, filename := "Inside category zoom", "category zoom (inside)", "inside-bar-zoom"
	if mode == interactivebar.ZoomSlider {
		label, title, filename = "Slider category zoom", "category zoom (slider)", "slider-bar-zoom"
	}
	return interactivebar.Config{
		Label: label, Caption: "The initial window shows 10% through 50% of the seven ordered categories.",
		XAxis:  interactiveBarCategories(),
		Series: []interactivebar.Series{interactiveBarSeries("Category A", 81), interactiveBarSeries("Category B", 82)},
		Zoom:   &interactivebar.Zoom{Mode: mode, StartPercent: 10, EndPercent: 50},
		Height: "460px", Options: controlledInteractiveBarOptions(title, filename),
	}
}

func sampleInteractiveBarMarkPoints() interactivebar.Config {
	categoryA := fixedInteractiveBarData(91)
	categoryA[0].Value = 100
	calculated := []interactivebar.PointReference{
		{Name: "Maximum", Statistic: interactivebar.StatisticMaximum},
		{Name: "Minimum", Statistic: interactivebar.StatisticMinimum},
	}
	return interactivebar.Config{
		Label: "Bar point references", Caption: "A named Monday point sits beside calculated minimum and maximum markers.",
		XAxis: interactiveBarCategories(),
		Series: []interactivebar.Series{
			{Name: "Category A", Data: categoryA, References: interactivebar.References{Points: append([]interactivebar.PointReference{{Name: "special mark", Coordinate: &interactivebar.Coordinate{Category: "Mon", Value: 100}, Label: &chart.LabelOptions{Show: new(true), Position: "inside"}}}, calculated...), ShowLabels: new(true)}},
			{Name: "Category B", Data: fixedInteractiveBarData(92), References: interactivebar.References{Points: calculated, ShowLabels: new(true)}},
		},
		Options: controlledInteractiveBarOptions("mark point options", "bar-point-references"),
	}
}

func sampleInteractiveBarMarkLines() interactivebar.Config {
	guides := []interactivebar.GuideReference{
		{Name: "Maximum", Statistic: interactivebar.StatisticMaximum},
		{Name: "Average", Statistic: interactivebar.StatisticAverage},
	}
	return interactivebar.Config{
		Label: "Bar guide references", Caption: "Maximum and average guides summarize both seven-value series.",
		XAxis: interactiveBarCategories(),
		Series: []interactivebar.Series{
			{Name: "Category A", Data: fixedInteractiveBarData(101), References: interactivebar.References{Lines: guides}},
			{Name: "Category B", Data: fixedInteractiveBarData(102), References: interactivebar.References{Lines: guides}},
		},
		Options: controlledInteractiveBarOptions("mark line options", "bar-guide-references"),
	}
}

func sampleInteractiveBarLargeCanvas() interactivebar.Config {
	return interactivebar.Config{
		Label: "Large bar canvas", Caption: "The upstream 1200 by 600 canvas becomes container-wide while preserving the 600-pixel height.",
		XAxis:  interactiveBarCategories(),
		Series: []interactivebar.Series{interactiveBarSeries("Category A", 111), interactiveBarSeries("Category B", 112)},
		Width:  "100%", Height: "600px",
		Options: controlledInteractiveBarOptions("large canvas size", "large-bar-canvas"),
	}
}

func interactiveChartBarCode() string {
	return `import (
  "github.com/araihu/goshtoso-charts/components/chart"
  interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

@interactivebar.Bar(interactivebar.Config{
  Label: "Basic bar example",
  XAxis: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Series: []interactivebar.Series{
    {Name: "Category A", Data: categoryA},
    {Name: "Category B", Data: categoryB},
  },
  Options: chart.ChartOptions{
    Title: &chart.TitleOptions{Text: "basic bar example", Subtitle: "This is the subtitle."},
    Tooltip: &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "axis"},
  },
})`
}

func interactiveBarAxesCode() string {
	return `Options: chart.ChartOptions{
  XAxis: &chart.AxisOptions{Name: "XAxisName", LabelSuffix: " x-unit", ShowSplitLine: chart.Bool(true)},
  YAxis: &chart.AxisOptions{Name: "YAxisName", LabelSuffix: " y-unit", ShowSplitLine: chart.Bool(true)},
}`
}

func interactiveBarLayoutCode() string {
	return `Orientation: interactivebar.OrientationHorizontal,
SeriesOptions: chart.SeriesOptions{Stack: "stackA", BarGap: "150%"},
Series: []interactivebar.Series{
  {Name: "Category A", Data: categoryA, Options: chart.SeriesOptions{BarWidth: "35"}},
  {Name: "Category B", Data: categoryB, Options: chart.SeriesOptions{BarWidth: "15%"}},
}`
}

func interactiveBarZoomCode() string {
	return `Zoom: &interactivebar.Zoom{
  Mode: interactivebar.ZoomSlider,
  StartPercent: 10,
  EndPercent: 50,
}`
}

func interactiveBarReferencesCode() string {
	return `References: interactivebar.References{
  Points: []interactivebar.PointReference{
    {Name: "Maximum", Statistic: interactivebar.StatisticMaximum},
    {Name: "special mark", Coordinate: &interactivebar.Coordinate{Category: "Mon", Value: 100}},
  },
  Lines: []interactivebar.GuideReference{{Name: "Average", Statistic: interactivebar.StatisticAverage}},
  ShowLabels: chart.Bool(true),
}`
}
