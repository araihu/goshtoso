package charts

import (
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/scatter"
)

const staticScatterUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"

type staticScatterCoverageEntry struct {
	Path      string
	SHA256    string
	Treatment string
}

func staticScatterUpstreamCoverage() []staticScatterCoverageEntry {
	return []staticScatterCoverageEntry{
		{Path: "examples/1-Painter/scatter_chart-1-basic/main.go", SHA256: "6bd838c49fc38d6b50be1b2c26e1845348de6a5bce3a4a7e637497b78ad61818", Treatment: "Basic categorical scatter with a missing observation"},
		{Path: "examples/1-Painter/scatter_chart-2-symbols/main.go", SHA256: "2667f6f260c63d56dcc22cb036b6b0408ea9da0f943757909d1436be7b9ad515", Treatment: "Per-series circle, diamond, square, and dot symbols"},
		{Path: "examples/1-Painter/scatter_chart-3-dense_data/main.go", SHA256: "0a50b43ccad6a96b3248d3e45e83add46e33b8b6ff98133e1f2597bdd46f49bb", Treatment: "Dense multi-value random walks with trends and maximum references"},
		{Path: "examples/1-Painter/scatter_chart-4-top_n_labels/main.go", SHA256: "cf92798819fbc010f44eaa406acabd337f16b52eec00793a10679e9c3b7cda81", Treatment: "Top-five value labels"},
		{Path: "examples/2-OptionFunc/scatter_chart-1-basic/main.go", SHA256: "a4528b8943edac99ab99f1632d328a34c64013e7551e3c13f61d1aa45844afd1", Treatment: "Basic data with circle symbols and integer formatting"},
	}
}

var staticScatterWeekLabels = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

func staticScatterBasicSeries() []scatter.Series {
	return []scatter.Series{
		{Name: "Email", Values: [][]float64{{120}, {132}, {101}, {}, {90}, {230}, {210}}},
		{Name: "Union Ads", Values: [][]float64{{220}, {182}, {191}, {234}, {290}, {330}, {310}}},
		{Name: "Video Ads", Values: [][]float64{{150}, {232}, {201}, {154}, {190}, {330}, {410}}},
		{Name: "Direct", Values: [][]float64{{320}, {332}, {301}, {334}, {390}, {330}, {320}}},
		{Name: "Search Engine", Values: [][]float64{{820}, {932}, {901}, {934}, {1290}, {1330}, {1320}}},
	}
}

func sampleBasicScatter() scatter.Config {
	return scatter.Config{
		Label:      "Basic scatter chart with one missing Email observation",
		Caption:    "Five series across Monday through Sunday; Email has no Thursday observation.",
		Categories: append([]string(nil), staticScatterWeekLabels...),
		Series:     staticScatterBasicSeries(),
		Options:    scatter.Options{Symbol: scatter.SymbolDot, Size: 4},
		Title:      scatter.TitleOptions{Text: "Scatter", FontSize: 16},
		Legend:     scatter.LegendOptions{Padding: scatter.Padding{Left: 100}},
		Width:      600,
		Height:     400,
		Controls:   chartcontrol.Options{Fullscreen: true},
		Export:     &chartcontrol.ExportOptions{Filename: "basic-scatter-chart", Background: chartcontrol.ExportBackgroundTransparent},
	}
}

func sampleSymbolScatter() scatter.Config {
	values := [][][]float64{
		{{120}, {132}, {101}, {95}, {90}, {230}, {210}},
		{{220}, {182}, {191}, {234}, {290}, {330}, {310}},
		{{150}, {232}, {201}, {154}, {190}, {330}, {410}},
		{{320}, {332}, {301}, {334}, {390}, {330}, {320}},
	}
	names := []string{"Email", "Union Ads", "Video Ads", "Direct"}
	symbols := []scatter.Symbol{scatter.SymbolCircle, scatter.SymbolDiamond, scatter.SymbolSquare, scatter.SymbolDot}
	series := make([]scatter.Series, len(names))
	for index := range series {
		series[index] = scatter.Series{Name: names[index], Values: values[index], Options: scatter.Options{Symbol: symbols[index]}}
	}
	return scatter.Config{
		Label:      "Scatter series with distinct point symbols",
		Caption:    "Circle, diamond, square, and dot markers distinguish the four series without relying on color alone.",
		Categories: append([]string(nil), staticScatterWeekLabels...),
		Series:     series,
		Options:    scatter.Options{Size: 4},
		Width:      600,
		Height:     400,
		Controls:   chartcontrol.Options{Fullscreen: true},
		Export:     &chartcontrol.ExportOptions{Filename: "scatter-series-symbols"},
	}
}

func sampleIntegerScatter() scatter.Config {
	cfg := sampleBasicScatter()
	cfg.Label = "Basic scatter chart with circle symbols and integer labels"
	cfg.Caption = "The same observations use hollow circles and whole-number axis formatting."
	cfg.Options = scatter.Options{Symbol: scatter.SymbolCircle, ValueFormat: scatter.ValueFormatInteger}
	cfg.Title = scatter.TitleOptions{Text: "Scatter", FontSize: 16}
	cfg.Legend = scatter.LegendOptions{Padding: scatter.Padding{Left: 100}}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "scatter-integer-labels"}
	return cfg
}

func basicScatterCode() string {
	return `@scatter.Scatter(scatter.Config{
  Label: "Basic scatter chart with one missing Email observation",
  Categories: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Series: []scatter.Series{
    {Name: "Email", Values: [][]float64{{120}, {132}, {101}, {}, {90}, {230}, {210}}},
    // Four more named series retain the same seven aligned categories.
  },
  Options: scatter.Options{Symbol: scatter.SymbolDot, Size: 4},
  Title: scatter.TitleOptions{Text: "Scatter", FontSize: 16},
  Legend: scatter.LegendOptions{Padding: scatter.Padding{Left: 100}},
  Width: 600, Height: 400,
})`
}

func symbolScatterCode() string {
	return `@scatter.Scatter(scatter.Config{
  Label: "Scatter series with distinct point symbols",
  Categories: week,
  Series: []scatter.Series{
    {Name: "Email", Values: email, Options: scatter.Options{Symbol: scatter.SymbolCircle}},
    {Name: "Union Ads", Values: union, Options: scatter.Options{Symbol: scatter.SymbolDiamond}},
    {Name: "Video Ads", Values: video, Options: scatter.Options{Symbol: scatter.SymbolSquare}},
    {Name: "Direct", Values: direct, Options: scatter.Options{Symbol: scatter.SymbolDot}},
  },
  Options: scatter.Options{Size: 4},
  Width: 600, Height: 400,
})`
}

func integerScatterCode() string {
	return `@scatter.Scatter(scatter.Config{
  Label: "Basic scatter chart with circle symbols and integer labels",
  Categories: week,
  Series: series,
  Options: scatter.Options{
    Symbol: scatter.SymbolCircle,
    ValueFormat: scatter.ValueFormatInteger,
  },
  Width: 600, Height: 400,
})`
}
