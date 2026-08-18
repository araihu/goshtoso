package charts

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactivecandlestick "github.com/araihu/goshtoso-charts/components/interactive/candlestick"
)

type candlestickSampleDatum struct {
	Category string
	Candle   interactivecandlestick.Candle
}

var interactiveCandlestickUpstreamData = []candlestickSampleDatum{
	{Category: "2018/1/24", Candle: interactivecandlestick.Candle{Open: 2320.26, Close: 2320.26, Low: 2287.3, High: 2362.94}},
	{Category: "2018/1/25", Candle: interactivecandlestick.Candle{Open: 2300, Close: 2291.3, Low: 2288.26, High: 2308.38}},
	{Category: "2018/1/28", Candle: interactivecandlestick.Candle{Open: 2295.35, Close: 2346.5, Low: 2295.35, High: 2346.92}},
	{Category: "2018/1/29", Candle: interactivecandlestick.Candle{Open: 2347.22, Close: 2358.98, Low: 2337.35, High: 2363.8}},
	{Category: "2018/1/30", Candle: interactivecandlestick.Candle{Open: 2360.75, Close: 2382.48, Low: 2347.89, High: 2383.76}},
	{Category: "2018/1/31", Candle: interactivecandlestick.Candle{Open: 2383.43, Close: 2385.42, Low: 2371.23, High: 2391.82}},
	{Category: "2018/2/1", Candle: interactivecandlestick.Candle{Open: 2377.41, Close: 2419.02, Low: 2369.57, High: 2421.15}},
	{Category: "2018/2/4", Candle: interactivecandlestick.Candle{Open: 2425.92, Close: 2428.15, Low: 2417.58, High: 2440.38}},
	{Category: "2018/2/5", Candle: interactivecandlestick.Candle{Open: 2411, Close: 2433.13, Low: 2403.3, High: 2437.42}},
	{Category: "2018/2/6", Candle: interactivecandlestick.Candle{Open: 2432.68, Close: 2434.48, Low: 2427.7, High: 2441.73}},
	{Category: "2018/2/7", Candle: interactivecandlestick.Candle{Open: 2430.69, Close: 2418.53, Low: 2394.22, High: 2433.89}},
	{Category: "2018/2/8", Candle: interactivecandlestick.Candle{Open: 2416.62, Close: 2432.4, Low: 2414.4, High: 2443.03}},
	{Category: "2018/2/18", Candle: interactivecandlestick.Candle{Open: 2441.91, Close: 2421.56, Low: 2415.43, High: 2444.8}},
	{Category: "2018/2/19", Candle: interactivecandlestick.Candle{Open: 2420.26, Close: 2382.91, Low: 2373.53, High: 2427.07}},
	{Category: "2018/2/20", Candle: interactivecandlestick.Candle{Open: 2383.49, Close: 2397.18, Low: 2370.61, High: 2397.94}},
	{Category: "2018/2/21", Candle: interactivecandlestick.Candle{Open: 2378.82, Close: 2325.95, Low: 2309.17, High: 2378.82}},
	{Category: "2018/2/22", Candle: interactivecandlestick.Candle{Open: 2322.94, Close: 2314.16, Low: 2308.76, High: 2330.88}},
	{Category: "2018/2/25", Candle: interactivecandlestick.Candle{Open: 2320.62, Close: 2325.82, Low: 2315.01, High: 2338.78}},
	{Category: "2018/2/26", Candle: interactivecandlestick.Candle{Open: 2313.74, Close: 2293.34, Low: 2289.89, High: 2340.71}},
	{Category: "2018/2/27", Candle: interactivecandlestick.Candle{Open: 2297.77, Close: 2313.22, Low: 2292.03, High: 2324.63}},
	{Category: "2018/2/28", Candle: interactivecandlestick.Candle{Open: 2322.32, Close: 2365.59, Low: 2308.92, High: 2366.16}},
	{Category: "2018/3/1", Candle: interactivecandlestick.Candle{Open: 2364.54, Close: 2359.51, Low: 2330.86, High: 2369.65}},
	{Category: "2018/3/4", Candle: interactivecandlestick.Candle{Open: 2332.08, Close: 2273.4, Low: 2259.25, High: 2333.54}},
	{Category: "2018/3/5", Candle: interactivecandlestick.Candle{Open: 2274.81, Close: 2326.31, Low: 2270.1, High: 2328.14}},
	{Category: "2018/3/6", Candle: interactivecandlestick.Candle{Open: 2333.61, Close: 2347.18, Low: 2321.6, High: 2351.44}},
	{Category: "2018/3/7", Candle: interactivecandlestick.Candle{Open: 2340.44, Close: 2324.29, Low: 2304.27, High: 2352.02}},
	{Category: "2018/3/8", Candle: interactivecandlestick.Candle{Open: 2326.42, Close: 2318.61, Low: 2314.59, High: 2333.67}},
	{Category: "2018/3/11", Candle: interactivecandlestick.Candle{Open: 2314.68, Close: 2310.59, Low: 2296.58, High: 2320.96}},
	{Category: "2018/3/12", Candle: interactivecandlestick.Candle{Open: 2309.16, Close: 2286.6, Low: 2264.83, High: 2333.29}},
	{Category: "2018/3/13", Candle: interactivecandlestick.Candle{Open: 2282.17, Close: 2263.97, Low: 2253.25, High: 2286.33}},
	{Category: "2018/3/14", Candle: interactivecandlestick.Candle{Open: 2255.77, Close: 2270.28, Low: 2253.31, High: 2276.22}},
	{Category: "2018/3/15", Candle: interactivecandlestick.Candle{Open: 2269.31, Close: 2278.4, Low: 2250, High: 2312.08}},
	{Category: "2018/3/18", Candle: interactivecandlestick.Candle{Open: 2267.29, Close: 2240.02, Low: 2239.21, High: 2276.05}},
	{Category: "2018/3/19", Candle: interactivecandlestick.Candle{Open: 2244.26, Close: 2257.43, Low: 2232.02, High: 2261.31}},
	{Category: "2018/3/20", Candle: interactivecandlestick.Candle{Open: 2257.74, Close: 2317.37, Low: 2257.42, High: 2317.86}},
	{Category: "2018/3/21", Candle: interactivecandlestick.Candle{Open: 2318.21, Close: 2324.24, Low: 2311.6, High: 2330.81}},
	{Category: "2018/3/22", Candle: interactivecandlestick.Candle{Open: 2321.4, Close: 2328.28, Low: 2314.97, High: 2332}},
	{Category: "2018/3/25", Candle: interactivecandlestick.Candle{Open: 2334.74, Close: 2326.72, Low: 2319.91, High: 2344.89}},
	{Category: "2018/3/26", Candle: interactivecandlestick.Candle{Open: 2318.58, Close: 2297.67, Low: 2281.12, High: 2319.99}},
	{Category: "2018/3/27", Candle: interactivecandlestick.Candle{Open: 2299.38, Close: 2301.26, Low: 2289, High: 2323.48}},
	{Category: "2018/3/28", Candle: interactivecandlestick.Candle{Open: 2273.55, Close: 2236.3, Low: 2232.91, High: 2273.55}},
	{Category: "2018/3/29", Candle: interactivecandlestick.Candle{Open: 2238.49, Close: 2236.62, Low: 2228.81, High: 2246.87}},
	{Category: "2018/4/1", Candle: interactivecandlestick.Candle{Open: 2229.46, Close: 2234.4, Low: 2227.31, High: 2243.95}},
	{Category: "2018/4/2", Candle: interactivecandlestick.Candle{Open: 2234.9, Close: 2227.74, Low: 2220.44, High: 2253.42}},
	{Category: "2018/4/3", Candle: interactivecandlestick.Candle{Open: 2232.69, Close: 2225.29, Low: 2217.25, High: 2241.34}},
	{Category: "2018/4/8", Candle: interactivecandlestick.Candle{Open: 2196.24, Close: 2211.59, Low: 2180.67, High: 2212.59}},
	{Category: "2018/4/9", Candle: interactivecandlestick.Candle{Open: 2215.47, Close: 2225.77, Low: 2215.47, High: 2234.73}},
	{Category: "2018/4/10", Candle: interactivecandlestick.Candle{Open: 2224.93, Close: 2226.13, Low: 2212.56, High: 2233.04}},
	{Category: "2018/4/11", Candle: interactivecandlestick.Candle{Open: 2236.98, Close: 2219.55, Low: 2217.26, High: 2242.48}},
	{Category: "2018/4/12", Candle: interactivecandlestick.Candle{Open: 2218.09, Close: 2206.78, Low: 2204.44, High: 2226.26}},
	{Category: "2018/4/15", Candle: interactivecandlestick.Candle{Open: 2199.91, Close: 2181.94, Low: 2177.39, High: 2204.99}},
	{Category: "2018/4/16", Candle: interactivecandlestick.Candle{Open: 2169.63, Close: 2194.85, Low: 2165.78, High: 2196.43}},
	{Category: "2018/4/17", Candle: interactivecandlestick.Candle{Open: 2195.03, Close: 2193.8, Low: 2178.47, High: 2197.51}},
	{Category: "2018/4/18", Candle: interactivecandlestick.Candle{Open: 2181.82, Close: 2197.6, Low: 2175.44, High: 2206.03}},
	{Category: "2018/4/19", Candle: interactivecandlestick.Candle{Open: 2201.12, Close: 2244.64, Low: 2200.58, High: 2250.11}},
	{Category: "2018/4/22", Candle: interactivecandlestick.Candle{Open: 2236.4, Close: 2242.17, Low: 2232.26, High: 2245.12}},
	{Category: "2018/4/23", Candle: interactivecandlestick.Candle{Open: 2242.62, Close: 2184.54, Low: 2182.81, High: 2242.62}},
	{Category: "2018/4/24", Candle: interactivecandlestick.Candle{Open: 2187.35, Close: 2218.32, Low: 2184.11, High: 2226.12}},
	{Category: "2018/4/25", Candle: interactivecandlestick.Candle{Open: 2213.19, Close: 2199.31, Low: 2191.85, High: 2224.63}},
	{Category: "2018/4/26", Candle: interactivecandlestick.Candle{Open: 2203.89, Close: 2177.91, Low: 2173.86, High: 2210.58}},
	{Category: "2018/5/2", Candle: interactivecandlestick.Candle{Open: 2170.78, Close: 2174.12, Low: 2161.14, High: 2179.65}},
	{Category: "2018/5/3", Candle: interactivecandlestick.Candle{Open: 2179.05, Close: 2205.5, Low: 2179.05, High: 2222.81}},
	{Category: "2018/5/6", Candle: interactivecandlestick.Candle{Open: 2212.5, Close: 2231.17, Low: 2212.5, High: 2236.07}},
	{Category: "2018/5/7", Candle: interactivecandlestick.Candle{Open: 2227.86, Close: 2235.57, Low: 2219.44, High: 2240.26}},
	{Category: "2018/5/8", Candle: interactivecandlestick.Candle{Open: 2242.39, Close: 2246.3, Low: 2235.42, High: 2255.21}},
	{Category: "2018/5/9", Candle: interactivecandlestick.Candle{Open: 2246.96, Close: 2232.97, Low: 2221.38, High: 2247.86}},
	{Category: "2018/5/10", Candle: interactivecandlestick.Candle{Open: 2228.82, Close: 2246.83, Low: 2225.81, High: 2247.67}},
	{Category: "2018/5/13", Candle: interactivecandlestick.Candle{Open: 2247.68, Close: 2241.92, Low: 2231.36, High: 2250.85}},
	{Category: "2018/5/14", Candle: interactivecandlestick.Candle{Open: 2238.9, Close: 2217.01, Low: 2205.87, High: 2239.93}},
	{Category: "2018/5/15", Candle: interactivecandlestick.Candle{Open: 2217.09, Close: 2224.8, Low: 2213.58, High: 2225.19}},
	{Category: "2018/5/16", Candle: interactivecandlestick.Candle{Open: 2221.34, Close: 2251.81, Low: 2210.77, High: 2252.87}},
	{Category: "2018/5/17", Candle: interactivecandlestick.Candle{Open: 2249.81, Close: 2282.87, Low: 2248.41, High: 2288.09}},
	{Category: "2018/5/20", Candle: interactivecandlestick.Candle{Open: 2286.33, Close: 2299.99, Low: 2281.9, High: 2309.39}},
	{Category: "2018/5/21", Candle: interactivecandlestick.Candle{Open: 2297.11, Close: 2305.11, Low: 2290.12, High: 2305.3}},
	{Category: "2018/5/22", Candle: interactivecandlestick.Candle{Open: 2303.75, Close: 2302.4, Low: 2292.43, High: 2314.18}},
	{Category: "2018/5/23", Candle: interactivecandlestick.Candle{Open: 2293.81, Close: 2275.67, Low: 2274.1, High: 2304.95}},
	{Category: "2018/5/24", Candle: interactivecandlestick.Candle{Open: 2281.45, Close: 2288.53, Low: 2270.25, High: 2292.59}},
	{Category: "2018/5/27", Candle: interactivecandlestick.Candle{Open: 2286.66, Close: 2293.08, Low: 2283.94, High: 2301.7}},
	{Category: "2018/5/28", Candle: interactivecandlestick.Candle{Open: 2293.4, Close: 2321.32, Low: 2281.47, High: 2322.1}},
	{Category: "2018/5/29", Candle: interactivecandlestick.Candle{Open: 2323.54, Close: 2324.02, Low: 2321.17, High: 2334.33}},
	{Category: "2018/5/30", Candle: interactivecandlestick.Candle{Open: 2316.25, Close: 2317.75, Low: 2310.49, High: 2325.72}},
	{Category: "2018/5/31", Candle: interactivecandlestick.Candle{Open: 2320.74, Close: 2300.59, Low: 2299.37, High: 2325.53}},
	{Category: "2018/6/3", Candle: interactivecandlestick.Candle{Open: 2300.21, Close: 2299.25, Low: 2294.11, High: 2313.43}},
	{Category: "2018/6/4", Candle: interactivecandlestick.Candle{Open: 2297.1, Close: 2272.42, Low: 2264.76, High: 2297.1}},
	{Category: "2018/6/5", Candle: interactivecandlestick.Candle{Open: 2270.71, Close: 2270.93, Low: 2260.87, High: 2276.86}},
	{Category: "2018/6/6", Candle: interactivecandlestick.Candle{Open: 2264.43, Close: 2242.11, Low: 2240.07, High: 2266.69}},
	{Category: "2018/6/7", Candle: interactivecandlestick.Candle{Open: 2242.26, Close: 2210.9, Low: 2205.07, High: 2250.63}},
	{Category: "2018/6/13", Candle: interactivecandlestick.Candle{Open: 2190.1, Close: 2148.35, Low: 2126.22, High: 2190.1}},
}

