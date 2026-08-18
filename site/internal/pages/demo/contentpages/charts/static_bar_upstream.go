package charts

import (
	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
)

const staticBarUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"

type staticBarCoverageEntry struct {
	Path      string
	SHA256    string
	Treatment string
}

func staticBarUpstreamCoverage() []staticBarCoverageEntry {
	return []staticBarCoverageEntry{
		{Path: "examples/1-Painter/bar_chart-1-basic/main.go", SHA256: "30f03d99e6f1394096d18ebed5edc04210096945c163719ce38fdd4c97368383", Treatment: "Basic vertical comparison"},
		{Path: "examples/1-Painter/bar_chart-2-size_margin/main.go", SHA256: "fc0d426db8e09dc2032bd0c8f6bf67dbcdd66668bac0547821e4094006a55c7b", Treatment: "Vertical thickness and group gap comparison"},
		{Path: "examples/1-Painter/bar_chart-3-label_position-round_caps/main.go", SHA256: "b29d5387ab867885d627c565ce11c2284dd046272a43a86dabb83549363696ce", Treatment: "Rounded caps and start/end value labels"},
		{Path: "examples/1-Painter/bar_chart-4-mark/main.go", SHA256: "544fea22c29db4225c7b10bb6d12137d484a4ca9b6c647dc29730a61ce4ced4c", Treatment: "Average lines and minimum/maximum points"},
		{Path: "examples/1-Painter/bar_chart-5-stacked/main.go", SHA256: "d35e6866b6e4d4071d3e09e26db77165fbd9aecefecf0e6e880bb35752e23119", Treatment: "Stacked totals, maximum line, and global maximum point"},
		{Path: "examples/1-Painter/horizontal_bar_chart-1-basic/main.go", SHA256: "735240dd8433bd2494ae019f272840a8ff2fcf5572166b78269e23cbff7111a0", Treatment: "Basic horizontal comparison"},
		{Path: "examples/1-Painter/horizontal_bar_chart-2-size_margin/main.go", SHA256: "34d9f682f5168830cfac75a4f35a4bcc2e8216ea25024142be892bc7784597da", Treatment: "Horizontal thickness and group gap comparison"},
		{Path: "examples/1-Painter/horizontal_bar_chart-3-mark/main.go", SHA256: "c2bd6eaf3f47d8bce333186aa0212204fe3c6ff67cc66236fc6fe1040d180cb4", Treatment: "Horizontal maximum reference lines"},
		{Path: "examples/1-Painter/horizontal_bar_chart-4-stacked/main.go", SHA256: "82ff73196a355aeda6ddb12b4e6d6cfc2d191dc32dfcd6e1830eb69a321b7b24", Treatment: "Stacked horizontal bars with exact value labels"},
		{Path: "examples/2-OptionFunc/bar_chart-1-basic/main.go", SHA256: "eeff9689b6279ecbbdf9475a8034e8f57aa582c29e8d74db74f34cf78b385ca8", Treatment: "Basic vertical comparison and references through option functions"},
		{Path: "examples/2-OptionFunc/horizontal_bar_chart-1-basic/main.go", SHA256: "84268867a32a2a4ff81cf29d2a56a48174d7c1d46c5edc8ac20a82c07e13fea4", Treatment: "Basic horizontal comparison through option functions"},
	}
}

var barMonthLabels = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

var barRainfallValues = []float64{2.0, 4.9, 7.0, 23.2, 25.6, 76.7, 135.6, 162.2, 32.6, 20.0, 6.4, 3.3}

var barEvaporationValues = []float64{2.6, 5.9, 9.0, 26.4, 28.7, 70.7, 175.6, 182.2, 48.7, 18.8, 6.0, 2.3}

func sampleBasicBar() bar.Config {
	return bar.Config{
		Label: "Monthly rainfall and evaporation", Caption: "Grouped monthly rainfall and evaporation values.",
		Title: "Bar Chart", Labels: append([]string(nil), barMonthLabels...),
		Series: []bar.Series{
			{Name: "Rainfall", Values: append([]float64(nil), barRainfallValues...)},
			{Name: "Evaporation", Values: append([]float64(nil), barEvaporationValues...)},
		},
		Legend: bar.LegendOptions{Placement: bar.LegendPlacementEnd, Overlay: true},
		Width:  600, Height: 400, Controls: chartcontrol.Options{Fullscreen: true},
		Export: &chartcontrol.ExportOptions{Filename: "monthly-rainfall-evaporation"},
	}
}

