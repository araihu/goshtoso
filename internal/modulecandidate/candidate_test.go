package modulecandidate

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildRejectsCandidateCommitFallback(t *testing.T) {
	_, err := Build(t.Context(), Config{
		Repository: filepath.Join(t.TempDir(), "missing-repository"),
		ModulePath: "github.com/araihu/goshtoso",
		Commit:     "c8639db109bcd23a77316848929a229e00c165c5",
		Subdir:     "",
		Output:     filepath.Join(t.TempDir(), "proxy"),
	})
	if err == nil || !strings.Contains(err.Error(), "resolve candidate commit") {
		t.Fatal("Build() accepted a missing exact candidate instead of failing closed")
	}
}

func TestBuildRejectsExplicitCandidateTreeMismatch(t *testing.T) {
	repository := t.TempDir()
	runGit(t, repository, "init", "-q")
	runGit(t, repository, "config", "user.name", "Module Candidate Test")
	runGit(t, repository, "config", "user.email", "module-candidate@example.invalid")
	runGit(t, repository, "commit", "--allow-empty", "-q", "-m", "candidate")
	commit := runGit(t, repository, "rev-parse", "HEAD")
	tree := runGit(t, repository, "rev-parse", "HEAD^{tree}")

	_, err := Build(t.Context(), Config{
		Repository:   repository,
		ModulePath:   "github.com/araihu/goshtoso",
		Commit:       commit,
		ExpectedTree: strings.Repeat("0", 40),
		Subdir:       "",
		Output:       filepath.Join(t.TempDir(), "proxy"),
	})
	if err == nil || !strings.Contains(err.Error(), "candidate tree mismatch") {
		t.Fatalf("Build() error = %v, want explicit tree mismatch (actual %s)", err, tree)
	}
}

