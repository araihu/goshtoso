package charts

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/pie"
)

const staticPieUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"

type staticPieCoverageEntry struct {
	Path      string
	SHA256    string
	Treatment string
}

func staticPieUpstreamCoverage() []staticPieCoverageEntry {
	return []staticPieCoverageEntry{
		{Path: "examples/1-Painter/doughnut_chart-1-basic/main.go", SHA256: "b97bca2322e90e2f03ab49aa77f683d0c58e027846b939e5a61100602dad1ebf", Treatment: "Basic doughnut presentation"},
		{Path: "examples/1-Painter/doughnut_chart-2-styles/main.go", SHA256: "5816db5dd035c8607b2929779353c32d2bca78ed5f6244b3fc04e65292ac3610", Treatment: "Outside labels, inside labels, and center total"},
		{Path: "examples/1-Painter/pie_chart-1-basic/main.go", SHA256: "06183e92e75445d89917af5dfd318c8b45f624c4efa6565b626a6aff6b3b128f", Treatment: "Basic pie presentation"},
		{Path: "examples/1-Painter/pie_chart-2-series_radius/main.go", SHA256: "54d85c6420a5e8f4fca7691c4969be80cc6bc52f8d4f10cbe5e499715875cbf6", Treatment: "Area-scaled slice radii"},
		{Path: "examples/1-Painter/pie_chart-3-gap/main.go", SHA256: "2392d1fd1a7644158626a261344e79b18bef2c3d802fa1cea8c3add413b980f6", Treatment: "Segment gap and hidden legend"},
		{Path: "examples/2-OptionFunc/doughnut_chart-1-basic/main.go", SHA256: "1936ff4508d6ef3967185e4076804bf53dc0bf8c64a254a569081fb1d399b453", Treatment: "Duplicate basic doughnut through option functions"},
		{Path: "examples/2-OptionFunc/pie_chart-1-basic/main.go", SHA256: "d09222d5febf104f07a81e05a4235d96004b61e5c032dd3513a501a840bbe9b7", Treatment: "Duplicate basic pie through option functions"},
	}
}

func upstreamBasicPieSlices() []pie.Slice {
	return []pie.Slice{
		{Name: "Search Engine", Value: 1048},
		{Name: "Direct", Value: 735},
		{Name: "Email", Value: 580},
		{Name: "Union Ads", Value: 484},
		{Name: "Video Ads", Value: 300},
	}
}

func upstreamStylePieSlices() []pie.Slice {
	return []pie.Slice{
		{Name: "Direct", Value: 1048},
		{Name: "Search Engine", Value: 735},
		{Name: "Referral", Value: 580},
		{Name: "Email", Value: 484},
		{Name: "Video Ads", Value: 300},
	}
}

func sampleBasicPie() pie.Config {
	return pie.Config{
		Label: "Pie Chart", Caption: "Five named channels shown as shares of one total.",
		Title:   pie.TitleOptions{Text: "Pie Chart", Subtitle: "(Fake Data)", Placement: pie.PlacementCenter, FontSize: 16, SubtitleFontSize: 10},
		Legend:  pie.LegendOptions{Orientation: pie.LegendVertical, LeftPercent: 80, VerticalPlacement: pie.VerticalPlacementBottom, FontSize: 10},
		Padding: pie.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}, Slices: upstreamBasicPieSlices(), Width: 600, Height: 400,
		Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "pie-chart", Background: chartcontrol.ExportBackgroundTransparent},
	}
}

func sampleAreaScaledPie() pie.Config {
	cfg := sampleBasicPie()
	cfg.Label = "Area-scaled Pie Chart"
	cfg.Caption = "Slice radii scale by the square root of each value while angles retain part-to-whole proportions."
	cfg.Radius = pie.RadiusOptions{OuterPixels: 120, Scale: pie.RadiusScaleArea}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "area-scaled-pie-chart"}
	return cfg
}

func sampleSegmentGapPie() pie.Config {
	return pie.Config{
		Label: "Pie Chart With Segment Gap", Caption: "Sixteen-pixel separation distinguishes adjacent slices without changing their values.",
		Title:  pie.TitleOptions{Text: "Pie Chart With Segment Gap", Placement: pie.PlacementCenter, FontSize: 16},
		Legend: pie.LegendOptions{Hidden: true}, SegmentGap: 16, Slices: upstreamBasicPieSlices(), Width: 600, Height: 400,
		Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "pie-chart-segment-gap"},
	}
}

func sampleBasicDoughnutChart() pie.Config {
	return pie.Config{
		Label: "Doughnut Chart", Caption: "The open center preserves the same five part-to-whole values.",
		Variant: pie.VariantDoughnut, InnerRadiusPercent: 60,
		Title:   pie.TitleOptions{Text: "Doughnut Chart", Subtitle: "(Fake Data)", Placement: pie.PlacementCenter, FontSize: 16, SubtitleFontSize: 10},
		Legend:  pie.LegendOptions{Orientation: pie.LegendVertical, LeftPercent: 80, VerticalPlacement: pie.VerticalPlacementBottom, FontSize: 10},
		Padding: pie.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}, Slices: upstreamBasicPieSlices(), Width: 600, Height: 400,
		Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "doughnut-chart"},
	}
}

