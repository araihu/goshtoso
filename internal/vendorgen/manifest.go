package vendorgen

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"net/url"
	"path"
	"strings"
)

const supportedManifestSchema = 1

// runtimeManifest is the canonical ordered inventory for every embedded
// JavaScript runtime asset. The JSON array order is the loader execution order.
type runtimeManifest struct {
	Schema       int                    `json:"schema"`
	Loader       runtimeAssetManifest   `json:"loader"`
	Dependencies []runtimeAssetManifest `json:"dependencies"`
}

type runtimeAssetManifest struct {
	Module              string `json:"module,omitempty"`
	Role                string `json:"role"`
	GoName              string `json:"go_name"`
	RoleGoName          string `json:"role_go_name,omitempty"`
	Name                string `json:"name"`
	Version             string `json:"version,omitempty"`
	File                string `json:"file,omitempty"`
	CDNURL              string `json:"cdn_url,omitempty"`
	PackageName         string `json:"package_name,omitempty"`
	ProvenanceURL       string `json:"provenance_url,omitempty"`
	LocalURL            string `json:"local_url,omitempty"`
	Homepage            string `json:"homepage,omitempty"`
	License             string `json:"license,omitempty"`
	LicenseFile         string `json:"license_file,omitempty"`
	LicenseURL          string `json:"license_url,omitempty"`
	LicenseIntegrity    string `json:"license_integrity,omitempty"`
	LicenseLocalURL     string `json:"-"`
	Purpose             string `json:"purpose"`
	Attribution         bool   `json:"attribution,omitempty"`
	Enabled             bool   `json:"enabled"`
	IncludeInMinimal    bool   `json:"include_in_minimal"`
	Defer               bool   `json:"defer"`
	WaitForWindowLoaded bool   `json:"wait_for_window_loaded,omitempty"`
	Integrity           string `json:"integrity,omitempty"`
}

func parseManifest(contents []byte) (runtimeManifest, error) {
	var manifest runtimeManifest
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return runtimeManifest{}, fmt.Errorf("parse runtime manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return runtimeManifest{}, fmt.Errorf("parse runtime manifest: trailing JSON value")
	}
	if manifest.Schema != supportedManifestSchema {
		return runtimeManifest{}, fmt.Errorf("runtime manifest schema = %d, want %d", manifest.Schema, supportedManifestSchema)
	}
	if err := validateRuntimeAsset(manifest.Loader, true); err != nil {
		return runtimeManifest{}, fmt.Errorf("loader: %w", err)
	}
	if manifest.Loader.RoleGoName == "" {
		manifest.Loader.RoleGoName = manifest.Loader.GoName
	}
	seenRoles := map[string]bool{manifest.Loader.Role: true}
	seenGoNames := map[string]bool{manifest.Loader.GoName: true}
	seenRoleGoNames := map[string]bool{manifest.Loader.RoleGoName: true}
	seenModules := make(map[string]bool)
	for index := range manifest.Dependencies {
		asset := &manifest.Dependencies[index]
		if err := validateRuntimeAsset(*asset, false); err != nil {
			return runtimeManifest{}, fmt.Errorf("dependencies[%d]: %w", index, err)
		}
		if asset.RoleGoName == "" {
			asset.RoleGoName = asset.GoName
		}
		if seenRoles[asset.Role] {
			return runtimeManifest{}, fmt.Errorf("dependencies[%d]: duplicate role %q", index, asset.Role)
		}
		if seenGoNames[asset.GoName] {
			return runtimeManifest{}, fmt.Errorf("dependencies[%d]: duplicate go_name %q", index, asset.GoName)
		}
		if seenRoleGoNames[asset.RoleGoName] {
			return runtimeManifest{}, fmt.Errorf("dependencies[%d]: duplicate role_go_name %q", index, asset.RoleGoName)
		}
		seenRoles[asset.Role] = true
		seenGoNames[asset.GoName] = true
		seenRoleGoNames[asset.RoleGoName] = true
		if asset.Module != "" {
			if seenModules[asset.Module] {
				return runtimeManifest{}, fmt.Errorf("dependencies[%d]: duplicate module %q", index, asset.Module)
			}
			seenModules[asset.Module] = true
			asset.LocalURL = urlPath(asset.Module, dep{Version: asset.Version, File: asset.File})
			asset.LicenseLocalURL = urlPath(asset.Module, dep{Version: asset.Version, File: asset.LicenseFile})
		}
	}
	return manifest, nil
}

func validateRuntimeAsset(asset runtimeAssetManifest, loader bool) error {
	if err := validateCommonRuntimeAsset(asset); err != nil {
		return err
	}
	if loader {
		return validateLoaderAsset(asset)
	}
	if asset.Module == "" {
		return validateFirstPartyAsset(asset)
	}
	return validateVendoredAsset(asset)
}

