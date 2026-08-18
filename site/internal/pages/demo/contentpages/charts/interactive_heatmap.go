package charts

const (
	interactiveHeatMapUpstreamPath     = "examples/heatmap.go"
	interactiveHeatMapUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveHeatMapUpstreamSHA256   = "c08b194eafa5e02e941ad91f7ff8402448bc77b407cc97903b19063d06dd6f14"
)

type interactiveHeatMapUpstreamSpan struct {
	Name   string
	Lines  string
	SHA256 string
}

var interactiveHeatMapUpstreamInventory = []interactiveHeatMapUpstreamSpan{
	{Name: "hmData", Lines: "15-44", SHA256: "4e16c58885e24812b235c9adb709b6357b4dfa62765a448b306b94f1b50f6774"},
	{Name: "weekDays", Lines: "46", SHA256: "f2105456e595925b61ad3e74deea94a9ca7532f0f70a6bf6a9ff72360e3620c3"},
	{Name: "dayHrs", Lines: "48-51", SHA256: "abe6227a0d8c6fd7abb77b023bdd385d6ca46f8ea0a0c4623bbfff9db4895f8d"},
	{Name: "genHeatMapData", Lines: "54-64", SHA256: "c2e58d4394f9f3e352048d7f6e8863b71419ef26b97ac8dfc223d75328219bcb"},
	{Name: "heatMapBase", Lines: "66-93", SHA256: "56319619a561ed145f00ae74b66c0d9ad7403f5f5d508c4c1def0742f97f9218"},
	{Name: "genHeatMapCalendarData", Lines: "95-107", SHA256: "9f4cb439c4297c6cb550fa7680e6497390a024b47d8e8404dac75d1bf3ed96a8"},
	{Name: "heatMapCalendar", Lines: "109-142", SHA256: "9ee96ab15abff25ed188f25eb885a215bcba6ac69a074c5e21a0280ef32db596"},
	{Name: "HeatmapExamples", Lines: "144", SHA256: "bdb7c3f6f5387d8c8ec512b899e7853b7572ff96c04d9ddcbe640d42b2c7b702"},
	{Name: "HeatmapExamples.Examples", Lines: "146-158", SHA256: "aebfa38bdcaaeb0b9e27457f83e925c7b42d65e3fd6352f99fe318e1cae7d286"},
}
