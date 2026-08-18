package e2eimpact

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type goPackage struct {
	ImportPath string
	Dir        string
	Imports    []string
}

type packageGraph struct {
	packages   map[string]goPackage
	reverse    map[string][]string
	identities map[string]string
	repoRoot   string
}

func loadPackageGraph(ctx context.Context, repoRoot string, known map[string]bool) (packageGraph, error) {
	command := exec.CommandContext(ctx, "go", "list", "-json", "./...", "./site/...")
	command.Dir = repoRoot
	output, err := command.StdoutPipe()
	if err != nil {
		return packageGraph{}, err
	}
	var stderr strings.Builder
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return packageGraph{}, err
	}
	graph := packageGraph{
		packages: make(map[string]goPackage), reverse: make(map[string][]string),
		identities: make(map[string]string), repoRoot: repoRoot,
	}
	decoder := json.NewDecoder(output)
	for {
		var pkg goPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return packageGraph{}, fmt.Errorf("decode go list: %w", err)
		}
		graph.packages[pkg.ImportPath] = pkg
	}
	if err := command.Wait(); err != nil {
		return packageGraph{}, fmt.Errorf("go list: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	for path, pkg := range graph.packages {
		for _, imported := range pkg.Imports {
			graph.reverse[imported] = append(graph.reverse[imported], path)
		}
		if identity, ok := packageIdentity(path); ok && known[identity] {
			graph.identities[path] = identity
		}
	}
	return graph, nil
}

func packageIdentity(importPath string) (string, bool) {
	const components = "/site/internal/pages/demo/componentpages/"
	if _, after, ok := strings.Cut(importPath, components); ok {
		remainder := after
		if remainder != "" && !strings.Contains(remainder, "/") {
			return remainder, true
		}
	}
	const examples = "/site/internal/pages/demo/examplepages/"
	if _, after, ok := strings.Cut(importPath, examples); ok {
		remainder := after
		if remainder != "" && remainder != "index" && !strings.Contains(remainder, "/") {
			return "example_" + remainder, true
		}
	}
	return "", false
}

func (graph packageGraph) packageForPath(path string) (string, bool) {
	absolute := filepath.Join(graph.repoRoot, filepath.FromSlash(path))
	best := ""
	bestLength := -1
	for importPath, pkg := range graph.packages {
		relative, err := filepath.Rel(pkg.Dir, absolute)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		if len(pkg.Dir) > bestLength {
			best = importPath
			bestLength = len(pkg.Dir)
		}
	}
	return best, best != ""
}

func (graph packageGraph) impactedIdentities(roots []string) map[string]string {
	reasons := make(map[string]string)
	queue := slices.Clone(roots)
	seen := make(map[string]bool)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		if identity, ok := graph.identities[current]; ok {
			reasons[identity] = fmt.Sprintf("%s: imports changed package", identity)
		}
		queue = append(queue, graph.reverse[current]...)
	}
	return reasons
}
