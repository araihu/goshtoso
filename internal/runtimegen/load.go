package runtimegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/araihu/assets/assetmeta"
)

func Load(reader io.Reader, inventory *assetmeta.Inventory) (Model, error) {
	document, err := assetmeta.Load[RootMetadata, ResourceMeta, DownloadMeta](reader, inventory)
	if err != nil {
		return Model{}, fmt.Errorf("load runtime overlay: %w", err)
	}

	loader, err := normalizeLocal(document.Metadata.Loader)
	if err != nil {
		return Model{}, fmt.Errorf("loader: %w", err)
	}
	model := Model{Loader: loader, Dependencies: make([]Asset, 0, len(document.Metadata.Order))}

	for index, entry := range document.Metadata.Order {
		hasDownload := entry.Download.Resource != "" || entry.Download.Download != ""
		hasLocal := strings.TrimSpace(entry.Local) != ""
		if hasDownload == hasLocal {
			return Model{}, fmt.Errorf("order[%d]: exactly one of download or local is required", index)
		}

		var asset Asset
		if hasDownload {
			asset, err = normalizeDownload(document, entry.Download)
		} else {
			local, ok := document.Metadata.FirstParty[entry.Local]
			if !ok {
				return Model{}, fmt.Errorf("order[%d]: unknown local asset %q", index, entry.Local)
			}
			asset, err = normalizeLocal(local)
		}
		if err != nil {
			return Model{}, fmt.Errorf("order[%d]: %w", index, err)
		}
		model.Dependencies = append(model.Dependencies, asset)
	}

	if len(model.Dependencies) == 0 {
		return Model{}, fmt.Errorf("order is empty")
	}
	if err := validateUniqueNames(model); err != nil {
		return Model{}, err
	}
	return model, nil
}

func normalizeDownload(document *assetmeta.Document[RootMetadata, ResourceMeta, DownloadMeta], ref assetmeta.Ref) (Asset, error) {
	if err := assetmeta.ValidateRefs(document.Inventory(), ref); err != nil {
		return Asset{}, err
	}
	resourceMetadata, ok := document.Resources[ref.Resource]
	if !ok {
		return Asset{}, fmt.Errorf("download %s has no metadata", ref)
	}
	downloadMetadata, ok := resourceMetadata.Downloads[ref.Download]
	if !ok {
		return Asset{}, fmt.Errorf("download %s has no metadata", ref)
	}
	metadata := downloadMetadata.Metadata
	if err := validateCommon(metadata.Role, metadata.GoName, metadata.RoleGoName, metadata.Name, metadata.Purpose); err != nil {
		return Asset{}, err
	}
	if metadata.WaitForWindowLoaded && metadata.Defer {
		return Asset{}, fmt.Errorf("wait_for_window_loaded and defer cannot both be true")
	}
	if strings.TrimSpace(metadata.PackageName) == "" {
		return Asset{}, fmt.Errorf("package_name is required")
	}
	if !absoluteHTTPSURL(metadata.Homepage) {
		return Asset{}, fmt.Errorf("homepage must be an absolute HTTPS URL")
	}
	if strings.TrimSpace(metadata.License) == "" {
		return Asset{}, fmt.Errorf("license is required")
	}
	if metadata.LicenseRef.Resource == "" || metadata.LicenseRef.Download == "" {
		return Asset{}, fmt.Errorf("license_ref is required")
	}
	if metadata.ProvenanceRef.Resource == "" || metadata.ProvenanceRef.Download == "" {
		return Asset{}, fmt.Errorf("provenance_ref is required")
	}
	if !metadata.Attribution {
		return Asset{}, fmt.Errorf("attribution is required")
	}
	if err := assetmeta.ValidateRefs(document.Inventory(), metadata.LicenseRef, metadata.ProvenanceRef); err != nil {
		return Asset{}, err
	}

	resolved, _ := document.Resolve(ref)
	license, _ := document.Resolve(metadata.LicenseRef)
	provenance, _ := document.Resolve(metadata.ProvenanceRef)
	localURL, err := acquisitionURL(resolved.Download.Path)
	if err != nil {
		return Asset{}, err
	}
	licenseLocalURL, err := acquisitionURL(license.Download.Path)
	if err != nil {
		return Asset{}, fmt.Errorf("license_ref: %w", err)
	}
	provenanceLocalURL, err := acquisitionURL(provenance.Download.Path)
	if err != nil {
		return Asset{}, fmt.Errorf("provenance_ref: %w", err)
	}

	roleGoName := metadata.RoleGoName
	if roleGoName == "" {
		roleGoName = metadata.GoName
	}
	return Asset{
		Resource: ref.Resource, Download: ref.Download,
		Role: metadata.Role, GoName: metadata.GoName, RoleGoName: roleGoName,
		Name: metadata.Name, Version: resolved.Resource.Version,
		URL: resolved.Download.URL, LocalURL: localURL,
		Integrity: resolved.Download.Integrity, Hash: resolved.Download.Hash,
		PackageName:   metadata.PackageName,
		ProvenanceURL: provenance.Download.URL, ProvenanceLocalURL: provenanceLocalURL,
		Homepage: metadata.Homepage, License: metadata.License,
		LicenseURL: license.Download.URL, LicenseLocalURL: licenseLocalURL,
		LicenseIntegrity: license.Download.Integrity, LicenseHash: license.Download.Hash,
		Purpose: metadata.Purpose, Attribution: metadata.Attribution,
		Enabled: metadata.Enabled, IncludeInMinimal: metadata.IncludeInMinimal,
		Defer: metadata.Defer, WaitForWindowLoaded: metadata.WaitForWindowLoaded,
	}, nil
}

