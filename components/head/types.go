package head

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
)

// OpenGraphType identifies the Open Graph object type for a page.
type OpenGraphType string

const (
	// OpenGraphTypeWebsite is the default type for site and documentation pages.
	OpenGraphTypeWebsite OpenGraphType = "website"
)

// TwitterCard identifies the X/Twitter Card presentation for a page.
type TwitterCard string

const (
	// TwitterCardSummary renders a compact summary card.
	TwitterCardSummary TwitterCard = "summary"
	// TwitterCardSummaryLargeImage renders a wide image-led summary card.
	TwitterCardSummaryLargeImage TwitterCard = "summary_large_image"
)

// SocialImage defines the required image portion of a complete social preview.
type SocialImage struct {
	// URL is the absolute HTTPS URL crawlers use to fetch the image.
	URL string
	// MIMEType is a parameter-free RFC 6838 image type, such as image/jpeg.
	MIMEType string
	// Width is the image width in pixels and must be positive.
	Width int
	// Height is the image height in pixels and must be positive.
	Height int
	// Alt describes the image for accessible social clients and must be non-empty.
	Alt string
}

// MetadataConfig defines route-specific document, Open Graph, and X/Twitter
// Card metadata. Render fails without a complete social contract.
type MetadataConfig struct {
	// Title is the non-empty route-specific document and social title.
	Title string
	// Description is the non-empty route-specific document and social summary.
	Description string
	// CanonicalURL is the route's absolute HTTPS canonical and Open Graph URL.
	CanonicalURL string
	// OpenGraphType is the Open Graph object type and defaults to website.
	OpenGraphType OpenGraphType
	// SiteName is the optional Open Graph site name.
	SiteName string
	// Locale is the optional Open Graph locale, such as en_US.
	Locale string
	// Image is the required social-preview image and structured metadata.
	Image SocialImage
	// TwitterCard selects the X/Twitter presentation and defaults to summary_large_image.
	TwitterCard TwitterCard
	// TwitterSite is the optional real project account handle, including @.
	TwitterSite string
}

func (cfg MetadataConfig) validated() (MetadataConfig, error) {
	cfg.Title = strings.TrimSpace(cfg.Title)
	if cfg.Title == "" {
		return MetadataConfig{}, errors.New("head.Metadata: Title is required")
	}
	cfg.Description = strings.TrimSpace(cfg.Description)
	if cfg.Description == "" {
		return MetadataConfig{}, errors.New("head.Metadata: Description is required")
	}

	var err error
	cfg.CanonicalURL, err = absoluteHTTPSURL("CanonicalURL", cfg.CanonicalURL)
	if err != nil {
		return MetadataConfig{}, err
	}
	cfg.Image.URL, err = absoluteHTTPSURL("Image.URL", cfg.Image.URL)
	if err != nil {
		return MetadataConfig{}, err
	}

	cfg.Image.MIMEType, err = imageMIMEType(cfg.Image.MIMEType)
	if err != nil {
		return MetadataConfig{}, err
	}
	if cfg.Image.Width <= 0 {
		return MetadataConfig{}, errors.New("head.Metadata: Image.Width must be positive")
	}
	if cfg.Image.Height <= 0 {
		return MetadataConfig{}, errors.New("head.Metadata: Image.Height must be positive")
	}
	cfg.Image.Alt = strings.TrimSpace(cfg.Image.Alt)
	if cfg.Image.Alt == "" {
		return MetadataConfig{}, errors.New("head.Metadata: Image.Alt is required")
	}

	cfg.OpenGraphType = OpenGraphType(strings.TrimSpace(string(cfg.OpenGraphType)))
	if cfg.OpenGraphType == "" {
		cfg.OpenGraphType = OpenGraphTypeWebsite
	}
	cfg.TwitterCard = TwitterCard(strings.TrimSpace(string(cfg.TwitterCard)))
	if cfg.TwitterCard == "" {
		cfg.TwitterCard = TwitterCardSummaryLargeImage
	}
	if cfg.TwitterCard != TwitterCardSummary && cfg.TwitterCard != TwitterCardSummaryLargeImage {
		return MetadataConfig{}, fmt.Errorf("head.Metadata: unsupported TwitterCard %q", cfg.TwitterCard)
	}

	cfg.SiteName = strings.TrimSpace(cfg.SiteName)
	cfg.Locale = strings.TrimSpace(cfg.Locale)
	cfg.TwitterSite = strings.TrimSpace(cfg.TwitterSite)
	return cfg, nil
}

