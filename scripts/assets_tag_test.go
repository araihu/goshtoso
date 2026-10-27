package scripts_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyAssetsTag(t *testing.T) {
	t.Parallel()
	revision := strings.Repeat("a", 40)
	tag := strings.Repeat("b", 40)
	for _, tc := range []struct {
		name, ref, annotated string
		valid                bool
	}{
		{"lightweight", fmt.Sprintf(`{"object":{"type":"commit","sha":%q}}`, revision), "", true},
		{"annotated", fmt.Sprintf(`{"object":{"type":"tag","sha":%q}}`, tag), fmt.Sprintf("{\n  \"sha\": %q,\n  \"object\": {\"type\":\"commit\",\"sha\":%q}\n}", tag, revision), true},
		{"wrong revision", fmt.Sprintf(`{"object":{"type":"commit","sha":%q}}`, tag), "", false},
		{"annotated wrong revision", fmt.Sprintf(`{"object":{"type":"tag","sha":%q}}`, tag), fmt.Sprintf(`{"sha":%q,"object":{"type":"commit","sha":%q}}`, revision, tag), false},
		{"non commit", fmt.Sprintf(`{"object":{"type":"tree","sha":%q}}`, revision), "", false},
		{"missing object", `{}`, "", false},
		{"invalid json", `not json`, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for name, content := range map[string]string{
				"ref.json": tc.ref, "tag.json": tc.annotated,
				"curl": `#!/usr/bin/env bash
set -euo pipefail
case "${!#}" in
  https://api.github.com/repos/araihu/assets/git/ref/tags/v0.2.6) cat "$FIXTURES/ref.json";;
  https://api.github.com/repos/araihu/assets/git/tags/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb) cat "$FIXTURES/tag.json";;
  *) exit 22;;
esac
`,
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("bash", "verify-assets-tag")
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "FIXTURES="+dir, "GH_TOKEN=fixture", "RELEASE=v0.2.6", "ASSETS_REVISION="+revision)
			output, err := cmd.CombinedOutput()
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v, output=%s", tc.valid, err, output)
			}
		})
	}
}
