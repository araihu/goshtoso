package modulecandidate

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Config pins a module candidate to one repository commit and tree.
type Config struct {
	Repository      string
	ModulePath      string
	Commit          string
	ExpectedTree    string
	Subdir          string
	Output          string
	DependencyProxy string
}

// Result identifies the exact source object accepted for a candidate build.
type Result struct {
	ModulePath   string `json:"modulePath"`
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	Tree         string `json:"tree"`
	InfoPath     string `json:"infoPath"`
	ModPath      string `json:"modPath"`
	ZipPath      string `json:"zipPath"`
	ListPath     string `json:"listPath"`
	ManifestPath string `json:"manifestPath"`
}

type manifest struct {
	SchemaVersion int                `json:"schemaVersion"`
	ModulePath    string             `json:"modulePath"`
	Version       string             `json:"version"`
	Commit        string             `json:"commit"`
	Tree          string             `json:"tree"`
	Artifacts     []manifestArtifact `json:"artifacts"`
}

type manifestArtifact struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	SHA256  string `json:"sha256"`
}

type graphModule struct {
	Path    string
	Version string
	Main    bool
	Replace *graphModule
}

// Build validates the exact candidate identity before any proxy materialization.
func Build(ctx context.Context, config Config) (Result, error) {
	if config.Subdir != "" {
		return Result{}, fmt.Errorf("subdir must be empty")
	}
	commit, err := gitObject(ctx, config.Repository, config.Commit+"^{commit}")
	if err != nil {
		return Result{}, fmt.Errorf("resolve candidate commit %q: %w", config.Commit, err)
	}
	if commit != config.Commit {
		return Result{}, fmt.Errorf("candidate commit mismatch: got %s, want %s", commit, config.Commit)
	}
	tree, err := gitObject(ctx, config.Repository, commit+"^{tree}")
	if err != nil {
		return Result{}, fmt.Errorf("resolve candidate tree for %s: %w", commit, err)
	}
	if config.ExpectedTree == "" {
		return Result{}, fmt.Errorf("expected candidate tree is required")
	}
	if tree != config.ExpectedTree {
		return Result{}, fmt.Errorf("candidate tree mismatch: got %s, want %s", tree, config.ExpectedTree)
	}
	commitTime, err := gitCommitTime(ctx, config.Repository, commit)
	if err != nil {
		return Result{}, err
	}
	version := "v0.0.0-" + commitTime.UTC().Format("20060102150405") + "-" + commit[:12]
	modBytes, err := gitBlob(ctx, config.Repository, commit+":go.mod")
	if err != nil {
		return Result{}, fmt.Errorf("read committed go.mod: %w", err)
	}
	modulePath, err := parseModulePath(modBytes)
	if err != nil {
		return Result{}, err
	}
	if config.ModulePath != modulePath {
		return Result{}, fmt.Errorf("candidate module path mismatch: got %s, want %s", modulePath, config.ModulePath)
	}
	files, err := candidateFiles(ctx, config.Repository, commit)
	if err != nil {
		return Result{}, err
	}
	zipBytes, err := moduleZip(ctx, config.Repository, modulePath, version, commitTime, files)
	if err != nil {
		return Result{}, err
	}
	infoBytes, err := json.Marshal(struct {
		Version string
		Time    time.Time
	}{Version: version, Time: commitTime.UTC()})
	if err != nil {
		return Result{}, fmt.Errorf("encode candidate info: %w", err)
	}
	infoBytes = append(infoBytes, '\n')
	result := Result{ModulePath: modulePath, Version: version, Commit: commit, Tree: tree}
	escaped := escapeModulePath(modulePath)
	versionDir := filepath.Join(config.Output, filepath.FromSlash(escaped), "@v")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create candidate proxy directory: %w", err)
	}
	result.InfoPath = filepath.Join(versionDir, version+".info")
	result.ModPath = filepath.Join(versionDir, version+".mod")
	result.ZipPath = filepath.Join(versionDir, version+".zip")
	result.ListPath = filepath.Join(versionDir, "list")
	candidateArtifacts := []struct {
		path string
		data []byte
	}{{result.InfoPath, infoBytes}, {result.ModPath, modBytes}, {result.ZipPath, zipBytes}, {result.ListPath, []byte(version + "\n")}}
	for _, artifact := range candidateArtifacts {
		if err := os.WriteFile(artifact.path, artifact.data, 0o644); err != nil {
			return Result{}, fmt.Errorf("write candidate artifact %s: %w", filepath.Base(artifact.path), err)
		}
	}
	manifestValue := manifest{SchemaVersion: 1, ModulePath: modulePath, Version: version, Commit: commit, Tree: tree}
	for _, artifact := range candidateArtifacts[:3] {
		manifestValue.Artifacts = append(manifestValue.Artifacts, manifestArtifact{Module: modulePath, Version: version, Kind: strings.TrimPrefix(filepath.Ext(artifact.path), "."), SHA256: sha256Bytes(artifact.data)})
	}
	goSumBytes, err := gitBlob(ctx, config.Repository, commit+":go.sum")
	if err != nil {
		goSumBytes = nil
	}
	graph, graphCache, graphCleanup, err := committedModuleGraph(ctx, modBytes, goSumBytes, config.DependencyProxy)
	if err != nil {
		return Result{}, err
	}
	defer graphCleanup()
	sums := parseGoSums(goSumBytes)
	for _, module := range graph {
		if module.Main {
			continue
		}
		if module.Replace != nil {
			return Result{}, fmt.Errorf("module graph replace is forbidden: %s => %s", module.Path, module.Replace.Path)
		}
		artifacts, err := mirrorDependency(config.Output, graphCache, module.Path, module.Version, sums)
		if err != nil {
			return Result{}, err
		}
		manifestValue.Artifacts = append(manifestValue.Artifacts, artifacts...)
	}
	sort.Slice(manifestValue.Artifacts, func(i, j int) bool {
		left, right := manifestValue.Artifacts[i], manifestValue.Artifacts[j]
		if left.Module != right.Module {
			return left.Module < right.Module
		}
		if left.Version != right.Version {
			return left.Version < right.Version
		}
		return left.Kind < right.Kind
	})
	manifestBytes, err := json.MarshalIndent(manifestValue, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("encode candidate manifest: %w", err)
	}
	manifestBytes = append(manifestBytes, '\n')
	result.ManifestPath = filepath.Join(config.Output, "module-candidate-manifest.json")
	if err := os.WriteFile(result.ManifestPath, manifestBytes, 0o644); err != nil {
		return Result{}, fmt.Errorf("write candidate manifest: %w", err)
	}
	return result, nil
}

