package assets

const (
	// StylesURL is the assets.Handler URL for Goshtoso's compiled component CSS.
	StylesURL = "/assets/styles.css"
	// DependencyLoaderURL is the assets.Handler URL for the ordered CDN/fallback loader.
	DependencyLoaderURL = "/assets/js/dependency-loader.js"
	// ComboboxURL is the assets.Handler URL for Goshtoso's combobox keyboard helper.
	ComboboxURL = "/assets/js/combobox.js"
	// ActionGroupURL is the assets.Handler URL for responsive ActionGroup measurement.
	ActionGroupURL = "/assets/js/action-group.js"
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
	// RuntimeRoleDependencyLoader identifies the ordered CDN/fallback bootstrap.
	RuntimeRoleDependencyLoader RuntimeAssetRole = "dependency-loader"
	// RuntimeRoleAlpineCollapse identifies the Alpine Collapse plugin.
	RuntimeRoleAlpineCollapse RuntimeAssetRole = "alpine-collapse"
	// RuntimeRoleAlpineFocus identifies the Alpine Focus plugin.
	RuntimeRoleAlpineFocus RuntimeAssetRole = "alpine-focus"
	// RuntimeRoleAlpineMask identifies the Alpine Mask plugin.
	RuntimeRoleAlpineMask RuntimeAssetRole = "alpine-mask"
	// RuntimeRoleAlpineJS identifies Alpine.js core.
	RuntimeRoleAlpineJS RuntimeAssetRole = "alpine"
	// RuntimeRoleHTMX identifies HTMX core.
	RuntimeRoleHTMX RuntimeAssetRole = "htmx"
	// RuntimeRoleCombobox identifies Goshtoso's combobox keyboard helper.
	RuntimeRoleCombobox RuntimeAssetRole = "combobox"
	// RuntimeRoleActionGroup identifies responsive ActionGroup measurement.
	RuntimeRoleActionGroup RuntimeAssetRole = "action-group"
)

// RuntimeAsset describes one stylesheet or script in Goshtoso's default head
// dependency contract. PrimaryURL is used by the CDN-first loader; LocalURL is
// the same-version embedded URL served by Handler. Integrity is SHA-384 SRI for
// dependencies whose primary and local bytes are required to match.
// IncludeInMinimal selects DependenciesMinimal membership. Defer is the direct
// local-script tag behavior. WaitForWindowLoaded is loader readiness behavior
// used before dynamically inserting the script.
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

// RuntimeManifest is Goshtoso's complete default embedded runtime/fallback
// contract. Dependencies are in execution order. Loader is separate because
// CDN-first rendering executes it to load Dependencies, while direct local
// rendering executes Dependencies and must not execute the loader as well.
type RuntimeManifest struct {
	Stylesheet   RuntimeAsset
	Loader       RuntimeAsset
	Dependencies []RuntimeAsset
}

// DefaultRuntimeManifest returns the complete default runtime/fallback
// contract. The returned value and dependency slice are caller-owned; mutation
// cannot affect later calls, assets.Handler, or head.Dependencies rendering.
// Mount Handler directly at /assets/ to serve every LocalURL.
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
		Loader: RuntimeAsset{
			Role:             RuntimeRoleDependencyLoader,
			Kind:             RuntimeAssetScript,
			PrimaryURL:       DependencyLoaderURL,
			LocalURL:         DependencyLoaderURL,
			Enabled:          true,
			IncludeInMinimal: true,
			Defer:            true,
		},
		Dependencies: []RuntimeAsset{
			{
				Role: RuntimeRoleAlpineCollapse, Kind: RuntimeAssetScript,
				PrimaryURL: AlpineCollapseCDNURL, LocalURL: AlpineCollapseURL,
				Integrity: AlpineCollapseIntegrity, Enabled: true, Defer: true,
			},
			{
				Role: RuntimeRoleAlpineFocus, Kind: RuntimeAssetScript,
				PrimaryURL: AlpineFocusCDNURL, LocalURL: AlpineFocusURL,
				Integrity: AlpineFocusIntegrity, Enabled: true, Defer: true,
			},
			{
				Role: RuntimeRoleAlpineMask, Kind: RuntimeAssetScript,
				PrimaryURL: AlpineMaskCDNURL, LocalURL: AlpineMaskURL,
				Integrity: AlpineMaskIntegrity, Enabled: true, Defer: true,
			},
			{
				Role: RuntimeRoleAlpineJS, Kind: RuntimeAssetScript,
				PrimaryURL: AlpineJSCDNURL, LocalURL: AlpineJSURL,
				Integrity: AlpineJSIntegrity, Enabled: true, IncludeInMinimal: true, Defer: true,
			},
			{
				Role: RuntimeRoleHTMX, Kind: RuntimeAssetScript,
				PrimaryURL: HTMXCDNURL, LocalURL: HTMXURL,
				Integrity: HTMXIntegrity, Enabled: true, IncludeInMinimal: true, WaitForWindowLoaded: true,
			},
			{
				Role: RuntimeRoleCombobox, Kind: RuntimeAssetScript,
				PrimaryURL: ComboboxURL, LocalURL: ComboboxURL,
				Enabled: true, IncludeInMinimal: true, Defer: true,
			},
			{
				Role: RuntimeRoleActionGroup, Kind: RuntimeAssetScript,
				PrimaryURL: ActionGroupURL, LocalURL: ActionGroupURL,
				Enabled: true, IncludeInMinimal: true, Defer: true,
			},
		},
	}
}
