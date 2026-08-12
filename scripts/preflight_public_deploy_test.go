package scripts_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreflightPublicDeployBuildsExactCandidate(t *testing.T) {
	if os.Getenv("MODULE_CANDIDATE_INTEGRATION") != "1" {
		t.Skip("set MODULE_CANDIDATE_INTEGRATION=1 with exact candidate identity")
	}
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "candidate-proxy")
	receipt := filepath.Join(t.TempDir(), "receipt.env")
	environment := []string{
		"MODULE_CANDIDATE_REPOSITORY=" + repoRoot,
		"MODULE_PATH=github.com/araihu/goshtoso",
		"MODULE_CANDIDATE_COMMIT=" + os.Getenv("MODULE_CANDIDATE_COMMIT"),
		"MODULE_CANDIDATE_TREE=" + os.Getenv("MODULE_CANDIDATE_TREE"),
		"MODULE_CANDIDATE_SUBDIR=",
		"MODULE_CANDIDATE_OUTPUT=" + output,
		"MODULE_CANDIDATE_DEPENDENCY_PROXY=" + os.Getenv("MODULE_CANDIDATE_DEPENDENCY_PROXY"),
		"MODULE_CANDIDATE_RECEIPT=" + receipt,
		"DEPLOY_TARGET=preview",
	}
	for _, key := range []string{"MODULE_CANDIDATE_COMMIT", "MODULE_CANDIDATE_TREE", "MODULE_CANDIDATE_DEPENDENCY_PROXY"} {
		if environmentValue(environment, key) == "" {
			t.Fatalf("integration environment is missing %s", key)
		}
	}
	if err := os.WriteFile(receipt, []byte(receiptFor(environment, "preview")), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("bash", filepath.Join(repoRoot, "scripts", "preflight-public-deploy.sh"))
	command.Env = append(os.Environ(), environment...)
	resultBytes, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("exact candidate preflight: %v\n%s", err, resultBytes)
	}
	var result struct {
		Version      string `json:"version"`
		Commit       string `json:"commit"`
		Tree         string `json:"tree"`
		ManifestPath string `json:"manifestPath"`
	}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("decode preflight result: %v\n%s", err, resultBytes)
	}
	if result.Version == "" || result.Commit != os.Getenv("MODULE_CANDIDATE_COMMIT") || result.Tree != os.Getenv("MODULE_CANDIDATE_TREE") || result.ManifestPath != filepath.Join(output, "module-candidate-manifest.json") {
		t.Fatalf("preflight result = %+v", result)
	}
}

func TestPreflightPublicDeployFailsClosedBeforeCandidateBuild(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(repoRoot, "scripts", "preflight-public-deploy.sh")
	base := []string{
		"MODULE_CANDIDATE_REPOSITORY=" + repoRoot,
		"MODULE_PATH=github.com/araihu/goshtoso",
		"MODULE_CANDIDATE_COMMIT=" + strings.Repeat("a", 40),
		"MODULE_CANDIDATE_TREE=" + strings.Repeat("b", 40),
		"MODULE_CANDIDATE_SUBDIR=",
		"MODULE_CANDIDATE_OUTPUT=" + filepath.Join(t.TempDir(), "proxy"),
		"MODULE_CANDIDATE_DEPENDENCY_PROXY=file:///dependencies",
		"MODULE_CANDIDATE_RECEIPT=" + filepath.Join(t.TempDir(), "receipt.env"),
		"DEPLOY_TARGET=preview",
	}
	for _, test := range []struct {
		name    string
		env     []string
		want    string
		receipt string
	}{
		{name: "absent commit", env: withoutEnvironment(base, "MODULE_CANDIDATE_COMMIT"), want: "required environment is absent: MODULE_CANDIDATE_COMMIT"},
		{name: "missing receipt", env: base, want: "candidate receipt is missing"},
		{name: "network fallback", env: replaceEnvironment(base, "MODULE_CANDIDATE_DEPENDENCY_PROXY", "https://proxy.golang.org,direct"), want: "dependency proxy must use file:// with no fallback"},
		{name: "target mismatch", env: base, receipt: receiptFor(base, "production"), want: "receipt mismatch for DEPLOY_TARGET"},
	} {
		t.Run(test.name, func(t *testing.T) {
			env := append([]string{"PATH=" + os.Getenv("PATH")}, test.env...)
			if test.receipt != "" {
				path := environmentValue(test.env, "MODULE_CANDIDATE_RECEIPT")
				if err := os.WriteFile(path, []byte(test.receipt), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.Command("bash", script)
			command.Env = env
			output, err := command.CombinedOutput()
			if err == nil || !bytes.Contains(output, []byte(test.want)) {
				t.Fatalf("preflight error=%v output=%s, want %q", err, output, test.want)
			}
		})
	}
}

func receiptFor(env []string, target string) string {
	return "MODULE_PATH=" + environmentValue(env, "MODULE_PATH") + "\n" +
		"MODULE_CANDIDATE_COMMIT=" + environmentValue(env, "MODULE_CANDIDATE_COMMIT") + "\n" +
		"MODULE_CANDIDATE_TREE=" + environmentValue(env, "MODULE_CANDIDATE_TREE") + "\n" +
		"MODULE_CANDIDATE_SUBDIR=" + environmentValue(env, "MODULE_CANDIDATE_SUBDIR") + "\n" +
		"DEPLOY_TARGET=" + target + "\n"
}

func environmentValue(env []string, key string) string {
	prefix := key + "="
	for _, value := range env {
		if strings.HasPrefix(value, prefix) {
			return strings.TrimPrefix(value, prefix)
		}
	}
	return ""
}

func withoutEnvironment(env []string, key string) []string {
	var result []string
	for _, value := range env {
		if !strings.HasPrefix(value, key+"=") {
			result = append(result, value)
		}
	}
	return result
}

func replaceEnvironment(env []string, key, replacement string) []string {
	result := withoutEnvironment(env, key)
	return append(result, key+"="+replacement)
}
