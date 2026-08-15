package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestScrollRegionBFullRunnerUsesExactManifestSelection keeps required CI from
// carrying a second, stale hand-maintained selection beside identities.json.
func TestScrollRegionBFullRunnerUsesExactManifestSelection(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	workspace := scrollRegionBFullRunnerWorkspace(t, repoRoot)
	script := filepath.Join(repoRoot, "scripts", "run-scrollregion-bfull.sh")
	command := exec.Command("bash", script)
	command.Dir = repoRoot
	command.Env = append(withoutEnvironment(os.Environ(), "GOWORK"),
		"GOWORK="+workspace,
		"GOSHTOSO_AXE_CORE_TGZ=/dev/null",
		"GOSHTOSO_SCROLLREGION_BFULL_LIST_ONLY=1",
		"GOSHTOSO_SCROLLREGION_BFULL_XVFB=off",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("required B-FULL list-only preflight: %v\n%s", err, output)
	}

	expectedCommand := exec.Command("go", "run", "./cmd/e2econstraints", "--print-specialized-tests=scrollregion_bfull")
	expectedCommand.Dir = filepath.Join(repoRoot, "site")
	expectedCommand.Env = append(withoutEnvironment(os.Environ(), "GOWORK"), "GOWORK="+workspace)
	expected, err := expectedCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("derive manifest-selected tests: %v\n%s", err, expected)
	}
	actual := selectedTestLines(output)
	if got, want := strings.Join(actual, "\n"), strings.Join(selectedTestLines(expected), "\n"); got != want {
		t.Fatalf("required runner selection differs from identities.json\nwant:\n%s\ngot:\n%s", want, got)
	}
}

// scrollRegionBFullRunnerWorkspace keeps the runner's nested site module on
// its real current-source workspace even when the outer standalone scripts
// package is intentionally tested with GOWORK=off. It is temporary test
// plumbing, never a repo-local go.work mutation.
func scrollRegionBFullRunnerWorkspace(t *testing.T, repoRoot string) string {
	t.Helper()
	directory := t.TempDir()
	workspace := filepath.Join(directory, "go.work")
	command := exec.Command("go", "work", "init", repoRoot, filepath.Join(repoRoot, "site"))
	command.Dir = directory
	command.Env = withoutEnvironment(os.Environ(), "GOWORK")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("create temporary root/site workspace: %v\n%s", err, output)
	}
	if _, err := os.Stat(workspace); err != nil {
		t.Fatalf("temporary root/site workspace missing: %v", err)
	}
	return workspace
}

func selectedTestLines(content []byte) []string {
	var tests []string
	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TestScrollRegion") && !strings.ContainsAny(line, " \t") {
			tests = append(tests, line)
		}
	}
	return tests
}

func TestScrollRegionBFullRunnerDoesNotCarryHandMaintainedSelection(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, "scripts", "run-scrollregion-bfull.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if regexp.MustCompile(`(?m)^tests=\(`).Match(content) {
		t.Fatal("required B-FULL runner carries a hand-maintained tests array instead of deriving identities.json selected_tests")
	}
}

func TestScrollRegionBFullRequiredTimeoutIsViableForLiteralClosure(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := os.ReadFile(filepath.Join(repoRoot, "scripts", "run-scrollregion-bfull.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(runner), "-timeout=60m") {
		t.Fatal("literal B-FULL runner must allow the measured full matrix to finish")
	}
	workflow, err := os.ReadFile(filepath.Join(repoRoot, ".github", "workflows", "required.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(workflow), "timeout-minutes: 75") {
		t.Fatal("required workflow must outlive the 60-minute Go B-FULL timeout")
	}
	if !strings.Contains(string(workflow), "ref: ${{ github.event.pull_request.head.sha || github.sha }}") {
		t.Fatal("required workflow must checkout the exact PR head, never the synthetic merge revision")
	}
}
