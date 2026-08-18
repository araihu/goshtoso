package charts

import (
	"time"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"
)

const (
	interactiveLineTimeUpstreamPath     = "examples/line.go (lineTime and generateLineItemsTwoAxis)"
	interactiveLineTimeUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
)

func sampleInteractiveLineTime() interactiveline.Config {
	axis := make([]time.Time, 0, 50)
	for offset := range 50 {
		// Day zero intentionally normalizes to January 31, as in source sample.
		axis = append(axis, time.Date(2025, time.February, offset, 0, 0, 0, 0, time.UTC))
	}
	return interactiveline.Config{
		Label: "Temporal X axis", Caption: "Fifty deterministic UTC time/value observations.",
		TimeAxis: &interactiveline.TimeAxis{
			Values: axis, Minimum: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		Series: []interactiveline.Series{{Name: "Category A", Data: deterministicLineTimeData(50)}},
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: "temporal X axis", Subtitle: "time.Date as X axis values"},
			Legend:   &chart.LegendOptions{Bottom: "0"},
			Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "axis"},
			YAxis:    &chart.AxisOptions{Min: chart.Float(0), Max: chart.Float(200)},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "temporal-x-axis"},
		},
	}
}

func deterministicLineTimeData(count int) []interactiveline.Data {
	values := make([]interactiveline.Data, count)
	state := uint32(7)
	for index := range values {
		state = state*1664525 + 1013904223
		values[index] = interactiveline.Data{Value: float64(100 + state%20)}
	}
	return values
}

func interactiveLineTimeCode() string {
	return `@interactiveline.Line(interactiveline.Config{
  Label: "Temporal X axis",
	  TimeAxis: &interactiveline.TimeAxis{
	    Minimum: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
    Values: []time.Time{time.Date(2025, time.February, 0, 0, 0, 0, 0, time.UTC)},
  },
  Series: []interactiveline.Series{{Name: "Category A", Data: []interactiveline.Data{{Value: 107}}}},
})`
}
