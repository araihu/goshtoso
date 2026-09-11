package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseImagePublication(t *testing.T) {
	script, err := os.ReadFile("publish-release-image")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, tag, latest, failure string
		wantError, wantLatest      bool
	}{
		{"latest stable", "v0.3.2", "v0.3.2", "", false, true},
		{"old rerun", "v0.3.1", "v0.3.2", "", false, false},
		{"prerelease", "v0.3.3-rc.1", "v0.3.2", "", true, false},
		{"unpublished", "v0.3.2", "v0.3.2", "release", true, false},
		{"smoke failure", "v0.3.2", "v0.3.2", "smoke", true, false},
		{"version push failure", "v0.3.2", "v0.3.2", "push", true, false},
		{"latest lookup failure", "v0.3.2", "v0.3.2", "latest", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, "scripts"), 0700); err != nil {
				t.Fatal(err)
			}
			files := map[string]string{
				"publish": string(script),
				"gh": `#!/usr/bin/env bash
set -eu
if [[ "$2" == */latest ]]; then
  [[ "$FAILURE" != latest ]]
  echo "$LATEST"
else
  [[ "$FAILURE" != release ]]
  echo "$RELEASE_TAG"
fi
`,
				"git": "#!/usr/bin/env bash\necho 123abc\n",
				"docker": `#!/usr/bin/env bash
set -eu
echo "$*" >> "$CALLS"
case "$1" in
  login) cat >/dev/null ;;
  push) [[ "$FAILURE" != push ]] ;;
  inspect) printf 'ghcr.io/araihu/goshtoso@sha256:%064d\n' 1 ;;
esac
`,
				"scripts/check-site-image": "#!/usr/bin/env bash\nset -eu\n[[ \"$2\" == \"$RELEASE_TAG\" ]]\n[[ \"$FAILURE\" != smoke ]]\n",
			}
			for name, body := range files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0700); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("bash", "publish")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"),
				"RELEASE_TAG="+tc.tag, "LATEST="+tc.latest, "FAILURE="+tc.failure,
				"GH_TOKEN=test", "GITHUB_ACTOR=test", "CALLS="+filepath.Join(dir, "calls"),
				"GITHUB_STEP_SUMMARY="+filepath.Join(dir, "summary"))
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.wantError {
				t.Fatalf("err=%v, output=%s", err, out)
			}
			calls, _ := os.ReadFile(filepath.Join(dir, "calls"))
			if got := strings.Contains(string(calls), "push ghcr.io/araihu/goshtoso:latest"); got != tc.wantLatest {
				t.Fatalf("latest push=%v, want %v; calls=%s", got, tc.wantLatest, calls)
			}
			if !tc.wantError {
				for _, want := range []string{"--build-arg GOSHTOSO_DOCS_VERSION=" + tc.tag, "push ghcr.io/araihu/goshtoso:" + tc.tag} {
					if !strings.Contains(string(calls), want) {
						t.Errorf("missing %q in %s", want, calls)
					}
				}
				if _, err := os.Stat(filepath.Join(dir, "image-goshtoso-release.txt")); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