func sampleBarGeometryComparison(horizontal bool) (bar.Config, bar.Config, bar.Config) {
	orientation := bar.OrientationVertical
	axisWord := "vertical"
	if horizontal {
		orientation = bar.OrientationHorizontal
		axisWord = "horizontal"
	}
	base := bar.Config{
		Label: "Default " + axisWord + " bar geometry", Caption: "Automatic bar thickness and group spacing.",
		Title: "Default", Orientation: orientation,
		Labels: append([]string(nil), barMonthLabels[:6]...),
		Series: []bar.Series{
			{Name: "Rainfall", Values: append([]float64(nil), barRainfallValues[:6]...)},
			{Name: "Evaporation", Values: append([]float64(nil), barEvaporationValues[:6]...)},
		},
		Legend: bar.LegendOptions{Hidden: true}, Width: 400, Height: 400,
		Export: &chartcontrol.ExportOptions{Filename: axisWord + "-bar-default-geometry"},
	}
	thin := cloneBarConfig(base)
	thin.Label, thin.Title, thin.Caption = "Thin "+axisWord+" bars", "Small Bar", "Each bar uses 15% of its allotted series slot."
	thin.Geometry.ThicknessRatio = .15
	thin.Export = &chartcontrol.ExportOptions{Filename: axisWord + "-bar-thin-geometry"}
	noGap := cloneBarConfig(base)
	noGap.Label, noGap.Title, noGap.Caption = "No-gap "+axisWord+" bars", "No Margin", "Grouped bars have no gap between series."
	zero := 0.0
	noGap.Geometry.GapRatio = &zero
	noGap.Export = &chartcontrol.ExportOptions{Filename: axisWord + "-bar-no-gap-geometry"}
	return base, thin, noGap
}

func sampleRoundedBarLabels(position bar.DataLabelPosition) bar.Config {
	positionName, title := "end", "Bar Chart Top Label"
	if position == bar.DataLabelPositionStart {
		positionName, title = "start", "Bar Chart Bottom Label"
	}
	tight := .02
	labels := bar.DataLabelOptions{Show: true, Format: bar.ValueFormatHumanized}
	return bar.Config{
		Label: "Rounded bars with " + positionName + " value labels", Caption: "Rounded value-end caps with exact labels anchored at the " + positionName + ".",
		Title: title, Labels: append([]string(nil), barMonthLabels[:6]...), LabelPosition: position,
		Series: []bar.Series{
			{Name: "Rainfall", Values: append([]float64(nil), barRainfallValues[3:9]...), Labels: labels},
			{Name: "Evaporation", Values: append([]float64(nil), barEvaporationValues[3:9]...), Labels: labels},
		},
		Geometry: bar.GeometryOptions{GapRatio: &tight, RoundedCaps: true}, Legend: bar.LegendOptions{Hidden: true},
		Width: 500, Height: 400, Export: &chartcontrol.ExportOptions{Filename: "rounded-bar-" + positionName + "-labels"},
	}
}

func sampleStackedBar() bar.Config {
	return bar.Config{
		Label: "Stacked monthly rainfall and evaporation", Caption: "Monthly totals retain each series contribution and the maximum stack total.",
		Labels: append([]string(nil), barMonthLabels...), Stacked: true,
		Series: []bar.Series{
			{Name: "Rainfall", Values: append([]float64(nil), barRainfallValues...), References: bar.References{MaximumLine: true, Format: bar.ValueFormatHumanized}},
			{Name: "Evaporation", Values: append([]float64(nil), barEvaporationValues...), References: bar.References{GlobalMaximum: true, PointPrefix: "Sum:", PointSize: 32, Format: bar.ValueFormatHumanized}},
		},
		Legend:  bar.LegendOptions{Placement: bar.LegendPlacementEnd, Overlay: true},
		Padding: bar.Padding{Top: 20, Right: 45, Bottom: 20, Left: 20}, Width: 600, Height: 400,
		Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "stacked-rainfall-evaporation"},
	}
}

func sampleHorizontalBarReferences() bar.Config {
	cfg := sampleHorizontalWorldPopulation()
	cfg.Label = "Horizontal world population with maximum reference lines"
	cfg.Caption = "Each reporting series includes its maximum-value reference line."
	cfg.Series[0].References = bar.References{MaximumLine: true}
	cfg.Series[1].References = bar.References{MaximumLine: true}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "horizontal-world-population-references"}
	return cfg
}