func absoluteHTTPSURL(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("head.Metadata: %s must be an absolute HTTPS URL: %w", field, err)
	}
	if !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" ||
		parsed.Hostname() == "" || parsed.Opaque != "" || parsed.User != nil {
		return "", fmt.Errorf("head.Metadata: %s must be an absolute HTTPS URL", field)
	}
	parsed.Scheme = "https"
	return parsed.String(), nil
}

func imageMIMEType(value string) (string, error) {
	mediaType, parameters, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("head.Metadata: Image.MIMEType must be a valid image media type: %w", err)
	}
	parts := strings.SplitN(mediaType, "/", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "image") ||
		!validRestrictedMediaName(parts[1]) {
		return "", errors.New("head.Metadata: Image.MIMEType must be an image media type")
	}
	if len(parameters) != 0 {
		return "", errors.New("head.Metadata: Image.MIMEType parameters are not supported")
	}
	return strings.ToLower(mediaType), nil
}

func validRestrictedMediaName(value string) bool {
	if len(value) == 0 || len(value) > 127 || !isASCIIAlphaNumeric(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if isASCIIAlphaNumeric(character) {
			continue
		}
		switch character {
		case '!', '#', '$', '&', '-', '^', '_', '.', '+':
		default:
			return false
		}
	}
	return true
}

func isASCIIAlphaNumeric(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9'
}

func (image SocialImage) widthContent() string {
	return strconv.Itoa(image.Width)
}

func (image SocialImage) heightContent() string {
	return strconv.Itoa(image.Height)
}

// Dependency identifies a third-party runtime managed by Dependencies.
type Dependency string

const (
	// DependencyAlpineJS identifies Alpine.js core.
	DependencyAlpineJS Dependency = "alpinejs"
	// DependencyAlpineCollapse identifies the Alpine Collapse plugin.
	DependencyAlpineCollapse Dependency = "alpinejs-collapse"
	// DependencyAlpineFocus identifies the Alpine Focus plugin.
	DependencyAlpineFocus Dependency = "alpinejs-focus"
	// DependencyAlpineMask identifies the Alpine Mask plugin.
	DependencyAlpineMask Dependency = "alpinejs-mask"
	// DependencyHTMX identifies HTMX core.
	DependencyHTMX Dependency = "htmx"
)

type config struct {
	initialized    bool
	manifest       assets.RuntimeManifest
	nonce          string
	localFallback  bool
	localRuntime   bool
	customManifest bool
	err            error
	loaderPayload  string
}

type runtimeAsset = assets.RuntimeAsset

// Option configures the dependency set emitted by Dependencies or
// DependenciesMinimal.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

type runtimeManifestOption struct {
	manifest assets.RuntimeManifest
}

func (runtimeManifestOption) apply(*config) {}

// WithRuntimeManifest selects a caller-defined runtime manifest as the
// dependency baseline. The manifest is snapshotted when this function is
// called. Use it at most once; all other options apply afterward regardless of
// its argument position.
func WithRuntimeManifest(manifest assets.RuntimeManifest) Option {
	return runtimeManifestOption{manifest: cloneRuntimeManifest(manifest)}
}

// WithDependencyCDNURL replaces the pinned CDN URL for one dependency. Empty
// URLs leave the version-matched default unchanged. When changing versions,
// also use WithDependencyLocalURL so the fallback remains version-matched.
func WithDependencyCDNURL(dependency Dependency, url string) Option {
	return optionFunc(func(cfg *config) {
		if url == "" {
			return
		}
		if source := cfg.source(dependency, "WithDependencyCDNURL"); source != nil {
			source.PrimaryURL = url
		}
	})
}

// WithDependencyLocalURL replaces the local fallback URL for one dependency.
// Empty URLs leave the embedded version-matched default unchanged.
func WithDependencyLocalURL(dependency Dependency, url string) Option {
	return optionFunc(func(cfg *config) {
		if url == "" {
			return
		}
		if source := cfg.source(dependency, "WithDependencyLocalURL"); source != nil {
			source.LocalURL = url
		}
	})
}

