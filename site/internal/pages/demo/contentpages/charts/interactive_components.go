package charts

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	interactiveboxplot "github.com/araihu/goshtoso-charts/components/interactive/boxplot"
	interactivegauge "github.com/araihu/goshtoso-charts/components/interactive/gauge"
	interactivegraph "github.com/araihu/goshtoso-charts/components/interactive/graph"
	interactiveheatmap "github.com/araihu/goshtoso-charts/components/interactive/heatmap"
	interactiveparallel "github.com/araihu/goshtoso-charts/components/interactive/parallel"
	interactivesankey "github.com/araihu/goshtoso-charts/components/interactive/sankey"
	interactivesunburst "github.com/araihu/goshtoso-charts/components/interactive/sunburst"
	interactivethemeriver "github.com/araihu/goshtoso-charts/components/interactive/themeriver"
	interactivetree "github.com/araihu/goshtoso-charts/components/interactive/tree"
	interactivetreemap "github.com/araihu/goshtoso-charts/components/interactive/treemap"
	interactivewordcloud "github.com/araihu/goshtoso-charts/components/interactive/wordcloud"
)

func controlledOptions(title, filename string) chart.ChartOptions {
	return chart.ChartOptions{
		Title:    &chart.TitleOptions{Text: title},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: filename},
	}
}

var interactiveHeatMapSourceData = [][3]int{
	{0, 0, 5}, {0, 1, 1}, {0, 2, 0}, {0, 3, 0}, {0, 4, 0}, {0, 5, 0},
	{0, 6, 0}, {0, 7, 0}, {0, 8, 0}, {0, 9, 0}, {0, 10, 0}, {0, 11, 2},
	{0, 12, 4}, {0, 13, 1}, {0, 14, 1}, {0, 15, 3}, {0, 16, 4}, {0, 17, 6},
	{0, 18, 4}, {0, 19, 4}, {0, 20, 3}, {0, 21, 3}, {0, 22, 2}, {0, 23, 5},
	{1, 0, 7}, {1, 1, 0}, {1, 2, 0}, {1, 3, 0}, {1, 4, 0}, {1, 5, 0},
	{1, 6, 0}, {1, 7, 0}, {1, 8, 0}, {1, 9, 0}, {1, 10, 5}, {1, 11, 2},
	{1, 12, 2}, {1, 13, 6}, {1, 14, 9}, {1, 15, 11}, {1, 16, 6}, {1, 17, 7},
	{1, 18, 8}, {1, 19, 12}, {1, 20, 5}, {1, 21, 5}, {1, 22, 7}, {1, 23, 2},
	{2, 0, 1}, {2, 1, 1}, {2, 2, 0}, {2, 3, 0}, {2, 4, 0}, {2, 5, 0}, {2, 6, 0},
	{2, 7, 0}, {2, 8, 0}, {2, 9, 0}, {2, 10, 3}, {2, 11, 2}, {2, 12, 1}, {2, 13, 9},
	{2, 14, 8}, {2, 15, 10}, {2, 16, 6}, {2, 17, 5}, {2, 18, 5}, {2, 19, 5},
	{2, 20, 7}, {2, 21, 4}, {2, 22, 2}, {2, 23, 4}, {3, 0, 7}, {3, 1, 3},
	{3, 2, 0}, {3, 3, 0}, {3, 4, 0}, {3, 5, 0}, {3, 6, 0}, {3, 7, 0},
	{3, 8, 1}, {3, 9, 0}, {3, 10, 5}, {3, 11, 4}, {3, 12, 7}, {3, 13, 14},
	{3, 14, 13}, {3, 15, 12}, {3, 16, 9}, {3, 17, 5}, {3, 18, 5}, {3, 19, 10},
	{3, 20, 6}, {3, 21, 4}, {3, 22, 4}, {3, 23, 1}, {4, 0, 1}, {4, 1, 3},
	{4, 2, 0}, {4, 3, 0}, {4, 4, 0}, {4, 5, 1}, {4, 6, 0}, {4, 7, 0},
	{4, 8, 0}, {4, 9, 2}, {4, 10, 4}, {4, 11, 4}, {4, 12, 2}, {4, 13, 4},
	{4, 14, 4}, {4, 15, 14}, {4, 16, 12}, {4, 17, 1}, {4, 18, 8}, {4, 19, 5},
	{4, 20, 3}, {4, 21, 7}, {4, 22, 3}, {4, 23, 0}, {5, 0, 2}, {5, 1, 1},
	{5, 2, 0}, {5, 3, 3}, {5, 4, 0}, {5, 5, 0}, {5, 6, 0}, {5, 7, 0}, {5, 8, 2},
	{5, 9, 0}, {5, 10, 4}, {5, 11, 1}, {5, 12, 5}, {5, 13, 10}, {5, 14, 5},
	{5, 15, 7}, {5, 16, 11}, {5, 17, 6}, {5, 18, 0}, {5, 19, 5}, {5, 20, 3},
	{5, 21, 4}, {5, 22, 2}, {5, 23, 0}, {6, 0, 1}, {6, 1, 0}, {6, 2, 0},
	{6, 3, 0}, {6, 4, 0}, {6, 5, 0}, {6, 6, 0}, {6, 7, 0}, {6, 8, 0},
	{6, 9, 0}, {6, 10, 1}, {6, 11, 0}, {6, 12, 2}, {6, 13, 1}, {6, 14, 3},
	{6, 15, 4}, {6, 16, 0}, {6, 17, 0}, {6, 18, 0}, {6, 19, 0}, {6, 20, 1},
	{6, 21, 2}, {6, 22, 2}, {6, 23, 6},
}

var interactiveHeatMapWeekDays = []string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"}

var interactiveHeatMapDayHours = []string{
	"12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a",
	"12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p",
}