func gitObject(ctx context.Context, repository, object string) (string, error) {
	if repository == "" || object == "^{commit}" {
		return "", fmt.Errorf("repository and commit are required")
	}
	command := exec.CommandContext(ctx, "git", "-C", repository, "rev-parse", "--verify", object)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func gitCommitTime(ctx context.Context, repository, commit string) (time.Time, error) {
	command := exec.CommandContext(ctx, "git", "-C", repository, "show", "-s", "--format=%cI", commit)
	output, err := command.CombinedOutput()
	if err != nil {
		return time.Time{}, fmt.Errorf("read candidate commit time: %w: %s", err, strings.TrimSpace(string(output)))
	}
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(string(output)))
	if err != nil {
		return time.Time{}, fmt.Errorf("parse candidate commit time: %w", err)
	}
	return value, nil
}

func gitBlob(ctx context.Context, repository, object string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", "-C", repository, "show", object)
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func parseModulePath(goMod []byte) (string, error) {
	for _, line := range strings.Split(string(goMod), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("committed go.mod has no module directive")
}

type candidateFile struct {
	path   string
	object string
}

func candidateFiles(ctx context.Context, repository, commit string) ([]candidateFile, error) {
	command := exec.CommandContext(ctx, "git", "-C", repository, "ls-tree", "-r", "-z", "--full-tree", commit)
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list candidate tree: %w", err)
	}
	var files []candidateFile
	var nested []string
	for _, entry := range bytes.Split(output, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		metadata, name, ok := bytes.Cut(entry, []byte{'\t'})
		fields := bytes.Fields(metadata)
		if !ok || len(fields) != 3 || string(fields[1]) != "blob" {
			continue
		}
		mode, filename := string(fields[0]), string(name)
		if mode != "100644" && mode != "100755" {
			continue
		}
		if filename != "go.mod" && strings.HasSuffix(filename, "/go.mod") {
			nested = append(nested, strings.TrimSuffix(filename, "go.mod"))
		}
		files = append(files, candidateFile{path: filename, object: string(fields[2])})
	}
	filtered := files[:0]
	for _, file := range files {
		if file.path == "vendor" || strings.HasPrefix(file.path, "vendor/") {
			continue
		}
		nestedFile := false
		for _, prefix := range nested {
			if strings.HasPrefix(file.path, prefix) {
				nestedFile = true
				break
			}
		}
		if !nestedFile {
			filtered = append(filtered, file)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].path < filtered[j].path })
	return filtered, nil
}

