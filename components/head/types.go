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

type dependencySource struct {
	cdnURL    string
	localURL  string
	integrity string
	enabled   bool
}

type config struct {
	initialized    bool
	stylesheetURL  string
	comboboxURL    string
	loaderURL      string
	nonce          string
	localFallback  bool
	localRuntime   bool
	alpineJS       dependencySource
	alpineCollapse dependencySource
	alpineFocus    dependencySource
	alpineMask     dependencySource
	htmx           dependencySource
}

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
			source.cdnURL = url
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
			source.localURL = url
		}
	})
}

// WithDependencyIntegrity replaces the Subresource Integrity value for one
// dependency. Pass an empty string to disable integrity for a custom source.
func WithDependencyIntegrity(dependency Dependency, integrity string) Option {
	return optionFunc(func(cfg *config) {
		if source := cfg.source(dependency); source != nil {
			source.integrity = integrity
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
			source.enabled = false
		}
	})
}

// WithStylesheetURL replaces the Goshtoso compiled stylesheet URL.
func WithStylesheetURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.stylesheetURL = url
		}
	})
}

// WithComboboxURL replaces the first-party combobox keyboard helper URL.
func WithComboboxURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.comboboxURL = url
		}
	})
}

// WithLoaderURL replaces the URL of Goshtoso's first-party dependency loader.
// Use it when the embedded loader is mirrored or copied to another asset path.
func WithLoaderURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url != "" {
			cfg.loaderURL = url
		}
	})
}

func newConfig(options []Option) config {
	cfg := config{
		initialized:   true,
		stylesheetURL: "/assets/styles.css",
		comboboxURL:   "/assets/js/combobox.js",
		loaderURL:     "/assets/js/dependency-loader.js",
		localFallback: true,
		alpineJS: dependencySource{
			cdnURL:    assets.AlpineJSCDNURL,
			localURL:  assets.AlpineJSURL,
			integrity: assets.AlpineJSIntegrity,
			enabled:   true,
		},
		alpineCollapse: dependencySource{
			cdnURL:    assets.AlpineCollapseCDNURL,
			localURL:  assets.AlpineCollapseURL,
			integrity: assets.AlpineCollapseIntegrity,
			enabled:   true,
		},
		alpineFocus: dependencySource{
			cdnURL:    assets.AlpineFocusCDNURL,
			localURL:  assets.AlpineFocusURL,
			integrity: assets.AlpineFocusIntegrity,
			enabled:   true,
		},
		alpineMask: dependencySource{
			cdnURL:    assets.AlpineMaskCDNURL,
			localURL:  assets.AlpineMaskURL,
			integrity: assets.AlpineMaskIntegrity,
			enabled:   true,
		},
		htmx: dependencySource{
			cdnURL:    assets.HTMXCDNURL,
			localURL:  assets.HTMXURL,
			integrity: assets.HTMXIntegrity,
			enabled:   true,
		},
	}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	if cfg.localRuntime {
		for _, source := range cfg.sources() {
			source.cdnURL = source.localURL
		}
		cfg.localFallback = false
	}
	return cfg
}

func (cfg *config) source(dependency Dependency) *dependencySource {
	switch dependency {
	case DependencyAlpineJS:
		return &cfg.alpineJS
	case DependencyAlpineCollapse:
		return &cfg.alpineCollapse
	case DependencyAlpineFocus:
		return &cfg.alpineFocus
	case DependencyAlpineMask:
		return &cfg.alpineMask
	case DependencyHTMX:
		return &cfg.htmx
	default:
		return nil
	}
}

func (cfg *config) sources() []*dependencySource {
	return []*dependencySource{
		&cfg.alpineCollapse,
		&cfg.alpineFocus,
		&cfg.alpineMask,
		&cfg.alpineJS,
		&cfg.htmx,
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
	appendSource := func(name string, source dependencySource, waitForWindowLoaded bool) {
		if !source.enabled {
			return
		}
		entry := loaderDependency{
			Name:                name,
			PrimaryURL:          source.cdnURL,
			Integrity:           source.integrity,
			WaitForWindowLoaded: waitForWindowLoaded,
		}
		if cfg.localFallback && source.cdnURL != source.localURL {
			entry.FallbackURL = source.localURL
		}
		dependencies = append(dependencies, entry)
	}

	if !minimal {
		appendSource("alpine-collapse", cfg.alpineCollapse, false)
		appendSource("alpine-focus", cfg.alpineFocus, false)
		appendSource("alpine-mask", cfg.alpineMask, false)
	}
	appendSource("alpine", cfg.alpineJS, false)
	appendSource("htmx", cfg.htmx, true)
	dependencies = append(dependencies, loaderDependency{
		Name:       "combobox",
		PrimaryURL: cfg.comboboxURL,
	})

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

func (cfg config) scriptAttributes(source dependencySource) templ.Attributes {
	attrs := templ.Attributes{}
	if source.integrity != "" {
		attrs["crossorigin"] = "anonymous"
		attrs["integrity"] = source.integrity
	}
	if cfg.nonce != "" {
		attrs["nonce"] = cfg.nonce
	}
	return attrs
}

func (cfg config) nonceAttributes() templ.Attributes {
	if cfg.nonce == "" {
		return nil
	}
	return templ.Attributes{"nonce": cfg.nonce}
}