func sampleInteractiveHeatMap() interactiveheatmap.Config {
	data := make([]interactiveheatmap.Data, len(interactiveHeatMapSourceData))
	for index, source := range interactiveHeatMapSourceData {
		data[index] = interactiveheatmap.Data{X: source[1], Y: source[0], Value: float64(source[2]), Missing: source[2] == 0}
	}
	options := controlledOptions("Basic heatmap example", "weekly-activity-heatmap")
	options.Legend = &chart.LegendOptions{Top: "32", Left: "center"}
	return interactiveheatmap.Config{
		Label: "Weekly activity by hour", Caption: "Seven days across twenty-four hourly buckets; empty source cells remain no-data cells.",
		XAxis: interactiveHeatMapDayHours, YAxis: interactiveHeatMapWeekDays,
		ValueRange: interactiveheatmap.ValueRange{Min: 0, Max: 10, Calculable: new(true)},
		SplitArea:  new(true),
		Series:     []interactiveheatmap.Series{{Name: "Activity", Data: data}},
		Height:     "34rem",
		Options:    options,
	}
}

func sampleInteractiveCalendarHeatMap() interactiveheatmap.Config {
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)
	random := rand.New(rand.NewSource(1))
	data := make([]interactiveheatmap.Data, 0, 366)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		value := random.Intn(21)
		data = append(data, interactiveheatmap.Data{Date: date, Value: float64(value), Missing: value == 0})
	}
	options := controlledOptions("Calendar heatmap example", "calendar-activity-heatmap")
	options.Legend = &chart.LegendOptions{Top: "32", Left: "center"}
	return interactiveheatmap.Config{
		Label: "Calendar activity", Caption: "A deterministic 366-day sequence preserves the source range and explicit no-data days.",
		Coordinate: interactiveheatmap.CoordinateCalendar,
		Calendar: &interactiveheatmap.Calendar{
			Start: start, End: end,
			Options: chart.CalendarOptions{
				Top: "100", Left: "30", Right: "30", CellSize: "auto", Orient: "horizontal",
				CellStyle: &chart.ItemStyle{BorderWidth: 0.5}, MonthLabel: &chart.CalendarLabelOptions{FontSize: 7},
			},
		},
		ValueRange: interactiveheatmap.ValueRange{Min: 0, Max: 20},
		Series:     []interactiveheatmap.Series{{Name: "Calendar activity", Data: data}},
		Height:     "34rem",
		Options:    options,
	}
}

func sampleInteractiveBoxPlot() interactiveboxplot.Config {
	return interactiveboxplot.Config{
		Label: "Request latency distribution", Caption: "Five-number latency summaries by environment.",
		Categories: []string{"Development", "Staging", "Production"},
		Series: []interactiveboxplot.Series{{
			Name: "Latency (ms)",
			Data: []interactiveboxplot.Data{
				{Min: 18, Q1: 31, Median: 42, Q3: 58, Max: 94},
				{Min: 22, Q1: 38, Median: 51, Q3: 73, Max: 116},
				{Min: 25, Q1: 44, Median: 62, Q3: 86, Max: 138},
			},
		}},
		Options: controlledOptions("Request latency", "request-latency-distribution"),
	}
}

func sampleInteractiveGauge() interactivegauge.Config {
	return interactivegauge.Config{
		Label: "Deployment completion", Caption: "Current rollout completion percentage.",
		Variant: interactivegauge.VariantProgress,
		Series: []interactivegauge.Series{{
			Name: "Rollout", Data: []interactivegauge.Data{{Name: "Complete", Value: 73}},
		}},
		Options: controlledOptions("Deployment completion", "deployment-completion"),
	}
}

func sampleInteractiveLiquidGauge(title string, shape interactivegauge.LiquidShape, showLabel, showOutline bool, animate *bool) interactivegauge.Config {
	liquid := interactivegauge.LiquidTreatment{Shape: shape, Animate: animate}
	if showLabel {
		liquid.Label = &interactivegauge.LiquidLabel{Show: new(true)}
	}
	if showOutline {
		liquid.Outline = &interactivegauge.LiquidOutline{Show: new(true)}
	}
	return interactivegauge.Config{
		Label: title, Caption: "Readings: 0.3, 0.4, and 0.5 in range 0 to 1.",
		Variant: interactivegauge.VariantLiquid, Max: 1, Height: "320px", Liquid: liquid,
		Series:  []interactivegauge.Series{{Name: "liquid", Data: sampleLiquidGaugeData()}},
		Options: controlledOptions(title, title),
	}
}

func sampleLiquidGaugeData() []interactivegauge.Data {
	return []interactivegauge.Data{
		{Name: "Wave 1", Value: .3},
		{Name: "Wave 2", Value: .4},
		{Name: "Wave 3", Value: .5},
	}
}

func sampleInteractiveGraph() interactivegraph.Config {
	return interactivegraph.Config{
		Label: "Service dependency graph", Caption: "Dependencies between edge, API, data, and notification services.",
		Categories: []interactivegraph.Category{{Name: "Entry"}, {Name: "Service"}, {Name: "Data"}},
		Nodes: []interactivegraph.Node{
			{Name: "Edge", Category: "Entry", Value: 8, Size: 44},
			{Name: "API", Category: "Service", Value: 12, Size: 52},
			{Name: "Worker", Category: "Service", Value: 7, Size: 38},
			{Name: "Database", Category: "Data", Value: 10, Size: 46},
			{Name: "Notifications", Category: "Service", Value: 5, Size: 34},
		},
		Links: []interactivegraph.Link{
			{Source: "Edge", Target: "API", Value: 12},
			{Source: "API", Target: "Database", Value: 10},
			{Source: "API", Target: "Worker", Value: 7},
			{Source: "Worker", Target: "Notifications", Value: 5},
		},
		Roam:    interactivegraph.RoamEnabled,
		Force:   &interactivegraph.ForceOptions{Repulsion: 180, Gravity: 0.08, EdgeLength: 110},
		Options: controlledOptions("Service dependencies", "service-dependencies"),
	}
}

