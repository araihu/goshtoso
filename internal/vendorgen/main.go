// Command vendorgen generates every consumer of the canonical embedded
// JavaScript manifest at assets/js/runtime/manifest.json.
//
// Modes:
//
//	go run ./cmd/vendorgen                // verify local bytes and regenerate views
//	go run ./cmd/vendorgen -check         // CI: verify bytes and fail on generated drift
//	go run ./cmd/vendorgen -verify-remote // fetch CDN URLs and verify canonical SRI
//	go run ./cmd/vendorgen -download      // verify and replace vendored bytes from CDN
//
// Run from the repository root.
package vendorgen

import (
	"crypto/sha512"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	manifestPath             = "assets/js/runtime/manifest.json"
	vendorConstantsPath      = "assets/vendor_gen.go"
	runtimeManifestPath      = "assets/runtime_manifest_gen.go"
	versionsCompatPath       = "assets/js/runtime/versions.json"
	runtimeAttributionsPath  = "site/internal/pages/demo/contentpages/legal/runtime_attributions_gen.go"
	runtimeDocumentationPath = "docs/RUNTIME_DEPENDENCIES.md"
	vendorRoot               = "assets/js/runtime"
)

// dep is one vendored third-party JavaScript dependency.
type dep struct {
	Version          string `json:"version"`
	File             string `json:"file"`
	URL              string `json:"url"`
	Integrity        string `json:"-"`
	PackageName      string `json:"-"`
	ProvenanceURL    string `json:"-"`
	LicenseFile      string `json:"-"`
	LicenseURL       string `json:"-"`
	LicenseIntegrity string `json:"-"`
}

type generatedArtifact struct {
	path     string
	contents string
}

func urlPath(module string, d dep) string {
	return "/assets/js/runtime/" + module + "/" + d.Version + "/" + d.File
}

func diskPath(module string, d dep) string {
	return filepath.Join(vendorRoot, module, d.Version, d.File)
}

func licenseDiskPath(module string, d dep) string {
	return filepath.Join(vendorRoot, module, d.Version, d.LicenseFile)
}

func integrityForBytes(contents []byte) string {
	sum := sha512.Sum384(contents)
	return "sha384-" + base64.StdEncoding.EncodeToString(sum[:])
}

func loadManifest() (runtimeManifest, error) {
	contents, err := os.ReadFile(manifestPath)
	if err != nil {
		return runtimeManifest{}, err
	}
	manifest, err := parseManifest(contents)
	if err != nil {
		return runtimeManifest{}, fmt.Errorf("parse %s: %w", manifestPath, err)
	}
	return manifest, nil
}

func verifyFiles(dependencies []vendoredDependency) error {
	for _, declared := range dependencies {
		path := diskPath(declared.Module, declared.Dependency)
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("vendored file missing: %s (run `just vendor-js`): %w", path, err)
		}
		if err := verifyBytes(declared.Module, declared.Dependency, contents); err != nil {
			return fmt.Errorf("vendored file %s: %w", path, err)
		}
		licensePath := licenseDiskPath(declared.Module, declared.Dependency)
		license, err := os.ReadFile(licensePath)
		if err != nil {
			return fmt.Errorf("vendored license missing: %s (run `just vendor-js`): %w", licensePath, err)
		}
		if got := integrityForBytes(license); got != declared.Dependency.LicenseIntegrity {
			return fmt.Errorf("vendored license %s: integrity = %q, want canonical %q", licensePath, got, declared.Dependency.LicenseIntegrity)
		}
	}
	return nil
}

func verifyEmbeddedInventory(manifest runtimeManifest) error {
	var paths []string
	entries, err := os.ReadDir("assets/js")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".js") {
			paths = append(paths, filepath.Join("assets/js", entry.Name()))
		}
	}
	if err := filepath.WalkDir(vendorRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".js") {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return err
	}
	return verifyInventoryPaths(manifest, paths)
}

