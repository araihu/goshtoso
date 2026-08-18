package charts

import (
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivepie "github.com/araihu/goshtoso-charts/components/interactive/pie"
)

const (
	interactivePieUpstreamPath     = "examples/pie.go"
	interactivePieUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactivePieUpstreamSHA256   = "a59bb6f11818d4175d033f025f00a58e6a191eff5acf30f0e0cd5f98cd493ada"

	pieCoverageExample = "example"
)

type interactivePieCoverageEntry struct {
	Name      string
	Status    string
	Treatment string
}

func interactivePieUpstreamCoverage() []interactivePieCoverageEntry {
	return []interactivePieCoverageEntry{
		{Name: "pieBase", Status: pieCoverageExample, Treatment: "basic seasonal distribution"},
		{Name: "pieShowLabel", Status: pieCoverageExample, Treatment: "visible name and value labels"},
		{Name: "pieRadius", Status: pieCoverageExample, Treatment: "donut radii"},
		{Name: "pieRadiusWithPadAngle", Status: pieCoverageExample, Treatment: "sector padding, offset center, vertical legend, hidden labels, and share tooltip"},
		{Name: "pieRoseArea", Status: pieCoverageExample, Treatment: "area rose"},
		{Name: "pieRoseRadius", Status: pieCoverageExample, Treatment: "radius rose"},
		{Name: "pieRoseAreaRadius", Status: pieCoverageExample, Treatment: "side-by-side area and radius roses"},
		{Name: "pieInPie", Status: pieCoverageExample, Treatment: "nested non-overlapping roses"},
		{Name: "pieWithDispatchAction", Status: pieCoverageExample, Treatment: "typed rotating emphasis and item tooltip"},
	}
}

type interactivePieSource struct {
	Path   string
	SHA256 string
	Scope  string
}

func interactivePieSupplementarySources() []interactivePieSource {
	return []interactivePieSource{
		{Path: "examples/page_center_layout.go", SHA256: "106456904719dfacfb13adcc1b9e66df83cf28a5a801539bad4d1958554166c9", Scope: "centered page-layout reference"},
		{Path: "examples/page_flex_layout.go", SHA256: "3113b7bdf78a2365ae62502fe86ab001f3ff3034b1d77752c693e95b28a0fd68", Scope: "flex page-layout reference"},
		{Path: "examples/page_none_layout.go", SHA256: "ce38424de2ffeb919661e536c7f44921de098ae14643d4f2975d8e72296c32f8", Scope: "unmanaged page-layout reference"},
	}
}

var interactivePieSeasons = []string{"Spring", "Summer", "Autumn", "Winter"}

// interactivePieData replaces upstream ambient randomness with fixed values
// from a local seed-1 sequence. Names, [0,100) domain, series call order, and
// chart geometry remain unchanged. The upstream "Autumn " typo is trimmed.
func interactivePieData(values [4]float64) []interactivepie.Data {
	data := make([]interactivepie.Data, len(interactivePieSeasons))
	for index, name := range interactivePieSeasons {
		data[index] = interactivepie.Data{Name: name, Value: values[index]}
	}
	return data
}

func interactivePieOptions(title, filename string) chart.ChartOptions {
	options := controlledOptions(title, filename)
	options.Legend = &chart.LegendOptions{Left: "center", Bottom: "0"}
	return options
}