func sampleInteractiveCandlestick() chart.Instance {
	cfg := interactiveCandlestickConfig()
	cfg.RootAttrs = templ.Attributes{"data-candlestick-variant": "baseline"}
	return interactivecandlestick.Candlestick(cfg)
}

func sampleInteractiveCandlestickInsideZoom() chart.Instance {
	cfg := interactiveCandlestickConfig()
	cfg.Label = "Inside zoom"
	cfg.Caption = "Drag or scroll inside the plot to inspect the same 50–100% date window."
	cfg.DataZoom = []interactivecandlestick.DataZoom{{Type: interactivecandlestick.DataZoomInside, StartPercent: 50, EndPercent: 100}}
	cfg.Options.Title.Text = cfg.Label
	cfg.RootAttrs = templ.Attributes{"data-candlestick-variant": "inside"}
	return interactivecandlestick.Candlestick(cfg)
}

func sampleInteractiveCandlestickBothZooms() chart.Instance {
	cfg := interactiveCandlestickConfig()
	cfg.Label = "Inside and slider zoom"
	cfg.Caption = "Inside gestures and the visible slider share the same 50–100% date window."
	cfg.DataZoom = []interactivecandlestick.DataZoom{
		{Type: interactivecandlestick.DataZoomInside, StartPercent: 50, EndPercent: 100},
		{StartPercent: 50, EndPercent: 100},
	}
	cfg.Options.Title.Text = cfg.Label
	cfg.RootAttrs = templ.Attributes{"data-candlestick-variant": "inside-slider"}
	return interactivecandlestick.Candlestick(cfg)
}

