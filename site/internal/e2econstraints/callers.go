package e2econstraints

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// BareE2ECommand identifies an active command that would silently stop finding
// the E2E package after the suite build gate is applied.
type BareE2ECommand struct {
	Path string
	Line int
	Text string
}

// FindBareE2ECommands scans executable and current contributor-facing files.
func FindBareE2ECommands(repoRoot string) ([]BareE2ECommand, error) {
	var findings []BareE2ECommand
	err := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if ignoredCommandDirectory(relative) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if !activeCommandFile(relative) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		scanner := bufio.NewScanner(file)
		for lineNumber := 1; scanner.Scan(); lineNumber++ {
			line := scanner.Text()
			if !strings.Contains(line, "go test") || !strings.Contains(line, "tests/e2e") ||
				strings.Contains(line, "grep -v /tests/e2e") {
				continue
			}
			if strings.Contains(line, "-tags") && strings.Contains(line, "e2e,") {
				continue
			}
			findings = append(findings, BareE2ECommand{Path: relative, Line: lineNumber, Text: strings.TrimSpace(line)})
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan %s: %w", relative, err)
		}
		return nil
	})
	slices.SortFunc(findings, func(a, b BareE2ECommand) int {
		if byPath := strings.Compare(a.Path, b.Path); byPath != 0 {
			return byPath
		}
		return a.Line - b.Line
	})
	return findings, err
}

func ignoredCommandDirectory(path string) bool {
	path = filepath.ToSlash(path)
	return path == ".git" || path == "docs/plans" || path == "docs/designs" ||
		path == "docs/audits" || strings.HasPrefix(path, ".git/") ||
		strings.HasPrefix(path, "docs/plans/") || strings.HasPrefix(path, "docs/designs/") ||
		strings.HasPrefix(path, "docs/audits/")
}

func activeCommandFile(path string) bool {
	base := filepath.Base(path)
	extension := filepath.Ext(path)
	return base == "Makefile" || strings.EqualFold(base, "justfile") || base == "Dockerfile" ||
		extension == ".md" || extension == ".yml" || extension == ".yaml" || extension == ".sh"
}
