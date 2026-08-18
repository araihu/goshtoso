package charts

import (
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveGeoUpstreamPath       = "examples/geo.go"
	interactiveGeoUpstreamRevision   = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveGeoUpstreamSHA256     = "3a6dbe86c34e5ea478b1dea5430c10cac9f7c4905264e12fc37654f0f5d4550a"
	interactiveGeoGeometryRevision   = "IBGE-MMD-2025"
	interactiveGeoBrazilSHA256       = "1b3719c82f6e2278a3e6ea8b7fc2e195460ee6a7de1546d0a8e05e6d0174bb3d"
	interactiveGeoSaoPauloSHA256     = "657dee960c4c4d991f5b0e6d59681152d5e2b9c48091e5094085a666c97ff317"
	interactiveGeoSemanticPaintClass = "text-on-surface-strong dark:text-on-surface-dark-strong"
)

type interactiveGeoVariant struct {
	name  string
	chart interactive.Instance
}

// Fixed values preserve upstream's [0,100) shape while coordinates identify
// Brazilian capitals from north to south.
var interactiveGeoBrazilPoints = []interactive.GeoPoint{
	{Name: "Manaus", Longitude: -60.02, Latitude: -3.12, Value: 81},
	{Name: "Recife", Longitude: -34.88, Latitude: -8.05, Value: 27},
	{Name: "Brasília", Longitude: -47.88, Latitude: -15.79, Value: 47},
	{Name: "Rio de Janeiro", Longitude: -43.17, Latitude: -22.91, Value: 59},
	{Name: "São Paulo", Longitude: -46.63, Latitude: -23.55, Value: 18},
	{Name: "Porto Alegre", Longitude: -51.23, Latitude: -30.03, Value: 63},
}

var interactiveGeoSaoPauloPoints = []interactive.GeoPoint{
	{Name: "São Paulo", Longitude: -46.63, Latitude: -23.55, Value: 12},
	{Name: "Campinas", Longitude: -47.06, Latitude: -22.91, Value: 76},
	{Name: "Ribeirão Preto", Longitude: -47.81, Latitude: -21.18, Value: 41},
}

func sampleInteractiveGeos() []interactiveGeoVariant {
	return []interactiveGeoVariant{
		{name: "effect-scatter", chart: sampleInteractiveGeoBrazil()},
		{name: "scatter", chart: sampleInteractiveGeoSaoPaulo()},
	}
}

func sampleInteractiveGeoBrazil() interactive.Instance {
	points := append([]interactive.GeoPoint(nil), interactiveGeoBrazilPoints...)
	points[0].Class = interactiveGeoSemanticPaintClass
	points[1].Color = "#dc2626"
	return interactive.Geo(interactive.GeoConfig{
		Label: "Brazil capitals", Caption: "Exact capital coordinates and illustrative values appear below the responsive chart.",
		Geometry:      interactive.GeoGeometryBrazil,
		GeometryPaint: interactive.GeoPaint{Class: interactiveGeoSemanticPaintClass},
		Series: []interactive.GeoSeries{{
			Name: "Brazil capitals", Kind: interactive.GeoSeriesEffectScatter, Color: "#7c3aed",
			Points: points, Ripple: &interactive.RippleOptions{Period: 4, Scale: 6, BrushType: "stroke"},
		}},
		Options: interactive.ChartOptions{
			Title: &interactive.TitleOptions{Text: "Brazil capitals"}, Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export: &chartcontrol.ExportOptions{Filename: "brazil-capitals"},
		},
	})
}

func sampleInteractiveGeoSaoPaulo() interactive.Instance {
	points := append([]interactive.GeoPoint(nil), interactiveGeoSaoPauloPoints...)
	points[1].Color = "#2563eb"
	points[2].Class = interactiveGeoSemanticPaintClass
	return interactive.Geo(interactive.GeoConfig{
		Label: "São Paulo cities", Caption: "Exact city coordinates appear over São Paulo municipality boundaries.",
		Geometry: interactive.GeoGeometrySaoPaulo, GeometryPaint: interactive.GeoPaint{Color: "#e2e8f0"},
		VisualRange: &interactive.GeoVisualRange{Min: 0, Max: 100, Calculable: interactive.Bool(true)},
		Series:      []interactive.GeoSeries{{Name: "São Paulo cities", Kind: interactive.GeoSeriesScatter, Class: interactiveGeoSemanticPaintClass, Points: points}},
		Options: interactive.ChartOptions{
			Title: &interactive.TitleOptions{Text: "São Paulo cities"}, Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export: &chartcontrol.ExportOptions{Filename: "sao-paulo-cities"},
		},
	})
}

func interactiveGeoCode() string {
	return `@interactive.Geo(interactive.GeoConfig{
  Label: "São Paulo cities",
  Geometry: interactive.GeoGeometrySaoPaulo,
  VisualRange: &interactive.GeoVisualRange{Min: 0, Max: 100},
  Series: []interactive.GeoSeries{{
    Name: "São Paulo cities",
    Kind: interactive.GeoSeriesScatter,
    Points: []interactive.GeoPoint{
      {Name: "São Paulo", Longitude: -46.63, Latitude: -23.55, Value: 12},
      {Name: "Campinas", Longitude: -47.06, Latitude: -22.91, Value: 76},
      {Name: "Ribeirão Preto", Longitude: -47.81, Latitude: -21.18, Value: 41},
    },
  }},
  Width: "100%", Height: "500px",
  Options: interactive.ChartOptions{
    Export: &chartcontrol.ExportOptions{Filename: "` + strings.ReplaceAll(strings.ToLower("São Paulo cities"), " ", "-") + `"},
  },
})`
}
