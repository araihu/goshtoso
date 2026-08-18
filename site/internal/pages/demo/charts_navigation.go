package demo

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/sidebar"
	sidebaricons "github.com/araihu/goshtoso/site/internal/demoicons/heroicons"
)

type chartsSidebarLink struct {
	ID          string
	Label       string
	Href        string
	Description string
}

type chartsSidebarGroup struct {
	Title string
	Items []chartsSidebarLink
}

func chartsDocsNavigation(active string) componentdocshell.Navigation {
	items := make([]sidebar.Item, 0, len(chartsSidebarTopLinks()))
	for _, link := range chartsSidebarTopLinks() {
		items = append(items, sidebar.Item{
			ID:        link.ID,
			Label:     link.Label,
			Href:      link.Href,
			Icon:      chartsSidebarIcon(link.ID),
			Active:    active == link.ID,
			LinkAttrs: navHxAttrs(link.Href, link.Label),
		})
	}

	sections := make([]sidebar.Section, 0, len(chartsSidebarGroups()))
	for _, group := range chartsSidebarGroups() {
		section := sidebar.Section{Title: group.Title}
		for _, link := range group.Items {
			section.Items = append(section.Items, sidebar.Item{
				ID:        link.ID,
				Label:     link.Label,
				Href:      link.Href,
				Active:    active == link.ID,
				LinkAttrs: navHxAttrs(link.Href, link.Label),
			})
		}
		sections = append(sections, section)
	}

	return componentdocshell.Navigation{
		Items:         items,
		SectionsTitle: "Documentation",
		Sections:      sections,
		DisableSearch: true,
	}
}

func chartsSidebarTopLinks() []chartsSidebarLink {
	return []chartsSidebarLink{
		{ID: "module-charts", Label: "Getting started", Href: "/modules/charts", Description: "Install Goshtoso Charts and render your first static or interactive chart."},
		{ID: "charts-chart-modes", Label: "Static and interactive", Href: "/modules/charts/docs/chart-modes", Description: "Choose between server-rendered SVG and browser-rendered charts."},
		{ID: "charts-chart-controls", Label: "Chart controls", Href: "/modules/charts/docs/chart-controls", Description: "Configure expand, fullscreen, export, and wrapper lifecycle."},
		{ID: "charts-theme-playground", Label: "Theme playground", Href: "/modules/charts/docs/theme-playground", Description: "Compare static and interactive charts across Goshtoso themes."},
		{ID: "charts-attributions", Label: "Attributions", Href: "/modules/charts/attributions", Description: "Review chart, runtime, and asset attributions."},
	}
}

