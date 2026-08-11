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

var (
	rawOutlineBoundaryRole = regexp.MustCompile(`(?:[[:alnum:]_\[\]=/-]+:)*(?:border|divide)-outline(?:-dark)?(?:-strong)?(?:/[[:digit:]]+)?`)
	controlBoundaryRole    = regexp.MustCompile(`(?:[[:alnum:]_\[\]=/-]+:)*border-control-outline(?:-dark)?`)
)

type requiredBoundarySite struct {
	family string
	name   string
	path   string
	light  int
	dark   int
}

// TestInteractiveBoundariesUseControlOutline separates required control
// boundaries from passive structural borders. A focus ring is additive and
// never satisfies the idle-boundary requirement.
func TestInteractiveBoundariesUseControlOutline(t *testing.T) {
	t.Parallel()

	required := []requiredBoundarySite{
		{family: "checkbox", name: "standard and container checkbox inputs", path: "checkbox/types.go", light: 2, dark: 2},
		{family: "radio", name: "radio input", path: "radio/types.go", light: 1, dark: 1},
		{family: "toggle", name: "switch track", path: "toggle/types.go", light: 1, dark: 1},
		{family: "file-input", name: "compact upload control", path: "fileinput/types.go", light: 1, dark: 1},
		{family: "file-input", name: "drop-zone idle state", path: "fileinput/fileinput.templ", light: 1, dark: 1},
		{family: "select", name: "native, disabled trigger, and default trigger", path: "select/types.go", light: 3, dark: 3},
		{family: "structured-input", name: "add action, text editor, and select editor", path: "structuredinput/structuredinput.templ", light: 4, dark: 4},
		{family: "palette-input", name: "preview, hex field, and swatch controls", path: "palette/palette.templ", light: 5, dark: 5},
		{family: "sidebar-search", name: "sidebar search field", path: "sidebar/sidebar.templ", light: 1, dark: 1},
		{family: "search-trigger", name: "trigger default and hover states", path: "search/types.go", light: 2, dark: 2},
		{family: "search-trigger", name: "dialog query boundary", path: "search/search.templ", light: 1, dark: 1},
		{family: "table-selection", name: "selection checkbox", path: "table/types.go", light: 1, dark: 1},
		{family: "table-selection", name: "filter query and select controls", path: "table/table.templ", light: 2, dark: 2},
		{family: "schema-form", name: "boolean, managed, and editable fields", path: "schemaform/schemaform.templ", light: 3, dark: 3},
	}

	decorativeLines := []struct {
		path string
		line string
	}{
		{path: "checkbox/checkbox.templ", line: `<label for={ cfg.ID } class="inline-flex min-w-52 items-center justify-between rounded-radius gap-3 border border-outline bg-surface-alt px-4 py-2 text-sm font-medium text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark has-checked:text-on-surface-strong dark:has-checked:text-on-surface-dark-strong has-disabled:cursor-not-allowed">`},
		{path: "checkbox/checkbox.templ", line: `<ul class="flex min-w-52 flex-col divide-y divide-outline overflow-clip rounded-radius border border-outline dark:divide-outline-dark dark:border-outline-dark">`},
		{path: "radio/radio.templ", line: `<div class="inline-flex w-fit overflow-clip rounded-radius border border-outline bg-surface-alt divide-x divide-outline dark:border-outline-dark dark:bg-surface-dark-alt dark:divide-outline-dark">`},
		{path: "radio/radio.templ", line: `<label for={ cfg.ID } class={ "inline-flex min-w-52 items-center justify-between gap-3 rounded-radius border border-outline bg-surface-alt px-4 py-2 text-sm font-medium text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark has-checked:text-on-surface-strong dark:has-checked:text-on-surface-dark-strong has-disabled:cursor-not-allowed " + cfg.RootClass }>`},
		{path: "radio/radio.templ", line: `<ul class="flex min-w-52 flex-col divide-y divide-outline overflow-clip rounded-radius border border-outline dark:divide-outline-dark dark:border-outline-dark">`},
		{path: "toggle/types.go", line: `base = "inline-flex min-w-52 items-center justify-between gap-3 rounded-radius border border-outline bg-surface-alt px-4 py-1.5 dark:border-outline-dark dark:bg-surface-dark-alt"`},
		{path: "fileinput/types.go", line: `base := "flex shrink-0 items-center border-l border-outline bg-surface px-3 py-2 font-medium text-primary dark:border-outline-dark dark:bg-surface-dark dark:text-primary-dark"`},
		{path: "select/select.templ", line: `class="absolute left-0 top-full z-30 mt-1 min-w-full overflow-hidden rounded-radius border border-outline bg-surface-alt shadow-lg dark:border-outline-dark dark:bg-surface-dark-alt"`},
		{path: "sidebar/types.go", line: `return "h-full w-full border-r border-outline bg-surface dark:border-outline-dark dark:bg-surface-dark flex flex-col"`},
		{path: "sidebar/sidebar.templ", line: `<div class="shrink-0 border-b border-outline dark:border-outline-dark p-4">`},
		{path: "sidebar/sidebar.templ", line: `<span class="flex items-center border-l border-outline py-2.5 pl-4 text-sm font-medium text-on-surface-muted cursor-not-allowed dark:border-outline-dark dark:text-on-surface-dark-muted">`},
		{path: "sidebar/sidebar.templ", line: `class="flex items-center gap-2 border-l border-outline py-2.5 pl-4 text-sm font-medium text-on-surface transition duration-200 hover:border-l-2 hover:border-outline-strong hover:text-on-surface-strong dark:border-outline-dark dark:text-on-surface-dark dark:hover:border-outline-dark-strong dark:hover:text-on-surface-dark-strong"`},
		{path: "sidebar/sidebar.templ", line: `class="flex items-center gap-2 border-l border-outline py-2.5 pl-4 text-sm font-medium text-on-surface transition duration-200 hover:border-l-2 hover:border-outline-strong hover:text-on-surface-strong focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:border-outline-dark dark:text-on-surface-dark dark:hover:border-outline-dark-strong dark:hover:text-on-surface-dark-strong dark:focus-visible:outline-primary-dark"`},
		{path: "search/search.templ", line: `class="flex max-h-[min(28rem,calc(100vh-12rem))] flex-col divide-y divide-outline overflow-hidden overflow-y-auto rounded-b-radius border border-outline bg-surface text-sm font-light dark:divide-outline-dark dark:border-outline-dark dark:bg-surface-dark-alt"`},
		{path: "search/search.templ", line: `<span class="shrink-0 rounded-radius border border-outline px-2 py-0.5 text-xs font-medium text-on-surface-muted dark:border-outline-dark dark:text-on-surface-dark-muted">{ item.Section }</span>`},
		{path: "search/search.templ", line: `<span class="shrink-0 rounded-radius border border-outline px-2 py-0.5 text-xs font-medium text-on-surface-muted dark:border-outline-dark dark:text-on-surface-dark-muted" x-text="item.section"></span>`},
		{path: "table/types.go", line: `base := "overflow-x-auto overflow-y-clip w-full rounded-radius border border-outline dark:border-outline-dark"`},
		{path: "table/types.go", line: `return "border-b border-outline bg-surface-alt text-sm text-on-surface-strong dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark-strong"`},
		{path: "table/types.go", line: `return "divide-y divide-outline dark:divide-outline-dark"`},
		{path: "table/types.go", line: `return base + " border border-outline dark:border-outline-dark"`},
		{path: "table/table.templ", line: `<div id={ cfg.filterBarID() } class="border border-outline rounded-radius mb-3 bg-surface dark:border-outline-dark dark:bg-surface-dark">`},
		{path: "table/table.templ", line: `<div class="relative h-6 w-11 rounded-full bg-outline/30 after:absolute after:left-[2px] after:top-[2px] after:size-5 after:rounded-full after:border after:border-outline/20 after:bg-surface after:transition-all peer-checked:bg-primary peer-checked:after:translate-x-full peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-primary dark:bg-outline-dark/30 dark:after:bg-surface-dark dark:peer-checked:bg-primary-dark"></div>`},
		{path: "table/table.templ", line: `class="flex items-center justify-between border-t border-outline px-4 py-3 dark:border-outline-dark"`},
		{path: "schemaform/schemaform.templ", line: `<fieldset class="rounded-radius border border-outline p-3 dark:border-outline-dark">`},
	}

	allowed := make(map[string]int, len(decorativeLines))
	for _, entry := range decorativeLines {
		allowed[entry.path+"\x00"+entry.line]++
	}
	seenAllowed := make(map[string]int, len(allowed))
	counts := make(map[string][2]int)
	var violations []string
	for _, family := range []string{"checkbox", "radio", "toggle", "fileinput", "select", "structuredinput", "palette", "sidebar", "search", "table", "schemaform"} {
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
				for _, role := range controlBoundaryRole.FindAllString(line, -1) {
					value := counts[filepath.ToSlash(path)]
					if strings.HasSuffix(role, "border-control-outline-dark") {
						value[1]++
					} else {
						value[0]++
					}
					counts[filepath.ToSlash(path)] = value
				}
				if len(rawOutlineBoundaryRole.FindAllString(line, -1)) == 0 {
					continue
				}
				key := filepath.ToSlash(path) + "\x00" + line
				if seenAllowed[key] < allowed[key] {
					seenAllowed[key]++
					continue
				}
				violations = append(violations, fmt.Sprintf("%s:%d: non-decorative outline boundary: %s", path, lineNumber, line))
			}
			return scanner.Err()
		})
		if err != nil {
			t.Fatalf("scan %s control boundaries: %v", family, err)
		}
	}

	for _, site := range required {
		got := counts[site.path]
		if got[0] != site.light || got[1] != site.dark {
			violations = append(violations, fmt.Sprintf("%s %s (%s): control boundary roles light/dark = %d/%d, want %d/%d", site.family, site.name, site.path, got[0], got[1], site.light, site.dark))
		}
	}
	for key, want := range allowed {
		if got := seenAllowed[key]; got != want {
			violations = append(violations, fmt.Sprintf("decorative allowlist mismatch: got %d, want %d for %q", got, want, key))
		}
	}
	sort.Strings(violations)
	if len(violations) > 0 {
		t.Fatalf("interactive boundaries require control-outline light/dark pairs; only exact decorative containers and dividers may retain outline roles:\n%s", strings.Join(violations, "\n"))
	}
}
