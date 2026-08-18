package charts

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/candlestick"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/funnel"
	"github.com/araihu/goshtoso-charts/components/heatmap"
	"github.com/araihu/goshtoso-charts/components/line"
	"github.com/araihu/goshtoso-charts/components/radar"
	"github.com/araihu/goshtoso-charts/components/scatter"
	charttable "github.com/araihu/goshtoso-charts/components/table"
	"github.com/araihu/goshtoso-charts/components/violin"
	"github.com/araihu/goshtoso/components/codeblock"
)

type deterministicLCG struct{ state uint64 }

func (generator *deterministicLCG) next() float64 {
	generator.state = generator.state*6364136223846793005 + 1442695040888963407
	return float64(generator.state>>33) / float64(1<<31)
}

func (generator *deterministicLCG) normal(mean, standardDeviation float64) float64 {
	first, second := generator.next(), generator.next()
	if first < 1e-10 {
		first = 1e-10
	}
	return mean + standardDeviation*math.Sqrt(-2*math.Log(first))*math.Cos(2*math.Pi*second)
}

func gettingStartedCodeBlock(language, label, code string) codeblock.Config {
	return codeblock.Config{Language: language, Label: label, Code: code}
}

const (
	gettingStartedInstallCode = `go get github.com/araihu/goshtoso-charts`
	gettingStartedAssetsCode  = `import chartassets "github.com/araihu/goshtoso-charts/assets"

mux.Handle("GET "+chartassets.Prefix, chartassets.Handler())`
	gettingStartedDependenciesCode = `import "github.com/araihu/goshtoso-charts/components/dependencies"

templ Layout() {
  <head>
    @dependencies.Dependencies()
  </head>
}`
	gettingStartedStaticCode = `@line.Line(line.Config{
  Label: "Request latency",
  Labels: []string{"Mon", "Tue", "Wed"},
  Series: []line.Series{{Name: "p95 (ms)", Values: []float64{42, 47, 44}}},
})`
	gettingStartedInteractiveCode = `import interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"

@interactivebar.Bar(interactivebar.Config{
  Label: "Deployments",
  XAxis: []string{"Mon", "Tue", "Wed"},
  Series: []interactivebar.Series{{
    Name: "Production",
    Data: []interactivebar.Data{{Value: 3}, {Value: 5}, {Value: 4}},
  }},
})`
)

func sampleLatency() line.Config {
	return line.Config{
		Label:    "HTTPS monitor latency in milliseconds",
		Caption:  "Median latency, last seven checks.",
		Labels:   []string{"08:00", "08:01", "08:02", "08:03", "08:04", "08:05", "08:06"},
		Series:   []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 900, 51, 2_000, 44, 46}}},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "https-monitor-latency"},
	}
}

func sampleDualAxisLine() line.Config {
	return line.Config{
		Label:   "Dual Axis Line",
		Caption: "Two series retain independent left and right scales.",
		Title:   line.Title{Text: "Dual Axis Line"},
		Labels:  []string{"A", "B", "C", "D", "E", "F", "G"},
		Series: []line.Series{
			{Name: "Left Series", Values: []float64{120, 132, 101, 134, 90, 230, 210}},
			{Name: "Right Series", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}, YAxisIndex: 1},
		},
		YAxes:    []line.Axis{{}, {}},
		Width:    600,
		Height:   400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "dual-axis-line"},
	}
}

func sampleAreaLine() line.Config {
	minimum := 0.0
	noGap := false
	return line.Config{
		Label:    "Line",
		Caption:  "A filled area emphasizes magnitude across the seven ordered labels.",
		Title:    line.Title{Text: "Line"},
		Labels:   []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Series:   []line.Series{{Name: "Email", Values: []float64{120, 132, 101, 134, 90, 230, 210}}},
		Area:     line.AreaOptions{Enabled: true, Opacity: 150.0 / 255.0},
		XAxis:    line.CategoryAxisOptions{BoundaryGap: &noGap},
		Legend:   line.LegendOptions{Padding: line.Padding{Top: 5, Bottom: 10}},
		YAxes:    []line.Axis{{Min: &minimum}},
		Width:    600,
		Height:   400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "filled-area-line"},
	}
}

func sampleDualAxisLineOverrides() line.Config {
	cfg := sampleDualAxisLine()
	cfg.Label = "Dual Axis Line caller presentation overrides"
	cfg.Series[0].Color = "#14532d"
	cfg.Series[1].Class = "caller-right-series"
	cfg.YAxes[0].Class = "caller-left-axis"
	cfg.YAxes[1].Color = "#7e22ce"
	cfg.Export = &chartcontrol.ExportOptions{Filename: "dual-axis-line-overrides"}
	return cfg
}

func sampleHorizontalWorldPopulation() bar.Config {
	return bar.Config{
		Label:       "World population by reporting series",
		Caption:     "Population values reported for 2011 and 2012.",
		Title:       "World Population",
		Orientation: bar.OrientationHorizontal,
		Labels:      []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
		Series: []bar.Series{
			{Name: "2011", Values: []float64{10, 30, 50, 70, 90, 110, 130}},
			{Name: "2012", Values: []float64{20, 40, 60, 80, 100, 120, 140}},
		},
		Padding:  bar.Padding{Top: 20, Right: 40, Bottom: 20, Left: 20},
		Width:    600,
		Height:   400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "world-population"},
	}
}

func sampleBarReferences() bar.Config {
	annotations := bar.References{Average: true, Minimum: true, Maximum: true, Format: bar.ValueFormatHumanized}
	return bar.Config{
		Label:   "Monthly rainfall and evaporation reference annotations",
		Caption: "Monthly values with average lines and minimum and maximum reference points; adjacent evidence lists every value and computed reference.",
		Labels:  []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		Series: []bar.Series{
			{Name: "Rainfall", Values: []float64{2.0, 4.9, 7.0, 23.2, 25.6, 76.7, 135.6, 162.2, 32.6, 20.0, 6.4, 3.3}, References: annotations},
			{Name: "Evaporation", Values: []float64{2.6, 5.9, 9.0, 26.4, 28.7, 70.7, 175.6, 182.2, 48.7, 18.8, 6.0, 2.3}, References: annotations},
		},
		Legend: bar.LegendOptions{Placement: bar.LegendPlacementEnd, Overlay: true},
		Width:  600, Height: 400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "monthly-rainfall-and-evaporation"},
	}
}

func sampleBasicRadar() radar.Config {
	return radar.Config{
		Label:   "Basic radar chart",
		Caption: "Allocated budget and actual spending across six shared dimensions.",
		Indicators: []radar.Indicator{
			{Name: "Sales", Max: 6500},
			{Name: "Administration", Max: 16000},
			{Name: "Information Technology", Max: 30000},
			{Name: "Customer Support", Max: 38000},
			{Name: "Development", Max: 52000},
			{Name: "Marketing", Max: 25000},
		},
		Series: []radar.Series{
			{Name: "Allocated Budget", Values: []float64{4200, 3000, 20000, 35000, 50000, 18000}},
			{Name: "Actual Spending", Values: []float64{5000, 14000, 28000, 26000, 42000, 21000}},
		},
		Title:  radar.TitleOptions{Text: "Basic Radar Chart", FontSize: 16},
		Legend: radar.LegendOptions{Horizontal: radar.PlacementEnd},
		Width:  600,
		Height: 400,
		RootAttrs: templ.Attributes{
			"data-static-radar-exhaustion": "1fe31b06",
			"data-goshtoso-candidate":      "radar-basic-0cf8dbdd72f6a398",
		},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-radar-chart"},
	}
}

