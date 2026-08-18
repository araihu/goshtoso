package charts

import (
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

const (
	interactiveScatterUpstreamRevision     = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveScatterUpstreamPath         = "examples/scatter.go"
	interactiveScatterUpstreamSHA256       = "a77ddbf7580210a842a3e1d3966ab62c3f229fdb1a33df8f319ef029bd4188b5"
	interactiveEffectScatterUpstreamPath   = "examples/effectscatter.go"
	interactiveEffectScatterUpstreamSHA256 = "1bf49dc5fb02b248ff6794aa549836b4c8fa02ddb89be6adc0c4574327673f1a"
)

type interactiveScatterCoverageEntry struct {
	Name      string
	Treatment string
}

func interactiveScatterUpstreamCoverage() []interactiveScatterCoverageEntry {
	return []interactiveScatterCoverageEntry{
		{Name: "scatterBase", Treatment: "basic two-series categorical scatter with round-rectangle symbols"},
		{Name: "scatterShowLabel", Treatment: "visible right-positioned point labels"},
		{Name: "scatterSplitLine", Treatment: "named axes with visible split lines"},
		{Name: "esBase", Treatment: "basic animated effect variant"},
		{Name: "esEffectStyle", Treatment: "per-series period, scale, and stroke/fill ripple styles"},
	}
}

type interactiveScatterSourceFunction struct {
	Path   string
	Name   string
	SHA256 string
	Role   string
}

// interactiveScatterSourceFunctions inventories every function and method in
// the two authoritative files. Hashes cover exact source spans from the func
// declaration through its closing brace at the pinned revision.
func interactiveScatterSourceFunctions() []interactiveScatterSourceFunction {
	return []interactiveScatterSourceFunction{
		{Path: interactiveScatterUpstreamPath, Name: "generateScatterItems", SHA256: "2cfe0abcb152c7020f5da65f8e22e616e665263a8e31918edc636499e08d4bb6", Role: "deterministic data generator adaptation"},
		{Path: interactiveScatterUpstreamPath, Name: "scatterBase", SHA256: "08faff249e4c7eaa65b602662f01896f4d745fda2c131526b3ac01b5354723b5", Role: "example"},
		{Path: interactiveScatterUpstreamPath, Name: "scatterShowLabel", SHA256: "99ead467aac7ae752e29e2912a3a7f3657c62fc05e0e0685074e1f2db5bc3623", Role: "example"},
		{Path: interactiveScatterUpstreamPath, Name: "scatterSplitLine", SHA256: "55d2f4d9d6e87a356894fcd71320c94a2ccc40abe4ba616587a759aa9f53acb7", Role: "example"},
		{Path: interactiveScatterUpstreamPath, Name: "ScatterExamples.Examples", SHA256: "39e747d19ef16f9ac2d20191242e80bf4bcb278d3db3c8f08f3e8891ad231ac9", Role: "page composition only"},
		{Path: interactiveEffectScatterUpstreamPath, Name: "generateEffectScatterItems", SHA256: "0f295b3eef4924158ea1b39bc0fc60ecc3b86afd61cf694943e9ebca4899a399", Role: "deterministic data generator adaptation"},
		{Path: interactiveEffectScatterUpstreamPath, Name: "esBase", SHA256: "c2cfce12547c08c27942e5aa51df2fd6a332a9ec9e9db092bdbbe60a6ba3bf69", Role: "example"},
		{Path: interactiveEffectScatterUpstreamPath, Name: "esEffectStyle", SHA256: "0fa720ed3f610334a6bf0cb8c4bcf8f33c79167f7f9033b2ced0996751316888", Role: "example"},
		{Path: interactiveEffectScatterUpstreamPath, Name: "EffectscatterExamples.Examples", SHA256: "844e71f0cb14680df92458b561662a41e4cb2ed3cbb83fdbdca9d3c9c376c19d", Role: "page composition only"},
	}
}

type interactiveScatterSource struct {
	Path   string
	SHA256 string
	Scope  string
}

func interactiveScatterSupplementarySources() []interactiveScatterSource {
	return []interactiveScatterSource{
		{Path: "examples/page_center_layout.go", SHA256: "106456904719dfacfb13adcc1b9e66df83cf28a5a801539bad4d1958554166c9", Scope: "centered page-layout reference"},
		{Path: "examples/page_flex_layout.go", SHA256: "3113b7bdf78a2365ae62502fe86ab001f3ff3034b1d77752c693e95b28a0fd68", Scope: "flex page-layout reference"},
		{Path: "examples/page_none_layout.go", SHA256: "ce38424de2ffeb919661e536c7f44921de098ae14643d4f2975d8e72296c32f8", Scope: "unmanaged page-layout reference"},
	}
}

var (
	interactiveScatterSports  = []string{"Swimming", "Surfing", "Shooting", "Skating", "Wrestling", "Diving"}
	interactiveScatterPlayers = []string{"Kobe", "Jordan", "Iverson", "LeBron", "Wade", "McGrady"}
)

func interactiveScatterData(values [6]float64, symbols bool) []interactivescatter.Data {
	data := make([]interactivescatter.Data, len(values))
	for index, value := range values {
		data[index] = interactivescatter.Data{Value: value}
		if symbols {
			data[index].Symbol = "roundRect"
			data[index].SymbolSize = 20
			data[index].SymbolRotate = 10
		}
	}
	return data
}

func interactiveScatterOptions(title, filename string) chart.ChartOptions {
	return chart.ChartOptions{
		Title:    &chart.TitleOptions{Text: title},
		Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "item"},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: filename},
	}
}

