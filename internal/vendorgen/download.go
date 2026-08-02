package vendorgen

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var fetchClient = &http.Client{Timeout: 30 * time.Second}

// downloadAll fetches every dep in the manifest into its versioned dir,
// verifies the bytes, and prunes stale version dirs. Run via `just vendor-js`.
func downloadAll(deps []vendoredDependency, stdout io.Writer) error {
	updates := make([]fileUpdate, 0, len(deps)*2)
	for _, declared := range deps {
		module, d := declared.Module, declared.Dependency
		url := strings.ReplaceAll(d.URL, "{v}", d.Version)
		body, err := fetch(url)
		if err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		if err := verifyBytes(module, d, body); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		license, err := fetch(strings.ReplaceAll(d.LicenseURL, "{v}", d.Version))
		if err != nil {
			return fmt.Errorf("%s license: %w", module, err)
		}
		if got := integrityForBytes(license); got != d.LicenseIntegrity {
			return fmt.Errorf("%s license integrity = %q, want canonical %q", module, got, d.LicenseIntegrity)
		}
		updates = append(updates,
			fileUpdate{path: diskPath(module, d), contents: body, mode: 0o644},
			fileUpdate{path: licenseDiskPath(module, d), contents: license, mode: 0o644},
		)
	}
	pruned, restorePruned, discardPruned, err := quarantineStaleVersions(deps)
	if err != nil {
		return err
	}
	if err := commitFileUpdates(updates); err != nil {
		restorePruned()
		return err
	}
	discardPruned()
	for _, declared := range deps {
		module, d := declared.Module, declared.Dependency
		if _, err := fmt.Fprintf(stdout, "vendorgen: fetched %s@%s -> %s and %s\n", module, d.Version, diskPath(module, d), licenseDiskPath(module, d)); err != nil {
			return err
		}
	}
	for _, label := range pruned {
		if _, err := fmt.Fprintf(stdout, "vendorgen: pruned stale %s\n", label); err != nil {
			return err
		}
	}
	return nil
}

// verifyRemote fetches every pinned CDN URL and proves it has the canonical
// integrity recorded in the manifest without changing the embedded files.
func verifyRemote(deps []vendoredDependency, stdout io.Writer) error {
	for _, declared := range deps {
		module, d := declared.Module, declared.Dependency
		url := strings.ReplaceAll(d.URL, "{v}", d.Version)
		body, err := fetch(url)
		if err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		if err := verifyBytes(module, d, body); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		provenance, err := fetch(strings.ReplaceAll(d.ProvenanceURL, "{v}", d.Version))
		if err != nil {
			return fmt.Errorf("%s provenance: %w", module, err)
		}
		var identity struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if err := json.Unmarshal(provenance, &identity); err != nil || identity.Name != d.PackageName || identity.Version != d.Version {
			return fmt.Errorf("%s provenance identity = %q@%q, want %q@%q", module, identity.Name, identity.Version, d.PackageName, d.Version)
		}
		license, err := fetch(strings.ReplaceAll(d.LicenseURL, "{v}", d.Version))
		if err != nil {
			return fmt.Errorf("%s license: %w", module, err)
		}
		if got := integrityForBytes(license); got != d.LicenseIntegrity {
			return fmt.Errorf("%s license integrity = %q, want canonical %q", module, got, d.LicenseIntegrity)
		}
		if _, err := fmt.Fprintf(stdout, "vendorgen: verified remote %s@%s (%s)\n", module, d.Version, d.Integrity); err != nil {
			return err
		}
	}
	return nil
}

func fetch(url string) ([]byte, error) {
	resp, err := fetchClient.Get(url) //nolint:gosec // url is validated from the committed manifest
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// markers maps each module to substrings the downloaded artifact must contain —
// a guard against a wrong/renamed upstream file. Only Alpine core and htmx core
// embed a `version:"X"` string; plugins/extensions are identified by the API
// token they register. {v} is substituted with the manifest version.
var markers = map[string][]string{
	"alpinejs":          {`version:"{v}"`},
	"alpinejs-collapse": {`directive("collapse"`},
	"alpinejs-focus":    {`magic("focus"`},
	"alpinejs-mask":     {`directive("mask"`},
	"htmx.org":          {`version:"{v}"`},
	"htmx-ext-sse":      {"defineExtension", "EventSource"},
	"htmx-ext-ws":       {"defineExtension", "WebSocket"},
}

// verifyBytes confirms the downloaded artifact carries this module's markers.
func verifyBytes(module string, d dep, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("empty download")
	}
	if got := integrityForBytes(body); got != d.Integrity {
		return fmt.Errorf("integrity = %q, want canonical %q", got, d.Integrity)
	}
	s := string(body)
	for _, m := range markers[module] {
		want := strings.ReplaceAll(m, "{v}", d.Version)
		if !strings.Contains(s, want) {
			return fmt.Errorf("marker %q not found in downloaded bytes", want)
		}
	}
	return nil
}

// pruneStaleVersions first moves every stale directory into one quarantine
// outside the embedded tree. A late discovery error restores prior moves; once
// all moves succeed, deleting the quarantine cannot expose a partial runtime.
func quarantineStaleVersions(deps []vendoredDependency) ([]string, func(), func(), error) {
	quarantine, err := os.MkdirTemp(filepath.Dir(vendorRoot), ".vendorgen-prune-*")
	if err != nil {
		return nil, nil, nil, err
	}
	type movedDir struct{ from, to, label string }
	var moved []movedDir
	restore := func() {
		for index := len(moved) - 1; index >= 0; index-- {
			_ = os.Rename(moved[index].to, moved[index].from)
		}
		_ = os.RemoveAll(quarantine)
	}
	for _, declared := range deps {
		dir := filepath.Join(vendorRoot, declared.Module)
		entries, readErr := os.ReadDir(dir)
		if os.IsNotExist(readErr) {
			continue
		}
		if readErr != nil {
			restore()
			return nil, nil, nil, readErr
		}
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == declared.Dependency.Version {
				continue
			}
			from := filepath.Join(dir, entry.Name())
			to := filepath.Join(quarantine, fmt.Sprintf("%06d", len(moved)))
			if err := os.Rename(from, to); err != nil {
				restore()
				return nil, nil, nil, err
			}
			moved = append(moved, movedDir{from: from, to: to, label: declared.Module + "/" + entry.Name()})
		}
	}
	labels := make([]string, 0, len(moved))
	for _, directory := range moved {
		labels = append(labels, directory.label)
	}
	discard := func() { _ = os.RemoveAll(quarantine) }
	return labels, restore, discard, nil
}