func sampleBasicCandlestick() candlestick.Config {
	return candlestick.Config{
		Label:      "Seven-day stock price",
		Caption:    "Daily open, high, low, and close values; Day 4 decreases while the other days increase.",
		Title:      "Candlestick Chart",
		SeriesName: "Stock Price",
		Data: []candlestick.Datum{
			{Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
			{Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
			{Label: "Day 3", Open: 112, High: 118, Low: 108, Close: 115},
			{Label: "Day 4", Open: 115, High: 120, Low: 104, Close: 108},
			{Label: "Day 5", Open: 108, High: 113, Low: 105, Close: 109},
			{Label: "Day 6", Open: 109, High: 116, Low: 106, Close: 114},
			{Label: "Day 7", Open: 114, High: 121, Low: 111, Close: 119},
		},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-candlestick-chart"},
	}
}

func sampleCandlestickOptionFuncBasic() candlestick.Config {
	cfg := sampleBasicCandlestick()
	cfg.Label = "Eight-day stock price"
	cfg.Caption = "The second upstream construction style extends the basic series with Day 8 while retaining the same renderer-neutral component."
	cfg.Title = "Basic Candlestick Chart"
	cfg.Data = append(append([]candlestick.Datum(nil), cfg.Data...), candlestick.Datum{Label: "Day 8", Open: 119, High: 125, Low: 116, Close: 122})
	cfg.Options = candlestick.Options{TitleFontSize: 18, YUnit: 1, Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}}
	cfg.Width, cfg.Height = 800, 600
	cfg.RootAttrs = templ.Attributes{"data-goshtoso-candidate": "candlestick-optionfunc-aad7ab0297061baa"}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "basic-candlestick-eight-days"}
	return cfg
}

func sampleCandlestickMultipleSeries() candlestick.Config {
	seriesData := func(base float64) []candlestick.Datum {
		return []candlestick.Datum{
			{Label: "Day 1", Open: base, High: base + 10, Low: base - 5, Close: base + 5},
			{Label: "Day 2", Open: base + 5, High: base + 15, Low: base, Close: base + 12},
			{Label: "Day 3", Open: base + 12, High: base + 18, Low: base + 8, Close: base + 15},
			// Upstream Low exceeds Close. Use the same corrected bearish range as the basic example.
			{Label: "Day 4", Open: base + 15, High: base + 20, Low: base + 4, Close: base + 8},
			{Label: "Day 5", Open: base + 8, High: base + 13, Low: base + 5, Close: base + 9},
		}
	}
	return candlestick.Config{
		Label:   "Three aligned stock series",
		Caption: "Stock A uses filled bodies, Stock B uses traditional hollow increases, and Stock C uses outlines. The invalid upstream Day 4 lows are corrected below their closes.",
		Series: []candlestick.Series{
			{Name: "Stock A", Data: seriesData(100), BodyStyle: candlestick.BodyStyleFilled},
			{Name: "Stock B", Data: seriesData(150), BodyStyle: candlestick.BodyStyleTraditional},
			{Name: "Stock C", Data: seriesData(200), BodyStyle: candlestick.BodyStyleOutline},
		},
		Options:   candlestick.Options{YUnit: 1, Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}},
		Width:     1000,
		Height:    700,
		RootAttrs: templ.Attributes{"data-static-candlestick-exhaustion": "1fe31b06", "data-goshtoso-candidate": "candlestick-multiple-f132f40ac3e920a8"},
		Controls:  chartcontrol.Options{Fullscreen: true},
		Export:    &chartcontrol.ExportOptions{Filename: "candlestick-multiple-series"},
	}
}

func sampleCandlestickGeometry() candlestick.Config {
	cfg := sampleCandlestickMultipleSeries()
	cfg.Label = "Candlestick geometry controls"
	cfg.Caption = "Narrower bodies, explicit inter-series spacing, thicker wicks, chart-level wick visibility, and a per-series override use typed geometry options."
	show, margin := true, .08
	cfg.Options.Geometry = candlestick.Geometry{CandleWidth: .58, WickWidth: 2, SeriesGap: &margin, ShowWicks: &show}
	hide := false
	cfg.Series[2].ShowWicks = &hide
	cfg.Export = &chartcontrol.ExportOptions{Filename: "candlestick-geometry"}
	return cfg
}

func sampleCandlestickAggregation() candlestick.Config {
	return candlestick.Config{
		Label:      "Candlestick aggregation",
		Caption:    "Fifteen one-minute observations appear above three five-minute windows. Each aggregated candle keeps the first open, highest high, lowest low, and final close in its window.",
		Title:      "1-Minute Candles (Before Aggregation)",
		SeriesName: "1-Minute",
		Data: []candlestick.Datum{
			{Label: "1", Open: 100, High: 102, Low: 99, Close: 101}, {Label: "2", Open: 101, High: 103, Low: 100, Close: 102},
			{Label: "3", Open: 102, High: 105, Low: 101, Close: 104}, {Label: "4", Open: 104, High: 106, Low: 103, Close: 105},
			{Label: "5", Open: 105, High: 107, Low: 104, Close: 106}, {Label: "6", Open: 106, High: 108, Low: 105, Close: 107},
			{Label: "7", Open: 107, High: 109, Low: 106, Close: 108}, {Label: "8", Open: 108, High: 110, Low: 107, Close: 109},
			{Label: "9", Open: 109, High: 111, Low: 108, Close: 110}, {Label: "10", Open: 110, High: 112, Low: 109, Close: 111},
			{Label: "11", Open: 111, High: 113, Low: 110, Close: 112}, {Label: "12", Open: 112, High: 114, Low: 111, Close: 113},
			{Label: "13", Open: 113, High: 115, Low: 112, Close: 114}, {Label: "14", Open: 114, High: 116, Low: 113, Close: 115},
			{Label: "15", Open: 115, High: 117, Low: 114, Close: 116},
		},
		Aggregation: candlestick.AggregationOptions{WindowSize: 5, Title: "5-Minute Aggregated Candles", SeriesName: "5-Minute"},
		Options:     candlestick.Options{TitleFontSize: 16, YUnit: 1, Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}},
		Width:       1200,
		Height:      800,
		RootAttrs:   templ.Attributes{"data-goshtoso-candidate": "candlestick-aggregation-ba7d1d31fef54f79"},
		Controls:    chartcontrol.Options{Fullscreen: true},
		Export:      &chartcontrol.ExportOptions{Filename: "candlestick-aggregation"},
	}
}

