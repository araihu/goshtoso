package charts

import (
	"math/rand"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivefunnel "github.com/araihu/goshtoso-charts/components/interactive/funnel"
)

const (
	interactiveFunnelUpstreamPath     = "examples/funnel.go"
	interactiveFunnelUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveFunnelUpstreamSHA256   = "c532e6490bad284b4b6a5dec20825359abc795a8ee9f3bb5febbcfb4e0cd2d55"
	interactiveFunnelSeed             = int64(1)
)

type interactiveFunnelCoverageEntry struct {
	Name      string
	Treatment string
}

func interactiveFunnelUpstreamCoverage() []interactiveFunnelCoverageEntry {
	return []interactiveFunnelCoverageEntry{
		{Name: "funnelBase", Treatment: "basic five-stage funnel"},
		{Name: "funnelShowLabel", Treatment: "visible stage labels positioned left"},
	}
}

type interactiveFunnelSourceSpan struct {
	Name   string
	Lines  string
	SHA256 string
	Role   string
}

func interactiveFunnelSourceSpans() []interactiveFunnelSourceSpan {
	return []interactiveFunnelSourceSpan{
		{Name: "dimensions", Lines: "13", SHA256: "bd5b4e6a9c429f461f802686cfe0539b660b85c12fc382656d0097d07d1f7e83", Role: "ordered stage labels"},
		{Name: "genFunnelKvItems", Lines: "15–21", SHA256: "9bf2059b03ae41f499ad99ac7ece0ac9deec70fe0fc3df8052310b1f8a64ac1c", Role: "random data helper adapted to a local fixed seed"},
		{Name: "funnelBase", Lines: "22–30", SHA256: "88c9efbc1bdda11af5c7cbad673b7cd568941ed9248a6428ec5ea5c45184fd45", Role: "basic example"},
		{Name: "funnelShowLabel", Lines: "33–47", SHA256: "ed198502f5b56653897070b7ba9b7fb862a8ceb4255ac4242cd817d0be0e23d7", Role: "left-label example"},
		{Name: "FunnelExamples.Examples", Lines: "51–63", SHA256: "6c55c7a63e033a5b0f6de045c16c31d16eee67a134000a1e240da88ffbb1ca97", Role: "page composition only"},
	}
}

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
