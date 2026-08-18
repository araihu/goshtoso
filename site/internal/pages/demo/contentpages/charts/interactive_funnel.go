package charts

import (
	"math/rand"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivefunnel "github.com/araihu/goshtoso-charts/components/interactive/funnel"
)

const (
	interactiveFunnelSeed = int64(1)
)

var interactiveFunnelDimensions = []string{"Visit", "Add", "Order", "Payment", "Deal"}

// fixedInteractiveFunnelData reproduces the helper's call order with a local
// seed. Stage order and the [0,50) domain are upstream behavior; concrete
// values are deterministic documentation fixtures, not upstream constants.
func fixedInteractiveFunnelData(callIndex int) []interactivefunnel.Data {
	if callIndex < 0 {
		panic("interactive Funnel call index must be nonnegative")
	}
	rng := rand.New(rand.NewSource(interactiveFunnelSeed))
	var data []interactivefunnel.Data
	for call := 0; call <= callIndex; call++ {
		data = make([]interactivefunnel.Data, len(interactiveFunnelDimensions))
		for index, name := range interactiveFunnelDimensions {
			data[index] = interactivefunnel.Data{Name: name, Value: float64(rng.Intn(50))}
		}
	}
	return data
}

func interactiveFunnelOptions(title, filename string) chart.ChartOptions {
	options := controlledOptions(title, filename)
	options.Legend = &chart.LegendOptions{Show: new(true), Left: "center", Bottom: "0"}
	options.Tooltip = &chart.TooltipOptions{Show: new(true), Trigger: "item"}
	return options
}

func sampleInteractiveFunnel() interactivefunnel.Config {
	return interactivefunnel.Config{
		Label:   "Basic five-stage funnel",
		Caption: "Five deterministic values preserve the upstream source sequence in the exact table and [0,50) value domain; the chart keeps the upstream default descending-by-value order.",
		Series:  []interactivefunnel.Series{{Name: "Analytics", Data: fixedInteractiveFunnelData(0)}},
		Width:   "100%", Height: "420px",
		Options: interactiveFunnelOptions("basic funnel example", "basic-five-stage-funnel"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractiveFunnelLabels() interactivefunnel.Config {
	return interactivefunnel.Config{
		Label:   "Funnel with left labels",
		Caption: "Every stage label remains visible to the left of its shape; the exact table keeps source order while the chart keeps the upstream default descending-by-value order.",
		Series:  []interactivefunnel.Series{{Name: "Analytics", Data: fixedInteractiveFunnelData(1)}},
		Width:   "100%", Height: "420px",
		Options:       interactiveFunnelOptions("show label", "funnel-left-labels"),
		SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true), Position: "left"}},
		Style:         charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func interactiveFunnelBaseCode() string {
	return `@interactivefunnel.Funnel(interactivefunnel.Config{
  Label: "Basic five-stage funnel",
  Series: []interactivefunnel.Series{{
    Name: "Analytics",
    Data: []interactivefunnel.Data{
      {Name: "Visit", Value: 31},
      {Name: "Add", Value: 37},
      {Name: "Order", Value: 47},
      {Name: "Payment", Value: 9},
      {Name: "Deal", Value: 31},
    },
  }},
  Options: chart.ChartOptions{
    Title: &chart.TitleOptions{Text: "basic funnel example"},
  },
})`
}

func interactiveFunnelLabelsCode() string {
	return `@interactivefunnel.Funnel(interactivefunnel.Config{
  Label: "Funnel with left labels",
  Series: []interactivefunnel.Series{{
    Name: "Analytics",
    Data: []interactivefunnel.Data{
      {Name: "Visit", Value: 18},
      {Name: "Add", Value: 25},
      {Name: "Order", Value: 40},
      {Name: "Payment", Value: 6},
      {Name: "Deal", Value: 0},
    },
  }},
  SeriesOptions: chart.SeriesOptions{
    Label: &chart.LabelOptions{
      Show:     chart.Bool(true),
      Position: "left",
    },
  },
})`
}
