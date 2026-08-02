// Package e2eimpact selects conservative Playwright build tags from a Git change range.
package e2eimpact

import "slices"

// Result is the stable machine-readable selector contract consumed by CI.
type Result struct {
	Mode         string   `json:"mode"`
	Tags         []string `json:"tags"`
	ChangedPaths []string `json:"changed_paths"`
	Reasons      []string `json:"reasons"`
}

func fullResult(paths []string, reasons ...string) Result {
	slices.Sort(paths)
	paths = slices.Compact(paths)
	slices.Sort(reasons)
	return Result{Mode: "full", Tags: []string{"full"}, ChangedPaths: paths, Reasons: reasons}
}