func TestBuildWritesDeterministicExactCommitModuleArtifacts(t *testing.T) {
	repository := t.TempDir()
	runGit(t, repository, "init", "-q")
	runGit(t, repository, "config", "user.name", "Module Candidate Test")
	runGit(t, repository, "config", "user.email", "module-candidate@example.invalid")
	writeCandidateFile(t, repository, "go.mod", "module example.com/candidate\n\ngo 1.26.5\n")
	writeCandidateFile(t, repository, "candidate.go", "package candidate\n\nconst Identity = \"exact\"\n")
	writeCandidateFile(t, repository, "site/go.mod", "module example.com/candidate/site\n\ngo 1.26.5\n")
	writeCandidateFile(t, repository, "site/ignored.go", "package ignored\n")
	writeCandidateFile(t, repository, "vendor/ignored/ignored.go", "package ignored\n")
	runGit(t, repository, "add", ".")
	commitTime := "2026-08-12T01:58:50Z"
	command := exec.Command("git", "commit", "-q", "-m", "candidate")
	command.Dir = repository
	command.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+commitTime, "GIT_COMMITTER_DATE="+commitTime)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("commit fixture: %v\n%s", err, output)
	}
	commit := runGit(t, repository, "rev-parse", "HEAD")
	tree := runGit(t, repository, "rev-parse", "HEAD^{tree}")

	first := buildCandidate(t, repository, commit, tree, filepath.Join(t.TempDir(), "proxy-one"))
	second := buildCandidate(t, repository, commit, tree, filepath.Join(t.TempDir(), "proxy-two"))
	wantVersion := "v0.0.0-20260812015850-" + commit[:12]
	if first.Version != wantVersion {
		t.Fatalf("Version = %q, want %q", first.Version, wantVersion)
	}
	if first.ModulePath != "example.com/candidate" {
		t.Fatalf("ModulePath = %q", first.ModulePath)
	}
	for _, artifact := range []struct {
		name        string
		firstPath   string
		secondPath  string
		wantContent []byte
	}{
		{"info", first.InfoPath, second.InfoPath, nil},
		{"mod", first.ModPath, second.ModPath, []byte("module example.com/candidate\n\ngo 1.26.5\n")},
		{"zip", first.ZipPath, second.ZipPath, nil},
		{"list", first.ListPath, second.ListPath, []byte(wantVersion + "\n")},
		{"manifest", first.ManifestPath, second.ManifestPath, nil},
	} {
		firstBytes := readCandidateFile(t, artifact.firstPath)
		secondBytes := readCandidateFile(t, artifact.secondPath)
		if !bytes.Equal(firstBytes, secondBytes) {
			t.Fatalf("%s bytes are not deterministic", artifact.name)
		}
		if artifact.wantContent != nil && !bytes.Equal(firstBytes, artifact.wantContent) {
			t.Fatalf("%s = %q, want %q", artifact.name, firstBytes, artifact.wantContent)
		}
	}
	var info struct {
		Version string    `json:"Version"`
		Time    time.Time `json:"Time"`
	}
	if err := json.Unmarshal(readCandidateFile(t, first.InfoPath), &info); err != nil {
		t.Fatal(err)
	}
	if info.Version != wantVersion || !info.Time.Equal(time.Date(2026, 8, 12, 1, 58, 50, 0, time.UTC)) {
		t.Fatalf("info = %+v", info)
	}
	archive, err := zip.OpenReader(first.ZipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	prefix := "example.com/candidate@" + wantVersion + "/"
	wantFiles := map[string]string{
		prefix + "go.mod":       "module example.com/candidate\n\ngo 1.26.5\n",
		prefix + "candidate.go": "package candidate\n\nconst Identity = \"exact\"\n",
	}
	if len(archive.File) != len(wantFiles) {
		t.Fatalf("zip contains %d files, want %d", len(archive.File), len(wantFiles))
	}
	for _, file := range archive.File {
		want, ok := wantFiles[file.Name]
		if !ok {
			t.Fatalf("unexpected zip member %q", file.Name)
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil || string(contents) != want {
			t.Fatalf("zip member %s = %q, err=%v", file.Name, contents, err)
		}
	}
}

func TestMirrorDependencyRejectsMissingAndTamperedArtifacts(t *testing.T) {
	modulePath, version := "example.com/dependency", "v1.2.3"
	goMod := []byte("module " + modulePath + "\n\ngo 1.26.5\n")
	moduleZip := candidateDependencyZip(t, modulePath, version, map[string]string{"dependency.go": "package dependency\n"})
	sums := map[string]string{}
	var err error
	sums[modulePath+" "+version+"/go.mod"], err = moduleH1(goMod, false)
	if err != nil {
		t.Fatal(err)
	}
	sums[modulePath+" "+version], err = moduleH1(moduleZip, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, missing := range []string{"info", "mod", "zip"} {
		t.Run("missing-"+missing, func(t *testing.T) {
			cache := t.TempDir()
			writeDependencyCache(t, cache, modulePath, version, goMod, moduleZip, missing)
			_, err := mirrorDependency(t.TempDir(), cache, modulePath, version, sums)
			if err == nil || !strings.Contains(err.Error(), "missing dependency "+missing+" artifact") {
				t.Fatalf("mirrorDependency() error = %v, want missing %s", err, missing)
			}
		})
	}

	for _, tampered := range []string{"mod", "zip"} {
		t.Run("tampered-"+tampered, func(t *testing.T) {
			cache := t.TempDir()
			writeDependencyCache(t, cache, modulePath, version, goMod, moduleZip, "")
			artifact := dependencyCacheArtifact(cache, modulePath, version, tampered)
			if err := os.WriteFile(artifact, []byte("tampered\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := mirrorDependency(t.TempDir(), cache, modulePath, version, sums)
			if err == nil || !strings.Contains(err.Error(), "tampered dependency") {
				t.Fatalf("mirrorDependency() error = %v, want tamper rejection", err)
			}
		})
	}
}

func TestBuildMirrorsOnlyAuthenticatedFileProxyDependencyGraph(t *testing.T) {
	modulePath, version := "example.com/dependency", "v1.2.3"
	goMod := []byte("module " + modulePath + "\n\ngo 1.26.5\n")
	moduleZip := candidateDependencyZip(t, modulePath, version, map[string]string{"dependency.go": "package dependency\n\nconst Value = \"fixture\"\n"})
	modSum, err := moduleH1(goMod, false)
	if err != nil {
		t.Fatal(err)
	}
	zipSum, err := moduleH1(moduleZip, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name   string
		mutate func(*testing.T, string)
		pass   bool
	}{
		{name: "authenticated", pass: true},
		{name: "missing-zip", mutate: func(t *testing.T, proxy string) {
			requireRemove(t, dependencyProxyArtifact(proxy, modulePath, version, "zip"))
		}},
		{name: "tampered-zip", mutate: func(t *testing.T, proxy string) {
			requireWrite(t, dependencyProxyArtifact(proxy, modulePath, version, "zip"), []byte("tampered\n"))
		}},
		{name: "tampered-mod", mutate: func(t *testing.T, proxy string) {
			requireWrite(t, dependencyProxyArtifact(proxy, modulePath, version, "mod"), []byte("module example.com/tampered\n"))
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			proxy := t.TempDir()
			writeDependencyProxy(t, proxy, modulePath, version, goMod, moduleZip)
			if test.mutate != nil {
				test.mutate(t, proxy)
			}
			repository := t.TempDir()
			runGit(t, repository, "init", "-q")
			runGit(t, repository, "config", "user.name", "Module Candidate Test")
			runGit(t, repository, "config", "user.email", "module-candidate@example.invalid")
			writeCandidateFile(t, repository, "go.mod", "module example.com/candidate\n\ngo 1.26.5\n\nrequire "+modulePath+" "+version+"\n")
			writeCandidateFile(t, repository, "go.sum", modulePath+" "+version+" "+zipSum+"\n"+modulePath+" "+version+"/go.mod "+modSum+"\n")
			writeCandidateFile(t, repository, "candidate.go", "package candidate\n\nimport _ \"example.com/dependency\"\n")
			runGit(t, repository, "add", ".")
			runGit(t, repository, "commit", "-q", "-m", "candidate")
			commit := runGit(t, repository, "rev-parse", "HEAD")
			tree := runGit(t, repository, "rev-parse", "HEAD^{tree}")
			output := filepath.Join(t.TempDir(), "output")
			proxyURL := (&url.URL{Scheme: "file", Path: proxy}).String()
			result, err := Build(t.Context(), Config{Repository: repository, ModulePath: "example.com/candidate", Commit: commit, ExpectedTree: tree, Output: output, DependencyProxy: proxyURL})
			if !test.pass {
				if err == nil {
					t.Fatal("Build() accepted missing or tampered dependency proxy artifact")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.ManifestPath == "" {
				t.Fatal("Build() did not write deterministic manifest")
			}
			for kind, want := range map[string][]byte{"mod": goMod, "zip": moduleZip} {
				got := readCandidateFile(t, filepath.Join(output, filepath.FromSlash(modulePath), "@v", version+"."+kind))
				if !bytes.Equal(got, want) {
					t.Fatalf("mirrored %s differs from authenticated proxy bytes", kind)
				}
			}
			var dependencyInfo struct{ Version string }
			if err := json.Unmarshal(readCandidateFile(t, filepath.Join(output, filepath.FromSlash(modulePath), "@v", version+".info")), &dependencyInfo); err != nil || dependencyInfo.Version != version {
				t.Fatalf("mirrored info = %+v, err=%v", dependencyInfo, err)
			}
		})
	}
}

func candidateDependencyZip(t *testing.T, modulePath, version string, files map[string]string) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for name, contents := range files {
		member, err := writer.Create(modulePath + "@" + version + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(member, contents); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func writeDependencyCache(t *testing.T, cache, modulePath, version string, goMod, moduleZip []byte, missing string) {
	t.Helper()
	artifacts := map[string][]byte{
		"info": []byte(`{"Version":"` + version + `","Time":"2026-08-12T00:00:00Z"}` + "\n"),
		"mod":  goMod,
		"zip":  moduleZip,
	}
	for kind, contents := range artifacts {
		if kind == missing {
			continue
		}
		filename := dependencyCacheArtifact(cache, modulePath, version, kind)
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeDependencyProxy(t *testing.T, proxy, modulePath, version string, goMod, moduleZip []byte) {
	t.Helper()
	for kind, contents := range map[string][]byte{
		"info": []byte(`{"Version":"` + version + `","Time":"2026-08-12T00:00:00Z"}` + "\n"),
		"mod":  goMod,
		"zip":  moduleZip,
	} {
		requireWrite(t, dependencyProxyArtifact(proxy, modulePath, version, kind), contents)
	}
	requireWrite(t, filepath.Join(proxy, filepath.FromSlash(modulePath), "@v", "list"), []byte(version+"\n"))
}

func dependencyProxyArtifact(proxy, modulePath, version, kind string) string {
	return filepath.Join(proxy, filepath.FromSlash(escapeModulePath(modulePath)), "@v", escapeModulePath(version)+"."+kind)
}

func requireWrite(t *testing.T, filename string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}

func requireRemove(t *testing.T, filename string) {
	t.Helper()
	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}
}

func dependencyCacheArtifact(cache, modulePath, version, kind string) string {
	return filepath.Join(cache, "cache", "download", filepath.FromSlash(escapeModulePath(modulePath)), "@v", escapeModulePath(version)+"."+kind)
}

func TestBuildRejectsSubdirBeforeWritingOutput(t *testing.T) {
	repository := t.TempDir()
	runGit(t, repository, "init", "-q")
	runGit(t, repository, "config", "user.name", "Module Candidate Test")
	runGit(t, repository, "config", "user.email", "module-candidate@example.invalid")
	runGit(t, repository, "commit", "--allow-empty", "-q", "-m", "candidate")
	commit := runGit(t, repository, "rev-parse", "HEAD")
	tree := runGit(t, repository, "rev-parse", "HEAD^{tree}")
	output := filepath.Join(t.TempDir(), "proxy")
	_, err := Build(t.Context(), Config{Repository: repository, ModulePath: "example.com/candidate", Commit: commit, ExpectedTree: tree, Subdir: "site", Output: output})
	if err == nil || !strings.Contains(err.Error(), "subdir must be empty") {
		t.Fatalf("Build() error = %v, want subdir rejection", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("output exists after rejected subdir: %v", err)
	}
}

func buildCandidate(t *testing.T, repository, commit, tree, output string) Result {
	t.Helper()
	result, err := Build(t.Context(), Config{Repository: repository, ModulePath: "example.com/candidate", Commit: commit, ExpectedTree: tree, Output: output})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func writeCandidateFile(t *testing.T, root, relative, contents string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readCandidateFile(t *testing.T, filename string) []byte {
	t.Helper()
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func runGit(t *testing.T, repository string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = repository
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}