func sampleInteractiveCandlestickYAxisZoom() chart.Instance {
	cfg := interactiveCandlestickConfig()
	cfg.Label = "Y-axis slider zoom"
	cfg.Caption = "A vertical slider narrows the visible value range from 50% to 100%."
	cfg.DataZoom = []interactivecandlestick.DataZoom{{Axis: interactivecandlestick.DataZoomYAxis, StartPercent: 50, EndPercent: 100}}
	cfg.Options.Title.Text = cfg.Label
	cfg.RootAttrs = templ.Attributes{"data-candlestick-variant": "y-axis"}
	return interactivecandlestick.Candlestick(cfg)
}

func sampleInteractiveCandlestickUpstreamStyle() chart.Instance {
	cfg := interactiveCandlestickConfig()
	cfg.Label = "Direction style and extrema"
	cfg.Caption = "Theme tokens reverse the direction treatment and mark the highest and lowest ranges."
	cfg.Series[0].Options.Rise = interactivecandlestick.DirectionStyle{
		Class: "goshtoso-charts-candlestick__direction--decreasing",
	}
	cfg.Series[0].Options.Fall = interactivecandlestick.DirectionStyle{
		Class: "goshtoso-charts-candlestick__direction--increasing",
	}
	cfg.Series[0].Options.Marks = interactivecandlestick.MarkOptions{Highest: true, Lowest: true, ShowLabel: new(true)}
	cfg.Options.Title.Text = cfg.Label
	cfg.RootAttrs = templ.Attributes{"data-candlestick-variant": "style"}
	return interactivecandlestick.Candlestick(cfg)
}