func sampleInteractiveSankey() interactivesankey.Config {
	return interactivesankey.Config{
		Label: "Request flow", Caption: "Requests flowing from ingress through services to outcomes.",
		Series: []interactivesankey.Series{{
			Name: "Requests",
			Nodes: []interactivesankey.Node{
				{Name: "Ingress"}, {Name: "API"}, {Name: "Cache"}, {Name: "Database"}, {Name: "Success"}, {Name: "Error"},
			},
			Links: []interactivesankey.Link{
				{Source: "Ingress", Target: "API", Value: 100},
				{Source: "API", Target: "Cache", Value: 48},
				{Source: "API", Target: "Database", Value: 45},
				{Source: "API", Target: "Error", Value: 7},
				{Source: "Cache", Target: "Success", Value: 48},
				{Source: "Database", Target: "Success", Value: 42},
				{Source: "Database", Target: "Error", Value: 3},
			},
		}},
		Layout:  interactivesankey.Layout{NodeWidth: 18, NodeGap: 14},
		Options: controlledOptions("Request flow", "request-flow"),
	}
}

func sampleInteractiveTree() interactivetree.Config {
	return interactivetree.Config{
		Label:   "Basic tree example",
		Caption: "One root with three branches. Node3 starts collapsed; expand it to reveal its children.",
		Roots: []*interactivetree.Node{{
			Name: "Root",
			Children: []*interactivetree.Node{
				{
					Name:     "Node1",
					Children: []*interactivetree.Node{{Name: "Child1"}},
				},
				{
					Name: "Node2",
					Children: []*interactivetree.Node{
						{Name: "Child1"},
						{Name: "Child2"},
						{Name: "Child3"},
					},
				},
				{
					Name: "Node3", Collapsed: new(true),
					Children: []*interactivetree.Node{
						{Name: "Child1"},
						{Name: "Child2"},
						{Name: "Child3"},
					},
				},
			},
		}},
		Orientation:  interactivetree.OrientationLeftToRight,
		InitialDepth: new(-1),
		NodeLabel:    &chart.LabelOptions{Show: new(true), Position: "top"},
		LeafLabel:    &chart.LabelOptions{Show: new(true), Position: "right"},
		Insets:       interactivetree.Insets{Left: "14%", Right: "14%", Top: "12%", Bottom: "12%"},
		Width:        "100%",
		Height:       "440px",
		Options:      controlledOptions("Basic tree example", "basic-tree-example"),
	}
}

func sampleInteractiveSunburst() interactivesunburst.Config {
	return interactivesunburst.Config{
		Label:   "Basic sunburst example",
		Caption: "Seven parent nodes, each paired with one child. Select a sector to focus the hierarchy; select the center to return.",
		Nodes: []*interactivesunburst.Node{
			{Name: "parent-0", Value: 0.81, Children: []*interactivesunburst.Node{{Name: "child-0", Value: 0.34}}},
			{Name: "parent-1", Value: 0.62, Children: []*interactivesunburst.Node{{Name: "child-1", Value: 0.57}}},
			{Name: "parent-2", Value: 0.45, Children: []*interactivesunburst.Node{{Name: "child-2", Value: 0.73}}},
			{Name: "parent-3", Value: 0.93, Children: []*interactivesunburst.Node{{Name: "child-3", Value: 0.28}}},
			{Name: "parent-4", Value: 0.38, Children: []*interactivesunburst.Node{{Name: "child-4", Value: 0.66}}},
			{Name: "parent-5", Value: 0.71, Children: []*interactivesunburst.Node{{Name: "child-5", Value: 0.49}}},
			{Name: "parent-6", Value: 0.54, Children: []*interactivesunburst.Node{{Name: "child-6", Value: 0.87}}},
		},
		LabelOptions: &chart.LabelOptions{Show: new(true), Position: "inside", FontSize: 10},
		InnerRadius:  16,
		OuterRadius:  88,
		Width:        "100%",
		Height:       "32rem",
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: "Basic sunburst example"},
			Legend:   &chart.LegendOptions{Show: new(false)},
			Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "item"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "basic-sunburst-example"},
		},
	}
}

func sampleInteractiveTreemap() interactivetreemap.Config {
	d3Children := make([]*interactivetreemap.Node, 40)
	for index := range d3Children {
		d3Children[index] = &interactivetreemap.Node{
			Name:  fmt.Sprintf("f%d", index),
			Value: float64(5 + (index*7+3)%10),
			Class: "file",
		}
	}
	return interactivetreemap.Config{
		Label:   "Basic treemap example",
		Caption: "File system usage in KB. Select a directory to focus it in the same chart; use the breadcrumb to return.",
		Nodes: []*interactivetreemap.Node{
			{
				Name: "d1", Class: "directory",
				Children: []*interactivetreemap.Node{{Name: "f1", Value: 1000, Class: "file"}},
			},
			{
				Name: "d2", Class: "directory",
				Children: []*interactivetreemap.Node{
					{Name: "f1", Value: 100, Class: "file"},
					{Name: "f2", Value: 300, Class: "file"},
					{Name: "f3", Value: 200, Class: "file"},
				},
			},
			{Name: "d3", Class: "directory", Children: d3Children},
			{Name: "f1", Value: 450, Class: "file"},
		},
		Navigation: interactivetreemap.NavigationDrillDown,
		Roam:       interactivetreemap.RoamEnabled,
		LabelOptions: &chart.LabelOptions{
			Show: new(true), Position: "inside", FontSize: 11,
		},
		UpperLabel: &chart.LabelOptions{Show: new(true), FontSize: 12},
		Breadcrumb: &interactivetreemap.Breadcrumb{Show: new(true), Height: 24, ItemGap: 8},
		NodeStyle:  interactivetreemap.NodeStyle{BorderWidth: 1, GapWidth: 1},
		LeafDepth:  new(1),
		Levels: []interactivetreemap.Level{
			{
				UpperLabel: &chart.LabelOptions{Show: new(true)},
				NodeStyle:  interactivetreemap.NodeStyle{BorderWidth: 1, GapWidth: 1},
			},
			{NodeStyle: interactivetreemap.NodeStyle{BorderWidth: 2, GapWidth: 1}},
			{
				NodeStyle:       interactivetreemap.NodeStyle{GapWidth: 1},
				ColorSaturation: &interactivetreemap.ColorRange{Min: 0.35, Max: 0.5},
			},
		},
		Width:  "100%",
		Height: "500px",
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: "Basic treemap example", Subtitle: "File system usage", Left: "center"},
			Legend:   &chart.LegendOptions{Show: new(false)},
			Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "item"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "basic-treemap-example"},
		},
	}
}

