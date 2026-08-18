package charts

import (
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

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
