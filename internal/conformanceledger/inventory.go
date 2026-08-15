package conformanceledger

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/internal/themeinventory"
)

var (
	kindDeclaration = regexp.MustCompile(`(?m)^\s*(Kind[A-Za-z0-9]+)\s+Kind\s*=\s*"([^"]+)"`)
	renderableEntry = regexp.MustCompile(`(?m)^\s*components\.(Kind[A-Za-z0-9]+):\s*([A-Za-z0-9_]+\.[A-Za-z0-9_]+)\(`)
	routePath       = regexp.MustCompile(`(?m)^\s*Path:\s*"(/components/[^"]+)"`)
	breakpointQuery = regexp.MustCompile(`@media\s*\(width\s*>=\s*([0-9]+(?:\.[0-9]+)?)rem\)`)
)

// lifecycleStateAuthorities is intentionally small and explicit. Each marker
// is owned by maintained component source, and DeriveInventory fails if a
// source change removes or renames a required dynamic state without updating
// its conformance mapping. These do not inflate the 347 configuration/default
// state Cartesian axis; they are a separately reconciled lifecycle axis.
var lifecycleStateAuthorities = []struct {
	Value  string
	Path   string
	Marker string
}{
	{Value: "button/lifecycle/disabled", Path: "components/button/button.templ", Marker: "disabled"},
	{Value: "button/lifecycle/loading", Path: "components/button/button.templ", Marker: "data-goshtoso-loading"},
	{Value: "button/lifecycle/hover-focus", Path: "components/button/types.go", Marker: "focus-visible"},
	{Value: "dropdown/lifecycle/open-closed-dismiss", Path: "components/dropdown/dropdown.templ", Marker: "closeAndFocus()"},
	{Value: "modal/lifecycle/open-closed-dismiss", Path: "components/modal/modal.templ", Marker: "keydown.esc.window"},
	{Value: "drawer/lifecycle/open-closed-dismiss", Path: "components/drawer/drawer.templ", Marker: "drawer:close-request"},
	{Value: "tooltip/lifecycle/hover-focus", Path: "components/tooltip/tooltip.templ", Marker: "peer-hover:opacity-100"},
	{Value: "tooltip/lifecycle/click-open-dismiss", Path: "components/tooltip/tooltip.templ", Marker: "showTooltip = false"},
	{Value: "search/lifecycle/open-closed-dismiss", Path: "components/search/search.templ", Marker: "closeSearch()"},
	{Value: "carousel/lifecycle/touch-reduced-motion", Path: "components/carousel/carousel.templ", Marker: "x-on:touchstart"},
	{Value: "form/lifecycle/error", Path: "components/form/errors.templ", Marker: "role=\"alert\""},
	{Value: "form/lifecycle/empty", Path: "components/form/errors.templ", Marker: "Empty Items renders nothing"},
	{Value: "avatar/lifecycle/loading-error", Path: "components/avatar/avatar.templ", Marker: "x-on:error"},
}

func DeriveInventory(repoRoot string) (Inventory, error) {
	packages, err := derivePackages(repoRoot)
	if err != nil {
		return Inventory{}, err
	}
	kinds, err := deriveKinds(repoRoot)
	if err != nil {
		return Inventory{}, err
	}
	renderables, err := deriveRenderables(repoRoot)
	if err != nil {
		return Inventory{}, err
	}
	routes, err := deriveRoutes(repoRoot)
	if err != nil {
		return Inventory{}, err
	}
	themes, err := deriveThemes(repoRoot)
	if err != nil {
		return Inventory{}, err
	}
	edges, err := deriveBreakpointEdges(repoRoot)
	if err != nil {
		return Inventory{}, err
	}

	states, err := deriveStateMetadata(repoRoot, kinds)
	if err != nil {
		return Inventory{}, err
	}
	lifecycleStates, err := deriveLifecycleStates(repoRoot)
	if err != nil {
		return Inventory{}, err
	}

	return Inventory{
		Packages:        packages,
		Renderables:     renderables,
		Kinds:           kinds,
		Routes:          routes,
		States:          states,
		LifecycleStates: lifecycleStates,
		Themes:          themes,
		BreakpointEdges: edges,
	}, nil
}

