package runtimegen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/araihu/assets/assetmeta"
)

var outputPaths = []string{
	"assets/vendor_gen.go",
	"assets/runtime_manifest_gen.go",
	"site/internal/pages/demo/contentpages/legal/runtime_attributions_gen.go",
	"docs/RUNTIME_DEPENDENCIES.md",
}

type generatedArtifact struct {
	path     string
	contents string
}

func Run(repoRoot string, inventory *assetmeta.Inventory, check bool, stdout io.Writer) error {
	if stdout == nil {
		return fmt.Errorf("stdout is nil")
	}
	overlayPath := filepath.Join(repoRoot, "assets", "runtime.overlay.yaml")
	file, err := os.Open(overlayPath)
	if err != nil {
		return fmt.Errorf("open assets/runtime.overlay.yaml: %w", err)
	}
	model, loadErr := Load(file, inventory)
	closeErr := file.Close()
	if loadErr != nil {
		return loadErr
	}
	if closeErr != nil {
		return fmt.Errorf("close assets/runtime.overlay.yaml: %w", closeErr)
	}

	artifacts := []generatedArtifact{
		{path: outputPaths[0], contents: renderVendorConstants(model)},
		{path: outputPaths[1], contents: renderRuntimeManifest(model)},
		{path: outputPaths[2], contents: renderRuntimeAttributions(model)},
		{path: outputPaths[3], contents: renderRuntimeDocumentation(model)},
	}
	if check {
		for _, artifact := range artifacts {
			existing, readErr := os.ReadFile(filepath.Join(repoRoot, artifact.path))
			if readErr != nil || string(existing) != artifact.contents {
				return fmt.Errorf("%s is stale; run `go run ./cmd/runtimegen` and commit", artifact.path)
			}
		}
		return nil
	}

	updates := make([]fileUpdate, 0, len(artifacts))
	for _, artifact := range artifacts {
		updates = append(updates, fileUpdate{
			path: filepath.Join(repoRoot, artifact.path), contents: []byte(artifact.contents), mode: 0o644,
		})
	}
	if err := commitFileUpdates(updates); err != nil {
		return err
	}
	for _, artifact := range artifacts {
		if _, err := fmt.Fprintf(stdout, "runtimegen: wrote %s\n", artifact.path); err != nil {
			return err
		}
	}
	return nil
}
