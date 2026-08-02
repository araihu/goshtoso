package e2eimpact

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

const zeroSHA = "0000000000000000000000000000000000000000"

// Change is one NUL-safe Git name-status record.
type Change struct {
	Status  string
	OldPath string
	NewPath string
}

func gitChanges(ctx context.Context, repoRoot, base, head string) ([]Change, error) {
	if base == "" || head == "" {
		return nil, fmt.Errorf("base and head revisions are required")
	}
	if base == zeroSHA {
		return nil, fmt.Errorf("base revision is the all-zero first-push SHA")
	}
	for _, revision := range []string{base, head} {
		command := exec.CommandContext(ctx, "git", "cat-file", "-e", revision+"^{commit}")
		command.Dir = repoRoot
		if output, err := command.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("revision %s is unavailable: %s", revision, strings.TrimSpace(string(output)))
		}
	}
	ancestor := exec.CommandContext(ctx, "git", "merge-base", "--is-ancestor", base, head)
	ancestor.Dir = repoRoot
	if err := ancestor.Run(); err != nil {
		return nil, fmt.Errorf("base %s is not an ancestor of head %s", base, head)
	}
	command := exec.CommandContext(ctx, "git", "diff", "--name-status", "-z", "-M", base, head)
	command.Dir = repoRoot
	data, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	return parseNameStatus(data)
}

func parseNameStatus(data []byte) ([]Change, error) {
	fields := bytes.Split(data, []byte{0})
	if len(fields) > 0 && len(fields[len(fields)-1]) == 0 {
		fields = fields[:len(fields)-1]
	}
	var changes []Change
	for index := 0; index < len(fields); {
		status := string(fields[index])
		index++
		if status == "" {
			return nil, fmt.Errorf("empty Git status record")
		}
		kind := status[0]
		if kind == 'R' || kind == 'C' {
			if index+1 >= len(fields) {
				return nil, fmt.Errorf("truncated Git %s record", status)
			}
			changes = append(changes, Change{Status: status, OldPath: string(fields[index]), NewPath: string(fields[index+1])})
			index += 2
			continue
		}
		if index >= len(fields) {
			return nil, fmt.Errorf("truncated Git %s record", status)
		}
		path := string(fields[index])
		index++
		changes = append(changes, Change{Status: status, OldPath: path, NewPath: path})
	}
	slices.SortFunc(changes, func(a, b Change) int {
		if byNew := strings.Compare(a.NewPath, b.NewPath); byNew != 0 {
			return byNew
		}
		return strings.Compare(a.OldPath, b.OldPath)
	})
	return changes, nil
}

func changedPaths(changes []Change) []string {
	var paths []string
	for _, change := range changes {
		paths = append(paths, change.OldPath)
		if change.NewPath != change.OldPath {
			paths = append(paths, change.NewPath)
		}
	}
	slices.Sort(paths)
	return slices.Compact(paths)
}
