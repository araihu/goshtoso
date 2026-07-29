package scripts_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestSiteModuleContractsSeparateCurrentSourceFromPinnedDependency(t *testing.T) {
	t.Parallel()

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(repoRoot, "scripts", "check-site-module")

	fixture := t.TempDir()
	rootDir := filepath.Join(fixture, "root")
	siteDir := filepath.Join(fixture, "site")
	proxyDir := filepath.Join(fixture, "proxy")
	modulePath := "example.com/library"
	version := "v0.0.1"

	writeFile(t, filepath.Join(rootDir, "go.mod"), "module "+modulePath+"\n\ngo 1.26.5\n")
	writeFile(t, filepath.Join(rootDir, "library.go"), "package library\n\nfunc NewAPI() string { return \"current\" }\n")
	writeFile(t, filepath.Join(siteDir, "go.mod"), fmt.Sprintf("module example.com/site\n\ngo 1.26.5\n\nrequire %s %s\n", modulePath, version))
	writeFile(t, filepath.Join(siteDir, "cmd", "server", "main.go"), `package main

import (
	"fmt"
	"example.com/library"
)

func main() { fmt.Println(library.NewAPI()) }
`)
	writeProxyModule(t, proxyDir, modulePath, version, "package library\n\nfunc OldAPI() string { return \"pinned\" }\n")

	proxyURL := (&url.URL{Scheme: "file", Path: proxyDir}).String()
	modCache := filepath.Join(fixture, "modcache")
	t.Cleanup(func() {
		_ = filepath.Walk(modCache, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			return os.Chmod(path, info.Mode()|0o200)
		})
	})
	env := append(os.Environ(),
		"GOSHTOSO_CONTRACT_ROOT="+rootDir,
		"GOSHTOSO_CONTRACT_SITE="+siteDir,
		"GOSHTOSO_CONTRACT_DEPENDENCY="+modulePath,
		"GOPROXY="+proxyURL,
		"GOSUMDB=off",
		"GOMODCACHE="+modCache,
		"GOCACHE="+filepath.Join(fixture, "gocache"),
	)

	currentOutput, currentErr := runContract(script, "current-source", env)
	if currentErr != nil {
		t.Fatalf("current-source contract should pass: %v\n%s", currentErr, currentOutput)
	}
	if !bytes.Contains(currentOutput, []byte("site current-source integration: PASS")) {
		t.Fatalf("current-source output missing success marker:\n%s", currentOutput)
	}

	pinnedEnv := append(append([]string{}, env...), "GOFLAGS=-mod=mod")
	pinnedOutput, pinnedErr := runContract(script, "pinned-dependency", pinnedEnv)
	if pinnedErr == nil {
		t.Fatalf("pinned-dependency contract should reject unavailable API:\n%s", pinnedOutput)
	}
	for _, want := range [][]byte{
		[]byte("undefined: library.NewAPI"),
		[]byte("site pinned-dependency deployability failed during non-E2E tests"),
	} {
		if !bytes.Contains(pinnedOutput, want) {
			t.Fatalf("pinned-dependency output missing %q:\n%s", want, pinnedOutput)
		}
	}
}

func runContract(script, mode string, env []string) ([]byte, error) {
	cmd := exec.Command(script, mode)
	cmd.Env = env
	return cmd.CombinedOutput()
}

func writeProxyModule(t *testing.T, proxyDir, modulePath, version, source string) {
	t.Helper()

	versionDir := filepath.Join(proxyDir, filepath.FromSlash(modulePath), "@v")
	writeFile(t, filepath.Join(versionDir, "list"), version+"\n")
	writeFile(t, filepath.Join(versionDir, version+".info"), fmt.Sprintf(
		"{\"Version\":%q,\"Time\":%q}\n",
		version,
		time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	))
	moduleFile := "module " + modulePath + "\n\ngo 1.26.5\n"
	writeFile(t, filepath.Join(versionDir, version+".mod"), moduleFile)

	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	zipFile, err := os.Create(filepath.Join(versionDir, version+".zip"))
	if err != nil {
		t.Fatal(err)
	}
	zipWriter := zip.NewWriter(zipFile)
	prefix := modulePath + "@" + version + "/"
	for name, content := range map[string]string{"go.mod": moduleFile, "library.go": source} {
		entry, createErr := zipWriter.Create(prefix + name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := entry.Write([]byte(content)); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
