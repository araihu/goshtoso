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

func TestSiteAndPublishedConsumerUseIndependentLibraries(t *testing.T) {
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

	writeFile(t, filepath.Join(rootDir, "go.mod"), "module "+modulePath+"\n\ngo 1.27.0\n")
	writeFile(t, filepath.Join(rootDir, "library.go"), "package library\n\nfunc NewAPI() string { return \"current\" }\n")
	writeFile(t, filepath.Join(siteDir, "go.mod"), fmt.Sprintf("module example.com/site\n\ngo 1.27.0\n\nrequire %s %s\n", modulePath, version))
	writeFile(t, filepath.Join(siteDir, "cmd", "server", "main.go"), `package main

import (
	"fmt"
	"example.com/library"
)

func main() { fmt.Println(library.NewAPI()) }
`)
	writeFile(t, filepath.Join(siteDir, "internal", "themes", "current_source", "doc.go"), "package current_source\n")
	writeFile(t, filepath.Join(siteDir, "internal", "themes", "current_source", "agreement_test.go"), `package current_source_test

import (
	"testing"

	"example.com/library"
)

func TestCurrentSourceOnlyFixture(t *testing.T) {
	if library.NewAPI() != "current" {
		t.Fatalf("unexpected current-source API")
	}
}
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

	consumerDir := filepath.Join(fixture, "consumer")
	writeFile(t, filepath.Join(consumerDir, "go.mod"), fmt.Sprintf("module example.com/consumer\n\ngo 1.27.0\n\nrequire %s %s\n", modulePath, version))
	consumerSource := "package main\nimport (\"fmt\"; \"example.com/library\")\nfunc main() { fmt.Println(library.OldAPI()) }\n"
	writeFile(t, filepath.Join(consumerDir, "main.go"), consumerSource)
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = consumerDir
	tidy.Env = append(env, "GOWORK=off")
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("tidy consumer: %v\n%s", err, out)
	}
	writeFile(t, filepath.Join(fixture, "go.work"), "go 1.27.0\nuse ./root\n")
	consumerEnv := append(env, "GOWORK="+filepath.Join(fixture, "go.work"), "GOSHTOSO_CONSUMER_FIXTURE="+consumerDir, "GOSHTOSO_CONSUMER_MODULE="+modulePath)
	publishedScript := filepath.Join(repoRoot, "scripts", "check-published-consumer")
	runConsumer := func(args ...string) ([]byte, error) {
		cmd := exec.Command(publishedScript, args...)
		cmd.Env = consumerEnv
		return cmd.CombinedOutput()
	}
	out, err := runConsumer()
	if err != nil || !bytes.Contains(out, []byte("published consumer: PASS")) {
		t.Fatalf("published consumer: %v\n%s", err, out)
	}
	manifestPath := filepath.Join(consumerDir, "go.mod")
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	writeProxyModule(t, proxyDir, modulePath, "v0.0.2", "package library\nfunc OldAPI() string { return \"second release\" }\n")
	out, err = runConsumer("v0.0.2")
	if err != nil || !bytes.Contains(out, []byte("PASS (example.com/library@v0.0.2)")) {
		t.Fatalf("release override: %v\n%s", err, out)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("release verification changed fixture manifest: %v", err)
	}
	// The published fixture must not accidentally pick up the checkout's NewAPI.
	writeFile(t, filepath.Join(consumerDir, "main.go"), string(bytes.ReplaceAll([]byte(consumerSource), []byte("OldAPI"), []byte("NewAPI"))))
	out, err = runConsumer()
	if err == nil || !bytes.Contains(out, []byte("undefined: library.NewAPI")) {
		t.Fatalf("consumer used checkout API: %v\n%s", err, out)
	}
	// An absolute replacement would remain usable after copying; reject it explicitly.
	f, err := os.OpenFile(filepath.Join(consumerDir, "go.mod"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprintf(f, "\nreplace %s => %s\n", modulePath, rootDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	out, err = runConsumer()
	if err == nil || !bytes.Contains(out, []byte("must not replace")) {
		t.Fatalf("consumer accepted replacement: %v\n%s", err, out)
	}

}

func TestSiteModuleContractsRunCurrentSourceOnlyThemeFixture(t *testing.T) {
	t.Parallel()

	fixture := t.TempDir()
	rootDir := filepath.Join(fixture, "root")
	siteDir := filepath.Join(fixture, "site")
	proxyDir := filepath.Join(fixture, "proxy")
	modulePath := "example.com/library"
	version := "v0.0.1"

	writeFile(t, filepath.Join(rootDir, "go.mod"), "module "+modulePath+"\n\ngo 1.27.0\n")
	writeFile(t, filepath.Join(rootDir, "library.go"), "package library\n\nfunc SharedAPI() string { return \"shared\" }\nfunc ThemeKey() string { return \"shared\" }\nfunc NewAPI() string { return \"current\" }\n")
	writeFile(t, filepath.Join(siteDir, "go.mod"), fmt.Sprintf("module example.com/site\n\ngo 1.27.0\n\nrequire %s %s\n", modulePath, version))
	writeFile(t, filepath.Join(siteDir, "cmd", "server", "main.go"), `package main

import (
	"fmt"
	"example.com/library"
)

func main() { fmt.Println(library.SharedAPI()) }
`)
	writeFile(t, filepath.Join(siteDir, "internal", "themes", "current_source", "doc.go"), "package current_source\n")
	writeFile(t, filepath.Join(siteDir, "internal", "themes", "catalog.go"), "package themes\n\nconst CatalogKey = \"shared\"\n")
	writeFile(t, filepath.Join(siteDir, "internal", "themes", "current_source", "agreement_test.go"), `package current_source_test

import (
	"os"
	"testing"

	demothemes "example.com/site/internal/themes"
	"example.com/library"
)

func TestCurrentSourceOnlyFixture(t *testing.T) {
	if library.NewAPI() != "current" {
		t.Fatalf("unexpected current-source API")
	}
	if demothemes.CatalogKey != library.ThemeKey() {
		t.Fatalf("theme key drift: demo=%q public=%q", demothemes.CatalogKey, library.ThemeKey())
	}
	sentinel := os.Getenv("GOSHTOSO_TEST_CURRENT_SOURCE_SENTINEL")
	if sentinel == "" {
		t.Fatal("current-source execution sentinel is not configured")
	}
	if err := os.WriteFile(sentinel, []byte("current-source fixture ran\n"), 0o600); err != nil {
		t.Fatalf("write current-source execution sentinel: %v", err)
	}
}
`)
	writeProxyModule(t, proxyDir, modulePath, version, "package library\n\nfunc SharedAPI() string { return \"pinned\" }\n")

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
		"GOSHTOSO_TEST_CURRENT_SOURCE_SENTINEL="+filepath.Join(fixture, "current-source-fixture-ran"),
		"GOPROXY="+proxyURL,
		"GOSUMDB=off",
		"GOMODCACHE="+modCache,
		"GOCACHE="+filepath.Join(fixture, "gocache"),
	)

	script := scriptPath(t)
	sentinelPath := assertCurrentSourceFixtureExecuted(t, script, env, fixture)
	assertCurrentSourceRejectsDemoDrift(t, script, env, siteDir, sentinelPath)

}

func assertCurrentSourceFixtureExecuted(t *testing.T, script string, env []string, fixture string) string {
	t.Helper()

	currentOutput, currentErr := runContract(script, "current-source", env)
	if currentErr != nil {
		t.Fatalf("current-source contract should run the agreement fixture: %v\n%s", currentErr, currentOutput)
	}
	if !bytes.Contains(currentOutput, []byte("current-source agreement fixture: included")) {
		t.Fatalf("current-source output missing fixture marker:\n%s", currentOutput)
	}
	if !outputContainsGoTestSuccess(currentOutput, "example.com/site/internal/themes/current_source") {
		t.Fatalf("current-source output missing executed fixture package:\n%s", currentOutput)
	}
	if !bytes.Contains(currentOutput, []byte("site current-source integration: PASS (3 non-E2E packages; server built)")) {
		t.Fatalf("current-source output missing exact package count:\n%s", currentOutput)
	}

	sentinelPath := filepath.Join(fixture, "current-source-fixture-ran")
	sentinel, readErr := os.ReadFile(sentinelPath)
	if readErr != nil {
		t.Fatalf("current-source fixture execution sentinel missing: %v\n%s", readErr, currentOutput)
	}
	if string(sentinel) != "current-source fixture ran\n" {
		t.Fatalf("unexpected current-source fixture execution sentinel %q", sentinel)
	}
	if err := os.Remove(sentinelPath); err != nil {
		t.Fatalf("reset current-source fixture execution sentinel: %v", err)
	}
	return sentinelPath
}

func assertCurrentSourceRejectsDemoDrift(t *testing.T, script string, env []string, siteDir, sentinelPath string) {
	t.Helper()

	demoCatalogPath := filepath.Join(siteDir, "internal", "themes", "catalog.go")
	demoCatalog, readErr := os.ReadFile(demoCatalogPath)
	if readErr != nil {
		t.Fatalf("read demo catalog for drift mutation: %v", readErr)
	}
	mutatedCatalog := bytes.Replace(demoCatalog, []byte("\"shared\""), []byte("\"demo-only\""), 1)
	if bytes.Equal(mutatedCatalog, demoCatalog) {
		t.Fatal("demo catalog drift mutation did not change a key")
	}
	if err := os.WriteFile(demoCatalogPath, mutatedCatalog, 0o644); err != nil {
		t.Fatalf("mutate demo catalog key: %v", err)
	}
	mutatedOutput, mutatedErr := runContract(script, "current-source", env)
	if mutatedErr == nil {
		t.Fatalf("current-source contract should reject the mutated demo key:\n%s", mutatedOutput)
	}
	for _, want := range [][]byte{
		[]byte("theme key drift: demo=\"demo-only\" public=\"shared\""),
		[]byte("site current-source integration failed during non-E2E tests"),
	} {
		if !bytes.Contains(mutatedOutput, want) {
			t.Fatalf("current-source drift output missing %q:\n%s", want, mutatedOutput)
		}
	}
	if _, err := os.Stat(sentinelPath); !os.IsNotExist(err) {
		t.Fatalf("current-source drift run did not fail before the execution sentinel; stat error: %v\n%s", err, mutatedOutput)
	}
	if err := os.WriteFile(demoCatalogPath, demoCatalog, 0o644); err != nil {
		t.Fatalf("restore demo catalog key after drift proof: %v", err)
	}
}

func scriptPath(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(repoRoot, "scripts", "check-site-module")
}

func runContract(script, mode string, env []string) ([]byte, error) {
	cmd := exec.Command(script, mode)
	cmd.Env = env
	return cmd.CombinedOutput()
}

func outputContainsGoTestSuccess(output []byte, packagePath string) bool {
	for line := range bytes.SplitSeq(output, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("ok ")) && bytes.Contains(line, []byte(packagePath)) {
			return true
		}
	}
	return false
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
	moduleFile := "module " + modulePath + "\n\ngo 1.27.0\n"
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
