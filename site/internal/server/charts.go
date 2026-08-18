package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	charts "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/charts"
)

type chartsPageRoute struct {
	Path        string
	Active      string
	Title       string
	Description string
	Render      func(*http.Request) templ.Component
}

func (s *Server) registerChartsRoutes() {
	for _, route := range chartsPageRoutes() {
		s.mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			s.handleChartsPage(w, r, route)
		})
	}
	s.mux.HandleFunc("/modules/charts/docs/theme-playground/frame", s.handleChartsThemePlaygroundFrame)
	s.mux.HandleFunc("/modules/charts/examples/live-availability/events", liveAvailabilityEvents)
}

func chartsPageRoutes() []chartsPageRoute {
	return []chartsPageRoute{
		chartsPage("/modules/charts", "module-charts", "Getting started", "Install Goshtoso Charts and render a first static or interactive chart.", charts.GettingStartedPage),
		chartsPage("/modules/charts/docs/chart-modes", "charts-chart-modes", "Static and interactive", "Choose the chart delivery model from the job the chart must do.", charts.ChartModesPage),
		{
			Path:        "/modules/charts/docs/chart-controls",
			Active:      "charts-chart-controls",
			Title:       "Chart controls",
			Description: "Configure expand, fullscreen, export, and wrapper lifecycle for charts.",
			Render: func(r *http.Request) templ.Component {
				return charts.ChartControlsPage(chartsFragmentRequest(r), charts.ParseChartControlExamples(r.URL.Query()))
			},
		},
		chartsPage("/modules/charts/docs/theme-playground", "charts-theme-playground", "Theme playground", "Compare static and interactive charts across Goshtoso themes.", charts.ThemePlaygroundPage),
		chartsPage("/modules/charts/attributions", "charts-attributions", "Attributions", "Chart, runtime, and asset attributions for Goshtoso Charts.", charts.AttributionsPage),

		chartsPage("/modules/charts/components/line", "charts-line", "Line chart", "Static/vector line chart examples rendered as SVG.", charts.LinePage),
		chartsPage("/modules/charts/components/bar", "charts-bar", "Bar chart", "Static/vector categorical comparisons in vertical, horizontal, grouped, or stacked layouts.", charts.BarPage),
		chartsPage("/modules/charts/components/pie", "charts-pie", "Pie chart", "Static/vector part-to-whole chart treatments.", charts.PiePage),
		chartsPage("/modules/charts/components/scatter", "charts-scatter", "Scatter chart", "Static/vector point relationships and distributions.", charts.ScatterPage),
		chartsPage("/modules/charts/components/radar", "charts-radar", "Radar chart", "Static/vector bounded multi-indicator profiles.", charts.RadarPage),
		chartsPage("/modules/charts/components/candlestick", "charts-candlestick", "Candlestick", "Static/vector open, high, low, and close ranges.", charts.CandlestickPage),
		chartsPage("/modules/charts/components/funnel", "charts-funnel", "Funnel chart", "Static/vector ordered stage values.", charts.FunnelPage),
		chartsPage("/modules/charts/components/heatmap", "charts-heatmap", "Heat map", "Static/vector values across a category grid.", charts.HeatMapPage),
		chartsPage("/modules/charts/components/table", "charts-table", "Table", "Chart-friendly tabular data presentation.", charts.TablePage),
		chartsPage("/modules/charts/components/violin", "charts-violin", "Violin chart", "Static/vector distribution shape and samples.", charts.ViolinPage),

		chartsPage("/modules/charts/components/interactive/bar", "charts-interactive-bar", "Interactive bar", "Interactive categorical comparisons with typed references and zoom.", charts.InteractiveBarPage),
		chartsPage("/modules/charts/components/interactive/line", "charts-interactive-line", "Interactive line", "Interactive trends, symbols, areas, and references.", charts.InteractiveLinePage),
		chartsPage("/modules/charts/components/interactive/scatter", "charts-interactive-scatter", "Interactive scatter", "Interactive point relationships and emphasis.", charts.InteractiveScatterPage),
		chartsPage("/modules/charts/components/interactive/scatter-3d", "charts-interactive-scatter-3d", "Interactive scatter 3D", "Interactive three-dimensional point clouds.", charts.InteractiveScatter3DPage),
		chartsPage("/modules/charts/components/interactive/bar-3d", "charts-interactive-bar-3d", "Interactive bar 3D", "Interactive three-dimensional bars.", charts.InteractiveBar3DPage),
		chartsPage("/modules/charts/components/interactive/surface-3d", "charts-interactive-surface-3d", "Interactive surface 3D", "Interactive mathematical surfaces.", charts.InteractiveSurface3DPage),
		chartsPage("/modules/charts/components/interactive/line-3d", "charts-interactive-line-3d", "Interactive line 3D", "Interactive three-dimensional paths.", charts.InteractiveLine3DPage),
		chartsPage("/modules/charts/components/interactive/pie", "charts-interactive-pie", "Interactive pie", "Interactive pie, donut, and rose treatments.", charts.InteractivePiePage),
		chartsPage("/modules/charts/components/interactive/radar", "charts-interactive-radar", "Interactive radar", "Interactive multi-indicator profiles.", charts.InteractiveRadarPage),
		chartsPage("/modules/charts/components/interactive/heatmap", "charts-interactive-heatmap", "Interactive heatmap", "Interactive Cartesian and calendar heatmaps.", charts.InteractiveHeatMapPage),
		chartsPage("/modules/charts/components/interactive/boxplot", "charts-interactive-boxplot", "Interactive box plot", "Interactive distribution summaries.", charts.InteractiveBoxPlotPage),
		chartsPage("/modules/charts/components/interactive/candlestick", "charts-interactive-candlestick", "Interactive candlestick", "Interactive OHLC ranges and zoom.", charts.InteractiveCandlestickPage),
		chartsPage("/modules/charts/components/interactive/gauge", "charts-interactive-gauge", "Interactive gauge", "Interactive progress and value gauges.", charts.InteractiveGaugePage),
		chartsPage("/modules/charts/components/interactive/funnel", "charts-interactive-funnel", "Interactive funnel", "Interactive ordered stage values.", charts.InteractiveFunnelPage),
		chartsPage("/modules/charts/components/interactive/graph", "charts-interactive-graph", "Interactive graph", "Interactive node and edge relationships.", charts.InteractiveGraphPage),
		chartsPage("/modules/charts/components/interactive/sankey", "charts-interactive-sankey", "Interactive Sankey", "Interactive weighted flows.", charts.InteractiveSankeyPage),
		chartsPage("/modules/charts/components/interactive/tree", "charts-interactive-tree", "Interactive tree", "Interactive hierarchical trees.", charts.InteractiveTreePage),
		chartsPage("/modules/charts/components/interactive/sunburst", "charts-interactive-sunburst", "Interactive sunburst", "Interactive radial hierarchies.", charts.InteractiveSunburstPage),
		chartsPage("/modules/charts/components/interactive/treemap", "charts-interactive-treemap", "Interactive treemap", "Interactive hierarchical area comparisons.", charts.InteractiveTreemapPage),
		chartsPage("/modules/charts/components/interactive/parallel", "charts-interactive-parallel", "Interactive parallel coordinates", "Interactive multivariate comparisons.", charts.InteractiveParallelPage),
		chartsPage("/modules/charts/components/interactive/theme-river", "charts-interactive-theme-river", "Interactive theme river", "Interactive change across aligned time streams.", charts.InteractiveThemeRiverPage),
		chartsPage("/modules/charts/components/interactive/word-cloud", "charts-interactive-word-cloud", "Interactive word cloud", "Interactive weighted word layouts.", charts.InteractiveWordCloudPage),
		chartsPage("/modules/charts/components/interactive/map", "charts-interactive-map", "Interactive map", "Interactive region maps.", charts.InteractiveMapPage),
		chartsPage("/modules/charts/components/interactive/geo", "charts-interactive-geo", "Interactive geo", "Interactive geographic coordinate views.", charts.InteractiveGeoPage),

		chartsPage("/modules/charts/examples/live-availability", "charts-live-availability", "Live availability", "A complete status snapshot updated over SSE.", charts.LiveAvailabilityExample),
	}
}

