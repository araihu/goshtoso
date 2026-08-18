package charts

import (
	"github.com/araihu/goshtoso-charts/components/chart"
	interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"
)

const (
	interactiveLineUpstreamPath     = "examples/line.go"
	interactiveLineUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveLineUpstreamSHA256   = "1f36444bd373eafde876af19746d6b0115a776fd7c019e5996bdf2d00ecd7b1c"

	lineCoverageExample     = "example"
	lineCoverageUnsupported = "unsupported"
)

type interactiveLineCoverageEntry struct {
	Name   string
	Status string
	Reason string
}

func interactiveLineUpstreamCoverage() []interactiveLineCoverageEntry {
	return []interactiveLineCoverageEntry{
		{Name: "lineBase", Status: lineCoverageExample},
		{Name: "lineShowLabel", Status: lineCoverageExample},
		{Name: "lineMarkPoint", Status: lineCoverageExample},
		{Name: "lineSplitLine", Status: lineCoverageExample},
		{Name: "lineNumerical", Status: lineCoverageExample},
		{Name: "lineTime", Status: lineCoverageExample},
		{Name: "lineStep", Status: lineCoverageExample},
		{Name: "lineSmooth", Status: lineCoverageExample},
		{Name: "lineArea", Status: lineCoverageExample},
		{Name: "lineSmoothArea", Status: lineCoverageExample},
		{Name: "lineOverlap", Status: lineCoverageUnsupported, Reason: "mixed-series composition requires a renderer-neutral composite chart API"},
		{Name: "lineMulti", Status: lineCoverageExample},
		{Name: "lineDemo", Status: lineCoverageExample},
		{Name: "lineSymbols", Status: lineCoverageExample},
	}
}

func interactiveLineCategories() []string {
	// Upstream includes a trailing space in Peach. Correct the obvious label typo.
	return []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"}
}

func fixedInteractiveLineData(values ...float64) []interactiveline.Data {
	data := make([]interactiveline.Data, len(values))
	for index, value := range values {
		data[index] = interactiveline.Data{Value: value}
	}
	return data
}

func controlledInteractiveLineOptions(title, filename string) chart.ChartOptions {
	options := controlledOptions(title, filename)
	options.Legend = &chart.LegendOptions{Bottom: "0"}
	return options
}

func sampleInteractiveLineBase() interactiveline.Config {
	return interactiveline.Config{
		Label: "Basic line example", Caption: "Six deterministic category values preserve the upstream shape.",
		XAxis:   interactiveLineCategories(),
		Series:  []interactiveline.Series{{Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230)}},
		Options: controlledInteractiveLineOptions("basic line example", "basic-line-example"),
	}
}

func sampleInteractiveLineLabels() interactiveline.Config {
	return interactiveline.Config{
		Label: "Visible point labels", Caption: "Every point exposes its value directly.",
		XAxis:         interactiveLineCategories(),
		Series:        []interactiveline.Series{{Name: "Category A", Data: fixedInteractiveLineData(150, 232, 201, 154, 190, 130)}},
		SeriesOptions: chart.SeriesOptions{ShowSymbol: new(true), Label: &chart.LabelOptions{Show: new(true)}},
		Options:       controlledInteractiveLineOptions("title and label options", "visible-point-labels"),
	}
}

func sampleInteractiveLineSymbols() interactiveline.Config {
	return interactiveline.Config{
		Label: "Diamond symbols", Caption: "Two smoothed series use visible diamond points.",
		XAxis: interactiveLineCategories(),
		Series: []interactiveline.Series{
			{Name: "Category A", Data: fixedInteractiveLineData(120, 182, 191, 234, 290, 330)},
			{Name: "Category B", Data: fixedInteractiveLineData(220, 282, 291, 334, 390, 430)},
		},
		SeriesOptions: chart.SeriesOptions{Smooth: new(true), ShowSymbol: new(true), Symbol: "diamond", SymbolSize: 15},
		Options:       controlledInteractiveLineOptions("symbol options", "diamond-symbols"),
	}
}

func sampleInteractiveLineMarkPoints() interactiveline.Config {
	return interactiveline.Config{
		Label: "Calculated point references", Caption: "Maximum, average, and minimum points are named beside exact values.",
		XAxis: interactiveLineCategories(),
		Series: []interactiveline.Series{{
			Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230),
			References: interactiveline.References{
				Points: []interactiveline.PointReference{
					{Name: "Maximum", Statistic: interactiveline.StatisticMaximum},
					{Name: "Average", Statistic: interactiveline.StatisticAverage},
					{Name: "Minimum", Statistic: interactiveline.StatisticMinimum},
				},
				ShowLabels: new(true),
			},
		}},
		Options: controlledInteractiveLineOptions("mark point options", "calculated-point-references"),
	}
}

