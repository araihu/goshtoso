package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const workerRevision = "1111111111111111111111111111111111111111"

func TestWorkerDeploymentSelectsVerifiedCurrentMain(t *testing.T) {
	script := workerWorkflowStep(t, "Select verified main revision")
	for _, tc := range []struct {
		name, revision, main, passed, apiFailure, output string
		wantError                                        bool
	}{
		{"verified main", workerRevision, workerRevision, "true", "", "deploy=true", false},
		{"failed CI", workerRevision, workerRevision, "false", "", "", true},
		{"superseded", workerRevision, strings.Repeat("2", 40), "false", "", "deploy=false", false},
		{"unavailable main", workerRevision, workerRevision, "true", "yes", "", true},
		{"invalid revision", "not-a-sha", workerRevision, "true", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := workerDeployFixture(t)
			output, err := runWorkerWorkflowStep(t, script, dir,
				"SOURCE_SHA="+tc.revision, "MAIN_SHA="+tc.main,
				"CI_PASSED="+tc.passed, "MAIN_API_FAILURE="+tc.apiFailure)
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v, output=%s", err, output)
			}
			selected, _ := os.ReadFile(filepath.Join(dir, "output"))
			if tc.output == "" && len(selected) != 0 || tc.output != "" && !strings.Contains(string(selected), tc.output) {
				t.Fatalf("selection=%q, want %q", selected, tc.output)
			}
		})
	}
}

func TestWorkerDeploymentRechecksMainBeforeUpload(t *testing.T) {
	script := workerWorkflowStep(t, "Deploy verified revision")
	for _, tc := range []struct {
		name, main, token string
		wantError, upload bool
	}{
		{"current", workerRevision, "test-token", false, true},
		{"main advanced during setup", strings.Repeat("2", 40), "test-token", false, false},
		{"missing credential", workerRevision, "", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := workerDeployFixture(t)
			output, err := runWorkerWorkflowStep(t, script, dir,
				"REVISION="+workerRevision, "MAIN_SHA="+tc.main, "CLOUDFLARE_API_TOKEN="+tc.token)
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v, output=%s", err, output)
			}
			calls, _ := os.ReadFile(filepath.Join(dir, "calls"))
			if got := len(calls) > 0; got != tc.upload {
				t.Fatalf("uploaded=%v, want %v", got, tc.upload)
			}
		})
	}
}

func workerWorkflowStep(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("../.github/workflows/deploy.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, step, ok := strings.Cut(string(data), "      - name: "+name+"\n")
	if !ok {
		t.Fatalf("missing workflow step %q", name)
	}
	step, _, _ = strings.Cut(step, "\n      - name:")
	_, script, ok := strings.Cut(step, "        run: |\n")
	if !ok {
		t.Fatalf("missing commands for %q", name)
	}
	return script
}

func workerDeployFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"git": "#!/usr/bin/env bash\nprintf '%s\\n' \"$SOURCE_SHA\"\n",
		"gh": `#!/usr/bin/env bash
set -eu
case "$1" in
  api) [[ "${MAIN_API_FAILURE:-}" != yes ]]; printf '%s\n' "$MAIN_SHA" ;;
  run) printf '%s\n' "$CI_PASSED" ;;
  *) exit 1 ;;
esac
`,
		"npm": "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \"$CALLS\"\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func runWorkerWorkflowStep(t *testing.T, script, dir string, env ...string) ([]byte, error) {
	t.Helper()
	command := exec.Command("bash", "-euc", script)
	command.Dir = dir
	command.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"),
		"GITHUB_REPOSITORY=araihu/goshtoso", "GITHUB_RUN_ID=1",
		"GITHUB_OUTPUT="+filepath.Join(dir, "output"),
		"GITHUB_STEP_SUMMARY="+filepath.Join(dir, "summary"), "CALLS="+filepath.Join(dir, "calls"))
	command.Env = append(command.Env, env...)
	return command.CombinedOutput()
}

func TestWorkerWorkflowKeepsCredentialsOnTrustedMainAndRetiresImagePublication(t *testing.T) {
	deploy, err := os.ReadFile("../.github/workflows/deploy.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, guard := range []string{
		"github.repository == 'araihu/goshtoso'", "github.ref == 'refs/heads/main'",
		"github.event.workflow_run.event == 'push'", "github.event.workflow_run.head_branch == 'main'",
		"github.event.workflow_run.head_repository.full_name == github.repository",
		"cancel-in-progress: false", "if: steps.source.outputs.deploy == 'true'",
		"if: steps.deploy.outputs.deployed == 'true'",
	} {
		if !strings.Contains(string(deploy), guard) {
			t.Fatalf("missing production guard %q", guard)
		}
	}
	release, err := os.ReadFile("../.github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"packages: write", "docker push", "publish-release-image", "image-goshtoso"} {
		if strings.Contains(string(deploy)+string(release), forbidden) {
			t.Fatalf("retired image publication remains: %s", forbidden)
		}
	}
	if _, err := os.Stat("publish-release-image"); !os.IsNotExist(err) {
		t.Fatal("retired image publication script remains")
	}
}
