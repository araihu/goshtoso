package head

import (
	"encoding/json"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
)

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
	initialized   bool
	manifest      assets.RuntimeManifest
	nonce         string
	localFallback bool
	localRuntime  bool
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

// WithDependencyCDNURL replaces the pinned CDN URL for one dependency. Empty
// URLs leave the version-matched default unchanged. When changing versions,
// also use WithDependencyLocalURL so the fallback remains version-matched.
func WithDependencyCDNURL(dependency Dependency, url string) Option {
	return optionFunc(func(cfg *config) {
		if url == "" {
			return
		}
		if source := cfg.source(dependency); source != nil {
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
		if source := cfg.source(dependency); source != nil {
			source.LocalURL = url
		}
	})
}

// WithDependencyIntegrity replaces the Subresource Integrity value for one
// dependency. Pass an empty string to disable integrity for a custom source.
func WithDependencyIntegrity(dependency Dependency, integrity string) Option {
	return optionFunc(func(cfg *config) {
		if source := cfg.source(dependency); source != nil {
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

// WithLocalRuntime makes the embedded version-matched assets primary and emits
// no CDN requests or fallback loader.
func WithLocalRuntime() Option {
	return optionFunc(func(cfg *config) {
		cfg.localRuntime = true
	})
}

// WithoutDependency omits a runtime that the application owns separately.
func WithoutDependency(dependency Dependency) Option {
	return optionFunc(func(cfg *config) {
		if source := cfg.source(dependency); source != nil {
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

// WithComboboxURL replaces the first-party combobox keyboard helper URL.
func WithComboboxURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.useStandaloneFirstParty()
			if source := cfg.asset(assets.RuntimeRoleCombobox); source != nil {
				source.PrimaryURL = url
				source.LocalURL = url
			}
		}
	})
}

// WithActionGroupURL replaces the first-party ActionGroup measurement helper URL.
func WithActionGroupURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.useStandaloneFirstParty()
			if source := cfg.asset(assets.RuntimeRoleActionGroup); source != nil {
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
	return newConfigFromManifest(assets.DefaultRuntimeManifest(), options)
}

func newConfigFromManifest(manifest assets.RuntimeManifest, options []Option) config {
	manifest.Dependencies = append([]assets.RuntimeAsset(nil), manifest.Dependencies...)
	cfg := config{
		initialized:   true,
		manifest:      manifest,
		localFallback: true,
	}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	if cfg.localRuntime {
		cfg.manifest.Stylesheet.PrimaryURL = cfg.manifest.Stylesheet.LocalURL
		cfg.manifest.Loader.PrimaryURL = cfg.manifest.Loader.LocalURL
		for index := range cfg.manifest.Dependencies {
			source := &cfg.manifest.Dependencies[index]
			source.PrimaryURL = source.LocalURL
		}
		cfg.localFallback = false
	}
	return cfg
}

func (cfg *config) source(dependency Dependency) *assets.RuntimeAsset {
	var role assets.RuntimeAssetRole
	switch dependency {
	case DependencyAlpineJS:
		role = assets.RuntimeRoleAlpineJS
	case DependencyAlpineCollapse:
		role = assets.RuntimeRoleAlpineCollapse
	case DependencyAlpineFocus:
		role = assets.RuntimeRoleAlpineFocus
	case DependencyAlpineMask:
		role = assets.RuntimeRoleAlpineMask
	case DependencyHTMX:
		role = assets.RuntimeRoleHTMX
	default:
		return nil
	}
	return cfg.asset(role)
}

func (cfg *config) asset(role assets.RuntimeAssetRole) *assets.RuntimeAsset {
	for index := range cfg.manifest.Dependencies {
		if cfg.manifest.Dependencies[index].Role == role {
			return &cfg.manifest.Dependencies[index]
		}
	}
	return nil
}

func (cfg *config) useStandaloneFirstParty() {
	if bundle := cfg.asset(assets.RuntimeRoleFirstParty); bundle != nil {
		bundle.Enabled = false
	}
	if combobox := cfg.asset(assets.RuntimeRoleCombobox); combobox != nil {
		combobox.Enabled = true
	}
	if actionGroup := cfg.asset(assets.RuntimeRoleActionGroup); actionGroup != nil {
		actionGroup.Enabled = true
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

func (cfg config) loaderAttributes(minimal bool) templ.Attributes {
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

	payload, err := json.Marshal(loaderConfig{Dependencies: dependencies})
	if err != nil {
		panic("head: marshal dependency loader configuration: " + err.Error())
	}
	attrs := templ.Attributes{"data-goshtoso-dependencies": string(payload)}
	if cfg.nonce != "" {
		attrs["nonce"] = cfg.nonce
	}
	return attrs
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
	attrs := templ.Attributes{}
	if source.Integrity != "" {
		attrs["crossorigin"] = "anonymous"
		attrs["integrity"] = source.Integrity
	}
	if cfg.nonce != "" {
		attrs["nonce"] = cfg.nonce
	}
	return attrs
}