// WithDependencyIntegrity replaces the Subresource Integrity value for one
// dependency. One value applies to both primary and fallback, so their bytes
// must match. Pass an empty string to disable integrity for a custom source.
func WithDependencyIntegrity(dependency Dependency, integrity string) Option {
	return optionFunc(func(cfg *config) {
		if source := cfg.source(dependency, "WithDependencyIntegrity"); source != nil {
			source.Integrity = integrity
		}
	})
}

// WithoutLocalFallback disables fallback and emits only the configured CDN
// sources. It has no effect when WithLocalRuntime is also set.
func WithoutLocalFallback() Option {
	return optionFunc(func(cfg *config) {
		cfg.localFallback = false
	})
}

// WithLocalRuntime makes the default embedded version-matched assets primary
// and emits no CDN requests or fallback loader. It cannot be combined with
// WithRuntimeManifest; custom local-only manifests must set PrimaryURL to the
// desired local URL and use WithoutLocalFallback.
func WithLocalRuntime() Option {
	return optionFunc(func(cfg *config) {
		cfg.localRuntime = true
	})
}

// WithoutDependency omits a runtime that the application owns separately.
func WithoutDependency(dependency Dependency) Option {
	return optionFunc(func(cfg *config) {
		if source := cfg.source(dependency, "WithoutDependency"); source != nil {
			source.Enabled = false
		}
	})
}

// WithStylesheetURL replaces the Goshtoso compiled stylesheet URL.
func WithStylesheetURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.manifest.Stylesheet.PrimaryURL = url
			cfg.manifest.Stylesheet.LocalURL = url
		}
	})
}

// WithComboboxURL selects legacy standalone compatibility mode, replacing the
// first-party bundle with both standalone helpers, then replaces the Combobox
// helper URL.
func WithComboboxURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.useStandaloneFirstParty("WithComboboxURL")
			if source := cfg.optionAsset(assets.RuntimeRoleCombobox, "WithComboboxURL"); source != nil {
				source.PrimaryURL = url
				source.LocalURL = url
			}
		}
	})
}

// WithActionGroupURL selects legacy standalone compatibility mode, replacing
// the first-party bundle with both standalone helpers, then replaces the
// ActionGroup measurement helper URL.
func WithActionGroupURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.useStandaloneFirstParty("WithActionGroupURL")
			if source := cfg.optionAsset(assets.RuntimeRoleActionGroup, "WithActionGroupURL"); source != nil {
				source.PrimaryURL = url
				source.LocalURL = url
			}
		}
	})
}

// WithLoaderURL replaces the URL of Goshtoso's first-party dependency loader.
// Use it when the embedded loader is mirrored or copied to another asset path.
func WithLoaderURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.manifest.Loader.PrimaryURL = url
			cfg.manifest.Loader.LocalURL = url
		}
	})
}

func newConfig(options []Option) config {
	manifest := assets.DefaultRuntimeManifest()
	manifestCount := 0
	for _, option := range options {
		if baseline, ok := option.(runtimeManifestOption); ok {
			manifest = cloneRuntimeManifest(baseline.manifest)
			manifestCount++
		}
	}
	cfg := baseConfig(manifest)
	cfg.customManifest = manifestCount > 0
	if manifestCount > 1 {
		cfg.err = errors.New("head: WithRuntimeManifest may be used once")
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if _, ok := option.(runtimeManifestOption); ok {
			continue
		}
		option.apply(&cfg)
	}
	finalizeConfig(&cfg)
	return cfg
}

func newConfigFromManifest(manifest assets.RuntimeManifest, options []Option) config {
	cfg := baseConfig(manifest)
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	finalizeConfig(&cfg)
	return cfg
}

func cloneRuntimeManifest(manifest assets.RuntimeManifest) assets.RuntimeManifest {
	manifest.Dependencies = append([]assets.RuntimeAsset(nil), manifest.Dependencies...)
	return manifest
}

func baseConfig(manifest assets.RuntimeManifest) config {
	return config{
		initialized:   true,
		manifest:      cloneRuntimeManifest(manifest),
		localFallback: true,
	}
}