func verifyInventoryPaths(manifest runtimeManifest, paths []string) error {
	expected := make(map[string]bool, len(manifest.Dependencies)+1)
	assets := append([]runtimeAssetManifest{manifest.Loader}, manifest.Dependencies...)
	for _, asset := range assets {
		expected[filepath.ToSlash(filepath.Join("assets", strings.TrimPrefix(asset.LocalURL, "/assets/")))] = true
	}
	actual := make(map[string]bool, len(paths))
	for _, path := range paths {
		actual[filepath.ToSlash(path)] = true
	}
	var missing, undeclared []string
	for path := range expected {
		if !actual[path] {
			missing = append(missing, path)
		}
	}
	for path := range actual {
		if !expected[path] {
			undeclared = append(undeclared, path)
		}
	}
	if len(missing) == 0 && len(undeclared) == 0 {
		return nil
	}
	sort.Strings(missing)
	sort.Strings(undeclared)
	return fmt.Errorf("embedded JavaScript inventory mismatch: missing=%v undeclared=%v", missing, undeclared)
}

func generatedArtifacts(manifest runtimeManifest) []generatedArtifact {
	return []generatedArtifact{
		{path: vendorConstantsPath, contents: renderVendorConstants(manifest)},
		{path: runtimeManifestPath, contents: renderRuntimeManifest(manifest)},
		{path: versionsCompatPath, contents: renderVersionsCompatibility(manifest)},
		{path: runtimeAttributionsPath, contents: renderRuntimeAttributions(manifest)},
		{path: runtimeDocumentationPath, contents: renderRuntimeDocumentation(manifest)},
	}
}

func checkArtifacts(artifacts []generatedArtifact) error {
	for _, artifact := range artifacts {
		existing, err := os.ReadFile(artifact.path)
		if err != nil || string(existing) != artifact.contents {
			return fmt.Errorf("::error::%s is stale — run `go run ./cmd/vendorgen` and commit", artifact.path)
		}
	}
	return nil
}

func writeArtifacts(artifacts []generatedArtifact, stdout io.Writer) error {
	updates := make([]fileUpdate, 0, len(artifacts))
	for _, artifact := range artifacts {
		updates = append(updates, fileUpdate{path: artifact.path, contents: []byte(artifact.contents), mode: 0o644})
	}
	if err := commitFileUpdates(updates); err != nil {
		return err
	}
	for _, artifact := range artifacts {
		if _, err := fmt.Fprintf(stdout, "vendorgen: wrote %s\n", artifact.path); err != nil {
			return err
		}
	}
	return nil
}

// Run verifies the canonical manifest and regenerates all of its consumers.
func Run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("vendorgen", flag.ContinueOnError)
	flags.SetOutput(stderr)
	check := flags.Bool("check", false, "fail if any generated manifest consumer would change")
	download := flags.Bool("download", false, "download pinned CDN bytes after verifying canonical integrity")
	verifyCDN := flags.Bool("verify-remote", false, "fetch every CDN URL and verify canonical integrity")
	if err := flags.Parse(args); err != nil {
		return err
	}

	manifest, err := loadManifest()
	if err != nil {
		return err
	}
	dependencies := vendoredDependencies(manifest)

	if *download {
		if err := downloadAll(dependencies, stdout); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}
	if err := verifyFiles(dependencies); err != nil {
		return err
	}
	if err := verifyEmbeddedInventory(manifest); err != nil {
		return err
	}
	if *verifyCDN {
		if err := verifyRemote(dependencies, stdout); err != nil {
			return fmt.Errorf("verify remote: %w", err)
		}
	}

	artifacts := generatedArtifacts(manifest)
	if *check {
		return checkArtifacts(artifacts)
	}
	if err := writeArtifacts(artifacts, stdout); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "vendorgen: verified %d canonical dependency hashes\n", len(dependencies))
	return err
}