func deriveLifecycleStates(repoRoot string) ([]SourceItem, error) {
	items := make([]SourceItem, 0, len(lifecycleStateAuthorities))
	seen := make(map[string]struct{}, len(lifecycleStateAuthorities))
	for _, authority := range lifecycleStateAuthorities {
		if _, duplicate := seen[authority.Value]; duplicate {
			return nil, fmt.Errorf("duplicate lifecycle state authority %s", authority.Value)
		}
		seen[authority.Value] = struct{}{}
		content, err := os.ReadFile(filepath.Join(repoRoot, authority.Path))
		if err != nil {
			return nil, fmt.Errorf("read lifecycle authority %s: %w", authority.Path, err)
		}
		if !strings.Contains(string(content), authority.Marker) {
			return nil, fmt.Errorf("unmapped lifecycle state %s: %s no longer contains %q", authority.Value, authority.Path, authority.Marker)
		}
		line, err := lineContaining(filepath.Join(repoRoot, authority.Path), authority.Marker)
		if err != nil {
			return nil, fmt.Errorf("locate lifecycle authority %s: %w", authority.Value, err)
		}
		if line == 0 {
			return nil, fmt.Errorf("unmapped lifecycle state %s: marker %q has no source line", authority.Value, authority.Marker)
		}
		items = append(items, SourceItem{Value: authority.Value, Source: SourceRef{Path: authority.Path, Symbol: fmt.Sprintf("%s:%d", authority.Marker, line)}})
	}
	sortSourceItems(items)
	return items, nil
}