func finalizeConfig(cfg *config) {
	if cfg.localRuntime {
		cfg.manifest.Stylesheet.PrimaryURL = cfg.manifest.Stylesheet.LocalURL
		cfg.manifest.Loader.PrimaryURL = cfg.manifest.Loader.LocalURL
		for index := range cfg.manifest.Dependencies {
			source := &cfg.manifest.Dependencies[index]
			source.PrimaryURL = source.LocalURL
		}
		cfg.localFallback = false
	}
}

func (cfg *config) source(dependency Dependency, option string) *assets.RuntimeAsset {
	role, ok := dependencyRole(dependency)
	if !ok {
		if cfg.customManifest {
			cfg.addError(fmt.Errorf("head: %s missing dependency role %q", option, dependency))
		}
		return nil
	}
	return cfg.optionAsset(role, option)
}

func dependencyRole(dependency Dependency) (assets.RuntimeAssetRole, bool) {
	switch dependency {
	case DependencyAlpineJS:
		return assets.RuntimeRoleAlpineJS, true
	case DependencyAlpineCollapse:
		return assets.RuntimeRoleAlpineCollapse, true
	case DependencyAlpineFocus:
		return assets.RuntimeRoleAlpineFocus, true
	case DependencyAlpineMask:
		return assets.RuntimeRoleAlpineMask, true
	case DependencyHTMX:
		return assets.RuntimeRoleHTMX, true
	default:
		return "", false
	}
}

func (cfg *config) asset(role assets.RuntimeAssetRole) *assets.RuntimeAsset {
	for index := range cfg.manifest.Dependencies {
		if cfg.manifest.Dependencies[index].Role == role {
			return &cfg.manifest.Dependencies[index]
		}
	}
	return nil
}

func (cfg *config) optionAsset(role assets.RuntimeAssetRole, option string) *assets.RuntimeAsset {
	source := cfg.asset(role)
	if source == nil && cfg.customManifest {
		cfg.addError(fmt.Errorf("head: %s missing dependency role %q", option, role))
	}
	return source
}

func (cfg *config) useStandaloneFirstParty(option string) {
	bundle := cfg.optionAsset(assets.RuntimeRoleFirstParty, option)
	combobox := cfg.optionAsset(assets.RuntimeRoleCombobox, option)
	actionGroup := cfg.optionAsset(assets.RuntimeRoleActionGroup, option)
	if bundle != nil {
		bundle.Enabled = false
	}
	if combobox != nil {
		combobox.Enabled = true
	}
	if actionGroup != nil {
		actionGroup.Enabled = true
	}
}

func (cfg *config) addError(err error) {
	if cfg.err == nil {
		cfg.err = err
	}
}

type loaderDependency struct {
	Name                string `json:"name"`
	PrimaryURL          string `json:"primary_url"`
	FallbackURL         string `json:"fallback_url,omitempty"`
	Integrity           string `json:"integrity,omitempty"`
	WaitForWindowLoaded bool   `json:"wait_for_window_loaded,omitempty"`
}

type loaderConfig struct {
	Dependencies []loaderDependency `json:"dependencies"`
}

func (cfg *config) prepare(minimal bool) error {
	if cfg.err != nil {
		return cfg.err
	}
	if err := cfg.validate(minimal); err != nil {
		return err
	}
	payload, err := json.Marshal(cfg.loaderConfig(minimal))
	if err != nil {
		return fmt.Errorf("head: marshal dependency loader configuration: %w", err)
	}
	cfg.loaderPayload = string(payload)
	return nil
}

func (cfg config) loaderConfig(minimal bool) loaderConfig {
	dependencies := make([]loaderDependency, 0, 6)
	for _, source := range cfg.manifest.Dependencies {
		if !source.Enabled || minimal && !source.IncludeInMinimal {
			continue
		}
		entry := loaderDependency{
			Name:                string(source.Role),
			PrimaryURL:          source.PrimaryURL,
			Integrity:           source.Integrity,
			WaitForWindowLoaded: source.WaitForWindowLoaded,
		}
		if cfg.localFallback && source.PrimaryURL != source.LocalURL {
			entry.FallbackURL = source.LocalURL
		}
		dependencies = append(dependencies, entry)
	}
	return loaderConfig{Dependencies: dependencies}
}

