package charts

import (
	"math"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveSurface3DBaseFormula  = "y = i / 60, x = j / 60, z = sin(x × π) × sin(y × π), with i and j from -60 through 59"
	interactiveSurface3DRoseFormula  = "y = i / 10, x = j / 10, z = sin(x² + y²) × x / π, with i and j from -30 through 29"
	interactiveSurface3DHeartFormula = "θ ∈ [0, 2π], φ ∈ [0, π]; hx = 13.5 × sin³(θ); c = 0.56 × ((1 + cos(θ)) / 2)²⁴; hz = 0.92 × (13cos(θ) − 5cos(2θ) − 2cos(3θ) − cos(4θ)) + c; s = sin(φ); x = s × hx; y = 5.2 × cos(φ); z = −1.5 + s × (hz + 1.5)"
	interactiveSurface3DHeartRows    = 49
	interactiveSurface3DHeartColumns = 65
)

var (
	interactiveSurface3DBasePoints  = generateInteractiveSurface3DBasePoints()
	interactiveSurface3DRosePoints  = generateInteractiveSurface3DRosePoints()
	interactiveSurface3DHeartPoints = generateInteractiveSurface3DHeartPoints()
)

func generateInteractiveSurface3DBasePoints() []interactive.Point3D {
	points := make([]interactive.Point3D, 0, 120*120)
	for i := -60; i < 60; i++ {
		y := float64(i) / 60
		for j := -60; j < 60; j++ {
			x := float64(j) / 60
			points = append(points, interactive.Point3D{
				X: x,
				Y: y,
				Z: math.Sin(x*math.Pi) * math.Sin(y*math.Pi),
			})
		}
	}
	return points
}

func generateInteractiveSurface3DRosePoints() []interactive.Point3D {
	points := make([]interactive.Point3D, 0, 60*60)
	for i := -30; i < 30; i++ {
		y := float64(i) / 10
		for j := -30; j < 30; j++ {
			x := float64(j) / 10
			points = append(points, interactive.Point3D{
				X: x,
				Y: y,
				Z: math.Sin(x*x+y*y) * x / math.Pi,
			})
		}
	}
	return points
}

func generateInteractiveSurface3DHeartPoints() []interactive.Point3D {
	points := make([]interactive.Point3D, 0, interactiveSurface3DHeartRows*interactiveSurface3DHeartColumns)
	for row := range interactiveSurface3DHeartRows {
		theta := 2 * math.Pi * float64(row) / float64(interactiveSurface3DHeartRows-1)
		outlineX := 13.5 * math.Pow(math.Sin(theta), 3)
		cleftLift := 0.56 * math.Pow((1+math.Cos(theta))/2, 24)
		outlineZ := 0.92*(13*math.Cos(theta)-5*math.Cos(2*theta)-2*math.Cos(3*theta)-math.Cos(4*theta)) + cleftLift
		for column := range interactiveSurface3DHeartColumns {
			phi := math.Pi * float64(column) / float64(interactiveSurface3DHeartColumns-1)
			scale := math.Sin(phi)
			points = append(points, interactive.Point3D{
				X: scale * outlineX,
				Y: 5.2 * math.Cos(phi),
				Z: -1.5 + scale*(outlineZ+1.5),
			})
		}
	}
	return points
}

func sampleInteractiveSurface3D(label, formula, filename string, points []interactive.Point3D) interactive.Instance {
	return interactive.Surface3D(interactive.Surface3DConfig{
		Label:   label,
		Caption: formula,
		Series: []interactive.Surface3DSeries{{
			Name: "surface3d", Points: append([]interactive.Point3D(nil), points...),
		}},
		VisualRange: &interactive.Surface3DVisualRange{
			Min: -3, Max: 3, Calculable: interactive.Bool(true), Palette: interactive.Surface3DPaletteColdToWarm,
		},
		DataSummary: interactive.Surface3DDataSummary{Formula: formula},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: label},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: filename},
		},
	})
}

func sampleInteractiveSurface3DHeart() interactive.Instance {
	return interactive.Surface3D(interactive.Surface3DConfig{
		Label:   "Rotating parametric heart",
		Caption: "A closed row-major mesh sampled entirely on the server.",
		Series: []interactive.Surface3DSeries{{
			Name:   "heart",
			Points: append([]interactive.Point3D(nil), interactiveSurface3DHeartPoints...),
			Mesh: &interactive.Surface3DMesh{
				Rows: interactiveSurface3DHeartRows, Columns: interactiveSurface3DHeartColumns,
			},
			Style: interactive.Surface3DSeriesStyle{
				Shading: interactive.Surface3DShadingLambert, Wireframe: interactive.Bool(false), Color: "#db2777",
			},
		}},
		Axes: &interactive.Surface3DAxes{
			X: interactive.Surface3DAxis{Name: "X", Show: interactive.Bool(false)},
			Y: interactive.Surface3DAxis{Name: "Y", Show: interactive.Bool(false)},
			Z: interactive.Surface3DAxis{Name: "Z", Show: interactive.Bool(false)},
		},
		Grid: interactive.Surface3DGrid{
			Width: 110, Height: 100, Depth: 36,
			View: &interactive.Surface3DView{AutoRotate: interactive.Bool(true), AutoRotateSpeed: 12},
		},
		DataSummary: interactive.Surface3DDataSummary{Formula: interactiveSurface3DHeartFormula},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: "Rotating parametric heart"},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: "parametric-heart"},
		},
	})
}

func interactiveSurface3DCode() string {
	return `@interactive.Surface3D(interactive.Surface3DConfig{
  Label: "Parametric heart",
  Series: []interactive.Surface3DSeries{{
    Name: "heart",
    Points: heartPoints, // Server-generated in row-major order.
    Mesh: &interactive.Surface3DMesh{Rows: 49, Columns: 65},
    Style: interactive.Surface3DSeriesStyle{
      Shading: interactive.Surface3DShadingLambert,
      Wireframe: interactive.Bool(false),
    },
  }},
  Grid: interactive.Surface3DGrid{View: &interactive.Surface3DView{
    AutoRotate: interactive.Bool(true),
  }},
  DataSummary: interactive.Surface3DDataSummary{
    Formula: "x = x(u,v); y = y(u,v); z = z(u,v)",
  },
})`
}
