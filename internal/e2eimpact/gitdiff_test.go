package e2eimpact

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseNameStatusHandlesModifyAddDeleteAndRename(t *testing.T) {
	data := []byte("M\x00components/button/types.go\x00A\x00new file.go\x00D\x00old.go\x00R100\x00before.go\x00after.go\x00")
	changes, err := parseNameStatus(data)
	require.NoError(t, err)
	require.Equal(t, []Change{
		{Status: "R100", OldPath: "before.go", NewPath: "after.go"},
		{Status: "M", OldPath: "components/button/types.go", NewPath: "components/button/types.go"},
		{Status: "A", OldPath: "new file.go", NewPath: "new file.go"},
		{Status: "D", OldPath: "old.go", NewPath: "old.go"},
	}, changes)
}

func TestGitChangesCoversMultiCommitPushAndRejectsUnsafeRanges(t *testing.T) {
	repository := t.TempDir()
	runGit(t, repository, "init", "-q")
	runGit(t, repository, "config", "user.email", "tests@example.com")
	runGit(t, repository, "config", "user.name", "Tests")
	writeFixture(t, repository, "first.txt", "first")
	runGit(t, repository, "add", ".")
	runGit(t, repository, "commit", "-qm", "first")
	base := runGit(t, repository, "rev-parse", "HEAD")
	primaryBranch := runGit(t, repository, "branch", "--show-current")
	runGit(t, repository, "checkout", "-qb", "feature")
	writeFixture(t, repository, "feature.txt", "feature")
	runGit(t, repository, "add", ".")
	runGit(t, repository, "commit", "-qm", "feature")
	featureHead := runGit(t, repository, "rev-parse", "HEAD")
	runGit(t, repository, "checkout", "-q", primaryBranch)
	writeFixture(t, repository, "second.txt", "second")
	runGit(t, repository, "add", ".")
	runGit(t, repository, "commit", "-qm", "second")
	writeFixture(t, repository, "third.txt", "third")
	runGit(t, repository, "add", ".")
	runGit(t, repository, "commit", "-qm", "third")
	head := runGit(t, repository, "rev-parse", "HEAD")

	changes, err := gitChanges(context.Background(), repository, base, head)
	require.NoError(t, err)
	require.Equal(t, []string{"second.txt", "third.txt"}, changedPaths(changes))
	mergeBase := runGit(t, repository, "merge-base", head, featureHead)
	require.Equal(t, base, mergeBase, "a branch-behind PR caller must pass this merge base")
	changes, err = gitChanges(context.Background(), repository, mergeBase, featureHead)
	require.NoError(t, err)
	require.Equal(t, []string{"feature.txt"}, changedPaths(changes))
	_, err = gitChanges(context.Background(), repository, zeroSHA, head)
	require.ErrorContains(t, err, "all-zero")
	_, err = gitChanges(context.Background(), repository, "missing", head)
	require.ErrorContains(t, err, "unavailable")
	_, err = gitChanges(context.Background(), repository, head, base)
	require.ErrorContains(t, err, "not an ancestor")
}

func runGit(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	return string(bytes.TrimSpace(output))
}

func writeFixture(t *testing.T, directory, name, contents string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o600))
}
