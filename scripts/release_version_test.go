package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/assets"
)

func TestReleasePublisherReadsTailwindVersion(t *testing.T) {
	module, err := os.ReadFile("../.dagger/src/index.ts")
	if err != nil {
		t.Fatal(err)
	}
	line := regexp.MustCompile(`(?m)^ver=\$\(sed .*\)$`).FindString(string(module))
	if line == "" {
		t.Fatal("release publisher version command not found")
	}
	// Decode the escaped backslashes in the TypeScript template literal, then
	// execute the actual publisher command against the manifest fixtures.
	command := strings.ReplaceAll(line, `\\`, `\`) + "\ntest -n \"$ver\"\nprintf '%s' \"$ver\""
	manifest, err := os.ReadFile("../muamba.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, manifest, want string
	}{
		{"repository manifest", string(manifest), assets.TailwindVersion()},
		{"unquoted", "resources:\n  other:\n    version: 1.2.3\n  tailwindcss:\n    version: 9.8.7\n  after:\n    version: 2.3.4\n", "9.8.7"},
		{"quoted", "resources:\n  tailwindcss:\n    version: \"9.8.7\"\n", "9.8.7"},
		{"missing", "resources:\n  other:\n    version: 1.2.3\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "muamba.yaml"), []byte(tc.manifest), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", "-euc", command)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			if tc.want == "" {
				if err == nil {
					t.Fatal("publisher accepted a missing Tailwind version")
				}
				return
			}
			if err != nil || string(output) != tc.want {
				t.Fatalf("publisher version = %q, err = %v; want %q", output, err, tc.want)
			}
		})
	}
}
