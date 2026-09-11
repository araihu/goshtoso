package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseTailwindVersion(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, manifest, want string
	}{
		{"unquoted", "  tailwindcss:\n    version: 4.3.3\n", "4.3.3"},
		{"double quoted", "  tailwindcss:\n    version: \"4.3.3\"\n", "4.3.3"},
		{"single quoted", "  tailwindcss:\n    version: '4.3.3'\n", "4.3.3"},
		{"prerelease", "  tailwindcss:\n    version: 4.4.0-beta.1\n", "4.4.0-beta.1"},
		{"missing resource", "  alpinejs:\n    version: 3.17.2\n", ""},
		{"missing version", "  tailwindcss:\n  alpinejs:\n    version: 3.17.2\n", ""},
		{"invalid version", "  tailwindcss:\n    version: latest\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "muamba.yaml")
			if err := os.WriteFile(path, []byte("resources:\n"+tc.manifest), 0o600); err != nil {
				t.Fatal(err)
			}
			output, err := exec.Command("bash", "release-tailwind-version", path).CombinedOutput()
			if tc.want == "" {
				if err == nil {
					t.Fatalf("expected rejection, got %q", output)
				}
				return
			}
			if err != nil || strings.TrimSpace(string(output)) != tc.want {
				t.Fatalf("version = %q, error = %v; want %q", output, err, tc.want)
			}
		})
	}
}
