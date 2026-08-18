package charts

import (
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactiveradar "github.com/araihu/goshtoso-charts/components/interactive/radar"
	"github.com/araihu/goshtoso-charts/components/line"
	"github.com/araihu/goshtoso-charts/components/pie"
	selectfield "github.com/araihu/goshtoso/components/select"
)

func themePlaygroundThemes() []selectfield.Option {
	themes := componentdocshell.DefaultThemes()
	for index := range themes {
		themes[index].Selected = themes[index].Value == "araihu"
	}
	return themes
}

func themePlaygroundShellData() string {
	return `componentDocShell({"persist":false,"theme":"araihu","colorScheme":"light"})`
}

func themePlaygroundStaticLine() line.Config {
	cfg := sampleBasicLine()
	cfg.Label = "Static line theme preview"
	cfg.Caption = ""
	cfg.Width = 480
	cfg.Height = 300
	cfg.Controls = chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}
	cfg.Export = nil
	return cfg
}

func themePlaygroundStaticPie() pie.Config {
	cfg := sampleBasicPie()
	cfg.Label = "Static pie theme preview"
	cfg.Caption = ""
	cfg.Width = 480
	cfg.Height = 300
	cfg.Controls = chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}
	cfg.Export = nil
	return cfg
}

func themePlaygroundInteractiveBar() interactive.BarConfig {
	cfg := sampleInteractiveBar()
	cfg.Label = "Interactive bar theme preview"
	cfg.Caption = ""
	cfg.Width = "100%"
	cfg.Height = "300px"
	cfg.Options.Controls = chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}
	cfg.Options.Export = nil
	return cfg
}

func themePlaygroundInteractiveRadar() interactiveradar.Config {
	cfg := sampleInteractiveRadarBase()
	cfg.Label = "Interactive radar theme preview"
	cfg.Caption = ""
	cfg.Width = "100%"
	cfg.Height = "300px"
	cfg.Options.Controls = chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}
	cfg.Options.Export = nil
	return cfg
}