func moduleZip(ctx context.Context, repository, modulePath, version string, commitTime time.Time, files []candidateFile) ([]byte, error) {
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	prefix := modulePath + "@" + version + "/"
	for _, file := range files {
		contents, err := gitBlob(ctx, repository, file.object)
		if err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("read candidate blob %s: %w", file.path, err)
		}
		header := &zip.FileHeader{Name: prefix + file.path, Method: zip.Deflate}
		header.SetMode(0o644)
		header.SetModTime(commitTime.UTC())
		member, err := writer.CreateHeader(header)
		if err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("create candidate zip member %s: %w", file.path, err)
		}
		if _, err := io.Copy(member, bytes.NewReader(contents)); err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("write candidate zip member %s: %w", file.path, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close candidate module zip: %w", err)
	}
	return output.Bytes(), nil
}

func escapeModulePath(modulePath string) string {
	var output strings.Builder
	for _, character := range modulePath {
		if character >= 'A' && character <= 'Z' {
			output.WriteByte('!')
			output.WriteRune(character + ('a' - 'A'))
		} else {
			output.WriteRune(character)
		}
	}
	return output.String()
}

func committedModuleGraph(ctx context.Context, goMod, goSum []byte, dependencyProxy string) ([]graphModule, string, func(), error) {
	directory, err := os.MkdirTemp("", "goshtoso-module-candidate-graph-*")
	if err != nil {
		return nil, "", func() {}, fmt.Errorf("create graph workspace: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(directory) }
	if err := os.WriteFile(filepath.Join(directory, "go.mod"), goMod, 0o600); err != nil {
		cleanup()
		return nil, "", func() {}, fmt.Errorf("write graph go.mod: %w", err)
	}
	if len(goSum) > 0 {
		if err := os.WriteFile(filepath.Join(directory, "go.sum"), goSum, 0o600); err != nil {
			cleanup()
			return nil, "", func() {}, fmt.Errorf("write graph go.sum: %w", err)
		}
	}
	moduleCache := filepath.Join(directory, "modcache")
	buildCache := filepath.Join(directory, "buildcache")
	proxy := "off"
	if dependencyProxy != "" {
		if !strings.HasPrefix(dependencyProxy, "file://") {
			cleanup()
			return nil, "", func() {}, fmt.Errorf("dependency proxy must be an explicit file:// URL")
		}
		proxy = dependencyProxy
	}
	command := exec.CommandContext(ctx, "go", "list", "-mod=readonly", "-m", "-json", "all")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off", "GOPROXY="+proxy, "GOSUMDB=off", "GONOSUMDB=*", "GOMODCACHE="+moduleCache, "GOCACHE="+buildCache)
	output, err := command.Output()
	if err != nil {
		cleanup()
		var exitErr *exec.ExitError
		if ok := errorAs(err, &exitErr); ok {
			return nil, "", func() {}, fmt.Errorf("enumerate committed module graph offline: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, "", func() {}, fmt.Errorf("enumerate committed module graph offline: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	var graph []graphModule
	for decoder.More() {
		var module graphModule
		if err := decoder.Decode(&module); err != nil {
			cleanup()
			return nil, "", func() {}, fmt.Errorf("decode committed module graph: %w", err)
		}
		graph = append(graph, module)
	}
	for _, module := range graph {
		if module.Main {
			continue
		}
		download := exec.CommandContext(ctx, "go", "mod", "download", module.Path+"@"+module.Version)
		download.Dir = directory
		download.Env = command.Env
		if output, err := download.CombinedOutput(); err != nil {
			cleanup()
			return nil, "", func() {}, fmt.Errorf("download resolved dependency %s@%s from file proxy: %w: %s", module.Path, module.Version, err, strings.TrimSpace(string(output)))
		}
	}
	return graph, moduleCache, cleanup, nil
}

func mirrorDependency(output, moduleCache, modulePath, version string, sums map[string]string) ([]manifestArtifact, error) {
	if moduleCache == "" {
		return nil, fmt.Errorf("module cache is required for dependency %s@%s", modulePath, version)
	}
	escapedModule := escapeModulePath(modulePath)
	escapedVersion := escapeModulePath(version)
	sourceDir := filepath.Join(moduleCache, "cache", "download", filepath.FromSlash(escapedModule), "@v")
	destinationDir := filepath.Join(output, filepath.FromSlash(escapedModule), "@v")
	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return nil, fmt.Errorf("create dependency proxy directory: %w", err)
	}
	var result []manifestArtifact
	for _, kind := range []string{"info", "mod", "zip"} {
		source := filepath.Join(sourceDir, escapedVersion+"."+kind)
		contents, err := os.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("missing dependency %s artifact %s: %w", kind, modulePath+"@"+version, err)
		}
		if kind == "mod" || kind == "zip" {
			key := modulePath + " " + version
			if kind == "mod" {
				key += "/go.mod"
			}
			want := sums[key]
			if want == "" {
				return nil, fmt.Errorf("missing committed go.sum authentication for %s", key)
			}
			got, err := moduleH1(contents, kind == "zip")
			if err != nil {
				return nil, fmt.Errorf("tampered dependency %s: hash artifact: %w", key, err)
			}
			if got != want {
				return nil, fmt.Errorf("tampered dependency %s: got %s, want %s", key, got, want)
			}
		}
		destination := filepath.Join(destinationDir, escapedVersion+"."+kind)
		if err := os.WriteFile(destination, contents, 0o644); err != nil {
			return nil, fmt.Errorf("write dependency artifact %s: %w", destination, err)
		}
		result = append(result, manifestArtifact{Module: modulePath, Version: version, Kind: kind, SHA256: sha256Bytes(contents)})
	}
	list := filepath.Join(destinationDir, "list")
	existing, _ := os.ReadFile(list)
	versions := strings.Fields(string(existing))
	versions = append(versions, version)
	sort.Strings(versions)
	versions = compactStrings(versions)
	if err := os.WriteFile(list, []byte(strings.Join(versions, "\n")+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("write dependency version list: %w", err)
	}
	return result, nil
}

func parseGoSums(contents []byte) map[string]string {
	result := map[string]string{}
	for _, line := range strings.Split(string(contents), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 {
			result[fields[0]+" "+fields[1]] = fields[2]
		}
	}
	return result
}

func moduleH1(contents []byte, archive bool) (string, error) {
	files := map[string][]byte{"go.mod": contents}
	if archive {
		reader, err := zip.NewReader(bytes.NewReader(contents), int64(len(contents)))
		if err != nil {
			return "", err
		}
		files = map[string][]byte{}
		for _, file := range reader.File {
			stream, err := file.Open()
			if err != nil {
				return "", err
			}
			data, err := io.ReadAll(stream)
			_ = stream.Close()
			if err != nil {
				return "", err
			}
			files[file.Name] = data
		}
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		fileHash := sha256.Sum256(files[name])
		_, _ = fmt.Fprintf(hash, "%s  %s\n", hex.EncodeToString(fileHash[:]), name)
	}
	return "h1:" + base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

func sha256Bytes(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func errorAs(err error, target any) bool {
	switch target := target.(type) {
	case **exec.ExitError:
		value, ok := err.(*exec.ExitError)
		if ok {
			*target = value
		}
		return ok
	default:
		return false
	}
}