func sampleDoughnutOutsideLabels() pie.Config {
	return pie.Config{
		Label: "Labels Outside", Caption: "Exterior labels remain connected to slices across a twenty-four-pixel segment gap.",
		Variant: pie.VariantDoughnut, InnerRadiusPercent: 60, SegmentGap: 24,
		Title: pie.TitleOptions{Text: "Labels Outside", Placement: pie.PlacementCenter}, Legend: pie.LegendOptions{Hidden: true},
		Padding: pie.Padding{Top: 10, Right: 10, Bottom: 15, Left: 10}, Slices: upstreamStylePieSlices(), Width: 600, Height: 400,
		Export: &chartcontrol.ExportOptions{Filename: "doughnut-labels-outside"},
	}
}

func sampleDoughnutInsideLabels() pie.Config {
	return pie.Config{
		Label: "Labels Inside", Caption: "Slice labels move into the enlarged center while exact values remain adjacent.",
		Variant: pie.VariantDoughnut, InnerRadiusPercent: 80, Labels: pie.LabelOptions{Placement: pie.LabelPlacementInside},
		Title: pie.TitleOptions{Text: "Labels Inside", Placement: pie.PlacementCenter}, Legend: pie.LegendOptions{Hidden: true},
		Padding: pie.Padding{Top: 10, Right: 10, Bottom: 15, Left: 10}, Slices: upstreamStylePieSlices(), Width: 400, Height: 400,
		Export: &chartcontrol.ExportOptions{Filename: "doughnut-labels-inside"},
	}
}

func sampleDoughnutCenterTotal() pie.Config {
	return pie.Config{
		Label: "Legend", Caption: "A compact total occupies the center; slice labels are hidden and the legend names every channel.",
		Variant: pie.VariantDoughnut, InnerRadiusPercent: 80, SegmentGap: 8, Labels: pie.LabelOptions{Hidden: true},
		Center: pie.CenterOptions{Content: pie.CenterContentTotal, Prefix: "Total Response: ", Format: pie.ValueFormatHumanized, Decimals: 2, FontSize: 12},
		Title:  pie.TitleOptions{Text: "Legend", Placement: pie.PlacementCenter}, Legend: pie.LegendOptions{VerticalPlacement: pie.VerticalPlacementBottom, Overlay: true},
		Padding: pie.Padding{Top: 10, Right: 10, Bottom: 15, Left: 10}, Slices: upstreamStylePieSlices(), Width: 400, Height: 400,
		RootAttrs: templ.Attributes{"data-goshtoso-candidate": "pie-doughnut-b97bca2322e90e2f", "data-static-pie-exhaustion": "1fe31b06"},
		Controls:  chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "doughnut-center-total"},
	}
}

func basicPieCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Pie Chart",
  Title: pie.TitleOptions{Text: "Pie Chart", Subtitle: "(Fake Data)", Placement: pie.PlacementCenter},
  Legend: pie.LegendOptions{Orientation: pie.LegendVertical, LeftPercent: 80, VerticalPlacement: pie.VerticalPlacementBottom},
  Slices: []pie.Slice{{Name: "Search Engine", Value: 1048}, {Name: "Direct", Value: 735}, {Name: "Email", Value: 580}, {Name: "Union Ads", Value: 484}, {Name: "Video Ads", Value: 300}},
  Width: 600, Height: 400,
})`
}

func areaScaledPieCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Area-scaled Pie Chart", Slices: slices,
  Radius: pie.RadiusOptions{OuterPixels: 120, Scale: pie.RadiusScaleArea},
  Width: 600, Height: 400,
})`
}

func segmentGapPieCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Pie Chart With Segment Gap", Slices: slices,
  SegmentGap: 16,
  Legend: pie.LegendOptions{Hidden: true},
  Width: 600, Height: 400,
})`
}

func basicDoughnutCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Doughnut Chart", Variant: pie.VariantDoughnut,
  InnerRadiusPercent: 60, Slices: slices,
  Width: 600, Height: 400,
})`
}

func doughnutLabelPlacementCode() string {
	return `outside := pie.Config{Label: "Labels Outside", Variant: pie.VariantDoughnut, InnerRadiusPercent: 60, SegmentGap: 24, Legend: pie.LegendOptions{Hidden: true}, Slices: slices}
inside := pie.Config{Label: "Labels Inside", Variant: pie.VariantDoughnut, InnerRadiusPercent: 80, Labels: pie.LabelOptions{Placement: pie.LabelPlacementInside}, Legend: pie.LegendOptions{Hidden: true}, Slices: slices}

@pie.Pie(outside)
@pie.Pie(inside)`
}

func doughnutCenterTotalCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Legend", Variant: pie.VariantDoughnut, InnerRadiusPercent: 80,
  SegmentGap: 8, Labels: pie.LabelOptions{Hidden: true},
  Center: pie.CenterOptions{Content: pie.CenterContentTotal, Prefix: "Total Response: ", Format: pie.ValueFormatHumanized, Decimals: 2, FontSize: 12},
  Legend: pie.LegendOptions{VerticalPlacement: pie.VerticalPlacementBottom, Overlay: true},
  Slices: slices,
})`
}