func interactiveCandlestickConfig() interactivecandlestick.Config {
	categories := make([]string, len(interactiveCandlestickUpstreamData))
	candles := make([]interactivecandlestick.Candle, len(interactiveCandlestickUpstreamData))
	for index, datum := range interactiveCandlestickUpstreamData {
		categories[index] = datum.Category
		candles[index] = datum.Candle
	}
	return interactivecandlestick.Config{
		Label:      "Candlestick example",
		Caption:    "Open, close, low, and high values from 24 January through 13 June 2018.",
		Categories: categories,
		Series: []interactivecandlestick.Series{{
			Name: "OHLC",
			Data: candles,
		}},
		DataZoom: []interactivecandlestick.DataZoom{
			{StartPercent: 50, EndPercent: 100},
		},
		Width: "100%", Height: "500px",
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: "Candlestick example", Left: "center", Top: "4%"},
			Legend:   &chart.LegendOptions{Left: "center", Top: "12%"},
			Tooltip:  &chart.TooltipOptions{Show: new(true), Trigger: "axis"},
			XAxis:    &chart.AxisOptions{Type: "category", SplitNumber: 20},
			YAxis:    &chart.AxisOptions{Type: "value", Scale: new(true)},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "candlestick-example"},
		},
		Style: charttheme.Style{Class: "max-w-full"},
	}
}

