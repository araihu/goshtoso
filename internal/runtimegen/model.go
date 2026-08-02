// Package runtimegen loads and generates Goshtoso runtime asset contracts.
package runtimegen

import "github.com/araihu/assets/assetmeta"

type RootMetadata struct {
	Loader     LocalAsset            `yaml:"loader"`
	Order      []OrderedEntry        `yaml:"order"`
	FirstParty map[string]LocalAsset `yaml:"first_party"`
}

type OrderedEntry struct {
	Download assetmeta.Ref `yaml:"download,omitempty"`
	Local    string        `yaml:"local,omitempty"`
}

type DownloadMeta struct {
	Role                string        `yaml:"role"`
	GoName              string        `yaml:"go_name"`
	RoleGoName          string        `yaml:"role_go_name,omitempty"`
	Name                string        `yaml:"name"`
	PackageName         string        `yaml:"package_name,omitempty"`
	Homepage            string        `yaml:"homepage,omitempty"`
	License             string        `yaml:"license,omitempty"`
	LicenseRef          assetmeta.Ref `yaml:"license_ref,omitempty"`
	ProvenanceRef       assetmeta.Ref `yaml:"provenance_ref,omitempty"`
	Purpose             string        `yaml:"purpose"`
	Attribution         bool          `yaml:"attribution,omitempty"`
	Enabled             bool          `yaml:"enabled"`
	IncludeInMinimal    bool          `yaml:"include_in_minimal"`
	Defer               bool          `yaml:"defer"`
	WaitForWindowLoaded bool          `yaml:"wait_for_window_loaded,omitempty"`
}

type ResourceMeta struct{}

type LocalAsset struct {
	Role             string `yaml:"role"`
	GoName           string `yaml:"go_name"`
	RoleGoName       string `yaml:"role_go_name,omitempty"`
	Name             string `yaml:"name"`
	LocalURL         string `yaml:"local_url"`
	Purpose          string `yaml:"purpose"`
	Enabled          bool   `yaml:"enabled"`
	IncludeInMinimal bool   `yaml:"include_in_minimal"`
	Defer            bool   `yaml:"defer"`
}

type Model struct {
	Loader       Asset
	Dependencies []Asset
}

type Asset struct {
	Resource            string
	Download            string
	Role                string
	GoName              string
	RoleGoName          string
	Name                string
	Version             string
	URL                 string
	LocalURL            string
	Integrity           string
	Hash                string
	PackageName         string
	ProvenanceURL       string
	ProvenanceLocalURL  string
	Homepage            string
	License             string
	LicenseURL          string
	LicenseLocalURL     string
	LicenseIntegrity    string
	LicenseHash         string
	Purpose             string
	Attribution         bool
	Enabled             bool
	IncludeInMinimal    bool
	Defer               bool
	WaitForWindowLoaded bool
}
