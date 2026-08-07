// Package iconpack generates consumer-local, attributed SVG icon packs from a
// verified Arai Hu Assets release boundary.
package iconpack

import (
	"context"
)

const (
	// OutputSchemaVersion identifies the generated manifest and provenance format.
	OutputSchemaVersion = 1
	toolName            = "goshtoso-iconpack"
)

// Options describes one deterministic icon-pack generation.
type Options struct {
	ReleaseRoot       string
	ReleaseArchive    string
	Release           string
	ArchiveSHA256     string
	CatalogSHA256     string
	ReleaseJSONSHA256 string
	ChecksumsSHA256   string
	Names             []string
	SelectionManifest string
	OutputDir         string
	Package           string
	ConstPrefix       string
	SpriteURL         string
	Check             bool
}

// Result identifies a verified or newly published output.
type Result struct {
	OutputDir     string
	Release       string
	CatalogSHA256 string
	SelectedCount int
	Published     bool
}

// Generate verifies the release boundary, selects exact canonical names, and
// atomically publishes one owned output directory.
func Generate(ctx context.Context, opts Options) (Result, error) {
	return generate(ctx, opts)
}