type parallelSampleRow struct {
	values [7]float64
	level  string
}

func sampleInteractiveParallel() interactiveparallel.Config {
	return interactiveparallel.Config{
		Label:   "Multi Series parallel coordinates",
		Caption: "Daily air-quality measurements for Beijing, Guangzhou, and Shanghai across seven numeric dimensions and one categorical level.",
		Dimensions: []interactiveparallel.Dimension{
			{Name: "Date", Range: &interactiveparallel.Range{Max: chart.Float(31)}, Inverse: true, NameLocation: interactiveparallel.NameStart},
			{Name: "AQI"},
			{Name: "PM2.5"},
			{Name: "PM10"},
			{Name: "CO"},
			{Name: "NO2"},
			{Name: "SO2"},
			{Name: "Level", Categories: []string{"Good", "Moderate", "Lightly", "Moderately", "Heavily", "Severely"}},
		},
		Series: []interactiveparallel.Series{
			{Name: "Beijing", Observations: parallelSampleObservations([]parallelSampleRow{
				{[7]float64{1, 55, 9, 56, 0.46, 18, 6}, "Moderate"},
				{[7]float64{2, 25, 11, 21, 0.65, 34, 9}, "Good"},
				{[7]float64{3, 56, 7, 63, 0.3, 14, 5}, "Moderate"},
				{[7]float64{4, 33, 7, 29, 0.33, 16, 6}, "Good"},
				{[7]float64{5, 42, 24, 44, 0.76, 40, 16}, "Good"},
				{[7]float64{6, 82, 58, 90, 1.77, 68, 33}, "Moderate"},
				{[7]float64{7, 74, 49, 77, 1.46, 48, 27}, "Moderate"},
				{[7]float64{8, 78, 55, 80, 1.29, 59, 29}, "Moderate"},
				{[7]float64{9, 267, 216, 280, 4.8, 108, 64}, "Heavily"},
				{[7]float64{10, 185, 127, 216, 2.52, 61, 27}, "Moderately"},
				{[7]float64{11, 39, 19, 38, 0.57, 31, 15}, "Good"},
				{[7]float64{12, 41, 11, 40, 0.43, 21, 7}, "Good"},
				{[7]float64{13, 64, 38, 74, 1.04, 46, 22}, "Moderate"},
				{[7]float64{14, 108, 79, 120, 1.7, 75, 41}, "Lightly"},
				{[7]float64{15, 108, 63, 116, 1.48, 44, 26}, "Lightly"},
				{[7]float64{16, 33, 6, 29, 0.34, 13, 5}, "Good"},
				{[7]float64{17, 94, 66, 110, 1.54, 62, 31}, "Moderate"},
				{[7]float64{18, 186, 142, 192, 3.88, 93, 79}, "Moderately"},
				{[7]float64{19, 57, 31, 54, 0.96, 32, 14}, "Moderate"},
				{[7]float64{20, 22, 8, 17, 0.48, 23, 10}, "Good"},
				{[7]float64{21, 39, 15, 36, 0.61, 29, 13}, "Good"},
			})},
			{Name: "Guangzhou", Observations: parallelSampleObservations([]parallelSampleRow{
				{[7]float64{1, 26, 37, 27, 1.163, 27, 13}, "Good"},
				{[7]float64{2, 85, 62, 71, 1.195, 60, 8}, "Moderate"},
				{[7]float64{3, 78, 38, 74, 1.363, 37, 7}, "Moderate"},
				{[7]float64{4, 21, 21, 36, 0.634, 40, 9}, "Good"},
				{[7]float64{5, 41, 42, 46, 0.915, 81, 13}, "Good"},
				{[7]float64{6, 56, 52, 69, 1.067, 92, 16}, "Moderate"},
				{[7]float64{7, 64, 30, 28, 0.924, 51, 2}, "Moderate"},
				{[7]float64{8, 55, 48, 74, 1.236, 75, 26}, "Moderate"},
				{[7]float64{9, 76, 85, 113, 1.237, 114, 27}, "Moderate"},
				{[7]float64{10, 91, 81, 104, 1.041, 56, 40}, "Moderate"},
				{[7]float64{11, 84, 39, 60, 0.964, 25, 11}, "Moderate"},
				{[7]float64{12, 64, 51, 101, 0.862, 58, 23}, "Moderate"},
				{[7]float64{13, 70, 69, 120, 1.198, 65, 36}, "Moderate"},
				{[7]float64{14, 77, 105, 178, 2.549, 64, 16}, "Moderate"},
				{[7]float64{15, 109, 68, 87, 0.996, 74, 29}, "Lightly"},
				{[7]float64{16, 73, 68, 97, 0.905, 51, 34}, "Moderate"},
				{[7]float64{17, 54, 27, 47, 0.592, 53, 12}, "Moderate"},
				{[7]float64{18, 51, 61, 97, 0.811, 65, 19}, "Moderate"},
				{[7]float64{19, 91, 71, 121, 1.374, 43, 18}, "Moderate"},
				{[7]float64{20, 73, 102, 182, 2.787, 44, 19}, "Moderate"},
				{[7]float64{21, 73, 50, 76, 0.717, 31, 20}, "Moderate"},
			})},
			{Name: "Shanghai", Observations: parallelSampleObservations([]parallelSampleRow{
				{[7]float64{1, 91, 45, 125, 0.82, 34, 23}, "Moderate"},
				{[7]float64{2, 65, 27, 78, 0.86, 45, 29}, "Moderate"},
				{[7]float64{3, 83, 60, 84, 1.09, 73, 27}, "Moderate"},
				{[7]float64{4, 109, 81, 121, 1.28, 68, 51}, "Lightly"},
				{[7]float64{5, 106, 77, 114, 1.07, 55, 51}, "Lightly"},
				{[7]float64{6, 109, 81, 121, 1.28, 68, 51}, "Lightly"},
				{[7]float64{7, 106, 77, 114, 1.07, 55, 51}, "Lightly"},
				{[7]float64{8, 89, 65, 78, 0.86, 51, 26}, "Moderate"},
				{[7]float64{9, 53, 33, 47, 0.64, 50, 17}, "Moderate"},
				{[7]float64{10, 80, 55, 80, 1.01, 75, 24}, "Moderate"},
				{[7]float64{11, 117, 81, 124, 1.03, 45, 24}, "Lightly"},
				{[7]float64{12, 99, 71, 142, 1.1, 62, 42}, "Moderate"},
				{[7]float64{13, 95, 69, 130, 1.28, 74, 50}, "Moderate"},
				{[7]float64{14, 116, 87, 131, 1.47, 84, 40}, "Lightly"},
				{[7]float64{15, 108, 80, 121, 1.3, 85, 37}, "Lightly"},
				{[7]float64{16, 134, 83, 167, 1.16, 57, 43}, "Lightly"},
				{[7]float64{17, 79, 43, 107, 1.05, 59, 37}, "Moderate"},
				{[7]float64{18, 71, 46, 89, 0.86, 64, 25}, "Moderate"},
				{[7]float64{19, 97, 71, 113, 1.17, 88, 31}, "Moderate"},
				{[7]float64{20, 84, 57, 91, 0.85, 55, 31}, "Moderate"},
				{[7]float64{21, 87, 63, 101, 0.9, 56, 41}, "Moderate"},
			})},
		},
		Width: "900px", Height: "500px",
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: "Multi Series"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "multi-series-parallel-coordinates"},
		},
		Style: charttheme.Style{Class: "max-w-full"},
	}
}

