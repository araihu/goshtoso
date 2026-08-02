package assets

const (
	// StylesURL is the assets.Handler URL for Goshtoso's compiled component CSS.
	StylesURL = "/assets/styles.css"
	// DependencyLoaderURL is the assets.Handler URL for the ordered CDN/fallback loader.
	DependencyLoaderURL = "/assets/js/dependency-loader.js"
	// ComboboxURL is the assets.Handler URL for Goshtoso's standalone Combobox runtime.
	ComboboxURL = "/assets/js/combobox.js"
	// ActionGroupURL is the assets.Handler URL for responsive ActionGroup measurement.
	ActionGroupURL = "/assets/js/action-group.js"
	// FirstPartyBundleURL is the minified reusable Goshtoso component-runtime bundle.
	FirstPartyBundleURL = "/assets/js/goshtoso.min.js"
	// DarkModeURL is the Alpine dark-mode store runtime served by Handler.
	DarkModeURL = "/assets/js/darkmode.js"
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
	// RuntimeRoleCombobox identifies Goshtoso's standalone Combobox runtime.
	RuntimeRoleCombobox RuntimeAssetRole = "combobox"
	// RuntimeRoleActionGroup identifies responsive ActionGroup measurement.
	RuntimeRoleActionGroup RuntimeAssetRole = "action-group"
	// RuntimeRoleFirstParty identifies the reusable Goshtoso component-runtime bundle.
	RuntimeRoleFirstParty RuntimeAssetRole = "first-party"
	// RuntimeRoleDarkMode identifies Goshtoso's Alpine dark-mode store.
	RuntimeRoleDarkMode RuntimeAssetRole = "dark-mode"
	// RuntimeRoleHTMXExtSSE identifies the HTMX Server-Sent Events extension.
	RuntimeRoleHTMXExtSSE RuntimeAssetRole = "htmx-ext-sse"
	// RuntimeRoleHTMXExtWS identifies the HTMX WebSocket extension.
	RuntimeRoleHTMXExtWS RuntimeAssetRole = "htmx-ext-ws"
)

// RuntimeAsset describes one stylesheet or script in Goshtoso's default head
// dependency contract. PrimaryURL is used by the CDN-first loader; LocalURL is
// the same-version embedded URL served by Handler. Integrity is SHA-384 SRI for
// dependencies whose primary and local bytes are required to match; one value
// applies to both URLs. A top-level Loader LocalURL is inventory, not an
// automatic fallback for the loader tag.
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
			// First-party globals must exist before Alpine scans x-data/x-init nodes.
			// Alpine.data providers register on alpine:init; their initializers do not
			// require HTMX, which keeps the established Alpine-before-HTMX order.
			{
				Role: RuntimeRoleFirstParty, Kind: RuntimeAssetScript,
				PrimaryURL: FirstPartyBundleURL, LocalURL: FirstPartyBundleURL,
				Enabled: true, IncludeInMinimal: true, Defer: true,
			},
			{
				Role: RuntimeRoleDarkMode, Kind: RuntimeAssetScript,
				PrimaryURL: DarkModeURL, LocalURL: DarkModeURL,
				IncludeInMinimal: true, Defer: true,
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
				Role: RuntimeRoleHTMXExtSSE, Kind: RuntimeAssetScript,
				PrimaryURL: HTMXExtSSECDNURL, LocalURL: HTMXExtSSEURL,
				Integrity: HTMXExtSSEIntegrity, IncludeInMinimal: true, Defer: true,
			},
			{
				Role: RuntimeRoleHTMXExtWS, Kind: RuntimeAssetScript,
				PrimaryURL: HTMXExtWSCDNURL, LocalURL: HTMXExtWSURL,
				Integrity: HTMXExtWSIntegrity, IncludeInMinimal: true, Defer: true,
			},
			{
				Role: RuntimeRoleCombobox, Kind: RuntimeAssetScript,
				PrimaryURL: ComboboxURL, LocalURL: ComboboxURL, IncludeInMinimal: true, Defer: true,
			},
			{
				Role: RuntimeRoleActionGroup, Kind: RuntimeAssetScript,
				PrimaryURL: ActionGroupURL, LocalURL: ActionGroupURL, IncludeInMinimal: true, Defer: true,
			},
		},
	}
}
