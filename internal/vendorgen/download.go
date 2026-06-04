package vendorgen

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// downloadAll fetches every dep in the manifest into its versioned dir,
// verifies the bytes, and prunes stale version dirs. Run via `just vendor-js`.
func downloadAll(deps map[string]dep, stdout io.Writer) error {
	for module, d := range deps {
		url := strings.ReplaceAll(d.URL, "{v}", d.Version)
		body, err := fetch(url)
		if err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		if err := verifyBytes(module, d, body); err != nil {
			return fmt.Errorf("%s: %w", module, err)
		}
		dst := diskPath(module, d)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, body, 0o644); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "vendorgen: fetched %s@%s -> %s\n", module, d.Version, dst); err != nil {
			return err
		}
		if err := pruneStale(module, d.Version, stdout); err != nil {
			return err
		}
	}
	return nil
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec // url is from the committed manifest
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
	"htmx.org":          {`version:"{v}"`},
	"htmx-ext-sse":      {"defineExtension", "EventSource"},
	"htmx-ext-ws":       {"defineExtension", "WebSocket"},
}

// verifyBytes confirms the downloaded artifact carries this module's markers.
func verifyBytes(module string, d dep, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("empty download")
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

// pruneStale removes assets/js/runtime/<module>/<otherVersion> dirs that are not
// the pinned version, so only the current version ships.
func pruneStale(module, keep string, stdout io.Writer) error {
	dir := filepath.Join(vendorRoot, module)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() != keep {
			if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(stdout, "vendorgen: pruned stale %s/%s\n", module, e.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}