func chartsSidebarGroups() []chartsSidebarGroup {
	return []chartsSidebarGroup{
		{
			Title: "Static / Vector",
			Items: []chartsSidebarLink{
				{ID: "charts-line", Label: "Line chart", Href: "/modules/charts/components/line", Description: "Server-rendered line chart SVG."},
				{ID: "charts-bar", Label: "Bar chart", Href: "/modules/charts/components/bar", Description: "Categorical comparisons in vertical, horizontal, grouped, or stacked layouts."},
				{ID: "charts-pie", Label: "Pie chart", Href: "/modules/charts/components/pie", Description: "Static part-to-whole chart treatments."},
				{ID: "charts-scatter", Label: "Scatter chart", Href: "/modules/charts/components/scatter", Description: "Static point relationships and distributions."},
				{ID: "charts-radar", Label: "Radar chart", Href: "/modules/charts/components/radar", Description: "Static bounded multi-indicator profiles."},
				{ID: "charts-candlestick", Label: "Candlestick", Href: "/modules/charts/components/candlestick", Description: "Static open, high, low, and close ranges."},
				{ID: "charts-funnel", Label: "Funnel chart", Href: "/modules/charts/components/funnel", Description: "Static ordered stage values."},
				{ID: "charts-heatmap", Label: "Heat map", Href: "/modules/charts/components/heatmap", Description: "Static values across a category grid."},
				{ID: "charts-table", Label: "Table", Href: "/modules/charts/components/table", Description: "Chart-friendly tabular data presentation."},
				{ID: "charts-violin", Label: "Violin chart", Href: "/modules/charts/components/violin", Description: "Static distribution shape and samples."},
			},
		},
		{
			Title: "Interactive / Cartesian",
			Items: []chartsSidebarLink{
				{ID: "charts-interactive-bar", Label: "Bar", Href: "/modules/charts/components/interactive/bar", Description: "Interactive categorical comparisons."},
				{ID: "charts-interactive-line", Label: "Line", Href: "/modules/charts/components/interactive/line", Description: "Interactive trends, symbols, areas, and references."},
				{ID: "charts-interactive-scatter", Label: "Scatter", Href: "/modules/charts/components/interactive/scatter", Description: "Interactive point relationships and emphasis."},
				{ID: "charts-interactive-candlestick", Label: "Candlestick", Href: "/modules/charts/components/interactive/candlestick", Description: "Interactive OHLC ranges and zoom."},
				{ID: "charts-interactive-heatmap", Label: "Heatmap", Href: "/modules/charts/components/interactive/heatmap", Description: "Interactive Cartesian and calendar heatmaps."},
			},
		},
		{
			Title: "Interactive / 3D",
			Items: []chartsSidebarLink{
				{ID: "charts-interactive-scatter-3d", Label: "Scatter 3D", Href: "/modules/charts/components/interactive/scatter-3d", Description: "Interactive three-dimensional point clouds."},
				{ID: "charts-interactive-bar-3d", Label: "Bar 3D", Href: "/modules/charts/components/interactive/bar-3d", Description: "Interactive three-dimensional bars."},
				{ID: "charts-interactive-surface-3d", Label: "Surface 3D", Href: "/modules/charts/components/interactive/surface-3d", Description: "Interactive mathematical surfaces."},
				{ID: "charts-interactive-line-3d", Label: "Line 3D", Href: "/modules/charts/components/interactive/line-3d", Description: "Interactive three-dimensional paths."},
			},
		},
		{
			Title: "Interactive / Statistical",
			Items: []chartsSidebarLink{
				{ID: "charts-interactive-pie", Label: "Pie", Href: "/modules/charts/components/interactive/pie", Description: "Interactive pie, donut, and rose treatments."},
				{ID: "charts-interactive-radar", Label: "Radar", Href: "/modules/charts/components/interactive/radar", Description: "Interactive multi-indicator profiles."},
				{ID: "charts-interactive-boxplot", Label: "Box plot", Href: "/modules/charts/components/interactive/boxplot", Description: "Interactive distribution summaries."},
				{ID: "charts-interactive-gauge", Label: "Gauge", Href: "/modules/charts/components/interactive/gauge", Description: "Interactive progress and value gauges."},
				{ID: "charts-interactive-funnel", Label: "Funnel", Href: "/modules/charts/components/interactive/funnel", Description: "Interactive ordered stage values."},
				{ID: "charts-interactive-parallel", Label: "Parallel coordinates", Href: "/modules/charts/components/interactive/parallel", Description: "Interactive multivariate comparisons."},
				{ID: "charts-interactive-theme-river", Label: "Theme river", Href: "/modules/charts/components/interactive/theme-river", Description: "Interactive change across aligned time streams."},
				{ID: "charts-interactive-word-cloud", Label: "Word cloud", Href: "/modules/charts/components/interactive/word-cloud", Description: "Interactive weighted word layouts."},
			},
		},
		{
			Title: "Interactive / Geographic",
			Items: []chartsSidebarLink{
				{ID: "charts-interactive-map", Label: "Map", Href: "/modules/charts/components/interactive/map", Description: "Interactive region maps."},
				{ID: "charts-interactive-geo", Label: "Geo", Href: "/modules/charts/components/interactive/geo", Description: "Interactive geographic coordinate views."},
			},
		},
		{
			Title: "Interactive / Relationships",
			Items: []chartsSidebarLink{
				{ID: "charts-interactive-graph", Label: "Graph", Href: "/modules/charts/components/interactive/graph", Description: "Interactive node and edge relationships."},
				{ID: "charts-interactive-sankey", Label: "Sankey", Href: "/modules/charts/components/interactive/sankey", Description: "Interactive weighted flows."},
				{ID: "charts-interactive-tree", Label: "Tree", Href: "/modules/charts/components/interactive/tree", Description: "Interactive hierarchical trees."},
				{ID: "charts-interactive-sunburst", Label: "Sunburst", Href: "/modules/charts/components/interactive/sunburst", Description: "Interactive radial hierarchies."},
				{ID: "charts-interactive-treemap", Label: "Treemap", Href: "/modules/charts/components/interactive/treemap", Description: "Interactive hierarchical area comparisons."},
			},
		},
		{
			Title: "Examples",
			Items: []chartsSidebarLink{
				{ID: "charts-live-availability", Label: "Live availability", Href: "/modules/charts/examples/live-availability", Description: "A complete status snapshot updated over SSE."},
			},
		},
	}
}

func chartsSidebarIcon(id string) templ.Component {
	symbol := sidebaricons.IconHeroiconsOptimized24OutlineQueueList
	switch id {
	case "module-charts":
		symbol = sidebaricons.IconHeroiconsOptimized24OutlineArrowDownTray
	case "charts-chart-controls":
		symbol = sidebaricons.IconHeroiconsOptimized24OutlineClipboardDocumentList
	case "charts-theme-playground":
		symbol = sidebaricons.IconHeroiconsOptimized24OutlineSwatch
	case "charts-attributions":
		symbol = sidebaricons.IconHeroiconsOptimized24OutlineHeart
	}
	return sidebaricons.Icon(sidebaricons.Config{Symbol: symbol, Decorative: true})
}
