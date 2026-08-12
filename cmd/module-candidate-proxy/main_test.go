package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/internal/modulecandidate"
)

func TestRunRequiresExplicitIdentityAndPrintsMachineResult(t *testing.T) {
	stdout := temporaryFile(t)
	stderr := temporaryFile(t)
	if code := run(nil, stdout, stderr, nil); code != 2 {
		t.Fatalf("run(nil) = %d, want usage exit 2", code)
	}
	if output := readTemporaryFile(t, stderr); !strings.Contains(output, "-repository is required") {
		t.Fatalf("stderr = %q", output)
	}

	stdout = temporaryFile(t)
	stderr = temporaryFile(t)
	args := []string{
		"-repository", "/candidate", "-module-path", "github.com/araihu/goshtoso",
		"-commit", strings.Repeat("a", 40), "-tree", strings.Repeat("b", 40),
		"-subdir", "", "-output", "/proxy", "-dependency-proxy", "file:///dependencies",
	}
	build := func(_ context.Context, config modulecandidate.Config) (modulecandidate.Result, error) {
		if config.Subdir != "" || config.DependencyProxy != "file:///dependencies" {
			t.Fatalf("config = %+v", config)
		}
		return modulecandidate.Result{ModulePath: config.ModulePath, Version: "v0.0.0-20260812015850-aaaaaaaaaaaa", Commit: config.Commit, Tree: config.ExpectedTree, ManifestPath: "/proxy/module-candidate-manifest.json"}, nil
	}
	if code := run(args, stdout, stderr, build); code != 0 {
		t.Fatalf("run() = %d, stderr=%q", code, readTemporaryFile(t, stderr))
	}
	var result modulecandidate.Result
	if err := json.Unmarshal([]byte(readTemporaryFile(t, stdout)), &result); err != nil {
		t.Fatal(err)
	}
	if result.Version == "" || result.ManifestPath != "/proxy/module-candidate-manifest.json" {
		t.Fatalf("result = %+v", result)
	}
}

func temporaryFile(t *testing.T) *os.File {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stream-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	return file
}

func readTemporaryFile(t *testing.T, file *os.File) string {
	t.Helper()
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}
