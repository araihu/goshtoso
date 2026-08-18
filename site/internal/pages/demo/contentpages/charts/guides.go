package charts

import demo "github.com/araihu/goshtoso/site/internal/pages/demo"

type chartModeComparison struct {
	Capability  string
	Static      string
	Interactive string
}

var chartModeComparisons = []chartModeComparison{
	{Capability: "Output and delivery", Static: "Inline SVG is rendered with the HTML response. The chart image needs no browser renderer.", Interactive: "A browser runtime initializes a canvas after the HTML response arrives."},
	{Capability: "JavaScript", Static: "None for the chart image. Expand, fullscreen, and export still use the shared control runtime when enabled.", Interactive: "Required for drawing, interaction, theme synchronization, export, and optional live updates."},
	{Capability: "Export", Static: "SVG and browser-rasterized PNG are available by default. Transparent output is supported.", Interactive: "PNG snapshots are available by default. SVG and transparent output are not supported."},
	{Capability: "Print and documents", Static: "Best when scalable marks, selectable text, and stable server output matter.", Interactive: "Prints as rendered pixels. Export a PNG or provide a static sibling when the artifact must be stable."},
	{Capability: "Exploration", Static: "No tooltips, zoom, or live state in the chart image.", Interactive: "Typed options can enable tooltips. Zoom exists only on chart types that expose it, currently candlestick."},
	{Capability: "Live data and themes", Static: "Theme tokens update SVG presentation through CSS without redrawing the chart.", Interactive: "The chart instance follows theme changes. Categorical Bar and Line can consume full-snapshot SSE updates."},
	{Capability: "Exact values", Static: "SVG geometry is still not a substitute for exact data. Keep the component's disclosure or an adjacent table when precision matters.", Interactive: "Canvas pixels are not an accessible data model. Keep caption, summaries, and an adjacent equivalent data view."},
	{Capability: "Failure mode", Static: "Invalid configuration fails during server rendering. If control JavaScript is blocked, the SVG remains visible but actions do not work.", Interactive: "Invalid configuration fails during server rendering. If runtime or initialization is blocked, the canvas cannot become a usable chart."},
}

const staticModeCode = `@line.Line(line.Config{
  Label:   "Weekly revenue",
  Caption: "Revenue for the last five business days.",
  Labels:  []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
  Series: []line.Series{{
    Name: "Revenue",
    Values: []float64{12, 18, 14, 24, 21},
  }},
})`

const interactiveModeCode = `@interactiveline.Line(interactiveline.Config{
  Label:   "Weekly revenue",
  Caption: "Hover for values; use the adjacent table for exact data.",
  XAxis:   []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
  Series: []interactiveline.Series{{
    Name: "Revenue",
    Data: []interactiveline.Data{
      {Value: 12}, {Value: 18}, {Value: 14}, {Value: 24}, {Value: 21},
    },
  }},
  Options: interactive.ChartOptions{
    Tooltip: &interactive.TooltipOptions{
      Show: interactive.Bool(true),
      Trigger: "axis",
    },
  },
})`

const controlsStaticCode = `@line.Line(line.Config{
  Label: "Weekly revenue",
  Labels: []string{"Mon", "Tue", "Wed"},
  Series: []line.Series{{Name: "Revenue", Values: []float64{12, 18, 14}}},
  Controls: chartcontrol.Options{
    Fullscreen: true,
  },
  Export: &chartcontrol.ExportOptions{
    Filename: "weekly-revenue",
    Formats: []chartcontrol.ExportFormat{
      chartcontrol.ExportSVG,
      chartcontrol.ExportPNG,
    },
    Background: chartcontrol.ExportBackgroundTransparent,
    PixelRatio: 2,
  },
})`

const controlsInteractiveCode = `@interactiveline.Line(interactiveline.Config{
  Label: "Weekly revenue",
  XAxis: []string{"Mon", "Tue", "Wed"},
  Series: []interactiveline.Series{{
    Name: "Revenue",
    Data: []interactiveline.Data{{Value: 12}, {Value: 18}, {Value: 14}},
  }},
  Options: interactive.ChartOptions{
    Controls: chartcontrol.Options{
      Expand: chartcontrol.Bool(false),
      Fullscreen: true,
    },
    Export: &chartcontrol.ExportOptions{
      Filename: "weekly-revenue",
      Formats: []chartcontrol.ExportFormat{chartcontrol.ExportPNG},
    },
  },
})`

const wrapperModeInitialCode = `@line.Line(line.Config{
  Label:     "Weekly revenue",
  Labels:    []string{"Mon", "Tue", "Wed"},
  RootAttrs: templ.Attributes{"id": "weekly-revenue"},
  Series: []line.Series{{
    Name:   "Revenue",
    Values: []float64{12, 18, 14},
  }},
  Controls: chartcontrol.Options{
    // Omit Mode for WrapperModeEnabled, the zero/default.
    Mode: chartcontrol.WrapperModeHidden,
  },
})`

const wrapperModeJavaScriptCode = `const chart = document.querySelector("#weekly-revenue");
const wrapper = chart.closest("[data-goshtoso-chart-wrapper]");
const showButton = document.querySelector("#show-weekly-revenue");

wrapper.addEventListener("goshtoso-charts:wrapper-mode-change", (event) => {
  console.log(event.detail.previousMode, event.detail.mode);
});

// The helper returns false for an unknown mode or a missing wrapper.
const changed = window.__goshtosoChartsControls.setWrapperMode(
  wrapper,
  "hidden",
  showButton,
);`

func wrapperModeAlpineCode() string {
	return demo.DisplayCode(`<div
  x-data="{
    mode: 'enabled',
    setMode(next, focusReturn) {
      const wrapper = this.$refs.chart
        .querySelector('[data-goshtoso-chart-wrapper]');
      wrapper.dispatchEvent(new CustomEvent(
        'goshtoso-charts:set-wrapper-mode',
        { bubbles: true, detail: { mode: next, focusReturn } },
      ));
    },
  }"
  @goshtoso-charts:wrapper-mode-change="mode = $event.detail.mode"
>
  <button type="button" @click="setMode('hidden', $el)">Hide chart</button>
  <button type="button" @click="setMode('enabled', $el)">Show chart</button>
  <div x-ref="chart">
    @line.Line(line.Config{
      Label:  "Weekly revenue",
      Labels: []string{"Mon", "Tue", "Wed"},
      Series: []line.Series{{
        Name: "Revenue", Values: []float64{12, 18, 14},
      }},
    })
  </div>
</div>`)
}

const wrapperModeHTMXCode = `templ WeeklyRevenue(mode chartcontrol.WrapperMode) {
  <div id="weekly-revenue-slot">
    @line.Line(line.Config{
      Label:  "Weekly revenue",
      Labels: []string{"Mon", "Tue", "Wed"},
      Series: []line.Series{{
        Name: "Revenue", Values: []float64{12, 18, 14},
      }},
      Controls: chartcontrol.Options{Mode: mode},
    })
  </div>
}

<button
  hx-get="/charts/weekly-revenue?wrapper=hidden"
  hx-target="#weekly-revenue-slot"
  hx-swap="outerHTML"
>Hide chart</button>`
