package charts

const staticRadarUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"

type staticRadarCoverageEntry struct {
	Path      string
	SHA256    string
	Treatment string
}

func staticRadarUpstreamCoverage() []staticRadarCoverageEntry {
	return []staticRadarCoverageEntry{
		{Path: "examples/1-Painter/radar_chart-1-basic/main.go", SHA256: "0cf8dbdd72f6a398b7c560b544a0d800570d17a620549753b726171f996254d4", Treatment: "Basic budget comparison through the Painter API"},
		{Path: "examples/2-OptionFunc/radar_chart-1-basic/main.go", SHA256: "39a9427d6bb3bcff7d7627943210e2b19b934785c03896e9beffb1a35462c78e", Treatment: "Same basic budget comparison through option functions"},
	}
}

type staticRadarAPISource struct {
	Path   string
	SHA256 string
}

func staticRadarAPISources() []staticRadarAPISource {
	return []staticRadarAPISource{
		{Path: "radar_chart.go", SHA256: "50f9e29787665a03ab744be3081b9582cbb2d4064025245818665923849e79d7"},
		{Path: "series.go", SHA256: "953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217"},
		{Path: "series_label.go", SHA256: "d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611"},
		{Path: "title.go", SHA256: "e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26"},
		{Path: "legend.go", SHA256: "eaad1144ff1c5af84049a2968e952caba2dac3970558d542c18c6a98ba6d515b"},
		{Path: "chart_option.go", SHA256: "0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858"},
	}
}