func parallelSampleObservations(rows []parallelSampleRow) []interactiveparallel.Observation {
	result := make([]interactiveparallel.Observation, len(rows))
	for index, row := range rows {
		values := make([]interactiveparallel.Value, 0, 8)
		for _, value := range row.values {
			values = append(values, interactiveparallel.Number(value))
		}
		values = append(values, interactiveparallel.Category(row.level))
		result[index] = interactiveparallel.Observation{Name: fmt.Sprintf("Day %d", index+1), Values: values}
	}
	return result
}

func sampleInteractiveThemeRiver() interactivethemeriver.Config {
	return interactivethemeriver.Config{
		Label: "ThemeRiver-SingleAxis-Time", Caption: "Six named streams across aligned daily values from 8–28 November 2015.",
		Streams: sampleThemeRiverStreams(), Layout: interactivethemeriver.Layout{BottomPercent: chart.Float(10)},
		Width: "100%", Height: "500px",
		Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "ThemeRiver-SingleAxis-Time"}, Tooltip: &chart.TooltipOptions{Show: new(true), Trigger: "axis"}, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "theme-river-single-axis-time"}},
	}
}

func sampleThemeRiverStreams() []interactivethemeriver.Stream {
	dates := make([]time.Time, 21)
	for index := range dates {
		dates[index] = time.Date(2015, time.November, 8+index, 0, 0, 0, 0, time.UTC)
	}
	values := [][]float64{
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{35, 36, 37, 22, 24, 26, 34, 21, 18, 45, 32, 35, 30, 28, 27, 26, 15, 30, 35, 42, 42},
		{21, 25, 27, 23, 24, 21, 35, 39, 40, 36, 33, 43, 40, 34, 28, 26, 37, 41, 46, 47, 41},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 4, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 4, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
	}
	names := []string{"DQ", "TY", "SS", "QG", "SY", "DD"}
	streams := make([]interactivethemeriver.Stream, len(names))
	for streamIndex, name := range names {
		points := make([]interactivethemeriver.Point, len(dates))
		for pointIndex, date := range dates {
			points[pointIndex] = interactivethemeriver.Point{Time: date, Value: values[streamIndex][pointIndex]}
		}
		streams[streamIndex] = interactivethemeriver.Stream{Name: name, Class: "stream-" + strings.ToLower(name), Points: points}
	}
	return streams
}

func sampleInteractiveWordCloud(title string, shape interactivewordcloud.Shape) interactivewordcloud.Config {
	return interactivewordcloud.Config{
		Label: title, Caption: "Twenty weighted terms with exact values available below the canvas.",
		Series: interactivewordcloud.Series{
			Name: "wordcloud", Words: sampleWordCloudWords(),
			Options: interactivewordcloud.SeriesOptions{
				Shape: shape, SizeRange: &interactivewordcloud.SizeRange{Min: 14, Max: 80},
			},
		},
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: title},
			Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "item"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: strings.ReplaceAll(strings.ToLower(title), " ", "-")},
		},
	}
}