func (cfg config) loaderAttributes() templ.Attributes {
	attrs := cfg.runtimeAttributes(cfg.manifest.Loader, true)
	attrs["data-goshtoso-dependencies"] = cfg.loaderPayload
	return attrs
}

func (cfg config) stylesheetAttributes() templ.Attributes {
	return cfg.runtimeAttributes(cfg.manifest.Stylesheet, false)
}

func (cfg config) runtimeAttributes(source runtimeAsset, nonce bool) templ.Attributes {
	attrs := templ.Attributes{}
	if source.Integrity != "" {
		attrs["crossorigin"] = "anonymous"
		attrs["integrity"] = source.Integrity
	}
	if cfg.nonce != "" {
		if nonce {
			attrs["nonce"] = cfg.nonce
		}
	}
	return attrs
}

func (cfg config) includes(source runtimeAsset, minimal bool) bool {
	return source.Enabled && (!minimal || source.IncludeInMinimal)
}

func (cfg config) localDependencies(minimal bool) []runtimeAsset {
	dependencies := make([]runtimeAsset, 0, len(cfg.manifest.Dependencies))
	for _, source := range cfg.manifest.Dependencies {
		if source.Enabled && (!minimal || source.IncludeInMinimal) {
			dependencies = append(dependencies, source)
		}
	}
	return dependencies
}

func (cfg config) scriptAttributes(source runtimeAsset) templ.Attributes {
	return cfg.runtimeAttributes(source, true)
}

func (cfg config) validate(minimal bool) error {
	if cfg.customManifest && cfg.localRuntime {
		return errors.New("head: custom RuntimeManifest cannot be combined with WithLocalRuntime")
	}
	if err := validateTopLevelAsset(cfg.manifest.Stylesheet, assets.RuntimeRoleStylesheet, assets.RuntimeAssetStylesheet, "stylesheet"); err != nil {
		return err
	}
	if err := validateTopLevelAsset(cfg.manifest.Loader, assets.RuntimeRoleDependencyLoader, assets.RuntimeAssetScript, "loader"); err != nil {
		return err
	}
	selectedDependencies, err := cfg.validateDependencies(minimal)
	if err != nil {
		return err
	}
	if err := cfg.validateKnownOrder(); err != nil {
		return err
	}
	if cfg.includes(cfg.manifest.Stylesheet, minimal) {
		if err := validateRequiredURL("stylesheet primary URL", cfg.manifest.Stylesheet.PrimaryURL); err != nil {
			return err
		}
	}
	loaderEnabled := !cfg.localRuntime && cfg.includes(cfg.manifest.Loader, minimal)
	if selectedDependencies > 0 && !cfg.localRuntime && !loaderEnabled {
		return errors.New("head: enabled dependencies require an enabled loader for CDN rendering")
	}
	if loaderEnabled {
		if err := validateRequiredURL("loader primary URL", cfg.manifest.Loader.PrimaryURL); err != nil {
			return err
		}
	}
	return nil
}

func validateTopLevelAsset(source runtimeAsset, role assets.RuntimeAssetRole, kind assets.RuntimeAssetKind, label string) error {
	if source.Role != role {
		return fmt.Errorf("head: %s role = %q, want %q", label, source.Role, role)
	}
	if source.Kind != kind {
		return fmt.Errorf("head: %s kind = %q, want %q", label, source.Kind, kind)
	}
	if source.Defer && kind == assets.RuntimeAssetStylesheet {
		return fmt.Errorf("head: %s Defer is unsupported", label)
	}
	if source.WaitForWindowLoaded {
		return fmt.Errorf("head: %s WaitForWindowLoaded is unsupported", label)
	}
	if err := validateOptionalURL(label+" primary URL", source.PrimaryURL); err != nil {
		return err
	}
	return validateOptionalURL(label+" local URL", source.LocalURL)
}