func sampleInteractiveLineSplitLines() interactiveline.Config {
	options := controlledInteractiveLineOptions("split line options", "split-line-options")
	options.YAxis = &chart.AxisOptions{ShowSplitLine: new(true)}
	return interactiveline.Config{
		Label: "Visible split lines", Caption: "Horizontal guides support scanning across labeled points.",
		XAxis:   interactiveLineCategories(),
		Series:  []interactiveline.Series{{Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230), Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}}}},
		Options: options,
	}
}

func sampleInteractiveLineNumerical() interactiveline.Config {
	xValues := make([]float64, 30)
	for index := range xValues {
		xValues[index] = float64(index)
	}
	return interactiveline.Config{
		Label: "Numerical x axis with guides", Caption: "Thirty ordered coordinates include two scale pieces, a highlighted range, and two guide lines.",
		ValueAxis: &interactiveline.ValueAxis{Values: xValues},
		Series: []interactiveline.Series{{
			Name: "Category A", Data: deterministicLineTimeData(30),
			Options: chart.SeriesOptions{Symbol: "triangle", SymbolSize: 10, AreaStyle: &chart.AreaStyle{}},
			References: interactiveline.References{
				Areas: []interactiveline.RangeReference{{Name: "Danger zone", StartX: 20, EndX: 25}},
				Lines: []interactiveline.GuideReference{
					{Name: "Danger level", Start: &interactiveline.Coordinate{X: 20, Y: 10}, End: &interactiveline.Coordinate{X: 25, Y: 50}},
					{Name: "Line of no return", X: chart.Float(28)},
				},
				ShowLabels: new(true), StartSymbol: "square", EndSymbol: "circle", SymbolSize: 10,
			},
		}},
		VisualScale: &interactiveline.VisualScale{Dimension: interactiveline.VisualDimensionX, Pieces: []interactiveline.VisualPiece{
			{GreaterThan: chart.Float(1), LessThan: chart.Float(7)},
			{GreaterThan: chart.Float(10), LessThan: chart.Float(15)},
		}},
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: "numerical X axis and accessories", Subtitle: "theme-aware ranges, guides, and scale pieces"},
			Legend:   &chart.LegendOptions{Bottom: "0"},
			YAxis:    &chart.AxisOptions{Max: chart.Float(200)},
			Controls: controlledInteractiveLineOptions("", "numerical-line").Controls,
			Export:   controlledInteractiveLineOptions("", "numerical-line").Export,
		},
	}
}

func sampleInteractiveLineStep() interactiveline.Config {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Step line", "The line changes after each category boundary."
	cfg.SeriesOptions.Step = "end"
	cfg.Options = controlledInteractiveLineOptions("step style", "step-line")
	return cfg
}

func sampleInteractiveLineSmooth() interactiveline.Config {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Smooth line", "A curved treatment emphasizes overall direction."
	cfg.SeriesOptions.Smooth = new(true)
	cfg.Options = controlledInteractiveLineOptions("smooth style", "smooth-line")
	return cfg
}

func sampleInteractiveLineArea() interactiveline.Config {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Area line", "A half-opacity area emphasizes magnitude; the marked interval identifies in-stock categories."
	cfg.SeriesOptions.Label = &chart.LabelOptions{Show: new(true)}
	cfg.SeriesOptions.AreaStyle = &chart.AreaStyle{Opacity: new(0.5)}
	cfg.Series[0].References = interactiveline.References{Areas: []interactiveline.RangeReference{{Name: "In stock", StartX: 2, EndX: 4}}, ShowLabels: new(true)}
	cfg.Options = controlledInteractiveLineOptions("area options", "area-line")
	return cfg
}

func sampleInteractiveLineSmoothArea() interactiveline.Config {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Smooth area", "A light area fill sits below the smoothed line."
	cfg.SeriesOptions.Label = &chart.LabelOptions{Show: new(true)}
	cfg.SeriesOptions.AreaStyle = &chart.AreaStyle{Opacity: new(0.2)}
	cfg.SeriesOptions.Smooth = new(true)
	cfg.Options = controlledInteractiveLineOptions("smooth area", "smooth-area-line")
	return cfg
}