func sampleWordCloudWords() []interactivewordcloud.Word {
	return []interactivewordcloud.Word{
		{Name: "Sam S Club", Value: 10000, Class: "retail"},
		{Name: "Macys", Value: 6181, Class: "retail", Color: "#ff8a3d"},
		{Name: "Amy Schumer", Value: 4386},
		{Name: "Jurassic World", Value: 4055},
		{Name: "Charter Communications", Value: 2467},
		{Name: "Chick Fil A", Value: 2244},
		{Name: "Planet Fitness", Value: 1898},
		{Name: "Pitch Perfect", Value: 1484},
		{Name: "Express", Value: 1689},
		{Name: "Home", Value: 1112},
		{Name: "Johnny Depp", Value: 985},
		{Name: "Lena Dunham", Value: 847},
		{Name: "Lewis Hamilton", Value: 582},
		{Name: "KXAN", Value: 555},
		{Name: "Mary Ellen Mark", Value: 550},
		{Name: "Farrah Abraham", Value: 462},
		{Name: "Rita Ora", Value: 366},
		{Name: "Serena Williams", Value: 282},
		{Name: "NCAA baseball tournament", Value: 273},
		{Name: "Point Break", Value: 265},
	}
}

func liveAvailabilityBar(label, caption string, states []int) interactive.BarConfig {
	categories := availabilityCategories(len(states))
	series := []interactive.BarSeries{
		{Name: "Healthy", Data: make([]interactive.BarData, len(states))},
		{Name: "Degraded", Data: make([]interactive.BarData, len(states))},
		{Name: "Down", Data: make([]interactive.BarData, len(states))},
	}
	for index, state := range states {
		if state >= 0 && state < len(series) {
			series[state].Data[index].Value = 1
		}
	}
	config := interactive.BarConfig{
		Label: label, Caption: caption, XAxis: categories, Series: series,
		Height: "240px",
		Style:  charttheme.Style{Palette: charttheme.PaletteStatus},
		Options: interactive.ChartOptions{
			Legend: &interactive.LegendOptions{Show: interactive.Bool(true)},
			XAxis: &interactive.AxisOptions{
				LabelInterval: interactive.Int(5), ShowFirstLabel: interactive.Bool(true), ShowLastLabel: interactive.Bool(true),
			},
			YAxis:     &interactive.AxisOptions{Min: interactive.Float(0), Max: interactive.Float(1), Show: interactive.Bool(false)},
			Animation: interactive.Bool(false),
			Controls:  chartcontrol.Options{Fullscreen: true},
			Export:    &chartcontrol.ExportOptions{Filename: "live-availability"},
		},
		SeriesOptions: interactive.SeriesOptions{Stack: "availability", BarWidth: "70%", BarGap: "0%"},
	}
	config.Live = &interactive.LiveData{URL: "/modules/charts/examples/live-availability/events", Event: "chart"}
	return config
}

func availabilityCategories(count int) []string {
	categories := make([]string, count)
	for index := range categories {
		categories[index] = fmt.Sprintf("-%ds", (count-1-index)*2)
	}
	return categories
}

func availabilityStates(step int) []int {
	states := make([]int, 36)
	for index := range states {
		switch phase := (step + index) % 24; {
		case phase >= 8 && phase <= 10:
			states[index] = 1
		case phase >= 17 && phase <= 19:
			states[index] = 2
		default:
			states[index] = 0
		}
	}
	return states
}

func sampleLiveAvailability() interactive.BarConfig {
	return liveAvailabilityBar(
		"Live availability from server-sent events",
		"Full snapshots stream through the renderer-neutral live-data contract.",
		availabilityStates(0),
	)
}

func interactiveChartLineCode() string {
	return `@interactiveline.Line(interactiveline.Config{
  Label: "Basic line example",
  XAxis: []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"},
  Series: []interactiveline.Series{{
    Name: "Category A",
    Data: []interactiveline.Data{{Value: 120}, {Value: 132}, {Value: 101}, {Value: 134}, {Value: 90}, {Value: 230}},
  }},
})`
}

func interactiveChartPieCode() string {
	return `@interactivepie.Pie(interactivepie.Config{
  Label: "Rose area",
  Series: []interactivepie.Series{{
    Name: "Seasons", InnerRadius: 40, OuterRadius: 75,
    RoseMode: interactivepie.RoseArea,
    LabelContent: interactivepie.LabelNameAndValue,
    Data: seasonalData,
  }},
})`
}

func interactiveChartHeatMapCode() string {
	return `@interactiveheatmap.HeatMap(interactiveheatmap.Config{
  Label: "Weekly activity by hour",
  XAxis: []string{
    "12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a",
    "12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p",
  },
  YAxis: []string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"},
  ValueRange: interactiveheatmap.ValueRange{
    Min: 0, Max: 10, Calculable: chart.Bool(true),
  },
  SplitArea: chart.Bool(true),
  Series: []interactiveheatmap.Series{{
    Name: "Activity",
    Data: []interactiveheatmap.Data{
      {X: 0, Y: 0, Value: 5},
      {X: 1, Y: 0, Value: 1},
      {X: 2, Y: 0, Missing: true},
    },
  }},
})`
}