func interactiveCandlestickCode() string {
	return `@interactivecandlestick.Candlestick(interactivecandlestick.Config{
  Label: "Candlestick example",
  Categories: tradingDates,
  Series: []interactivecandlestick.Series{{
    Name: "OHLC",
    Data: []interactivecandlestick.Candle{
      {Open: 2320.26, Close: 2320.26, Low: 2287.3, High: 2362.94},
      // Remaining observations stay aligned with Categories.
    },
  }},
  DataZoom: []interactivecandlestick.DataZoom{
    {StartPercent: 50, EndPercent: 100},
  },
})`
}

func interactiveCandlestickInsideZoomCode() string {
	return `cfg.DataZoom = []interactivecandlestick.DataZoom{{
  Type: interactivecandlestick.DataZoomInside,
  StartPercent: 50, EndPercent: 100,
}}
@interactivecandlestick.Candlestick(cfg)`
}

func interactiveCandlestickBothZoomsCode() string {
	return `cfg.DataZoom = []interactivecandlestick.DataZoom{
  {Type: interactivecandlestick.DataZoomInside, StartPercent: 50, EndPercent: 100},
  {StartPercent: 50, EndPercent: 100},
}
@interactivecandlestick.Candlestick(cfg)`
}

func interactiveCandlestickYAxisZoomCode() string {
	return `cfg.DataZoom = []interactivecandlestick.DataZoom{{
  Axis: interactivecandlestick.DataZoomYAxis,
  StartPercent: 50, EndPercent: 100,
}}
@interactivecandlestick.Candlestick(cfg)`
}

func interactiveCandlestickStyleCode() string {
	return `cfg.Series[0].Options.Rise = interactivecandlestick.DirectionStyle{
  Class: "goshtoso-charts-candlestick__direction--decreasing",
}
cfg.Series[0].Options.Fall = interactivecandlestick.DirectionStyle{
  Class: "goshtoso-charts-candlestick__direction--increasing",
}
cfg.Series[0].Options.Marks = interactivecandlestick.MarkOptions{
  Highest: true, Lowest: true, ShowLabel: chart.Bool(true),
}
@interactivecandlestick.Candlestick(cfg)`
}