func sampleInteractiveLineMulti() interactiveline.Config {
	return interactiveline.Config{
		Label: "Four line comparison", Caption: "Four aligned series preserve the upstream multi-line shape.",
		XAxis: interactiveLineCategories(),
		Series: []interactiveline.Series{
			{Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230)},
			{Name: "Category B", Data: fixedInteractiveLineData(220, 182, 191, 234, 290, 330)},
			{Name: "Category C", Data: fixedInteractiveLineData(150, 232, 201, 154, 190, 130)},
			{Name: "Category D", Data: fixedInteractiveLineData(320, 332, 301, 334, 390, 430)},
		},
		Options: controlledInteractiveLineOptions("multi lines", "multi-lines"),
	}
}

func sampleInteractiveLineDemo() interactiveline.Config {
	options := controlledInteractiveLineOptions("Search Time: Hash table vs Binary search", "search-time-comparison")
	options.XAxis = &chart.AxisOptions{Name: "Elements"}
	options.YAxis = &chart.AxisOptions{Name: "Cost time (ns)", ShowSplitLine: new(true)}
	return interactiveline.Config{
		Label: "Search time comparison", Caption: "Two labeled series compare cost across increasing element counts.",
		XAxis: []string{"10e1", "10e2", "10e3", "10e4", "10e5", "10e6", "10e7"},
		Series: []interactiveline.Series{
			{Name: "map", Data: fixedInteractiveLineData(19, 31, 43, 57, 72, 118, 127), Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true), Position: "bottom"}, Smooth: new(true)}, References: interactiveline.References{Lines: []interactiveline.GuideReference{{Name: "Average", Statistic: interactiveline.StatisticAverage}}, ShowLabels: new(true)}},
			{Name: "slice", Data: fixedInteractiveLineData(24.9, 34.9, 48.1, 58.3, 69.7, 123, 131), Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true), Position: "top"}, Smooth: new(true)}, References: interactiveline.References{Lines: []interactiveline.GuideReference{{Name: "Average", Statistic: interactiveline.StatisticAverage}}, ShowLabels: new(true)}},
		},
		Options: options,
	}
}

func interactiveLineLabelsCode() string {
	return `@interactiveline.Line(interactiveline.Config{
  Label: "Visible point labels",
  XAxis: []string{"Apple", "Banana", "Peach"},
  Series: []interactiveline.Series{{
    Name: "Category A",
    Data: []interactiveline.Data{{Value: 150}, {Value: 232}, {Value: 201}},
  }},
  SeriesOptions: chart.SeriesOptions{
    Label: &chart.LabelOptions{Show: chart.Bool(true)},
    ShowSymbol: chart.Bool(true), Symbol: "diamond", SymbolSize: 15,
  },
})`
}

func interactiveLineReferencesCode() string {
	return `References: interactiveline.References{
  Points: []interactiveline.PointReference{
    {Name: "Maximum", Statistic: interactiveline.StatisticMaximum},
    {Name: "Average", Statistic: interactiveline.StatisticAverage},
    {Name: "Minimum", Statistic: interactiveline.StatisticMinimum},
  },
  ShowLabels: chart.Bool(true),
}`
}

func interactiveLineNumericalCode() string {
	return `@interactiveline.Line(interactiveline.Config{
  Label: "Numerical x axis with guides",
  ValueAxis: &interactiveline.ValueAxis{Values: []float64{0, 1, 2}},
  Series: []interactiveline.Series{{
    Name: "Category A",
    Data: []interactiveline.Data{{Value: 107}, {Value: 112}, {Value: 118}},
    References: interactiveline.References{
      Areas: []interactiveline.RangeReference{{Name: "Danger zone", StartX: 1, EndX: 2}},
    },
  }},
  VisualScale: &interactiveline.VisualScale{
    Dimension: interactiveline.VisualDimensionX,
    Pieces: []interactiveline.VisualPiece{{GreaterThan: chart.Float(0), LessThan: chart.Float(2)}},
  },
})`
}

func interactiveLineShapesCode() string {
	return `SeriesOptions: chart.SeriesOptions{
  Step: "end",
  Smooth: chart.Bool(true),
  AreaStyle: &chart.AreaStyle{Opacity: chart.Float(0.2)},
}`
}

func interactiveLineMultiCode() string {
	return `Series: []interactiveline.Series{
  {Name: "Category A", Data: categoryA},
  {Name: "Category B", Data: categoryB},
  {Name: "Category C", Data: categoryC},
  {Name: "Category D", Data: categoryD},
}`
}