func sampleCandlestickPatterns() candlestick.Config {
	return candlestick.Config{
		Label: "Candlestick patterns", Caption: "Ten OHLC observations demonstrate normal, doji, hammer, engulfing, inverted-hammer, and recovery candles. Pattern names and exact OHLC values remain available in text.",
		Title: "Candlestick Patterns", SeriesName: "Stock Price with Patterns",
		Data: []candlestick.Datum{
			{Label: "1", Open: 100, High: 110, Low: 95, Close: 105}, {Label: "2", Open: 105, High: 108, Low: 102, Close: 105.1},
			{Label: "3", Open: 108, High: 109, Low: 98, Close: 107}, {Label: "4", Open: 107, High: 108, Low: 103, Close: 104},
			{Label: "5", Open: 102, High: 115, Low: 101, Close: 113}, {Label: "6", Open: 113, High: 125, Low: 112, Close: 114},
			{Label: "7", Open: 114, High: 118, Low: 113, Close: 117}, {Label: "8", Open: 119, High: 120, Low: 108, Close: 110},
			{Label: "9", Open: 110, High: 113, Low: 107, Close: 109.9}, {Label: "10", Open: 109, High: 118, Low: 108, Close: 116},
		},
		Patterns: candlestick.PatternOptions{Selection: candlestick.PatternSelectionAll, References: []candlestick.CloseReferenceType{candlestick.CloseReferenceAverage, candlestick.CloseReferenceMinimum}},
		Options:  candlestick.Options{TitleFontSize: 16, YUnit: 1, Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}},
		Width:    900, Height: 650, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "candlestick-patterns"},
	}
}

func sampleCandlestickCorePatterns() candlestick.Config {
	cfg := sampleCandlestickPatterns()
	cfg.Label, cfg.Title, cfg.SeriesName, cfg.Width, cfg.Height = "Core candlestick patterns", "Core Patterns", "Important Patterns", 800, 400
	cfg.Patterns = candlestick.PatternOptions{Selection: candlestick.PatternSelectionCore}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "candlestick-patterns-core"}
	return cfg
}

func sampleCandlestickPatternLabels() candlestick.Config {
	cfg := sampleCandlestickPatterns()
	cfg.Label, cfg.Title, cfg.SeriesName, cfg.Width, cfg.Height = "Candlestick pattern labels", "Custom Pattern Formatting", "Custom Format", 800, 400
	cfg.Patterns = candlestick.PatternOptions{Selection: candlestick.PatternSelectionAll, PreferLabels: true, Label: candlestick.PatternLabelStyle{Text: candlestick.PatternLabelTextNameWithCount, Color: "#ffffff", BackgroundColor: "#1d4ed8", FontSize: 8, CornerRadius: 2}}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "candlestick-pattern-labels"}
	return cfg
}

func sampleCandlestickBullishPatterns() candlestick.Config {
	cfg := sampleCandlestickPatterns()
	cfg.Label, cfg.Title, cfg.SeriesName, cfg.Width, cfg.Height = "Bullish candlestick patterns", "Bullish Patterns", "Bullish Only", 800, 400
	cfg.Patterns = candlestick.PatternOptions{Selection: candlestick.PatternSelectionBullish}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "candlestick-patterns-bullish"}
	return cfg
}

func sampleCandlestickBollingerBands() candlestick.Config {
	return candlestick.Config{
		Label:      "Candlestick Chart with Bollinger Bands",
		Caption:    "Close-price Bollinger upper, simple moving average middle, and Bollinger lower lines use a centered five-period window.",
		Title:      "Candlestick Chart with Bollinger Bands",
		SeriesName: "Price",
		Data: []candlestick.Datum{
			{Label: "1", Open: 100, High: 110, Low: 95, Close: 105},
			{Label: "2", Open: 105, High: 115, Low: 100, Close: 112},
			{Label: "3", Open: 112, High: 118, Low: 108, Close: 115},
			{Label: "4", Open: 115, High: 125, Low: 110, Close: 120},
			{Label: "5", Open: 120, High: 130, Low: 115, Close: 125},
			{Label: "6", Open: 125, High: 135, Low: 120, Close: 130},
			{Label: "7", Open: 130, High: 140, Low: 125, Close: 135},
			{Label: "8", Open: 135, High: 145, Low: 130, Close: 140},
			{Label: "9", Open: 140, High: 150, Low: 135, Close: 145},
			{Label: "10", Open: 145, High: 155, Low: 140, Close: 150},
			{Label: "11", Open: 150, High: 160, Low: 145, Close: 148},
			{Label: "12", Open: 148, High: 153, Low: 143, Close: 146},
			{Label: "13", Open: 146, High: 151, Low: 141, Close: 144},
			{Label: "14", Open: 144, High: 149, Low: 139, Close: 142},
			{Label: "15", Open: 142, High: 147, Low: 137, Close: 140},
			{Label: "16", Open: 140, High: 145, Low: 135, Close: 138},
			{Label: "17", Open: 138, High: 143, Low: 133, Close: 136},
			{Label: "18", Open: 136, High: 141, Low: 131, Close: 134},
			{Label: "19", Open: 134, High: 139, Low: 129, Close: 132},
			{Label: "20", Open: 132, High: 137, Low: 127, Close: 130},
		},
		TrendLines: []candlestick.TrendLine{
			{Type: candlestick.TrendTypeBollingerUpper, Period: 5},
			{Type: candlestick.TrendTypeSimpleMovingAverage, Period: 5},
			{Type: candlestick.TrendTypeBollingerLower, Period: 5},
		},
		Options: candlestick.Options{
			TitleFontSize: 18,
			YUnit:         1,
			Padding:       candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
		},
		Width: 800, Height: 600,
		RootAttrs: templ.Attributes{"data-goshtoso-candidate": "candlestick-bollinger-fc218c7fedf84c7a"},
		Controls:  chartcontrol.Options{Fullscreen: true},
		Export:    &chartcontrol.ExportOptions{Filename: "candlestick-bollinger-bands"},
	}
}

func sampleCandlestickPresentationOverrides() candlestick.Config {
	cfg := sampleCandlestickBollingerBands()
	cfg.Label = "Candlestick caller presentation overrides"
	cfg.Options.Increasing.Color = "#14532d"
	cfg.Options.Decreasing.Class = "caller-decreasing-candles"
	cfg.TrendLines[0].Color = "#1d4ed8"
	cfg.TrendLines[1].Class = "caller-middle-band"
	cfg.Export = &chartcontrol.ExportOptions{Filename: "candlestick-presentation-overrides"}
	return cfg
}

func sampleBasicFunnel() funnel.Config {
	return funnel.Config{
		Label:   "Basic funnel",
		Caption: "Seven ordered stages, with every exact value available below the chart.",
		Title:   "Funnel",
		Stages: []funnel.Stage{
			{Label: "Show", Value: 100},
			{Label: "Click", Value: 80},
			{Label: "Visit", Value: 60},
			{Label: "Inquiry", Value: 40},
			{Label: "Order", Value: 20},
			{Label: "Pay", Value: 10},
			{Label: "Cancel", Value: 2},
		},
		Options:   funnel.Options{Legend: funnel.Legend{Padding: funnel.Padding{Left: 100}}},
		RootAttrs: templ.Attributes{"data-static-funnel-source": "a54875c89cd0be43"},
		Controls:  chartcontrol.Options{Fullscreen: true},
		Export:    &chartcontrol.ExportOptions{Filename: "basic-funnel"},
	}
}