func sampleHorizontalStackedBar() bar.Config {
	labels := bar.DataLabelOptions{Show: true}
	return bar.Config{
		Label: "Stacked horizontal values", Caption: "Stacked 2011 and 2012 values with exact labels and a hidden numeric axis.",
		Title: "Some Numbers", Orientation: bar.OrientationHorizontal, Stacked: true,
		Labels: []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
		Series: []bar.Series{
			{Name: "2011", Values: []float64{10, 30, 50, 70, 90, 110, 130}, Labels: labels},
			{Name: "2012", Values: []float64{20, 40, 60, 80, 100, 120, 140}, Labels: labels},
		},
		ValueAxis: bar.ValueAxisOptions{Hidden: true}, Padding: bar.Padding{Top: 20, Right: 20, Bottom: 0, Left: 20},
		Width: 600, Height: 400, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "stacked-horizontal-values"},
	}
}

func cloneBarConfig(cfg bar.Config) bar.Config {
	result := cfg
	result.Labels = append([]string(nil), cfg.Labels...)
	result.Series = append([]bar.Series(nil), cfg.Series...)
	for index := range result.Series {
		result.Series[index].Values = append([]float64(nil), cfg.Series[index].Values...)
	}
	return result
}

func basicBarCode() string {
	return `@bar.Bar(bar.Config{
  Label: "Monthly rainfall and evaporation",
  Title: "Bar Chart",
  Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
  Series: []bar.Series{
    {Name: "Rainfall", Values: rainfall},
    {Name: "Evaporation", Values: evaporation},
  },
  Legend: bar.LegendOptions{Placement: bar.LegendPlacementEnd, Overlay: true},
  Width: 600, Height: 400,
})`
}

func geometryBarCode(horizontal bool) string {
	orientation := ""
	if horizontal {
		orientation = "Orientation: bar.OrientationHorizontal,\n  "
	}
	return `gap := 0.0
base := bar.Config{
  Label: "Bar geometry", ` + orientation + `Labels: months,
  Series: []bar.Series{{Name: "Rainfall", Values: rainfall}, {Name: "Evaporation", Values: evaporation}},
}
thin := base
thin.Geometry = bar.GeometryOptions{ThicknessRatio: 0.15}
noGap := base
noGap.Geometry = bar.GeometryOptions{GapRatio: &gap}

@bar.Bar(base)
@bar.Bar(thin)
@bar.Bar(noGap)`
}

func roundedBarLabelsCode() string {
	return `tightGap := 0.02
labels := bar.DataLabelOptions{Show: true, Format: bar.ValueFormatHumanized}
base := bar.Config{
  Label: "Rounded bars with exact value labels", Labels: months,
  Series: []bar.Series{{Name: "Rainfall", Values: rainfall, Labels: labels}, {Name: "Evaporation", Values: evaporation, Labels: labels}},
  Geometry: bar.GeometryOptions{GapRatio: &tightGap, RoundedCaps: true},
}
start := base
start.LabelPosition = bar.DataLabelPositionStart

@bar.Bar(base)
@bar.Bar(start)`
}

func stackedBarCode() string {
	return `@bar.Bar(bar.Config{
  Label: "Stacked monthly rainfall and evaporation", Labels: months, Stacked: true,
  Series: []bar.Series{
    {Name: "Rainfall", Values: rainfall, References: bar.References{MaximumLine: true, Format: bar.ValueFormatHumanized}},
    {Name: "Evaporation", Values: evaporation, References: bar.References{GlobalMaximum: true, PointPrefix: "Sum:", PointSize: 32, Format: bar.ValueFormatHumanized}},
  },
  Padding: bar.Padding{Top: 20, Right: 45, Bottom: 20, Left: 20},
})`
}

func horizontalBarReferencesCode() string {
	return `@bar.Bar(bar.Config{
  Label: "Horizontal world population with maximum reference lines",
  Orientation: bar.OrientationHorizontal,
  Labels: []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
  Series: []bar.Series{
    {Name: "2011", Values: values2011, References: bar.References{MaximumLine: true}},
    {Name: "2012", Values: values2012, References: bar.References{MaximumLine: true}},
  },
})`
}

func horizontalStackedBarCode() string {
	return `labels := bar.DataLabelOptions{Show: true}
@bar.Bar(bar.Config{
  Label: "Stacked horizontal values", Orientation: bar.OrientationHorizontal, Stacked: true,
  Labels: []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
  Series: []bar.Series{
    {Name: "2011", Values: values2011, Labels: labels},
    {Name: "2012", Values: values2012, Labels: labels},
  },
  ValueAxis: bar.ValueAxisOptions{Hidden: true},
})`
}