func validateCommonRuntimeAsset(asset runtimeAssetManifest) error {
	for field, value := range map[string]string{
		"role": asset.Role, "go_name": asset.GoName, "name": asset.Name, "purpose": asset.Purpose,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	if !token.IsIdentifier(asset.GoName) || !ast.IsExported(asset.GoName) {
		return fmt.Errorf("go_name must be an exported Go identifier")
	}
	if asset.RoleGoName != "" && (!token.IsIdentifier(asset.RoleGoName) || !ast.IsExported(asset.RoleGoName)) {
		return fmt.Errorf("role_go_name must be an exported Go identifier")
	}
	return nil
}

func validateLoaderAsset(asset runtimeAssetManifest) error {
	if !safeEmbeddedJavaScriptURL(asset.LocalURL) {
		return fmt.Errorf("local_url must be a safe /assets/js/*.js URL")
	}
	if declaresVendoredFields(asset) {
		return fmt.Errorf("loader cannot be vendored or attributed")
	}
	return nil
}

func validateFirstPartyAsset(asset runtimeAssetManifest) error {
	if !safeEmbeddedJavaScriptURL(asset.LocalURL) {
		return fmt.Errorf("first-party local_url must be a safe /assets/js/*.js URL")
	}
	if declaresVendoredFields(asset) {
		return fmt.Errorf("first-party dependency cannot declare vendored or attribution fields")
	}
	return nil
}

func validateVendoredAsset(asset runtimeAssetManifest) error {
	if asset.Version == "" || asset.File == "" || asset.CDNURL == "" || asset.Integrity == "" || asset.PackageName == "" || asset.ProvenanceURL == "" {
		return fmt.Errorf("vendored dependency requires version, file, cdn_url, integrity, package_name, and provenance_url")
	}
	if !safePathSegment(asset.Module) || !safePathSegment(asset.Version) || !safePathSegment(asset.File) || !safePathSegment(asset.LicenseFile) {
		return fmt.Errorf("module, version, file, and license_file must be safe path segments")
	}
	if asset.LicenseFile == asset.File {
		return fmt.Errorf("license_file must differ from file")
	}
	for field, template := range map[string]string{"cdn_url": asset.CDNURL, "provenance_url": asset.ProvenanceURL, "license_url": asset.LicenseURL} {
		if err := validateVersionedHTTPSURL(field, template, asset.Version); err != nil {
			return err
		}
	}
	for field, integrity := range map[string]string{"integrity": asset.Integrity, "license_integrity": asset.LicenseIntegrity} {
		if err := validateSHA384Integrity(field, integrity); err != nil {
			return err
		}
	}
	if !asset.Attribution || asset.Homepage == "" || asset.License == "" {
		return fmt.Errorf("vendored dependency requires attribution, homepage, and license")
	}
	if !absoluteHTTPSURL(asset.Homepage) {
		return fmt.Errorf("homepage must be an absolute HTTPS URL")
	}
	return nil
}

func declaresVendoredFields(asset runtimeAssetManifest) bool {
	return asset.Module != "" || asset.Version != "" || asset.File != "" || asset.CDNURL != "" ||
		asset.Integrity != "" || asset.PackageName != "" || asset.ProvenanceURL != "" ||
		asset.Attribution || asset.Homepage != "" || asset.License != "" || asset.LicenseFile != "" ||
		asset.LicenseURL != "" || asset.LicenseIntegrity != ""
}

func validateVersionedHTTPSURL(field, template, version string) error {
	if strings.Count(template, "{v}") != 1 || !absoluteHTTPSURL(strings.ReplaceAll(template, "{v}", version)) {
		return fmt.Errorf("%s must contain exactly one {v} and expand to an absolute HTTPS URL", field)
	}
	return nil
}

func validateSHA384Integrity(field, integrity string) error {
	encodedIntegrity, found := strings.CutPrefix(integrity, "sha384-")
	decodedIntegrity, err := base64.StdEncoding.DecodeString(encodedIntegrity)
	if !found || err != nil || len(decodedIntegrity) != 48 {
		return fmt.Errorf("%s must be a SHA-384 SRI value", field)
	}
	return nil
}

func safeEmbeddedJavaScriptURL(value string) bool {
	return strings.HasPrefix(value, "/assets/js/") && strings.HasSuffix(value, ".js") && path.Clean(value) == value && !strings.ContainsAny(value, `%?#\`)
}

func safePathSegment(value string) bool {
	return value != "" && value != "." && value != ".." && path.Base(value) == value && !strings.ContainsAny(value, `%?#\`)
}

func absoluteHTTPSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

type vendoredDependency struct {
	Module     string
	Dependency dep
}

func vendoredDependencies(manifest runtimeManifest) []vendoredDependency {
	dependencies := make([]vendoredDependency, 0, len(manifest.Dependencies))
	for _, asset := range manifest.Dependencies {
		if asset.Module == "" {
			continue
		}
		dependencies = append(dependencies, vendoredDependency{Module: asset.Module, Dependency: dep{
			Version:          asset.Version,
			File:             asset.File,
			URL:              asset.CDNURL,
			Integrity:        asset.Integrity,
			PackageName:      asset.PackageName,
			ProvenanceURL:    asset.ProvenanceURL,
			LicenseFile:      asset.LicenseFile,
			LicenseURL:       asset.LicenseURL,
			LicenseIntegrity: asset.LicenseIntegrity,
		}})
	}
	return dependencies
}