func sampleCompactFunnel() funnel.Config {
	return funnel.Config{
		Label:   "Compact five-stage funnel",
		Caption: "Five ordered stages retain the supplementary source sequence and exact values.",
		Title:   "Funnel",
		Stages: []funnel.Stage{
			{Label: "Show", Value: 100},
			{Label: "Click", Value: 80},
			{Label: "Visit", Value: 60},
			{Label: "Inquiry", Value: 40},
			{Label: "Order", Value: 20},
		},
		RootAttrs: templ.Attributes{"data-static-funnel-source": "f2a990a30d9a89a7"},
		Controls:  chartcontrol.Options{Fullscreen: true},
		Export:    &chartcontrol.ExportOptions{Filename: "compact-five-stage-funnel"},
	}
}

func samplePeopleTable() charttable.Config {
	return charttable.Config{
		Label:   "People directory",
		Caption: "Three people, their age, address, tags, and available action.",
		Columns: []charttable.Column{
			{Header: "Name", Span: 2},
			{Header: "Age", Span: 1},
			{Header: "Address", Span: 3},
			{Header: "Tag", Span: 2},
			{Header: "Action", Span: 2},
		},
		Rows: [][]string{
			{"John Brown", "32", "New York No. 1 Lake Park", "nice, developer", "Send Mail"},
			{"Jim Green", "42", "London No. 1 Lake Park", "wow", "Send Mail"},
			{"Joe Black", "32", "Sidney No. 1 Lake Park", "cool, teacher", "Send Mail"},
		},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "people-directory"},
	}
}

func sampleMarketTable() charttable.Config {
	return charttable.Config{
		Label: "Market changes",
		Columns: []charttable.Column{
			{Header: "Name"},
			{Header: "Price", Align: charttable.AlignEnd},
			{Header: "Change", Align: charttable.AlignEnd},
		},
		Rows: [][]string{
			{"Datadog Inc", "97.32", "-7.49%"},
			{"Hashicorp Inc", "28.66", "-9.25%"},
			{"Gitlab Inc", "51.63", "+4.32%"},
		},
		Colors: charttable.Colors{
			Surface: "#1c1c20", HeaderBackground: "#505050", HeaderText: "#ffffff",
			Text: "#ffffff", RowBackgrounds: []string{"#1c1c20"},
		},
		CellStyle: func(cell charttable.Cell) charttable.CellAppearance {
			if cell.Column != 2 {
				return charttable.CellAppearance{}
			}
			if strings.HasPrefix(cell.Value, "+") {
				return charttable.CellAppearance{BackgroundColor: "#b33514"}
			}
			return charttable.CellAppearance{BackgroundColor: "#217c32"}
		},
		Export: &chartcontrol.ExportOptions{Filename: "market-changes"},
	}
}

func sampleReadableRadar() radar.Config {
	cfg := sampleBasicRadar()
	cfg.Label = "Readable radar values and compact layout"
	cfg.Caption = "The same budget comparison uses typed labels, compact formatting, logical placement, and caller colors."
	cfg.Indicators = append([]radar.Indicator(nil), cfg.Indicators...)
	for index := range cfg.Indicators {
		cfg.Indicators[index].Label.FontSize = 10
	}
	cfg.Indicators[0].Min = 1000
	cfg.Series = append([]radar.Series(nil), cfg.Series...)
	for index := range cfg.Series {
		cfg.Series[index].Values = append([]float64(nil), cfg.Series[index].Values...)
	}
	cfg.Series[0].Options.LabelFontSize = 9
	cfg.Series[1].Options = radar.SeriesOptions{ValueLabels: radar.ValueLabelsHidden, ValueFormat: radar.ValueFormatInteger}
	cfg.Style = charttheme.Style{
		Palette: charttheme.PaletteAraiHu,
		Colors:  []string{"#365314", "#c2410c"},
		Class:   "radar-explicit-override",
	}
	cfg.Options = radar.Options{RadiusPercent: 44, ValueLabels: radar.ValueLabelsShown, ValueFormat: radar.ValueFormatHumanized}
	cfg.Title = radar.TitleOptions{Text: "Basic Radar Chart", Subtext: "Values at each vertex", Horizontal: radar.PlacementCenter, FontSize: 16, SubtextFontSize: 12}
	cfg.Legend = radar.LegendOptions{Orientation: radar.LegendVertical, Horizontal: radar.PlacementEnd, Alignment: radar.AlignmentEnd, FontSize: 10, Overlay: true}
	cfg.Padding = radar.Padding{Top: 24, Right: 84, Bottom: 24, Left: 24}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "readable-radar-values", Background: chartcontrol.ExportBackgroundTransparent}
	return cfg
}