// Fixed values record a local seed-1 sequence in original upstream call order.
// Categories, series names, symbol geometry, and the [0,100) domain remain exact;
// the upstream trailing space in "Shooting " is corrected.
func sampleInteractiveScatter() interactivescatter.Config {
	return interactivescatter.Config{
		Label: "Basic sports scatter", Caption: "Six sports across Category A and Category B; exact scores follow the chart.",
		XAxis: interactiveScatterSports,
		Series: []interactivescatter.Series{
			{Name: "Category A", Data: interactiveScatterData([6]float64{81, 87, 47, 59, 81, 18}, true)},
			{Name: "Category B", Data: interactiveScatterData([6]float64{25, 40, 56, 0, 94, 11}, true)},
		},
		Width: "100%", Height: "420px", Options: interactiveScatterOptions("basic scatter example", "basic-sports-scatter"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractiveScatterLabels() interactivescatter.Config {
	config := interactivescatter.Config{
		Label: "Sports scatter with labels", Caption: "Visible labels repeat point values to the right of each symbol.", XAxis: interactiveScatterSports,
		Series: []interactivescatter.Series{
			{Name: "Category A", Data: interactiveScatterData([6]float64{62, 89, 28, 74, 11, 45}, true)},
			{Name: "Category B", Data: interactiveScatterData([6]float64{37, 6, 95, 66, 28, 58}, true)},
		},
		Width: "100%", Height: "420px", Options: interactiveScatterOptions("label options", "labeled-sports-scatter"),
		SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true), Position: "right"}},
		Style:         charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
	return config
}

func sampleInteractiveScatterSplitLines() interactivescatter.Config {
	options := interactiveScatterOptions("splitline options", "split-line-sports-scatter")
	options.XAxis = &chart.AxisOptions{Name: "Sports", ShowSplitLine: new(true)}
	options.YAxis = &chart.AxisOptions{Name: "Score", ShowSplitLine: new(true)}
	return interactivescatter.Config{
		Label: "Sports scatter with split lines", Caption: "Named axes and split lines support comparison across six sports.", XAxis: interactiveScatterSports,
		Series: []interactivescatter.Series{
			{Name: "Player A", Data: interactiveScatterData([6]float64{47, 47, 87, 88, 90, 15}, true)},
			{Name: "Player B", Data: interactiveScatterData([6]float64{41, 8, 87, 31, 29, 56}, true)},
		},
		Width: "100%", Height: "420px", Options: options, Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractiveEffectScatter() interactivescatter.Config {
	return interactivescatter.Config{
		Label: "Basic dunk effect scatter", Caption: "A ripple emphasizes each player without changing the six exact dunk values.", Variant: interactivescatter.VariantEffect,
		XAxis: interactiveScatterPlayers, Series: []interactivescatter.Series{{Name: "Dunk", Data: interactiveScatterData([6]float64{37, 31, 85, 26, 13, 90}, false)}},
		Width: "100%", Height: "420px", Options: interactiveScatterOptions("basic EffectScatter example", "basic-dunk-effect-scatter"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractiveEffectScatterStyles() interactivescatter.Config {
	return interactivescatter.Config{
		Label: "Dunk and shoot ripple styles", Caption: "Dunk uses a slower, larger stroke ripple; Shoot uses a faster, smaller fill ripple.", Variant: interactivescatter.VariantEffect,
		XAxis: interactiveScatterPlayers,
		Series: []interactivescatter.Series{
			{Name: "Dunk", Data: interactiveScatterData([6]float64{94, 63, 33, 47, 78, 24}, false), Ripple: &chart.RippleOptions{Period: 4, Scale: 10, BrushType: "stroke"}},
			{Name: "Shoot", Data: interactiveScatterData([6]float64{59, 53, 57, 21, 89, 99}, false), Ripple: &chart.RippleOptions{Period: 3, Scale: 6, BrushType: "fill"}},
		},
		Width: "100%", Height: "420px", Options: interactiveScatterOptions("wave style", "styled-player-ripples"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func interactiveChartScatterCode() string {
	return `@interactivescatter.Scatter(interactivescatter.Config{
  Label: "Sports scores",
  XAxis: []string{"Swimming", "Surfing", "Shooting", "Skating", "Wrestling", "Diving"},
  Series: []interactivescatter.Series{{
    Name: "Category A",
    Data: []interactivescatter.Data{{Value: 81, Symbol: "roundRect", SymbolSize: 20, SymbolRotate: 10}},
  }},
  Width: "100%",
  Height: "420px",
})`
}
