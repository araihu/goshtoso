package components

import (
	"math"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
	"github.com/araihu/goshtoso-charts/components/line"
)

const chartsShowcaseLineFormula = "t = i / 1000; x = (1 + 0.25 × cos(75 × t)) × cos(t); y = (1 + 0.25 × cos(75 × t)) × sin(t); z = t + 2 × sin(75 × t)"

type chartsShowcaseKind string

const (
	chartsShowcaseStatic      chartsShowcaseKind = "static"
	chartsShowcaseInteractive chartsShowcaseKind = "interactive"
	chartsShowcaseLine3D      chartsShowcaseKind = "line-3d"
)

func parseChartsShowcaseKind(value string) chartsShowcaseKind {
	switch chartsShowcaseKind(value) {
	case chartsShowcaseStatic, chartsShowcaseInteractive, chartsShowcaseLine3D:
		return chartsShowcaseKind(value)
	default:
		return chartsShowcaseLine3D
	}
}

func chartsShowcaseTitle(kind chartsShowcaseKind) string {
	switch kind {
	case chartsShowcaseStatic:
		return "Static Goshtoso Charts showcase"
	case chartsShowcaseInteractive:
		return "Interactive Goshtoso Charts showcase"
	default:
		return "Interactive 3D Goshtoso Charts showcase"
	}
}

func chartsShowcaseHeading(kind chartsShowcaseKind) string {
	switch kind {
	case chartsShowcaseStatic:
		return "Server-rendered SVG"
	case chartsShowcaseInteractive:
		return "Interactive line chart"
	default:
		return "Charts for every use case"
	}
}

// ChartsShowcasePageForQuery resolves the bounded public showcase variant.
func ChartsShowcasePageForQuery(value string) templ.Component {
	return ChartsShowcasePageFor(parseChartsShowcaseKind(value))
}

// ChartsShowcaseFrameForQuery resolves the bounded public showcase frame.
func ChartsShowcaseFrameForQuery(value string) templ.Component {
	return ChartsShowcaseFrameFor(parseChartsShowcaseKind(value))
}

func chartsShowcaseStaticLine() line.Instance {
	return line.Line(line.Config{
		Label:   "Weekly request latency",
		Caption: "A server-rendered SVG chart with no browser chart runtime.",
		Labels:  []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Series: []line.Series{
			{Name: "p50", Values: []float64{28, 31, 29, 35, 33, 30, 27}},
			{Name: "p95", Values: []float64{52, 61, 57, 70, 66, 58, 49}},
		},
		Width: 900, Height: 360,
		Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted},
		Export:   &chartcontrol.ExportOptions{Disabled: true},
	})
}

func chartsShowcaseInteractiveLine() interactive.Instance {
	return interactive.Line(interactive.LineConfig{
		Label:   "Live deployment duration trend",
		Caption: "Hover and inspect values rendered by the local interactive runtime.",
		XAxis:   []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Series: []interactive.LineSeries{
			{Name: "Production", Data: []interactive.LineData{{Value: 8}, {Value: 12}, {Value: 10}, {Value: 17}, {Value: 13}, {Value: 9}, {Value: 11}}},
			{Name: "Staging", Data: []interactive.LineData{{Value: 5}, {Value: 7}, {Value: 6}, {Value: 9}, {Value: 8}, {Value: 6}, {Value: 7}}},
		},
		Width: "100%", Height: "20rem",
		Options: interactive.ChartOptions{
			Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted},
			Export:   &chartcontrol.ExportOptions{Disabled: true},
		},
	})
}

func chartsShowcaseLine() interactive.Instance {
	points := make([]interactive.Point3D, 0, 25000)
	for index := 0; index < 25000; index++ {
		t := float64(index) / 1000
		radius := 1 + 0.25*math.Cos(75*t)
		points = append(points, interactive.Point3D{
			X: radius * math.Cos(t),
			Y: radius * math.Sin(t),
			Z: t + 2*math.Sin(75*t),
		})
	}

	return interactive.Line3D(interactive.Line3DConfig{
		Label:   "A parametric line rendered by Goshtoso Charts",
		Caption: "",
		Series: []interactive.Line3DSeries{{
			Name: "line3D", Points: points,
		}},
		VisualRange: &interactive.Line3DVisualRange{
			Min: 0, Max: 30, Calculable: interactive.Bool(false),
		},
		Grid: interactive.Line3DGrid{
			View: &interactive.Line3DView{AutoRotate: interactive.Bool(true)},
		},
		DataSummary: interactive.Line3DDataSummary{
			Formula:   chartsShowcaseLineFormula,
			Parameter: "t", ParameterMin: 0, ParameterMax: 24.999,
		},
		Width:  "100%",
		Height: "22rem",
		Options: interactive.ChartOptions{
			Animation: interactive.Bool(false),
			Legend:    &interactive.LegendOptions{Show: interactive.Bool(false)},
			Controls:  chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted},
			Export:    &chartcontrol.ExportOptions{Disabled: true},
		},
		RootAttrs: templ.Attributes{"data-showcase-component": "line-3d"},
	})
}
