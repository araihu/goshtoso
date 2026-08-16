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

var rawBackdropDisplayColor = regexp.MustCompile(`(?:[[:alnum:]_\[\]=/-]+:)*bg-(?:black|white)(?:/[[:digit:]]+)?|#000000`)
var authoredClassToken = regexp.MustCompile(`[[:alnum:]_:\[\]=/.-]+`)

func isAuthoredComponentSource(path string) bool {
	if strings.HasSuffix(path, "_templ.go") || strings.HasSuffix(path, "_test.go") {
		return false
	}
	return strings.HasSuffix(path, ".templ") || strings.HasSuffix(path, ".go")
}

type backdropElevationUse struct {
	path  string
	line  string
	token string
}

func (use backdropElevationUse) key() string {
	return use.path + "\x00" + use.line + "\x00" + use.token
}

// TestBackdropAndElevationUseGovernedRoles rejects the current raw backdrop
// and elevation primitives. Palette's three literal color displays are data,
// not component-surface styling, and are the only exact allowlist.
func TestBackdropAndElevationUseGovernedRoles(t *testing.T) {
	t.Parallel()

	allowedPaletteData := map[string]int{
		(backdropElevationUse{path: "palette/palette.templ", line: `placeholder="#000000"`, token: "#000000"}).key(): 1,
		(backdropElevationUse{path: "palette/palette.templ", line: `class="h-5 w-full rounded-sm border border-control-outline bg-white transition-transform hover:scale-125 hover:ring-2 hover:ring-primary focus:scale-125 dark:border-control-outline-dark dark:hover:ring-primary-dark"`, token: "bg-white"}).key(): 1,
		(backdropElevationUse{path: "palette/palette.templ", line: `class="h-5 w-full rounded-sm border border-control-outline bg-black transition-transform hover:scale-125 hover:ring-2 hover:ring-primary focus:scale-125 dark:border-control-outline-dark dark:hover:ring-primary-dark"`, token: "bg-black"}).key(): 1,
	}

	namedRaw := map[string]*regexp.Regexp{
		"drawer/types.go":   regexp.MustCompile(`bg-black/40|dark:bg-black/60|shadow-xl`),
		"modal/modal.templ": regexp.MustCompile(`bg-black/20`),
		"sidebar/types.go":  regexp.MustCompile(`bg-black/50`),
		"search/search.templ": regexp.MustCompile(
			`bg-surface-dark/55|dark:bg-black/60`,
		),
		"search/types.go": regexp.MustCompile(`shadow-2xl|shadow-black/20`),
		"kbd/types.go":    regexp.MustCompile(`shadow-sm|shadow-outline/30|dark:shadow-black/20`),
	}

	seenAllowed := make(map[string]int, len(allowedPaletteData))
	var violations []string
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
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

		normalizedPath := strings.TrimPrefix(filepath.ToSlash(path), "./")
		scanner := bufio.NewScanner(file)
		for lineNumber := 1; scanner.Scan(); lineNumber++ {
			line := strings.TrimSpace(scanner.Text())
			if matcher := namedRaw[normalizedPath]; matcher != nil {
				for _, token := range matcher.FindAllString(line, -1) {
					violations = append(violations, fmt.Sprintf("%s:%d: raw %s", normalizedPath, lineNumber, token))
				}
			}

			for _, token := range rawBackdropDisplayColor.FindAllString(line, -1) {
				use := backdropElevationUse{path: normalizedPath, line: line, token: token}
				if allowedPaletteData[use.key()] > seenAllowed[use.key()] {
					seenAllowed[use.key()]++
					continue
				}
				if matcher := namedRaw[normalizedPath]; matcher != nil && matcher.MatchString(token) {
					continue
				}
				violations = append(violations, fmt.Sprintf("%s:%d: ungoverned raw display color %s", normalizedPath, lineNumber, token))
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatalf("scan authored component sources: %v", err)
	}

	for allowed, want := range allowedPaletteData {
		if got := seenAllowed[allowed]; got != want {
			violations = append(violations, fmt.Sprintf("Palette displayed-data allowlist mismatch: got %d, want %d for %q", got, want, allowed))
		}
	}
	sort.Strings(violations)
	if len(violations) > 0 {
		t.Fatalf("backdrop and elevation styling requires governed semantic roles; only exact Palette displayed-data literals are allowed:\n%s", strings.Join(violations, "\n"))
	}
}

func TestBackdropAndElevationSemanticRoleProvenance(t *testing.T) {
	t.Parallel()

	requiredSourceRoles := []struct {
		path  string
		role  string
		count int
	}{
		{path: "drawer/types.go", role: "bg-backdrop/40", count: 1},
		{path: "drawer/types.go", role: "dark:bg-backdrop/60", count: 1},
		{path: "drawer/types.go", role: "shadow-elevation-raised", count: 1},
		{path: "modal/modal.templ", role: "bg-backdrop/20", count: 2},
		{path: "sidebar/types.go", role: "bg-backdrop/50", count: 1},
		{path: "search/search.templ", role: "bg-backdrop-surface/55", count: 1},
		{path: "search/search.templ", role: "dark:bg-backdrop/60", count: 1},
		{path: "search/types.go", role: "shadow-elevation-overlay", count: 1},
		{path: "kbd/types.go", role: "shadow-elevation-control", count: 1},
		{path: "kbd/types.go", role: "dark:shadow-elevation-control-dark", count: 1},
	}
	for _, required := range requiredSourceRoles {
		content, err := os.ReadFile(required.path)
		if err != nil {
			t.Fatalf("read %s: %v", required.path, err)
		}
		got := 0
		for _, token := range authoredClassToken.FindAllString(string(content), -1) {
			if token == required.role {
				got++
			}
		}
		if got != required.count {
			t.Errorf("%s semantic role %q count = %d, want %d", required.path, required.role, got, required.count)
		}
	}

	mappings := []string{
		`--color-backdrop: var(--color-black);`,
		`--color-backdrop-surface: var(--color-surface-dark);`,
		`--shadow-elevation-control: 0 1px 3px 0 color-mix(in srgb, var(--color-outline) 30%, transparent), 0 1px 2px -1px color-mix(in srgb, var(--color-outline) 30%, transparent);`,
		`--shadow-elevation-control-dark: 0 1px 3px 0 color-mix(in srgb, var(--color-black) 20%, transparent), 0 1px 2px -1px color-mix(in srgb, var(--color-black) 20%, transparent);`,
		`--shadow-elevation-raised: 0 20px 25px -5px color-mix(in srgb, var(--color-black) 10%, transparent), 0 8px 10px -6px color-mix(in srgb, var(--color-black) 10%, transparent);`,
		`--shadow-elevation-overlay: 0 25px 50px -12px color-mix(in srgb, var(--color-black) 20%, transparent);`,
	}
	for _, artifact := range []string{"../all-themes.css", "../assets/goshtoso-theme.css"} {
		content, err := os.ReadFile(artifact)
		if err != nil {
			t.Fatalf("read %s: %v", artifact, err)
		}
		for _, mapping := range mappings {
			if got := strings.Count(string(content), mapping); got != 1 {
				t.Errorf("%s mapping %q count = %d, want 1", artifact, mapping, got)
			}
		}
	}
}
