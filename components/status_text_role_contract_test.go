package components_test

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var rawStatusTextRole = regexp.MustCompile(`(?:[[:alnum:]_\[\]=/-]+:)*text-(?:info|success|warning|danger)(?:/[[:digit:]]+)?`)

type rawStatusTextUse struct {
	path string
	line string
	role string
}

func (use rawStatusTextUse) key() string {
	return use.path + "\x00" + use.line + "\x00" + use.role
}

// TestStatusTextUsesReadableSemanticRoles rejects base status colors on text.
// Base status roles remain valid only for these exact currentColor icon uses.
func TestStatusTextUsesReadableSemanticRoles(t *testing.T) {
	t.Parallel()

	allowedIcons := map[string]int{
		(rawStatusTextUse{path: "toast/toast.templ", line: `{{ iconBgClass := "bg-info/15 text-info" }}`, role: "text-info"}).key():                                                                                                      1,
		(rawStatusTextUse{path: "toast/toast.templ", line: `{{ iconBgClass = "bg-success/15 text-success" }}`, role: "text-success"}).key():                                                                                              1,
		(rawStatusTextUse{path: "toast/toast.templ", line: `{{ iconBgClass = "bg-warning/15 text-warning" }}`, role: "text-warning"}).key():                                                                                              1,
		(rawStatusTextUse{path: "toast/toast.templ", line: `{{ iconBgClass = "bg-danger/15 text-danger" }}`, role: "text-danger"}).key():                                                                                                 1,
		(rawStatusTextUse{path: "toast/types.go", line: `return "bg-info/15 text-info"`, role: "text-info"}).key():                                                                                                                       2,
		(rawStatusTextUse{path: "toast/types.go", line: `return "bg-success/15 text-success"`, role: "text-success"}).key():                                                                                                              1,
		(rawStatusTextUse{path: "toast/types.go", line: `return "bg-warning/15 text-warning"`, role: "text-warning"}).key():                                                                                                              1,
		(rawStatusTextUse{path: "toast/types.go", line: `return "bg-danger/15 text-danger"`, role: "text-danger"}).key():                                                                                                                 1,
		(rawStatusTextUse{path: "form/errors.templ", line: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-5 shrink-0 text-danger mt-0.5" aria-hidden="true">`, role: "text-danger"}).key(): 1,
	}

	families := []string{"toast", "banner", "schemaform", "radio", "dropdown", "navbar", "form"}
	seenAllowed := make(map[string]int, len(allowedIcons))
	var violations []string
	for _, family := range families {
		err := filepath.WalkDir(family, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !isAuthoredComponentSource(path) {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for lineNumber := 1; scanner.Scan(); lineNumber++ {
				line := strings.TrimSpace(scanner.Text())
				for _, match := range rawStatusTextRole.FindAllStringIndex(line, -1) {
					if hasRoleSuffix(line, match[1]) {
						continue
					}
					use := rawStatusTextUse{path: filepath.ToSlash(path), line: line, role: line[match[0]:match[1]]}
					if allowedIcons[use.key()] > seenAllowed[use.key()] {
						seenAllowed[use.key()]++
						continue
					}
					violations = append(violations, fmt.Sprintf("%s:%d: %s in %s", path, lineNumber, use.role, line))
				}
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s status roles: %v", family, err)
		}
	}

	for allowed, want := range allowedIcons {
		if got := seenAllowed[allowed]; got != want {
			violations = append(violations, fmt.Sprintf("icon allowlist mismatch: got %d, want %d for %q", got, want, allowed))
		}
	}
	sort.Strings(violations)
	if len(violations) > 0 {
		t.Fatalf("base status text roles require semantic *-text light/dark pairs; only exact currentColor icons are allowed:\n%s", strings.Join(violations, "\n"))
	}
}

func isAuthoredComponentSource(path string) bool {
	if strings.HasSuffix(path, ".templ") {
		return true
	}
	return strings.HasSuffix(path, ".go") &&
		!strings.HasSuffix(path, "_templ.go") &&
		!strings.HasSuffix(path, "_test.go")
}

func hasRoleSuffix(line string, matchEnd int) bool {
	if matchEnd >= len(line) {
		return false
	}
	next := line[matchEnd]
	return next == '-' || next == '_' || next >= '0' && next <= '9' || next >= 'A' && next <= 'Z' || next >= 'a' && next <= 'z'
}