func interactiveCalendarHeatMapCode() string {
	return `// Add "time" to the templ file's Go import block.
@interactiveheatmap.HeatMap(interactiveheatmap.Config{
  Label: "Calendar activity",
  Coordinate: interactiveheatmap.CoordinateCalendar,
  Calendar: &interactiveheatmap.Calendar{
    Start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
    End: time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
    Options: chart.CalendarOptions{
      Top: "100", Left: "30", Right: "30",
      CellSize: "auto", Orient: "horizontal",
      CellStyle: &chart.ItemStyle{BorderWidth: 0.5},
      MonthLabel: &chart.CalendarLabelOptions{FontSize: 7},
    },
  },
  ValueRange: interactiveheatmap.ValueRange{Min: 0, Max: 20},
  Series: []interactiveheatmap.Series{{
    Name: "Calendar activity",
    Data: []interactiveheatmap.Data{
      {Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Value: 12},
      {Date: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC), Missing: true},
      {Date: time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC), Value: 7},
    },
  }},
})`
}

func interactiveChartBoxPlotCode() string {
	return `@interactiveboxplot.BoxPlot(interactiveboxplot.Config{
  Label: "Request latency distribution",
  Categories: []string{"Development", "Staging", "Production"},
  Series: []interactiveboxplot.Series{{
    Name: "Latency (ms)",
    Data: []interactiveboxplot.Data{
      {Min: 18, Q1: 31, Median: 42, Q3: 58, Max: 94},
      {Min: 22, Q1: 38, Median: 51, Q3: 73, Max: 116},
      {Min: 25, Q1: 44, Median: 62, Q3: 86, Max: 138},
    },
  }},
	Options: chart.ChartOptions{Animation: chart.Bool(false)},
})`
}

func interactiveChartGaugeCode() string {
	return `@interactivegauge.Gauge(interactivegauge.Config{
  Label: "Deployment completion",
  Variant: interactivegauge.VariantProgress,
  Scale: interactivegauge.Scale{}, // Theme-aware cold-to-warm default.
  Series: []interactivegauge.Series{{
    Name: "Rollout",
    Data: []interactivegauge.Data{{Name: "Complete", Value: 73}},
  }},
})`
}

func interactiveGaugeLiquidCode() string {
	return `@interactivegauge.Gauge(interactivegauge.Config{
  Label: "Bounded fill",
  Variant: interactivegauge.VariantLiquid,
  Min: 0,
  Max: 1,
  Series: []interactivegauge.Series{{
    Name: "liquid",
    Data: []interactivegauge.Data{
      {Name: "Wave 1", Value: .3},
      {Name: "Wave 2", Value: .4},
      {Name: "Wave 3", Value: .5},
    },
  }},
  Liquid: interactivegauge.LiquidTreatment{
    Shape:   interactivegauge.LiquidShapeDiamond,
    Animate: chart.Bool(true),
    Label:   &interactivegauge.LiquidLabel{Show: chart.Bool(true)},
  },
})`
}

func liveAvailabilityCode() string {
	return `@interactive.Bar(interactive.BarConfig{
  Label: "Live availability",
  XAxis: rollingCategories, // 36 evenly spaced buckets.
  Series: oneHotStateSeries, // Healthy, Degraded, Down; equal lengths.
  SeriesOptions: interactive.SeriesOptions{
    Stack: "availability", BarWidth: "70%", BarGap: "0%",
  },
	Options: interactive.ChartOptions{Animation: interactive.Bool(false)},
  Live: &interactive.LiveData{URL: "/modules/charts/examples/live-availability/events", Event: "chart"},
})`
}

func interactiveGraphCode() string {
	return `@interactivegraph.Graph(interactivegraph.Config{
  Label: "Service dependencies",
  Nodes: []interactivegraph.Node{{Name: "API"}, {Name: "Database"}},
  Links: []interactivegraph.Link{{Source: "API", Target: "Database", Value: 10}},
  Layout: interactivegraph.LayoutForce,
  Roam: interactivegraph.RoamEnabled,
  Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Service dependencies"}},
})`
}

func interactiveSankeyCode() string {
	return `@interactivesankey.Sankey(interactivesankey.Config{
  Label: "Request flow",
  Series: []interactivesankey.Series{{
    Name: "Requests",
    Nodes: []interactivesankey.Node{{Name: "Ingress"}, {Name: "Success"}},
    Links: []interactivesankey.Link{{Source: "Ingress", Target: "Success", Value: 90}},
  }},
  Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Request flow"}},
})`
}

func interactiveTreeCode() string {
	return `@interactivetree.Tree(interactivetree.Config{
  Label: "Basic tree example",
  Caption: "One root with three branches; Node3 starts collapsed.",
  Roots: []*interactivetree.Node{{
    Name: "Root",
    Children: []*interactivetree.Node{
      {Name: "Node1", Children: []*interactivetree.Node{{Name: "Child1"}}},
      {Name: "Node2", Children: []*interactivetree.Node{
        {Name: "Child1"}, {Name: "Child2"}, {Name: "Child3"},
      }},
      {Name: "Node3", Collapsed: chart.Bool(true), Children: []*interactivetree.Node{
        {Name: "Child1"}, {Name: "Child2"}, {Name: "Child3"},
      }},
    },
  }},
  Orientation: interactivetree.OrientationLeftToRight,
  InitialDepth: chart.Int(-1),
  NodeLabel: &chart.LabelOptions{Show: chart.Bool(true), Position: "top"},
  LeafLabel: &chart.LabelOptions{Show: chart.Bool(true), Position: "right"},
  Insets: interactivetree.Insets{Left: "14%", Right: "14%", Top: "12%", Bottom: "12%"},
  Width: "100%",
  Height: "440px",
})`
}

