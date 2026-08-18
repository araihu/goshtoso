package charts

import (
	"math"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const interactiveLine3DFormula = "t = i / 1000; x = (1 + 0.25 × cos(75 × t)) × cos(t); y = (1 + 0.25 × cos(75 × t)) × sin(t); z = t + 2 × sin(75 × t), with i from 0 through 24999"

var interactiveLine3DPoints = generateInteractiveLine3DPoints()

func generateInteractiveLine3DPoints() []interactive.Point3D {
	points := make([]interactive.Point3D, 0, 25000)
	for i := range 25000 {
		t := float64(i) / 1000
		radius := 1 + 0.25*math.Cos(75*t)
		points = append(points, interactive.Point3D{
			X: radius * math.Cos(t),
			Y: radius * math.Sin(t),
			Z: t + 2*math.Sin(75*t),
		})
	}
	return points
}

func sampleInteractiveLine3D(label, filename string, autoRotate bool) interactive.Instance {
	grid := interactive.Line3DGrid{}
	if autoRotate {
		grid.View = &interactive.Line3DView{AutoRotate: interactive.Bool(true)}
	}
	return interactive.Line3D(interactive.Line3DConfig{
		Label:   label,
		Caption: interactiveLine3DFormula,
		Series: []interactive.Line3DSeries{{
			Name: "line3D", Points: append([]interactive.Point3D(nil), interactiveLine3DPoints...),
		}},
		VisualRange: &interactive.Line3DVisualRange{
			Min: 0, Max: 30, Calculable: interactive.Bool(true),
		},
		Grid: grid,
		DataSummary: interactive.Line3DDataSummary{
			Formula: interactiveLine3DFormula, Parameter: "t", ParameterMin: 0, ParameterMax: 24.999,
		},
		Options: interactive.ChartOptions{
			Title:    &interactive.TitleOptions{Text: label},
			Controls: chartcontrol.Options{},
			Export:   &chartcontrol.ExportOptions{Filename: filename},
		},
	})
}

func interactiveLine3DCode() string {
	return `@interactive.Line3D(interactive.Line3DConfig{
  Label: "basic line3d example",
  Series: []interactive.Line3DSeries{{
    Name: "line3D",
    Points: points,
  }},
  VisualRange: &interactive.Line3DVisualRange{
    Max: 30,
    Calculable: interactive.Bool(true),
  },
  Grid: interactive.Line3DGrid{
    View: &interactive.Line3DView{AutoRotate: interactive.Bool(false)},
  },
  DataSummary: interactive.Line3DDataSummary{
    Formula: "t = i / 1000; x = ...; y = ...; z = ...",
    Parameter: "t", ParameterMin: 0, ParameterMax: 24.999,
  },
})`
}
