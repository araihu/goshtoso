package charts

import (
	"fmt"
	"math/rand"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const interactiveScatter3DSeed = int64(20260728)

var interactiveScatter3DBasicPoints = deterministicScatter3DPoints()

func deterministicScatter3DPoints() []interactive.Point3D {
	rng := rand.New(rand.NewSource(interactiveScatter3DSeed))
	points := make([]interactive.Point3D, 80)
	for index := range points {
		points[index] = interactive.Point3D{
			Name: fmt.Sprintf("point-%02d", index+1),
			X:    float64(rng.Int63() % 100),
			Y:    float64(rng.Int63() % 100),
			Z:    float64(rng.Int63() % 100),
		}
	}
	return points
}

func sampleInteractiveScatter3DBasic() interactive.Instance {
	return interactive.Scatter3D(interactive.Scatter3DConfig{
		Label:   "basic Scatter3D example",
		Caption: "Eighty deterministic points in the 0 through 99 domain; expand the adjacent disclosure for every exact coordinate.",
		Series:  []interactive.Scatter3DSeries{{Name: "scatter3d", Points: append([]interactive.Point3D(nil), interactiveScatter3DBasicPoints...)}},
		VisualRange: &interactive.Scatter3DVisualRange{
			Max: 100, Calculable: interactive.Bool(true), Palette: interactive.Scatter3DPaletteColdToWarm,
		},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: "basic Scatter3D example"},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: "basic-scatter3d-example"},
		},
	})
}

func sampleInteractiveScatter3DStyled() interactive.Instance {
	return interactive.Scatter3D(interactive.Scatter3DConfig{
		Label:   "user-defined item style",
		Caption: "Three exact points with caller-selected colors and named axes.",
		Axes: &interactive.Scatter3DAxes{
			X: interactive.Scatter3DAxis{Name: "MY-X-AXIS", Show: interactive.Bool(true)},
			Y: interactive.Scatter3DAxis{Name: "MY-Y-AXIS"},
			Z: interactive.Scatter3DAxis{Name: "MY-Z-AXIS"},
		},
		Series: []interactive.Scatter3DSeries{{Name: "scatter3d", Points: []interactive.Point3D{
			{Name: "point1", X: 10, Y: 10, Z: 10, Color: "green"},
			{Name: "point2", X: 15, Y: 15, Z: 15, Color: "blue"},
			{Name: "point3", X: 20, Y: 20, Z: 20, Color: "red"},
		}}},
		Options: interactive.ChartOptions{
			Title:  &interactive.TitleOptions{Text: "user-defined item style"},
			Export: &chartcontrol.ExportOptions{Filename: "scatter3d-item-style"},
		},
	})
}

func interactiveScatter3DCode() string {
	return `@interactive.Scatter3D(interactive.Scatter3DConfig{
  Label: "Three-dimensional observations",
  Axes: &interactive.Scatter3DAxes{
    X: interactive.Scatter3DAxis{Name: "X", Show: interactive.Bool(true)},
    Y: interactive.Scatter3DAxis{Name: "Y"},
    Z: interactive.Scatter3DAxis{Name: "Z"},
  },
  Series: []interactive.Scatter3DSeries{{
    Name: "observations",
    Options: interactive.Scatter3DSeriesOptions{Class: "observation-series"},
    Points: []interactive.Point3D{
      {Name: "first", X: 10, Y: 10, Z: 10, Color: "green"},
      {Name: "second", X: 15, Y: 15, Z: 15, Class: "highlighted-point"},
    },
  }},
  Width: "100%", Height: "38rem",
})`
}
