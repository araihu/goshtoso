package charts

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/pie"
)

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