func (cfg config) validateDependencies(minimal bool) (int, error) {
	seen := make(map[assets.RuntimeAssetRole]struct{}, len(cfg.manifest.Dependencies))
	selected := 0
	for _, source := range cfg.manifest.Dependencies {
		if source.Role == "" {
			return 0, errors.New("head: dependency role is empty")
		}
		if !safeRuntimeRole(source.Role) {
			return 0, fmt.Errorf("head: dependency role %q is unsafe", source.Role)
		}
		if _, exists := seen[source.Role]; exists {
			return 0, fmt.Errorf("head: duplicate dependency role %q", source.Role)
		}
		seen[source.Role] = struct{}{}
		if source.Kind != assets.RuntimeAssetScript {
			return 0, fmt.Errorf("head: dependency %s kind = %q, want %q", source.Role, source.Kind, assets.RuntimeAssetScript)
		}
		if err := validateOptionalURL("dependency "+string(source.Role)+" primary URL", source.PrimaryURL); err != nil {
			return 0, err
		}
		if err := validateOptionalURL("dependency "+string(source.Role)+" local URL", source.LocalURL); err != nil {
			return 0, err
		}
		if !cfg.includes(source, minimal) {
			continue
		}
		selected++
		if cfg.localRuntime {
			if err := validateRequiredURL("dependency "+string(source.Role)+" local URL", source.LocalURL); err != nil {
				return 0, err
			}
			continue
		}
		if err := validateRequiredURL("dependency "+string(source.Role)+" primary URL", source.PrimaryURL); err != nil {
			return 0, err
		}
		if cfg.localFallback {
			if err := validateRequiredURL("dependency "+string(source.Role)+" local URL", source.LocalURL); err != nil {
				return 0, err
			}
		}
	}
	return selected, nil
}

func (cfg config) validateKnownOrder() error {
	positions := make(map[assets.RuntimeAssetRole]int, len(cfg.manifest.Dependencies))
	for index, source := range cfg.manifest.Dependencies {
		if source.Enabled {
			positions[source.Role] = index
		}
	}
	for _, pair := range [][2]assets.RuntimeAssetRole{
		{assets.RuntimeRoleAlpineCollapse, assets.RuntimeRoleAlpineJS},
		{assets.RuntimeRoleAlpineFocus, assets.RuntimeRoleAlpineJS},
		{assets.RuntimeRoleAlpineMask, assets.RuntimeRoleAlpineJS},
		{assets.RuntimeRoleFirstParty, assets.RuntimeRoleAlpineJS},
		{assets.RuntimeRoleDarkMode, assets.RuntimeRoleAlpineJS},
		{assets.RuntimeRoleHTMX, assets.RuntimeRoleHTMXExtSSE},
		{assets.RuntimeRoleHTMX, assets.RuntimeRoleHTMXExtWS},
	} {
		before, beforeEnabled := positions[pair[0]]
		after, afterEnabled := positions[pair[1]]
		if beforeEnabled && afterEnabled && before > after {
			return fmt.Errorf("head: %s must precede %s", pair[0], pair[1])
		}
	}
	_, bundleEnabled := positions[assets.RuntimeRoleFirstParty]
	_, comboboxEnabled := positions[assets.RuntimeRoleCombobox]
	_, actionGroupEnabled := positions[assets.RuntimeRoleActionGroup]
	if bundleEnabled && (comboboxEnabled || actionGroupEnabled) {
		return errors.New("head: first-party bundle cannot be combined with standalone combobox or action-group runtimes")
	}
	return nil
}

func safeRuntimeRole(role assets.RuntimeAssetRole) bool {
	for index, character := range string(role) {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || index > 0 && (character == '-' || character == '_' || character == '.') {
			continue
		}
		return false
	}
	return true
}

func validateOptionalURL(label, rawURL string) error {
	if rawURL == "" {
		return nil
	}
	return validateRequiredURL(label, rawURL)
}

func validateRequiredURL(label, rawURL string) error {
	if rawURL == "" || rawURL != strings.TrimSpace(rawURL) || strings.ContainsAny(rawURL, "\\\r\n\t") {
		return fmt.Errorf("head: %s %q is not a safe HTTP(S) or relative URL", label, rawURL)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("head: %s %q is invalid: %w", label, rawURL, err)
	}
	if parsed.Scheme == "" {
		if parsed.Host != "" || strings.HasPrefix(rawURL, "//") || parsed.Path == "" {
			return fmt.Errorf("head: %s %q is not a safe relative URL", label, rawURL)
		}
		return nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("head: %s %q is not a safe HTTP(S) URL", label, rawURL)
	}
	return nil
}
