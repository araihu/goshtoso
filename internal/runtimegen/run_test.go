package runtimegen

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitFileUpdatesRollsBackOnReplaceFailure(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.go")
	second := filepath.Join(root, "second.go")
	for path, contents := range map[string]string{first: "old first\n", second: "old second\n"} {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	originalRename := renameFile
	renameFile = func(oldPath, newPath string) error {
		if newPath == second && strings.Contains(filepath.Base(oldPath), ".runtimegen-stage-") {
			return errors.New("injected replacement failure")
		}
		return os.Rename(oldPath, newPath)
	}
	t.Cleanup(func() { renameFile = originalRename })

	err := commitFileUpdates([]fileUpdate{
		{path: first, contents: []byte("new first\n"), mode: 0o644},
		{path: second, contents: []byte("new second\n"), mode: 0o644},
	})
	if err == nil || !strings.Contains(err.Error(), "injected replacement failure") {
		t.Fatalf("error = %v", err)
	}
	for path, want := range map[string]string{first: "old first\n", second: "old second\n"} {
		got, readErr := os.ReadFile(path)
		if readErr != nil || string(got) != want {
			t.Fatalf("%s = %q, %v; want %q", path, got, readErr, want)
		}
	}
}

func TestRunGeneratesAndChecksEveryArtifact(t *testing.T) {
	root := t.TempDir()
	overlayPath := filepath.Join(root, "assets", "runtime.overlay.yaml")
	if err := os.MkdirAll(filepath.Dir(overlayPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overlayPath, []byte(validOverlay), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := Run(root, fixtureInventory(t), false, &stdout); err != nil {
		t.Fatal(err)
	}
	for _, output := range outputPaths {
		contents, err := os.ReadFile(filepath.Join(root, output))
		if err != nil || len(contents) == 0 {
			t.Fatalf("output %s: bytes=%d err=%v", output, len(contents), err)
		}
	}
	if !strings.Contains(stdout.String(), "runtimegen: wrote assets/vendor_gen.go") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if err := Run(root, fixtureInventory(t), true, &bytes.Buffer{}); err != nil {
		t.Fatalf("check after generation: %v", err)
	}

	stale := filepath.Join(root, outputPaths[0])
	if err := os.WriteFile(stale, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run(root, fixtureInventory(t), true, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), outputPaths[0]+" is stale") {
		t.Fatalf("stale check error = %v", err)
	}
	contents, err := os.ReadFile(stale)
	if err != nil || string(contents) != "stale\n" {
		t.Fatalf("check mode wrote stale output: %q, %v", contents, err)
	}
}

func TestRunRejectsMissingInputsWithoutPartialOutputs(t *testing.T) {
	root := t.TempDir()
	err := Run(root, fixtureInventory(t), false, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "assets/runtime.overlay.yaml") {
		t.Fatalf("error = %v", err)
	}
	for _, output := range outputPaths {
		if _, statErr := os.Stat(filepath.Join(root, output)); !os.IsNotExist(statErr) {
			t.Fatalf("partial output %s exists: %v", output, statErr)
		}
	}
}
