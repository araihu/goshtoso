package charts

const staticHeatMapUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"

type staticHeatMapCoverageEntry struct {
	Path      string
	SHA256    string
	Treatment string
}

func staticHeatMapUpstreamCoverage() []staticHeatMapCoverageEntry {
	return []staticHeatMapCoverageEntry{{
		Path:      "examples/1-Painter/heat_map-1-basic/main.go",
		SHA256:    "c39a3d85a0df126da5d099a60e1491ae424d0768260ee738b7932288f1bf687f",
		Treatment: "Basic five-by-five matrix",
	}}
}

type staticHeatMapSourceSpan struct {
	Path   string
	Lines  string
	Name   string
	SHA256 string
	Role   string
}

func staticHeatMapSourceSpans() []staticHeatMapSourceSpan {
	const path = "examples/1-Painter/heat_map-1-basic/main.go"
	return []staticHeatMapSourceSpan{
		{Path: path, Lines: "14-22", Name: "writeFile", SHA256: "4582ad5f11c031d2b70604df82f08d30ff21d977b6affb87bf0ade5eb7ae88ce", Role: "Filesystem output helper; no chart behavior"},
		{Path: path, Lines: "24-51", Name: "main", SHA256: "227e252486af0ec1e89a2c6920253c42bc26ff2f0366907409ca9af320319fd9", Role: "Five-by-five data, centered title, named axes, and 600x400 presentation"},
		{Path: path, Lines: "25-31", Name: "matrix literal", SHA256: "85c0b5b518e8d8d8689b6579e1377224dba3f87f844e69a685203740d59e4705", Role: "Exact twenty-five-value matrix"},
		{Path: path, Lines: "33-37", Name: "chart options", SHA256: "44f2aeb070244f692526b134d79f35d364f4cf588dc30cb044d2d8f10f635182", Role: "Data binding, centered title, and X/Y axis titles"},
		{Path: path, Lines: "39-50", Name: "painter and output", SHA256: "a593149de21fcd6622c896759f082626f4557d2c3fb7d158a350d8bd15b4bea7", Role: "600x400 PNG rendering and filesystem delivery"},
	}
}

type staticHeatMapAPISource struct {
	Path   string
	SHA256 string
}

func staticHeatMapAPISources() []staticHeatMapAPISource {
	return []staticHeatMapAPISource{
		{Path: "heat_map.go", SHA256: "dd9b80660b9e0c0b11e5ec9ef00f5f10005a7edc1061da08f928cea15324b23c"},
		{Path: "series_label.go", SHA256: "d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611"},
		{Path: "title.go", SHA256: "e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26"},
		{Path: "axis.go", SHA256: "72d8ed4c1253122a3ac49e76fbf00ff905a75aff26bcc9e4b6aea48325372a07"},
		{Path: "chart_option.go", SHA256: "0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858"},
		{Path: "painter.go", SHA256: "f4ac102e9b21623765e2fdfe4c0910a03265bc751b9f5d019ae41e80611be959"},
	}
}
