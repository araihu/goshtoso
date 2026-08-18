package charts

import (
	"math/rand"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

const (
	interactiveBarUpstreamPath     = "examples/bar.go"
	interactiveBarUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveBarUpstreamSHA256   = "dcda545f978fdd055ecff5a6050b2ad9dc8cf9fe350bd7e4768952e8068fc9f9"

	barCoverageExample     = "example"
	barCoverageUnsupported = "unsupported"
)

type interactiveBarCoverageEntry struct {
	Name      string
	Status    string
	Treatment string
	Reason    string
}

func interactiveBarUpstreamCoverage() []interactiveBarCoverageEntry {
	return []interactiveBarCoverageEntry{
		{Name: "barBasic", Status: barCoverageExample, Treatment: "basic presentation"},
		{Name: "barTitle", Status: barCoverageExample, Treatment: "basic presentation"},
		{Name: "barTooltip", Status: barCoverageExample, Treatment: "basic presentation"},
		{Name: "barSetToolbox", Status: barCoverageExample, Treatment: "shared controls, PNG export, and exact-value disclosure"},
		{Name: "barShowLabel", Status: barCoverageExample, Treatment: "visible value labels"},
		{Name: "barXYName", Status: barCoverageExample, Treatment: "axis names and units"},
		{Name: "barXYFormatter", Status: barCoverageExample, Treatment: "literal axis units"},
		{Name: "barColor", Status: barCoverageExample, Treatment: "explicit color override"},
		{Name: "barSplitLine", Status: barCoverageExample, Treatment: "axis names and units"},
		{Name: "barGap", Status: barCoverageExample, Treatment: "bar width and gap"},
		{Name: "barDataZoomInside", Status: barCoverageExample, Treatment: "inside category zoom"},
		{Name: "barDataZoomSlider", Status: barCoverageExample, Treatment: "slider category zoom"},
		{Name: "barReverse", Status: barCoverageExample, Treatment: "horizontal orientation"},
		{Name: "barStack", Status: barCoverageExample, Treatment: "stacked series"},
		{Name: "barMarkPoints", Status: barCoverageExample, Treatment: "point references"},
		{Name: "barMarkLines", Status: barCoverageExample, Treatment: "guide references"},
		{Name: "barOverlap", Status: barCoverageUnsupported, Reason: "mixed Bar, Line, and Scatter composition requires a renderer-neutral composite chart API"},
		{Name: "barSize", Status: barCoverageExample, Treatment: "large responsive canvas"},
		{Name: "barWidth", Status: barCoverageExample, Treatment: "bar width and gap"},
	}
}

type interactiveBarSource struct {
	Path   string
	SHA256 string
	Scope  string
}

func interactiveBarSupplementarySources() []interactiveBarSource {
	return []interactiveBarSource{
		{Path: "examples/page_center_layout.go", SHA256: "106456904719dfacfb13adcc1b9e66df83cf28a5a801539bad4d1958554166c9", Scope: "page layout reference"},
		{Path: "examples/page_flex_layout.go", SHA256: "3113b7bdf78a2365ae62502fe86ab001f3ff3034b1d77752c693e95b28a0fd68", Scope: "page layout reference"},
		{Path: "examples/page_none_layout.go", SHA256: "ce38424de2ffeb919661e536c7f44921de098ae14643d4f2975d8e72296c32f8", Scope: "page layout reference"},
		{Path: "examples/themes.go", SHA256: "843c478c63b9cf3ab13b1e13518ea98912332bb34caf0dae5d48343fabd121a0", Scope: "site theme and chart-token reference"},
		{Path: "examples/renderer.go", SHA256: "c4956db261f554c6a161c0d25baa7dbd7c2c179523997d297020cd55916e6a3f", Scope: "private renderer integration; not a chart option"},
		{Path: "examples/bar3d.go", SHA256: "110b3b85f2528d76eb8271b64f1facd81a974e30ecc0dd77319d5a409ff64275", Scope: "separate existing Bar 3D component"},
	}
}

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
