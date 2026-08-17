// Package iconpack generates consumer-local, attributed SVG icon packs from a
// verified Arai Hu Assets release boundary or a consumer-supplied source
// manifest and icon-pack root.
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
	// ConfigPath selects the Muamba-backed .iconpack.yaml contract. It is
	// mutually exclusive with the legacy release/source flags below.
	ConfigPath          string
	IconpackLockPath    string
	Trust               bool
	AllowHTTP           bool
	ReleaseRoot         string
	ReleaseArchive      string
	Release             string
	ArchiveSHA256       string
	CatalogSHA256       string
	ReleaseJSONSHA256   string
	ChecksumsSHA256     string
	SourceRoot          string
	SourceArchive       string
	SourceArchiveSHA256 string
	SourceManifest      string
	Names               []string
	SelectionManifest   string
	OutputDir           string
	Package             string
	ConstPrefix         string
	SpriteURL           string
	IconifyPrefix       string
	Check               bool
}

// Result identifies a verified or newly published output.
type Result struct {
	OutputDir     string
	Release       string
	CatalogSHA256 string
	SelectedCount int
	Published     bool
}

// Generate verifies the selected source boundary, selects exact canonical
// names, and atomically publishes one owned output directory. A source
// manifest makes non-Arai-Hu icon packs first-class inputs while retaining the
// Arai Hu Assets release mode for catalog-backed packs.
func Generate(ctx context.Context, opts Options) (Result, error) {
	return generate(ctx, opts)
}