func sampleInteractivePie() interactivepie.Config {
	return interactivepie.Config{
		Label: "Basic seasonal pie", Caption: "Four seasonal values shown as shares of one total.",
		Series: []interactivepie.Series{{
			Name: "Seasons", Data: interactivePieData([4]float64{81, 87, 47, 59}),
		}},
		Width: "100%", Height: "420px",
		Options: interactivePieOptions("Basic pie example", "basic-seasonal-pie"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieLabels() interactivepie.Config {
	return interactivepie.Config{
		Label: "Seasonal values with labels", Caption: "Each sector names its season and exact value; the adjacent table also gives shares.",
		Series: []interactivepie.Series{{
			Name: "Seasons", LabelContent: interactivepie.LabelNameAndValue,
			Data:    interactivePieData([4]float64{81, 18, 25, 40}),
			Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}},
		}},
		Width: "100%", Height: "420px", Options: interactivePieOptions("label options", "seasonal-pie-labels"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieRadius() interactivepie.Config {
	return interactivepie.Config{
		Label: "Seasonal donut", Caption: "A 40% inner radius opens the center while preserving proportional sector angles.",
		Series: []interactivepie.Series{{
			Name: "Seasons", InnerRadius: 40, OuterRadius: 75, LabelContent: interactivepie.LabelNameAndValue,
			Data:    interactivePieData([4]float64{56, 0, 94, 11}),
			Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}},
		}},
		Width: "100%", Height: "420px", Options: interactivePieOptions("radius options", "seasonal-donut"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePiePadded() interactivepie.Config {
	options := interactivePieOptions("radius and pad-angle options", "padded-seasonal-donut")
	options.Legend = &chart.LegendOptions{
		Orient: "vertical", Right: "20%", Top: "center",
		Padding: &chart.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
	}
	options.Tooltip = &chart.TooltipOptions{Show: new(true), Trigger: "item"}
	return interactivepie.Config{
		Label: "Offset padded donut", Caption: "Five-degree sector gaps, an offset center, and a right-side legend preserve the upstream composition.",
		Series: []interactivepie.Series{{
			Name: "Seasons", InnerRadius: 40, OuterRadius: 75, Center: &interactivepie.Center{X: 40, Y: 50}, PadAngle: 5,
			Data:    interactivePieData([4]float64{62, 89, 28, 74}),
			Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(false)}},
		}},
		Width: "100%", Height: "420px", Options: options, TooltipContent: interactivepie.TooltipNameAndShare,
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieRoseArea() interactivepie.Config {
	return interactivepie.Config{
		Label: "Rose area", Caption: "Equal angles map seasonal values to sector area.",
		Series: []interactivepie.Series{{
			Name: "Area", InnerRadius: 40, OuterRadius: 75, RoseMode: interactivepie.RoseArea,
			LabelContent: interactivepie.LabelNameAndValue,
			Data:         interactivePieData([4]float64{11, 45, 37, 6}),
			Options:      chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}},
		}},
		Width: "100%", Height: "420px",
		Options: interactivePieOptions("Rose area", "rose-area"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieRoseRadius() interactivepie.Config {
	return interactivepie.Config{
		Label: "Rose radius", Caption: "Proportional angles remain while seasonal values also change radius.",
		Series: []interactivepie.Series{{
			Name: "Radius", InnerRadius: 30, OuterRadius: 75, RoseMode: interactivepie.RoseRadius,
			LabelContent: interactivepie.LabelNameAndValue,
			Data:         interactivePieData([4]float64{95, 66, 28, 58}),
			Options:      chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}},
		}},
		Width: "100%", Height: "420px",
		Options: interactivePieOptions("Rose radius", "rose-radius"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieNested() interactivepie.Config {
	return interactivepie.Config{
		Label: "Nested seasonal pie", Caption: "A thin outer area ring surrounds an inner radius rose at one shared center.",
		Series: []interactivepie.Series{
			{
				Name: "Outer area", InnerRadius: 50, OuterRadius: 55, RoseMode: interactivepie.RoseArea,
				LabelContent: interactivepie.LabelNameAndValue,
				Data:         interactivePieData([4]float64{87, 31, 29, 56}),
				Options:      chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}},
			},
			{
				Name: "Inner radius", OuterRadius: 45, RoseMode: interactivepie.RoseRadius,
				Data: interactivePieData([4]float64{37, 31, 85, 26}),
			},
		},
		Width: "100%", Height: "440px",
		Options: interactivePieOptions("Pie in pie", "nested-seasonal-pie"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePiePairedRoses() interactivepie.Config {
	return interactivepie.Config{
		Label: "Area and radius roses", Caption: "Two centers compare equal-angle area encoding with proportional-angle radius encoding.",
		Series: []interactivepie.Series{
			{
				Name: "Area", InnerRadius: 30, OuterRadius: 75, Center: &interactivepie.Center{X: 25, Y: 50}, RoseMode: interactivepie.RoseArea,
				LabelContent: interactivepie.LabelNameAndValue, Data: interactivePieData([4]float64{47, 47, 87, 88}),
				Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}},
			},
			{
				Name: "Radius", InnerRadius: 30, OuterRadius: 75, Center: &interactivepie.Center{X: 75, Y: 50}, RoseMode: interactivepie.RoseRadius,
				LabelContent: interactivepie.LabelNameAndValue, Data: interactivePieData([4]float64{90, 15, 41, 8}),
				Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: new(true)}},
			},
		},
		Width: "100%", Height: "440px", Options: interactivePieOptions("roseType(area) vs roseType(radius)", "paired-seasonal-roses"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieAutoEmphasis() interactivepie.Config {
	options := interactivePieOptions("dispatch action", "rotating-seasonal-emphasis")
	options.Title.Right = "40%"
	options.Legend = &chart.LegendOptions{Orient: "vertical", Left: "left", Top: "center"}
	options.Tooltip = &chart.TooltipOptions{Show: new(true), Trigger: "item"}
	return interactivepie.Config{
		Label: "Rotating seasonal emphasis", Caption: "One sector at a time receives emphasis and a tooltip; reduced-motion preference stops the cycle.",
		Series: []interactivepie.Series{{
			Name: "Seasons", OuterRadius: 55, Center: &interactivepie.Center{X: 50, Y: 60},
			LabelContent: interactivepie.LabelNameAndValue, Data: interactivePieData([4]float64{13, 90, 94, 63}),
			Options: chart.SeriesOptions{
				Label:    &chart.LabelOptions{Show: new(true)},
				Emphasis: &chart.EmphasisOptions{ItemStyle: &chart.ItemStyle{ShadowBlur: 10, ShadowColor: "rgba(0, 0, 0, 0.5)"}},
			},
		}},
		Width: "100%", Height: "440px", Options: options, TooltipContent: interactivepie.TooltipNameAndShare,
		AutoEmphasis: &interactivepie.AutoEmphasisOptions{IntervalMilliseconds: 1000},
		Style:        charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieSelected() interactivepie.Config {
	options := interactivePieOptions("selected sector", "selected-seasonal-sector")
	options.Controls = chartcontrol.Options{Fullscreen: true}
	return interactivepie.Config{
		Label: "Selectable seasonal donut", Caption: "Spring starts selected; readers can toggle any sector without changing the underlying values.",
		Series: []interactivepie.Series{{
			Name: "Seasons", InnerRadius: 35, OuterRadius: 70, Selectable: true,
			Data: []interactivepie.Data{
				{Name: "Spring", Value: 81, Selected: true}, {Name: "Summer", Value: 18},
				{Name: "Autumn", Value: 25}, {Name: "Winter", Value: 40},
			},
		}},
		Width: "100%", Height: "420px", Options: options, Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}