func interactiveSunburstCode() string {
	return `@interactivesunburst.Sunburst(interactivesunburst.Config{
  Label: "Basic sunburst example",
  Caption: "Seven parent nodes, each paired with one child.",
  Nodes: []*interactivesunburst.Node{
    {Name: "parent-0", Value: 0.81, Children: []*interactivesunburst.Node{
      {Name: "child-0", Value: 0.34},
    }},
	{Name: "parent-1", Value: 0.62, Children: []*interactivesunburst.Node{{Name: "child-1", Value: 0.57}}},
	{Name: "parent-2", Value: 0.45, Children: []*interactivesunburst.Node{{Name: "child-2", Value: 0.73}}},
	{Name: "parent-3", Value: 0.93, Children: []*interactivesunburst.Node{{Name: "child-3", Value: 0.28}}},
	{Name: "parent-4", Value: 0.38, Children: []*interactivesunburst.Node{{Name: "child-4", Value: 0.66}}},
	{Name: "parent-5", Value: 0.71, Children: []*interactivesunburst.Node{{Name: "child-5", Value: 0.49}}},
    {Name: "parent-6", Value: 0.54, Children: []*interactivesunburst.Node{
      {Name: "child-6", Value: 0.87},
    }},
  },
  Navigation: interactivesunburst.NavigationDrillDown,
  Sort: interactivesunburst.SortDescending,
  LabelOptions: &chart.LabelOptions{Show: chart.Bool(true), Position: "inside", FontSize: 10},
  InnerRadius: 16,
  OuterRadius: 88,
  Width: "100%",
  Height: "32rem",
  Style: charttheme.Style{Class: "max-w-full"},
  RootAttrs: templ.Attributes{"data-chart-purpose": "hierarchy"},
})`
}

func interactiveTreemapCode() string {
	return `@interactivetreemap.Treemap(interactivetreemap.Config{
  Label: "Basic treemap example",
  Caption: "File system usage in KB. Select a directory; use the breadcrumb to return.",
  Nodes: []*interactivetreemap.Node{
    {Name: "d1", Class: "directory", Children: []*interactivetreemap.Node{
      {Name: "f1", Value: 1000, Class: "file"},
    }},
    {Name: "d2", Class: "directory", Children: []*interactivetreemap.Node{
      {Name: "f1", Value: 100}, {Name: "f2", Value: 300}, {Name: "f3", Value: 200},
    }},
    {Name: "d3", Class: "directory", Children: deterministicFiles},
    {Name: "f1", Value: 450, Class: "file"},
  },
  Navigation: interactivetreemap.NavigationDrillDown,
  Roam: interactivetreemap.RoamEnabled,
  LabelOptions: &chart.LabelOptions{Show: chart.Bool(true), Position: "inside"},
  UpperLabel: &chart.LabelOptions{Show: chart.Bool(true)},
  Breadcrumb: &interactivetreemap.Breadcrumb{Show: chart.Bool(true)},
  NodeStyle: interactivetreemap.NodeStyle{BorderWidth: 1, GapWidth: 1},
  LeafDepth: chart.Int(1),
  Width: "100%",
  Height: "500px",
})`
}

func interactiveParallelCode() string {
	return `@interactiveparallel.Parallel(interactiveparallel.Config{
  Label: "Multi Series parallel coordinates",
  Dimensions: []interactiveparallel.Dimension{
    {Name: "Date", Range: &interactiveparallel.Range{Max: chart.Float(31)},
      Inverse: true, NameLocation: interactiveparallel.NameStart},
    {Name: "AQI"}, {Name: "PM2.5"}, {Name: "PM10"}, {Name: "CO"},
    {Name: "NO2"}, {Name: "SO2"},
    {Name: "Level", Categories: []string{
      "Good", "Moderate", "Lightly", "Moderately", "Heavily", "Severely",
    }},
  },
  Series: []interactiveparallel.Series{
    {Name: "Beijing", Observations: beijingObservations},
    {Name: "Guangzhou", Observations: guangzhouObservations},
    {Name: "Shanghai", Observations: shanghaiObservations},
  },
  Width: "900px", Height: "500px",
  Options: chart.ChartOptions{
    Title: &chart.TitleOptions{Text: "Multi Series"},
  },
  Style: charttheme.Style{Class: "max-w-full"},
})`
}

func interactiveThemeRiverCode() string {
	return `@interactivethemeriver.ThemeRiver(interactivethemeriver.Config{
  Label: "ThemeRiver-SingleAxis-Time",
  Caption: "Six named streams across aligned daily values.",
  Streams: []interactivethemeriver.Stream{
    {Name: "DQ", Class: "stream-dq", Points: []interactivethemeriver.Point{
      {Time: time.Date(2015, time.November, 8, 0, 0, 0, 0, time.UTC), Value: 10},
      {Time: time.Date(2015, time.November, 9, 0, 0, 0, 0, time.UTC), Value: 15},
      // Remaining aligned dates continue through 28 November.
    }},
    // TY, SS, QG, SY, and DD use the same aligned dates.
  },
  Layout: interactivethemeriver.Layout{BottomPercent: chart.Float(10)},
  Width: "100%", Height: "500px",
  Options: chart.ChartOptions{
    Title: &chart.TitleOptions{Text: "ThemeRiver-SingleAxis-Time"},
    Tooltip: &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "axis"},
  },
})`
}

func interactiveWordCloudCode() string {
	return `@interactivewordcloud.WordCloud(interactivewordcloud.Config{
  Label: "star shape",
  Series: interactivewordcloud.Series{
    Name: "wordcloud",
    Words: []interactivewordcloud.Word{
      {Name: "Sam S Club", Value: 10000, Class: "retail"},
      {Name: "Macys", Value: 6181, Color: "#ff8a3d"},
    },
    Options: interactivewordcloud.SeriesOptions{
      Shape: interactivewordcloud.ShapeStar,
      SizeRange: &interactivewordcloud.SizeRange{Min: 14, Max: 80},
    },
  },
})`
}