func derivePackages(repoRoot string) ([]SourceItem, error) {
	componentsRoot := filepath.Join(repoRoot, "components")
	var packages []SourceItem
	err := filepath.WalkDir(componentsRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasGo := false
		for _, child := range entries {
			if !child.IsDir() && strings.HasSuffix(child.Name(), ".go") && !strings.HasSuffix(child.Name(), "_test.go") {
				hasGo = true
				break
			}
		}
		if !hasGo {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		packages = append(packages, SourceItem{
			Value:  "github.com/araihu/goshtoso/" + filepath.ToSlash(rel),
			Source: SourceRef{Path: filepath.ToSlash(rel), Symbol: "package " + filepath.Base(path)},
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("derive component packages: %w", err)
	}
	sortSourceItems(packages)
	return packages, nil
}

func deriveKinds(repoRoot string) ([]SourceItem, error) {
	path := filepath.Join(repoRoot, "components", "component.go")
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read kind authority: %w", err)
	}
	matches := kindDeclaration.FindAllStringSubmatch(string(source), -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("kind authority %s yielded no Kinds", path)
	}
	items := make([]SourceItem, 0, len(matches))
	for _, match := range matches {
		items = append(items, SourceItem{Value: match[2], Source: SourceRef{Path: "components/component.go", Symbol: match[1]}})
	}
	return items, nil
}

func deriveRenderables(repoRoot string) ([]SourceItem, error) {
	files := []string{
		"components/composition_identity_test.go",
		"components/display_identity_test.go",
		"components/feedback_navigation_identity_test.go",
		"components/input_identity_test.go",
	}
	var items []SourceItem
	seen := map[string]struct{}{}
	for _, relative := range files {
		source, err := os.ReadFile(filepath.Join(repoRoot, relative))
		if err != nil {
			return nil, fmt.Errorf("read renderable authority %s: %w", relative, err)
		}
		for _, match := range renderableEntry.FindAllStringSubmatch(string(source), -1) {
			if _, duplicate := seen[match[2]]; duplicate {
				return nil, fmt.Errorf("duplicate renderable %s", match[2])
			}
			seen[match[2]] = struct{}{}
			items = append(items, SourceItem{Value: match[2], Source: SourceRef{Path: relative, Symbol: match[1]}})
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("renderable authorities yielded no renderables")
	}
	sortSourceItems(items)
	return items, nil
}

func deriveRoutes(repoRoot string) ([]SourceItem, error) {
	relative := "site/internal/pages/catalog/catalog.go"
	source, err := os.ReadFile(filepath.Join(repoRoot, relative))
	if err != nil {
		return nil, fmt.Errorf("read route authority: %w", err)
	}
	matches := routePath.FindAllStringSubmatch(string(source), -1)
	items := make([]SourceItem, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if _, duplicate := seen[match[1]]; duplicate {
			return nil, fmt.Errorf("duplicate route %s in %s", match[1], relative)
		}
		seen[match[1]] = struct{}{}
		items = append(items, SourceItem{Value: match[1], Source: SourceRef{Path: relative, Symbol: strings.TrimPrefix(match[1], "/")}})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("route authority %s yielded no component routes", relative)
	}
	return items, nil
}

func deriveThemes(repoRoot string) ([]SourceItem, error) {
	relative := "all-themes.css"
	source, err := os.ReadFile(filepath.Join(repoRoot, relative))
	if err != nil {
		return nil, fmt.Errorf("read theme authority: %w", err)
	}
	themes, err := themeinventory.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parse theme authority: %w", err)
	}
	items := make([]SourceItem, 0, len(themes))
	for _, theme := range themes {
		items = append(items, SourceItem{Value: theme.Key, Source: SourceRef{Path: relative, Symbol: "[data-theme=" + theme.Key + "]"}})
	}
	return items, nil
}

func deriveBreakpointEdges(repoRoot string) ([]int, error) {
	relative := "assets/styles.css"
	source, err := os.ReadFile(filepath.Join(repoRoot, relative))
	if err != nil {
		return nil, fmt.Errorf("read compiled breakpoint authority: %w", err)
	}
	seen := map[int]struct{}{}
	for _, match := range breakpointQuery.FindAllStringSubmatch(string(source), -1) {
		rem, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return nil, fmt.Errorf("parse breakpoint %q: %w", match[1], err)
		}
		pixels := int(rem * 16)
		if float64(pixels) != rem*16 {
			return nil, fmt.Errorf("breakpoint %srem is not an integer pixel edge", match[1])
		}
		seen[pixels] = struct{}{}
	}
	edges := make([]int, 0, len(seen))
	for edge := range seen {
		edges = append(edges, edge)
	}
	sort.Ints(edges)
	if len(edges) == 0 {
		return nil, fmt.Errorf("compiled breakpoint authority yielded no width edges")
	}
	return edges, nil
}

func deriveStateMetadata(repoRoot string, kinds []SourceItem) ([]SourceItem, error) {
	states := make([]SourceItem, 0, len(kinds))
	for _, kind := range kinds {
		states = append(states, SourceItem{
			Value:  kind.Value + "/default",
			Source: kind.Source,
		})
	}

	componentsRoot := filepath.Join(repoRoot, "components")
	err := filepath.WalkDir(componentsRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "types.go" {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse state metadata %s: %w", path, err)
		}
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return fmt.Errorf("relativize state metadata %s: %w", path, err)
		}
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			activeType := ""
			for _, spec := range gen.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if typeName, ok := value.Type.(*ast.Ident); ok {
					activeType = typeName.Name
				}
				if activeType == "" || !ast.IsExported(activeType) {
					continue
				}
				for _, name := range value.Names {
					if ast.IsExported(name.Name) {
						states = append(states, SourceItem{Value: file.Name.Name + "/" + activeType + "." + name.Name, Source: SourceRef{Path: filepath.ToSlash(relative), Symbol: name.Name}})
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("derive state metadata: %w", err)
	}
	sortSourceItems(states)
	for index := 1; index < len(states); index++ {
		if states[index].Value == states[index-1].Value {
			return nil, fmt.Errorf("duplicate state %s from %s and %s", states[index].Value, states[index-1].Source.Path, states[index].Source.Path)
		}
	}
	return states, nil
}

func sortSourceItems(items []SourceItem) {
	sort.Slice(items, func(left, right int) bool { return items[left].Value < items[right].Value })
}

// lineContaining returns a stable source symbol for text-based authorities.
func lineContaining(path, needle string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		if strings.Contains(scanner.Text(), needle) {
			return line, nil
		}
	}
	return 0, scanner.Err()
}
