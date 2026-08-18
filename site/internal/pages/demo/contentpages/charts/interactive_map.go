package charts

import (
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveMapUpstreamPath     = "examples/map.go"
	interactiveMapUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveMapUpstreamSHA256   = "3b59b5cb7ed392f3fa436d51fd420704ab2e82e439c95b226d35d12b913cf9da"
	interactiveMapGeometryRevision = "IBGE-MMD-2025"
	interactiveMapGeometrySHA256   = "1b3719c82f6e2278a3e6ea8b7fc2e195460ee6a7de1546d0a8e05e6d0174bb3d"
)

type interactiveMapVariant struct {
	variant interactive.MapVariant
	chart   interactive.Instance
}

// Fixed illustrative values preserve upstream's typed [0,150) value shape.
// Names and UF codes match IBGE's official state identity table.
var interactiveMapBrazilRegions = []interactive.MapRegion{
	{Name: "Rondônia", Code: "RO", Value: 42}, {Name: "Acre", Code: "AC", Value: 28},
	{Name: "Amazonas", Code: "AM", Value: 81}, {Name: "Roraima", Code: "RR", Value: 19},
	{Name: "Pará", Code: "PA", Value: 96}, {Name: "Amapá", Code: "AP", Value: 24},
	{Name: "Tocantins", Code: "TO", Value: 37}, {Name: "Maranhão", Code: "MA", Value: 73},
	{Name: "Piauí", Code: "PI", Value: 48}, {Name: "Ceará", Code: "CE", Value: 102},
	{Name: "Rio Grande do Norte", Code: "RN", Value: 55}, {Name: "Paraíba", Code: "PB", Value: 61},
	{Name: "Pernambuco", Code: "PE", Value: 118}, {Name: "Alagoas", Code: "AL", Value: 52},
	{Name: "Sergipe", Code: "SE", Value: 35}, {Name: "Bahia", Code: "BA", Value: 134},
	{Name: "Minas Gerais", Code: "MG", Value: 126}, {Name: "Espírito Santo", Code: "ES", Value: 58},
	{Name: "Rio de Janeiro", Code: "RJ", Value: 121}, {Name: "São Paulo", Code: "SP", Value: 146},
	{Name: "Paraná", Code: "PR", Value: 109}, {Name: "Santa Catarina", Code: "SC", Value: 87},
	{Name: "Rio Grande do Sul", Code: "RS", Value: 112}, {Name: "Mato Grosso do Sul", Code: "MS", Value: 46},
	{Name: "Mato Grosso", Code: "MT", Value: 64}, {Name: "Goiás", Code: "GO", Value: 92},
	{Name: "Distrito Federal", Code: "DF", Value: 76},
}

func sampleInteractiveMaps() []interactiveMapVariant {
	return []interactiveMapVariant{
		{interactive.MapVariantBasic, sampleInteractiveMap("Brazil states", interactive.MapVariantBasic)},
		{interactive.MapVariantLabels, sampleInteractiveMap("State labels", interactive.MapVariantLabels)},
		{interactive.MapVariantScale, sampleInteractiveMap("Brazil value scale", interactive.MapVariantScale)},
		{interactive.MapVariantRegional, sampleInteractiveMap("Southeast focus", interactive.MapVariantRegional)},
		{interactive.MapVariantTheme, sampleInteractiveMap("Theme-aware Brazil", interactive.MapVariantTheme)},
	}
}

func sampleInteractiveMap(title string, variant interactive.MapVariant) interactive.Instance {
	regions := append([]interactive.MapRegion(nil), interactiveMapBrazilRegions...)
	if variant == interactive.MapVariantRegional {
		for index := range regions {
			if regions[index].Code == "SP" {
				regions[index].Class = interactiveGeoSemanticPaintClass
			}
			if regions[index].Code == "RJ" {
				regions[index].Color = "#dc2626"
			}
		}
	}
	return interactive.Map(interactive.MapConfig{
		Label: title, Caption: "Illustrative typed values for all 26 states and the Federal District appear below the responsive map.",
		Geometry: interactive.MapGeometryBrazil, Variant: variant,
		Series: interactive.MapSeries{Name: "Brazil state values", Regions: regions},
		Scale:  &interactive.MapScale{Min: 0, Max: 150, Calculable: interactive.Bool(true)},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: title},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: strings.ReplaceAll(strings.ToLower(title), " ", "-")},
		},
	})
}

func interactiveMapCode() string {
	return `@interactive.Map(interactive.MapConfig{
  Label: "Brazil states",
  Geometry: interactive.MapGeometryBrazil,
  Variant: interactive.MapVariantScale,
  Series: interactive.MapSeries{
    Name: "Brazil state values",
    Regions: brazilStateValues, // all 26 states plus DF, with Name, UF Code, and Value
  },
  Scale: &interactive.MapScale{
    Min: 0, Max: 150, Calculable: interactive.Bool(true),
  },
  Width: "100%", Height: "500px",
})`
}
