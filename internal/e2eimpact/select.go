package e2eimpact

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// SelectNameStatus calculates the safe E2E scope from a NUL-delimited
// `git diff --name-status -z -M` stream supplied by a trusted Git boundary.
// Keeping Git metadata outside the source Directory makes Dagger calls work
// from both ordinary clones and linked worktrees.
func SelectNameStatus(ctx context.Context, repoRoot string, data []byte) Result {
	changes, err := parseNameStatus(data)
	if err != nil {
		return fullResult(nil, "invalid Git name-status stream: "+err.Error())
	}
	return selectChanges(ctx, repoRoot, changes)
}

type identityRecord struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

type identityManifest struct {
	Identities     []identityRecord `json:"identities"`
	FullOnly       identityRecord   `json:"full_only"`
	known          map[string]bool
	fileIdentities map[string][]string
}

func loadIdentityManifest(repoRoot string) (identityManifest, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "site/tests/e2e/identities.json"))
	if err != nil {
		return identityManifest{}, err
	}
	var manifest identityManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return identityManifest{}, err
	}
	manifest.known = make(map[string]bool, len(manifest.Identities))
	manifest.fileIdentities = make(map[string][]string)
	for _, identity := range manifest.Identities {
		if manifest.known[identity.Name] {
			return identityManifest{}, fmt.Errorf("duplicate identity %q", identity.Name)
		}
		manifest.known[identity.Name] = true
		for _, file := range identity.Files {
			manifest.fileIdentities[file] = append(manifest.fileIdentities[file], identity.Name)
		}
	}
	return manifest, nil
}

// Select calculates the safe E2E scope for base..head.
func Select(ctx context.Context, repoRoot, base, head string) Result {
	changes, err := gitChanges(ctx, repoRoot, base, head)
	if err != nil {
		return fullResult(nil, "unsafe Git range: "+err.Error())
	}
	return selectChanges(ctx, repoRoot, changes)
}

func selectChanges(ctx context.Context, repoRoot string, changes []Change) Result {
	paths := changedPaths(changes)
	absoluteRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fullResult(paths, "repository root unavailable: "+err.Error())
	}
	repoRoot = absoluteRoot
	manifest, err := loadIdentityManifest(repoRoot)
	if err != nil {
		return fullResult(paths, "identity manifest unavailable: "+err.Error())
	}
	if err := validateHandlerOwnership(repoRoot, manifest.known); err != nil {
		return fullResult(paths, "handler ownership invalid: "+err.Error())
	}
	graph, err := loadPackageGraph(ctx, repoRoot, manifest.known)
	if err != nil {
		return fullResult(paths, "Go package graph unavailable: "+err.Error())
	}
	classified := classifyChanges(changes, graph, manifest)
	if len(classified.full) > 0 {
		return fullResult(paths, classified.full...)
	}
	reasons := graph.impactedIdentities(classified.roots)
	maps.Copy(reasons, classified.direct)
	if len(reasons) == 0 {
		return fullResult(paths, "no focused E2E identity selected")
	}
	tags := make([]string, 0, len(reasons))
	reasonList := make([]string, 0, len(reasons))
	for identity, reason := range reasons {
		tags = append(tags, identity)
		reasonList = append(reasonList, reason)
	}
	slices.Sort(tags)
	slices.Sort(reasonList)
	return Result{Mode: "focused", Tags: tags, ChangedPaths: paths, Reasons: reasonList}
}