func chartsPage(path, active, title, description string, render func(bool) templ.Component) chartsPageRoute {
	return chartsPageRoute{
		Path:        path,
		Active:      active,
		Title:       title,
		Description: description,
		Render: func(r *http.Request) templ.Component {
			return render(chartsFragmentRequest(r))
		},
	}
}

func chartsFragmentRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true"
}

func (s *Server) handleChartsPage(w http.ResponseWriter, r *http.Request, route chartsPageRoute) {
	if route.Active == "charts-chart-controls" && chartsFragmentRequest(r) {
		examples := charts.ParseChartControlExamples(r.URL.Query())
		if example, ok := charts.ChartControlExampleForTarget(examples, r.Header.Get("HX-Target")); ok {
			renderChartsComponent(w, r, example)
			return
		}
	}

	meta := demo.PageMeta{
		Title:       route.Title,
		Description: route.Description,
		Path:        route.Path,
		Type:        "TechArticle",
	}
	content := route.Render(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if chartsFragmentRequest(r) {
		_ = demo.ComponentDocsFragment(meta, route.Active, content, storageAllowed(r)).Render(r.Context(), w)
		return
	}
	_ = demo.ComponentDocsLayout(meta, route.Active, content, storageAllowed(r)).Render(r.Context(), w)
}

func (s *Server) handleChartsThemePlaygroundFrame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = charts.ThemePlaygroundFrame().Render(r.Context(), w)
}

func renderChartsComponent(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "render chart fragment", http.StatusInternalServerError)
	}
}

func liveAvailabilityEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if err := writeAvailabilityEvent(w, availabilitySnapshot(time.Now(), 0)); err != nil {
		return
	}
	flusher.Flush()
	if r.Context().Err() != nil {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	step := 1
	for {
		select {
		case <-r.Context().Done():
			return
		case now := <-ticker.C:
			if err := writeAvailabilityEvent(w, availabilitySnapshot(now, step)); err != nil {
				return
			}
			flusher.Flush()
			step++
		}
	}
}

func writeAvailabilityEvent(w http.ResponseWriter, snapshot interactive.CartesianSnapshot) error {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: chart\ndata: %s\n\n", payload)
	return err
}

func availabilitySnapshot(now time.Time, step int) interactive.CartesianSnapshot {
	const bucketCount = 36
	const bucketWidth = 2 * time.Second
	categories := make([]string, bucketCount)
	series := []interactive.CartesianSnapshotSeries{
		{Name: "Healthy", Values: make([]float64, bucketCount)},
		{Name: "Degraded", Values: make([]float64, bucketCount)},
		{Name: "Down", Values: make([]float64, bucketCount)},
	}
	end := now.Truncate(bucketWidth)
	for index := range categories {
		bucketTime := end.Add(time.Duration(index-bucketCount+1) * bucketWidth)
		categories[index] = bucketTime.Format("15:04:05")
		state := 0
		switch phase := (step + index) % 24; {
		case phase >= 8 && phase <= 10:
			state = 1
		case phase >= 17 && phase <= 19:
			state = 2
		}
		series[state].Values[index] = 1
	}
	return interactive.CartesianSnapshot{Categories: categories, Series: series}
}