func normalizeLocal(local LocalAsset) (Asset, error) {
	if err := validateCommon(local.Role, local.GoName, local.RoleGoName, local.Name, local.Purpose); err != nil {
		return Asset{}, err
	}
	if !safeEmbeddedJavaScriptURL(local.LocalURL) {
		return Asset{}, fmt.Errorf("local_url must be a safe /assets/js/*.js URL")
	}
	roleGoName := local.RoleGoName
	if roleGoName == "" {
		roleGoName = local.GoName
	}
	return Asset{
		Role: local.Role, GoName: local.GoName, RoleGoName: roleGoName,
		Name: local.Name, LocalURL: local.LocalURL, Purpose: local.Purpose,
		Enabled: local.Enabled, IncludeInMinimal: local.IncludeInMinimal, Defer: local.Defer,
	}, nil
}

func validateCommon(role, goName, roleGoName, name, purpose string) error {
	for field, value := range map[string]string{
		"role": role, "go_name": goName, "name": name, "purpose": purpose,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	if !token.IsIdentifier(goName) || !ast.IsExported(goName) {
		return fmt.Errorf("go_name must be an exported Go identifier")
	}
	if roleGoName != "" && (!token.IsIdentifier(roleGoName) || !ast.IsExported(roleGoName)) {
		return fmt.Errorf("role_go_name must be an exported Go identifier")
	}
	return nil
}

func validateUniqueNames(model Model) error {
	roles := map[string]bool{model.Loader.Role: true}
	goNames := map[string]bool{model.Loader.GoName: true}
	roleGoNames := map[string]bool{model.Loader.RoleGoName: true}
	for index, asset := range model.Dependencies {
		if roles[asset.Role] {
			return fmt.Errorf("dependencies[%d]: duplicate role %q", index, asset.Role)
		}
		if goNames[asset.GoName] {
			return fmt.Errorf("dependencies[%d]: duplicate go_name %q", index, asset.GoName)
		}
		if roleGoNames[asset.RoleGoName] {
			return fmt.Errorf("dependencies[%d]: duplicate role_go_name %q", index, asset.RoleGoName)
		}
		roles[asset.Role] = true
		goNames[asset.GoName] = true
		roleGoNames[asset.RoleGoName] = true
	}
	return nil
}

func acquisitionURL(filePath string) (string, error) {
	trimmed, ok := strings.CutPrefix(filePath, "assets/")
	if !ok || !safeRelativePath(trimmed) {
		return "", fmt.Errorf("path %q must be below assets", filePath)
	}
	return "/assets/" + trimmed, nil
}

func safeEmbeddedJavaScriptURL(value string) bool {
	return strings.HasPrefix(value, "/assets/js/") && strings.HasSuffix(value, ".js") && path.Clean(value) == value && !strings.ContainsAny(value, `%?#\`)
}

func safeRelativePath(value string) bool {
	return value != "" && value != ".." && !strings.HasPrefix(value, "/") && path.Clean(value) == value && !strings.HasPrefix(value, "../") && !strings.ContainsAny(value, `%?#\`)
}

func absoluteHTTPSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}
