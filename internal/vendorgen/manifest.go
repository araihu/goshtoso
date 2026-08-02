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
	LocalURL            string `json:"local_url,omitempty"`
	Homepage            string `json:"homepage,omitempty"`
	License             string `json:"license,omitempty"`
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
			asset.LocalURL = urlPath(asset.Module, dep{Version: asset.Version, File: asset.File})
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
	if asset.Module != "" || asset.Version != "" || asset.File != "" || asset.CDNURL != "" || asset.Attribution {
		return fmt.Errorf("loader cannot be vendored or attributed")
	}
	return nil
}

func validateFirstPartyAsset(asset runtimeAssetManifest) error {
	if !safeEmbeddedJavaScriptURL(asset.LocalURL) {
		return fmt.Errorf("first-party local_url must be a safe /assets/js/*.js URL")
	}
	if asset.Version != "" || asset.File != "" || asset.CDNURL != "" || asset.Attribution {
		return fmt.Errorf("first-party dependency cannot declare vendored or attribution fields")
	}
	return nil
}

func validateVendoredAsset(asset runtimeAssetManifest) error {
	if asset.Version == "" || asset.File == "" || asset.CDNURL == "" || asset.Integrity == "" {
		return fmt.Errorf("vendored dependency requires version, file, cdn_url, and integrity")
	}
	if !safePathSegment(asset.Module) || !safePathSegment(asset.Version) || !safePathSegment(asset.File) {
		return fmt.Errorf("module, version, and file must be safe path segments")
	}
	if !absoluteHTTPSURL(strings.ReplaceAll(asset.CDNURL, "{v}", asset.Version)) {
		return fmt.Errorf("cdn_url must expand to an absolute HTTPS URL")
	}
	encodedIntegrity, found := strings.CutPrefix(asset.Integrity, "sha384-")
	decodedIntegrity, err := base64.StdEncoding.DecodeString(encodedIntegrity)
	if !found || err != nil || len(decodedIntegrity) != 48 {
		return fmt.Errorf("integrity must be a SHA-384 SRI value")
	}
	if !asset.Attribution || asset.Homepage == "" || asset.License == "" {
		return fmt.Errorf("vendored dependency requires attribution, homepage, and license")
	}
	if !absoluteHTTPSURL(asset.Homepage) {
		return fmt.Errorf("homepage must be an absolute HTTPS URL")
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

func vendoredDependencies(manifest runtimeManifest) map[string]dep {
	dependencies := make(map[string]dep)
	for _, asset := range manifest.Dependencies {
		if asset.Module == "" {
			continue
		}
		dependencies[asset.Module] = dep{
			Version:   asset.Version,
			File:      asset.File,
			URL:       asset.CDNURL,
			Integrity: asset.Integrity,
		}
	}
	return dependencies
}
