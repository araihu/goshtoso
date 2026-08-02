package assets

const (
	// StylesURL is the assets.Handler URL for Goshtoso's compiled component CSS.
	StylesURL = "/assets/styles.css"
)

// RuntimeAssetKind describes how a runtime manifest asset is included in HTML.
type RuntimeAssetKind string

const (
	// RuntimeAssetStylesheet identifies a stylesheet link.
	RuntimeAssetStylesheet RuntimeAssetKind = "stylesheet"
	// RuntimeAssetScript identifies an executable script.
	RuntimeAssetScript RuntimeAssetKind = "script"
)

// RuntimeAssetRole identifies one asset's purpose without requiring callers to
// infer execution semantics from its URL.
type RuntimeAssetRole string

const (
	// RuntimeRoleStylesheet identifies Goshtoso's compiled CSS.
	RuntimeRoleStylesheet RuntimeAssetRole = "stylesheet"
)

// RuntimeAsset describes one stylesheet or script in Goshtoso's default head
// dependency contract.
// PrimaryURL is used by the CDN-first loader; LocalURL is
// the same-version embedded URL served by Handler. Integrity is SHA-384 SRI for
// dependencies whose primary and local bytes are required to match; one value
// applies to both URLs. A top-level Loader LocalURL is inventory, not an
// automatic fallback for the loader tag.
// IncludeInMinimal selects DependenciesMinimal membership. On dependency
// entries, Defer controls direct local-script tags and WaitForWindowLoaded
// controls loader readiness before dynamically inserting the script. For
// top-level assets, only Loader.Defer is supported; Stylesheet.Defer and either
// top-level WaitForWindowLoaded value are rejected.
type RuntimeAsset struct {
	Role                RuntimeAssetRole
	Kind                RuntimeAssetKind
	PrimaryURL          string
	LocalURL            string
	Integrity           string
	Enabled             bool
	IncludeInMinimal    bool
	Defer               bool
	WaitForWindowLoaded bool
}

// RuntimeAssetMetadata describes the identity, provenance, licensing, and
// purpose of a runtime role without changing RuntimeAsset's loading contract.
// First-party assets leave Version and License empty; use GoshtosoVersion for
// the linked module identity.
type RuntimeAssetMetadata struct {
	Role          RuntimeAssetRole
	Name          string
	Version       string
	PackageName   string
	ProvenanceURL string
	Homepage      string
	License       string
	LicenseURL    string
	Purpose       string
}

// DefaultRuntimeMetadata returns generated metadata in loader execution order.
// The returned slice is caller-owned.
func DefaultRuntimeMetadata() []RuntimeAssetMetadata {
	metadata := defaultRuntimeMetadata()
	return append([]RuntimeAssetMetadata(nil), metadata...)
}

// RuntimeManifest is a typed embedded runtime/fallback contract. Dependencies
// are in declared loader execution order. Disabled dependencies remain public
// inventory and can be enabled in a caller-owned copy. Loader is separate
// because CDN-first rendering executes it to load enabled Dependencies, while
// direct local rendering must not execute the loader.
type RuntimeManifest struct {
	Stylesheet   RuntimeAsset
	Loader       RuntimeAsset
	Dependencies []RuntimeAsset
}

// DefaultRuntimeManifest returns the complete pinned runtime/fallback contract.
// The returned value and dependency slice are caller-owned; mutation cannot
// affect later calls, Handler, or head.Dependencies rendering. Mount Handler
// directly at /assets/ to serve every LocalURL. Pinned versions are the tested
// combination; changing this manifest configures loading but does not guarantee
// compatibility with arbitrary dependency versions.
func DefaultRuntimeManifest() RuntimeManifest {
	return RuntimeManifest{
		Stylesheet: RuntimeAsset{
			Role:             RuntimeRoleStylesheet,
			Kind:             RuntimeAssetStylesheet,
			PrimaryURL:       StylesURL,
			LocalURL:         StylesURL,
			Enabled:          true,
			IncludeInMinimal: true,
		},
		Loader:       defaultRuntimeLoader(),
		Dependencies: defaultRuntimeDependencies(),
	}
}