func sampleBasicHeatMap() heatmap.Config {
	return heatmap.Config{
		Label: "Basic heat map", Caption: "Values across five X and Y categories.", Title: "Heat Map Chart",
		XAxis: heatmap.Axis{Title: "X-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
		YAxis: heatmap.Axis{Title: "Y-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
		Rows: [][]float64{
			{4.4, 4.9, 7.0, 7.5, 4.3},
			{2.6, 5.9, 9.0, 6.4, 2.3},
			{3.3, 6.4, 7.0, 4.9, 3.2},
			{1.9, 6.0, 9.0, 5.9, 2.6},
			{4.4, 5.9, 7.0, 6.4, 4.6},
		},
		ValueRange: heatmap.ValueRange{Min: 1.9, Max: 9.0},
		Width:      600,
		Height:     400,
		RootAttrs: templ.Attributes{
			"data-static-heatmap-exhaustion": "1fe31b06",
			"data-static-heatmap-source":     "c39a3d85a0df126d",
			"data-goshtoso-candidate":        "heatmap-basic-c39a3d85a0df126d",
		},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-heat-map"},
	}
}

func sampleBasicHeatMapOverride() heatmap.Config {
	cfg := sampleBasicHeatMap()
	cfg.Label = "Caller-styled pinned heat map with value labels"
	cfg.Caption = "The same pinned matrix with typed labels, spacing, and a caller-defined cold-to-warm scale."
	cfg.TitleOptions = heatmap.TitleOptions{Subtext: "Same source data, caller presentation", FontSize: 16, SubtextFontSize: 10}
	cfg.XAxis.TitleFontSize, cfg.XAxis.LabelFontSize, cfg.XAxis.LabelCount = 11, 9, 5
	cfg.YAxis.TitleFontSize, cfg.YAxis.LabelFontSize, cfg.YAxis.LabelCount = 11, 9, 5
	cfg.Padding = heatmap.Padding{Top: 20, Right: 24, Bottom: 24, Left: 28}
	cfg.ValueLabels = heatmap.ValueLabelOptions{Show: true, Format: heatmap.ValueFormatExact, Decimals: 1, TrailingZeros: true, FontSize: 10}
	cfg.Gradient = heatmap.Gradient{Stops: []heatmap.GradientStop{
		{At: 0, Color: "#0e7490", Class: "scale-cold"},
		{At: 0.5, Color: "#fbbf24", Class: "scale-middle"},
		{At: 1, Color: "#e11d48", Class: "scale-warm"},
	}}
	cfg.RootAttrs = templ.Attributes{
		"data-static-heatmap-exhaustion": "1fe31b06",
		"data-static-heatmap-source":     "c39a3d85a0df126d",
		"data-static-heatmap-api":        "dd9b80660b9e0c0b",
		"data-goshtoso-candidate":        "heatmap-options-dd9b80660b9e0c0b",
	}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "caller-styled-heat-map"}
	return cfg
}

func sampleDistributionShapes() violin.Config {
	const sampleCount = 200
	generator := &deterministicLCG{state: 42}
	series := []violin.Series{
		{Name: "Normal", Samples: make([]float64, sampleCount)},
		{Name: "Right Skewed", Samples: make([]float64, sampleCount)},
		{Name: "Bimodal", Samples: make([]float64, sampleCount)},
		{Name: "Tight", Samples: make([]float64, sampleCount)},
	}
	for index := range sampleCount {
		series[0].Samples[index] = generator.normal(50, 10)
		series[1].Samples[index] = 30 + math.Exp(generator.normal(0, 1))*10
		if generator.next() < .45 {
			series[2].Samples[index] = generator.normal(35, 6)
		} else {
			series[2].Samples[index] = generator.normal(65, 6)
		}
		series[3].Samples[index] = generator.normal(50, 3)
	}
	for index := range series {
		series[index].Marks = violin.MarkLines{Mean: true, Median: true}
		series[index].Statistics = violin.Statistics{Quantiles: []float64{.25, .75}}
		series[index].Class = "distribution-" + []string{"normal", "right-skewed", "bimodal", "tight"}[index]
	}
	return violin.Config{
		Label:   "Distribution shapes from deterministic samples",
		Caption: "Four 200-sample distributions; lines mark each mean and median, with quartiles in the adjacent summary.",
		Title:   "Distribution Shapes",
		Series:  series,
		Density: violin.Distribution{Points: 80},
		Padding: violin.Padding{Top: 5, Right: 5, Bottom: 50, Left: 5},
		Width:   1200, Height: 800,
	}
}

func sampleDenseScatter() scatter.Config {
	const dataPointCount = 1000
	categories := make([]string, dataPointCount)
	for index := range categories {
		categories[index] = fmt.Sprintf("foo %d", index)
	}
	values := denseScatterValues(rand.New(rand.NewSource(20260728)), 3, dataPointCount, 10)
	zero, maximum := 0.0, 280.0
	noGap := false
	return scatter.Config{
		Label:      "Dense scatter data",
		Caption:    "Three bounded random-walk series across 1,000 categories; sampled axis labels, trend lines, and maximum references summarize the shape. Applications own any adjacent exact-data view.",
		Categories: categories,
		Width:      600, Height: 400,
		Options: scatter.Options{Size: 0.5, Trend: scatter.TrendLine{Kind: scatter.TrendSimpleMovingAverage, Period: 100}, ValueFormat: scatter.ValueFormatHumanized},
		Series: []scatter.Series{
			{Name: "One", Values: values[0], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
			{Name: "Two", Values: values[1], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
			{Name: "Three", Values: values[2]},
		},
		Title:   scatter.TitleOptions{Text: "Dense Scatter Chart Demo", Placement: scatter.PlacementCenter},
		Legend:  scatter.LegendOptions{Orientation: scatter.LegendVertical, Placement: scatter.PlacementRight, Alignment: scatter.AlignmentRight, FontSize: 6},
		XAxis:   scatter.CategoryAxisOptions{BoundaryGap: &noGap, LabelCount: 10, LabelFontSize: 6, LabelRotation: 45},
		YAxis:   scatter.ValueAxisOptions{Min: &zero, Max: &maximum, Unit: 10, LabelSkip: 1, LabelFontSize: 6},
		Padding: scatter.Padding{Top: 16, Right: 32, Bottom: 16, Left: 16},
		RootAttrs: templ.Attributes{
			"data-goshtoso-candidate":        "scatter-dense-0a50b43ccad6a96b",
			"data-static-scatter-exhaustion": "1fe31b06",
		},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "dense-scatter-data"},
	}
}

func sampleTopNScatter() scatter.Config {
	values := []float64{15.2, 18.5, 22.1, 19.8, 25.4, 21.3, 17.9, 32.6, 28.1, 24.7, 31.5, 29.3, 26.8, 35.2, 41.7, 38.9, 33.1, 29.6, 27.4, 30.8, 36.3, 42.1, 39.5, 44.8, 48.3, 45.6, 40.2, 37.9, 34.5, 26.1}
	categories := make([]string, len(values))
	for index := range values {
		categories[index] = fmt.Sprintf("Day %d", index+1)
	}
	zero, maximum := 0.0, 50.0
	return scatter.Config{
		Label:      "Website traffic over 30 days with peak-day labels",
		Caption:    "Daily visitors in thousands. Exact values and the selected five labels remain available below the chart.",
		Categories: categories,
		Series: []scatter.Series{{
			Name: "Daily Visitors (k)", Values: [][]float64{{15.2}, {18.5}, {22.1}, {19.8}, {25.4}, {21.3}, {17.9}, {32.6}, {28.1}, {24.7}, {31.5}, {29.3}, {26.8}, {35.2}, {41.7}, {38.9}, {33.1}, {29.6}, {27.4}, {30.8}, {36.3}, {42.1}, {39.5}, {44.8}, {48.3}, {45.6}, {40.2}, {37.9}, {34.5}, {26.1}},
		}},
		Options: scatter.Options{TopNLabels: scatter.TopNLabels{Count: 5, FontSize: 16, Color: "var(--color-chart-danger)"}},
		Title:   scatter.TitleOptions{Text: "Website Traffic Over 30 Days - Peak Days Highlighted", Subtext: "(Only top 5 traffic days show labels)"},
		Legend:  scatter.LegendOptions{Hidden: true},
		YAxis:   scatter.ValueAxisOptions{Min: &zero, Max: &maximum},
		Padding: scatter.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
		Width:   800, Height: 500,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "website-traffic-top-labels"},
	}
}

func denseScatterValues(rng *rand.Rand, seriesCount, pointCount int, maxVariationPercentage float64) [][][]float64 {
	data := make([][][]float64, seriesCount)
	for seriesIndex := range data {
		data[seriesIndex] = make([][]float64, pointCount)
		for pointIndex := range pointCount {
			if pointIndex == 0 {
				data[seriesIndex][pointIndex] = []float64{rng.Float64() * 100}
				continue
			}
			previous := data[seriesIndex][pointIndex-1][0]
			variation := previous * maxVariationPercentage / 100
			minimum, maximum := previous-variation, previous+variation
			values := []float64{minimum + rng.Float64()*(maximum-minimum)}
			if pointIndex%2 == 0 {
				values = append(values, minimum+rng.Float64()*(maximum-minimum))
			}
			if pointIndex%10 == 0 {
				values = append(values, minimum+rng.Float64()*(maximum-minimum))
			}
			data[seriesIndex][pointIndex] = values
		}
	}
	return data
}

func lineCode() string {
	return `@line.Line(line.Config{
  Label: "HTTPS monitor latency in milliseconds",
  Labels: []string{"08:00", "08:01", "08:02"},
  Series: []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 51}}},
})`
}

func dualAxisLineCode() string {
	return `@line.Line(line.Config{
  Label: "Dual Axis Line",
  Title: line.Title{Text: "Dual Axis Line"},
  Labels: []string{"A", "B", "C", "D", "E", "F", "G"},
  Series: []line.Series{
    {Name: "Left Series", Values: []float64{120, 132, 101, 134, 90, 230, 210}},
    {Name: "Right Series", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}, YAxisIndex: 1},
  },
  YAxes: []line.Axis{{}, {}},
  Width: 600,
  Height: 400,
})`
}

func areaLineCode() string {
	return `minimum := 0.0
noGap := false

@line.Line(line.Config{
  Label: "Line",
  Title: line.Title{Text: "Line"},
  Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Series: []line.Series{{Name: "Email", Values: []float64{120, 132, 101, 134, 90, 230, 210}}},
  Area: line.AreaOptions{Enabled: true, Opacity: 150.0 / 255.0},
  XAxis: line.CategoryAxisOptions{BoundaryGap: &noGap},
  Legend: line.LegendOptions{Padding: line.Padding{Top: 5, Bottom: 10}},
  YAxes: []line.Axis{{Min: &minimum}},
  Width: 600,
  Height: 400,
})`
}

func dualAxisLineOverridesCode() string {
	return `cfg := dualAxisLineConfig()
cfg.Series[0].Color = "#14532d"
cfg.Series[1].Class = "caller-right-series"
cfg.YAxes[0].Class = "caller-left-axis"
cfg.YAxes[1].Color = "#7e22ce"

@line.Line(cfg)`
}

func horizontalBarCode() string {
	return `@bar.Bar(bar.Config{
  Label: "World population by reporting series",
  Title: "World Population",
  Orientation: bar.OrientationHorizontal,
  Labels: []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
  Series: []bar.Series{
    {Name: "2011", Values: []float64{10, 30, 50, 70, 90, 110, 130}},
    {Name: "2012", Values: []float64{20, 40, 60, 80, 100, 120, 140}},
  },
  Padding: bar.Padding{Top: 20, Right: 40, Bottom: 20, Left: 20},
  Width: 600, Height: 400,
})`
}

func barReferencesCode() string {
	return `references := bar.References{
  Average: true, Minimum: true, Maximum: true,
  Format: bar.ValueFormatHumanized,
}

@bar.Bar(bar.Config{
  Label: "Monthly rainfall and evaporation reference annotations",
  Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
  Series: []bar.Series{
    {Name: "Rainfall", Values: rainfall, References: references},
    {Name: "Evaporation", Values: evaporation, References: references},
  },
  Width: 600, Height: 400,
})`
}

func scatterCode() string {
	return `@scatter.Scatter(scatter.Config{
	Label: "Dense scatter data",
	Categories: labels,
	Width: 600, Height: 400,
	Options: scatter.Options{Size: 0.5, Trend: scatter.TrendLine{
		Kind: scatter.TrendSimpleMovingAverage, Period: 100,
	}},
	Series: []scatter.Series{
		{Name: "One", Values: values[0], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
		{Name: "Two", Values: values[1], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
		{Name: "Three", Values: values[2]},
	},
})`
}

func topNScatterCode() string {
	return `@scatter.Scatter(scatter.Config{
	Label: "Website traffic over 30 days with peak-day labels",
	Categories: days,
	Series: []scatter.Series{{Name: "Daily Visitors (k)", Values: values}},
	Options: scatter.Options{TopNLabels: scatter.TopNLabels{
		Count: 5, FontSize: 16, Color: "var(--color-chart-danger)",
	}},
	Title: scatter.TitleOptions{Text: "Website Traffic Over 30 Days - Peak Days Highlighted", Subtext: "(Only top 5 traffic days show labels)"},
	Legend: scatter.LegendOptions{Hidden: true},
	YAxis: scatter.ValueAxisOptions{Min: &zero, Max: &maximum},
	Padding: scatter.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
	Width: 800, Height: 500,
})`
}

func radarCode() string {
	return `@radar.Radar(radar.Config{
  Label: "Basic radar chart",
  Indicators: []radar.Indicator{
    {Name: "Sales", Max: 6500},
    {Name: "Administration", Max: 16000},
    {Name: "Information Technology", Max: 30000},
    {Name: "Customer Support", Max: 38000},
    {Name: "Development", Max: 52000},
    {Name: "Marketing", Max: 25000},
  },
  Series: []radar.Series{
    {Name: "Allocated Budget", Values: []float64{4200, 3000, 20000, 35000, 50000, 18000}},
    {Name: "Actual Spending", Values: []float64{5000, 14000, 28000, 26000, 42000, 21000}},
  },
  Title: radar.TitleOptions{Text: "Basic Radar Chart", FontSize: 16},
  Legend: radar.LegendOptions{Horizontal: radar.PlacementEnd},
  Width: 600, Height: 400,
})`
}

func radarReadableCode() string {
	return `cfg := sampleBasicRadar()
cfg.Label = "Readable radar values and compact layout"
cfg.Caption = "The same budget comparison uses typed labels, compact formatting, logical placement, and caller colors."
cfg.Indicators[0].Min = 1000
for index := range cfg.Indicators {
  cfg.Indicators[index].Label.FontSize = 10
}
cfg.Options = radar.Options{
  RadiusPercent: 44,
  ValueLabels: radar.ValueLabelsShown,
  ValueFormat: radar.ValueFormatHumanized,
}
cfg.Series[0].Options.LabelFontSize = 9
cfg.Series[1].Options = radar.SeriesOptions{
  ValueLabels: radar.ValueLabelsHidden,
  ValueFormat: radar.ValueFormatInteger,
}
cfg.Title = radar.TitleOptions{
  Text: "Basic Radar Chart", Subtext: "Values at each vertex",
  Horizontal: radar.PlacementCenter, FontSize: 16, SubtextFontSize: 12,
}
cfg.Legend = radar.LegendOptions{
  Orientation: radar.LegendVertical, Horizontal: radar.PlacementEnd,
  Alignment: radar.AlignmentEnd, FontSize: 10, Overlay: true,
}
cfg.Padding = radar.Padding{Top: 24, Right: 84, Bottom: 24, Left: 24}
cfg.Style = charttheme.Style{
  Palette: charttheme.PaletteAraiHu,
  Colors: []string{"#365314", "#c2410c"},
  Class: "radar-explicit-override",
}
cfg.Export = &chartcontrol.ExportOptions{
  Filename: "readable-radar-values",
  Background: chartcontrol.ExportBackgroundTransparent,
}
@radar.Radar(cfg)`
}

func candlestickCode() string {
	return `@candlestick.Candlestick(candlestick.Config{
  Label: "Seven-day stock price",
  Title: "Candlestick Chart",
  SeriesName: "Stock Price",
  Data: []candlestick.Datum{
    {Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
    {Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
    {Label: "Day 3", Open: 112, High: 118, Low: 108, Close: 115},
    {Label: "Day 4", Open: 115, High: 120, Low: 104, Close: 108},
    {Label: "Day 5", Open: 108, High: 113, Low: 105, Close: 109},
    {Label: "Day 6", Open: 109, High: 116, Low: 106, Close: 114},
    {Label: "Day 7", Open: 114, High: 121, Low: 111, Close: 119},
  },
})`
}

func candlestickExtendedBasicCode() string {
	return `cfg := sampleBasicCandlestick()
cfg.Label = "Eight-day stock price"
cfg.Data = append(cfg.Data,
  candlestick.Datum{Label: "Day 8", Open: 119, High: 125, Low: 116, Close: 122},
)
cfg.Options = candlestick.Options{
  TitleFontSize: 18,
  YUnit: 1,
  Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
}
cfg.Width, cfg.Height = 800, 600
@candlestick.Candlestick(cfg)`
}

func candlestickMultipleSeriesCode() string {
	return `@candlestick.Candlestick(candlestick.Config{
  Label: "Three aligned stock series",
  Series: []candlestick.Series{
    {Name: "Stock A", Data: stockA, BodyStyle: candlestick.BodyStyleFilled},
    {Name: "Stock B", Data: stockB, BodyStyle: candlestick.BodyStyleTraditional},
    {Name: "Stock C", Data: stockC, BodyStyle: candlestick.BodyStyleOutline},
  },
  Options: candlestick.Options{
    YUnit: 1,
    Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
  },
  Width: 1000, Height: 700,
})`
}

func candlestickGeometryCode() string {
	return `showWicks, hideWicks, seriesGap := true, false, 0.08
cfg := sampleCandlestickMultipleSeries()
cfg.Options.Geometry = candlestick.Geometry{
  CandleWidth: 0.58,
  WickWidth: 2,
  SeriesGap: &seriesGap,
  ShowWicks: &showWicks,
}
cfg.Series[2].ShowWicks = &hideWicks
@candlestick.Candlestick(cfg)`
}

func candlestickBollingerCode() string {
	return `@candlestick.Candlestick(candlestick.Config{
  Label: "Candlestick Chart with Bollinger Bands",
  Title: "Candlestick Chart with Bollinger Bands",
  SeriesName: "Price",
  Data: []candlestick.Datum{
    {Label: "1", Open: 100, High: 110, Low: 95, Close: 105},
    {Label: "2", Open: 105, High: 115, Low: 100, Close: 112},
    {Label: "3", Open: 112, High: 118, Low: 108, Close: 115},
    {Label: "4", Open: 115, High: 125, Low: 110, Close: 120},
    {Label: "5", Open: 120, High: 130, Low: 115, Close: 125},
    {Label: "6", Open: 125, High: 135, Low: 120, Close: 130},
    {Label: "7", Open: 130, High: 140, Low: 125, Close: 135},
    {Label: "8", Open: 135, High: 145, Low: 130, Close: 140},
    {Label: "9", Open: 140, High: 150, Low: 135, Close: 145},
    {Label: "10", Open: 145, High: 155, Low: 140, Close: 150},
    {Label: "11", Open: 150, High: 160, Low: 145, Close: 148},
    {Label: "12", Open: 148, High: 153, Low: 143, Close: 146},
    {Label: "13", Open: 146, High: 151, Low: 141, Close: 144},
    {Label: "14", Open: 144, High: 149, Low: 139, Close: 142},
    {Label: "15", Open: 142, High: 147, Low: 137, Close: 140},
    {Label: "16", Open: 140, High: 145, Low: 135, Close: 138},
    {Label: "17", Open: 138, High: 143, Low: 133, Close: 136},
    {Label: "18", Open: 136, High: 141, Low: 131, Close: 134},
    {Label: "19", Open: 134, High: 139, Low: 129, Close: 132},
    {Label: "20", Open: 132, High: 137, Low: 127, Close: 130},
  },
  TrendLines: []candlestick.TrendLine{
    {Type: candlestick.TrendTypeBollingerUpper, Period: 5},
    {Type: candlestick.TrendTypeSimpleMovingAverage, Period: 5},
    {Type: candlestick.TrendTypeBollingerLower, Period: 5},
  },
  Options: candlestick.Options{
    TitleFontSize: 18,
    YUnit: 1,
    Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
  },
  Width: 800,
  Height: 600,
})`
}

func candlestickAggregationCode() string {
	return `@candlestick.Candlestick(candlestick.Config{
  Label: "Candlestick aggregation",
  Title: "1-Minute Candles (Before Aggregation)",
  SeriesName: "1-Minute",
  Data: []candlestick.Datum{
    {Label: "1", Open: 100, High: 102, Low: 99, Close: 101},
    {Label: "2", Open: 101, High: 103, Low: 100, Close: 102},
    {Label: "3", Open: 102, High: 105, Low: 101, Close: 104},
    {Label: "4", Open: 104, High: 106, Low: 103, Close: 105},
    {Label: "5", Open: 105, High: 107, Low: 104, Close: 106},
    {Label: "6", Open: 106, High: 108, Low: 105, Close: 107},
    {Label: "7", Open: 107, High: 109, Low: 106, Close: 108},
    {Label: "8", Open: 108, High: 110, Low: 107, Close: 109},
    {Label: "9", Open: 109, High: 111, Low: 108, Close: 110},
    {Label: "10", Open: 110, High: 112, Low: 109, Close: 111},
    {Label: "11", Open: 111, High: 113, Low: 110, Close: 112},
    {Label: "12", Open: 112, High: 114, Low: 111, Close: 113},
    {Label: "13", Open: 113, High: 115, Low: 112, Close: 114},
    {Label: "14", Open: 114, High: 116, Low: 113, Close: 115},
    {Label: "15", Open: 115, High: 117, Low: 114, Close: 116},
  },
  Aggregation: candlestick.AggregationOptions{
    WindowSize: 5,
    Title: "5-Minute Aggregated Candles",
    SeriesName: "5-Minute",
  },
  Options: candlestick.Options{
    TitleFontSize: 16,
    YUnit: 1,
    Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
  },
  Width: 1200,
  Height: 800,
})`
}

func candlestickPresentationCode() string {
	return `cfg := sampleCandlestickBollingerBands()
cfg.Options.Increasing.Color = "#14532d"
cfg.Options.Decreasing.Class = "caller-decreasing-candles"
cfg.TrendLines[0].Color = "#1d4ed8"
cfg.TrendLines[1].Class = "caller-middle-band"
@candlestick.Candlestick(cfg)`
}

func candlestickPatternsCode() string {
	return `@candlestick.Candlestick(candlestick.Config{
  Label: "Candlestick patterns",
  Title: "Candlestick Patterns",
  SeriesName: "Stock Price with Patterns",
  Data: []candlestick.Datum{
    {Label: "1", Open: 100, High: 110, Low: 95, Close: 105},
    {Label: "2", Open: 105, High: 108, Low: 102, Close: 105.1},
    {Label: "3", Open: 108, High: 109, Low: 98, Close: 107},
    {Label: "4", Open: 107, High: 108, Low: 103, Close: 104},
    {Label: "5", Open: 102, High: 115, Low: 101, Close: 113},
    {Label: "6", Open: 113, High: 125, Low: 112, Close: 114},
    {Label: "7", Open: 114, High: 118, Low: 113, Close: 117},
    {Label: "8", Open: 119, High: 120, Low: 108, Close: 110},
    {Label: "9", Open: 110, High: 113, Low: 107, Close: 109.9},
    {Label: "10", Open: 109, High: 118, Low: 108, Close: 116},
  },
  Patterns: candlestick.PatternOptions{
    Selection: candlestick.PatternSelectionAll,
    References: []candlestick.CloseReferenceType{
      candlestick.CloseReferenceAverage,
      candlestick.CloseReferenceMinimum,
    },
  },
  Options: candlestick.Options{TitleFontSize: 16, YUnit: 1, Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}},
  Width: 900, Height: 650,
})`
}

func candlestickPatternVariantCode() string {
	return `cfg.Patterns.Selection = candlestick.PatternSelectionCore
// Other groups: Bullish, Bearish, Reversal, or Trend.
// Enabled selects an explicit ordered subset instead of a group.
cfg.Patterns.DojiThreshold = 0.05
cfg.Patterns.ShadowTolerance = 0.01
cfg.Patterns.ShadowRatio = 2
cfg.Patterns.EngulfingMinSize = 1
cfg.Patterns.PreferLabels = true
cfg.Patterns.Label = candlestick.PatternLabelStyle{
  Text: candlestick.PatternLabelTextNameWithCount,
  Color: "#ffffff", BackgroundColor: "#1d4ed8", FontSize: 8, CornerRadius: 2,
}
@candlestick.Candlestick(cfg)`
}

func funnelCode() string {
	return `@funnel.Funnel(funnel.Config{
  Label: "Basic funnel",
  Title: "Funnel",
  Stages: []funnel.Stage{
    {Label: "Show", Value: 100},
    {Label: "Click", Value: 80},
    {Label: "Visit", Value: 60},
    {Label: "Inquiry", Value: 40},
    {Label: "Order", Value: 20},
    {Label: "Pay", Value: 10},
    {Label: "Cancel", Value: 2},
  },
  Options: funnel.Options{
    Legend: funnel.Legend{Padding: funnel.Padding{Left: 100}},
  },
})`
}

func compactFunnelCode() string {
	return `@funnel.Funnel(funnel.Config{
  Label: "Compact five-stage funnel",
  Title: "Funnel",
  Stages: []funnel.Stage{
    {Label: "Show", Value: 100},
    {Label: "Click", Value: 80},
    {Label: "Visit", Value: 60},
    {Label: "Inquiry", Value: 40},
    {Label: "Order", Value: 20},
  },
})`
}

func tableCode() string {
	return `@table.Table(table.Config{
  Label: "People directory",
  Columns: []table.Column{
    {Header: "Name", Span: 2},
    {Header: "Age"},
    {Header: "Address", Span: 3},
    {Header: "Tag", Span: 2},
    {Header: "Action", Span: 2},
  },
  Rows: [][]string{
    {"John Brown", "32", "New York No. 1 Lake Park", "nice, developer", "Send Mail"},
    {"Jim Green", "42", "London No. 1 Lake Park", "wow", "Send Mail"},
    {"Joe Black", "32", "Sidney No. 1 Lake Park", "cool, teacher", "Send Mail"},
  },
})`
}

func marketTableCode() string {
	return `cfg := table.Config{
  Label: "Market changes",
  Columns: []table.Column{
    {Header: "Name"},
    {Header: "Price", Align: table.AlignEnd},
    {Header: "Change", Align: table.AlignEnd},
  },
  Rows: [][]string{
    {"Datadog Inc", "97.32", "-7.49%"},
    {"Hashicorp Inc", "28.66", "-9.25%"},
    {"Gitlab Inc", "51.63", "+4.32%"},
  },
  Colors: table.Colors{
    Surface: "#1c1c20", HeaderBackground: "#505050",
    HeaderText: "#ffffff", Text: "#ffffff",
  },
  CellStyle: marketChangeStyle,
}
@table.Table(cfg)`
}

func heatMapCode() string {
	return `@heatmap.HeatMap(heatmap.Config{
  Label: "Basic heat map",
  Title: "Heat Map Chart",
  XAxis: heatmap.Axis{Title: "X-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
  YAxis: heatmap.Axis{Title: "Y-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
  Rows: [][]float64{
    {4.4, 4.9, 7.0, 7.5, 4.3},
    {2.6, 5.9, 9.0, 6.4, 2.3},
    {3.3, 6.4, 7.0, 4.9, 3.2},
    {1.9, 6.0, 9.0, 5.9, 2.6},
    {4.4, 5.9, 7.0, 6.4, 4.6},
  },
  ValueRange: heatmap.ValueRange{Min: 1.9, Max: 9},
  Width: 600,
  Height: 400,
})`
}

func heatMapOverrideCode() string {
	return `cfg := sampleBasicHeatMap()
cfg.TitleOptions = heatmap.TitleOptions{
  Subtext: "Same source data, caller presentation",
  FontSize: 16,
  SubtextFontSize: 10,
}
cfg.XAxis.TitleFontSize, cfg.XAxis.LabelFontSize, cfg.XAxis.LabelCount = 11, 9, 5
cfg.YAxis.TitleFontSize, cfg.YAxis.LabelFontSize, cfg.YAxis.LabelCount = 11, 9, 5
cfg.Padding = heatmap.Padding{Top: 20, Right: 24, Bottom: 24, Left: 28}
cfg.ValueLabels = heatmap.ValueLabelOptions{
  Show: true,
  Format: heatmap.ValueFormatExact,
  Decimals: 1,
  TrailingZeros: true,
  FontSize: 10,
}
cfg.Gradient = heatmap.Gradient{
  Stops: []heatmap.GradientStop{
    {At: 0, Color: "#0e7490", Class: "scale-cold"},
    {At: 0.5, Color: "#fbbf24", Class: "scale-middle"},
    {At: 1, Color: "#e11d48", Class: "scale-warm"},
  },
}
@heatmap.HeatMap(cfg)`
}

func violinCode() string {
	return `@violin.Violin(violin.Config{
  Label: "Distribution shapes from deterministic samples",
  Caption: "Four 200-sample distributions with exact adjacent statistics.",
  Title: "Distribution Shapes",
  Series: []violin.Series{
    {Name: "Normal", Samples: normalSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
    {Name: "Right Skewed", Samples: rightSkewedSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
    {Name: "Bimodal", Samples: bimodalSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
    {Name: "Tight", Samples: tightSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
  },
  Density: violin.Distribution{Points: 80},
  Padding: violin.Padding{Top: 5, Right: 5, Bottom: 50, Left: 5},
  Width: 1200, Height: 800,
})`
}
