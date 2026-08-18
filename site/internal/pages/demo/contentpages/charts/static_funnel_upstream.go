package charts

const staticFunnelUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"

type staticFunnelCoverageEntry struct {
	Path      string
	SHA256    string
	Treatment string
}

func staticFunnelUpstreamCoverage() []staticFunnelCoverageEntry {
	return []staticFunnelCoverageEntry{
		{Path: "examples/1-Painter/funnel_chart-1-basic/main.go", SHA256: "a54875c89cd0be43fa7c0614520a11c489562bbc72f8c577a314eb3f24f75a6d", Treatment: "Basic seven-stage funnel"},
		{Path: "examples/2-OptionFunc/funnel_chart-1-basic/main.go", SHA256: "332ffeb340e1236170f41a3a93a46499897af75df261ed60f7ac01dd35ae4893", Treatment: "Same seven-stage funnel through option functions"},
	}
}

type staticFunnelSourceSpan struct {
	Path   string
	Lines  string
	Name   string
	SHA256 string
	Role   string
}

func staticFunnelSourceSpans() []staticFunnelSourceSpan {
	return []staticFunnelSourceSpan{
		{Path: "examples/1-Painter/funnel_chart-1-basic/main.go", Lines: "14-22", Name: "writeFile", SHA256: "9a3e255a47b40c36123b225677519f09a62c7f0f38cd1341515d30ce687f1fc5", Role: "Filesystem output helper; no chart behavior"},
		{Path: "examples/1-Painter/funnel_chart-1-basic/main.go", Lines: "24-46", Name: "main", SHA256: "82bd67be63d88d6ef2312890964ed595ff9a12990ee89bc4edc85e4bc93f9892", Role: "Seven-stage data, title, legend padding, and 600x400 presentation"},
		{Path: "examples/2-OptionFunc/funnel_chart-1-basic/main.go", Lines: "14-22", Name: "writeFile", SHA256: "9a3e255a47b40c36123b225677519f09a62c7f0f38cd1341515d30ce687f1fc5", Role: "Filesystem output helper; no chart behavior"},
		{Path: "examples/2-OptionFunc/funnel_chart-1-basic/main.go", Lines: "24-44", Name: "main", SHA256: "1289983e007306c2d3b1d44136aaa35ec56691a5d05f502a3c718d6d356b69a3", Role: "Duplicate seven-stage presentation through option functions"},
		{Path: "examples/2-OptionFunc/web-1/main.go", Lines: "145-547", Name: "indexHandler", SHA256: "c9970d3ac7aab849623793deee832f1c1fec5569c4964e3b84b61857424b8068", Role: "Mixed-family web composition containing the supplementary Funnel literal"},
		{Path: "examples/2-OptionFunc/web-1/main.go", Lines: "450-487", Name: "Funnel literal", SHA256: "f2a990a30d9a89a7d982bc79522c1fb11b2611173ac54c31a19382a543b1ba45", Role: "Five-stage supplementary dataset and default presentation"},
	}
}

type staticFunnelAPISource struct {
	Path   string
	SHA256 string
}

func staticFunnelAPISources() []staticFunnelAPISource {
	return []staticFunnelAPISource{
		{Path: "funnel_chart.go", SHA256: "d8c1d8e5ee84ae0f534da3046bd7d43973f7afbf9298e8b8d86c4200f0db6cfb"},
		{Path: "series.go", SHA256: "953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217"},
		{Path: "series_label.go", SHA256: "d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611"},
		{Path: "title.go", SHA256: "e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26"},
		{Path: "legend.go", SHA256: "eaad1144ff1c5af84049a2968e952caba2dac3970558d542c18c6a98ba6d515b"},
		{Path: "chart_option.go", SHA256: "0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858"},
	}
}
