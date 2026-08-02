package e2econstraints

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var identityPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Identity records the source files and runnable tests selected by one build tag.
type Identity struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
	Tests []string `json:"tests"`
}

// Manifest is the checked-in contract for focused and full-only E2E tests.
type Manifest struct {
	Identities []Identity `json:"identities"`
	FullOnly   Identity   `json:"full_only"`
}

// SourceFile is one parsed E2E source file and its Go build expression.
type SourceFile struct {
	Name  string
	Expr  constraint.Expr
	Tests []string
}

// Suite is the statically parsed E2E build-constraint inventory.
type Suite struct {
	Files []SourceFile
}

// LoadManifest reads and validates the checked-in identity names.
func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read identity manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode identity manifest: %w", err)
	}
	seen := make(map[string]struct{}, len(manifest.Identities))
	for i, identity := range manifest.Identities {
		if !identityPattern.MatchString(identity.Name) {
			return Manifest{}, fmt.Errorf("identity %d has invalid name %q", i, identity.Name)
		}
		if _, duplicate := seen[identity.Name]; duplicate {
			return Manifest{}, fmt.Errorf("duplicate identity %q", identity.Name)
		}
		seen[identity.Name] = struct{}{}
	}
	if len(manifest.Identities) == 0 {
		return Manifest{}, fmt.Errorf("identity manifest is empty")
	}
	return manifest, nil
}

// InspectSuite parses build constraints with the standard library and discovers Test functions via AST.
func InspectSuite(dir string) (Suite, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Suite{}, fmt.Errorf("read E2E directory: %w", err)
	}
	var suite Suite
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		expr, err := readBuildConstraint(path)
		if err != nil {
			return Suite{}, err
		}
		tests, err := discoverTests(path)
		if err != nil {
			return Suite{}, err
		}
		suite.Files = append(suite.Files, SourceFile{Name: entry.Name(), Expr: expr, Tests: tests})
	}
	slices.SortFunc(suite.Files, func(a, b SourceFile) int { return strings.Compare(a.Name, b.Name) })
	return suite, nil
}

func readBuildConstraint(path string) (constraint.Expr, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "//go:build ") {
			expr, err := constraint.Parse(line)
			if err != nil {
				return nil, fmt.Errorf("parse %s constraint: %w", filepath.Base(path), err)
			}
			return expr, nil
		}
		if line != "" && !strings.HasPrefix(line, "//") {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return nil, fmt.Errorf("%s: missing //go:build constraint", filepath.Base(path))
}

func discoverTests(path string) ([]string, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	var tests []string
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name.Name == "TestMain" ||
			!strings.HasPrefix(function.Name.Name, "Test") {
			continue
		}
		tests = append(tests, function.Name.Name)
	}
	slices.Sort(tests)
	return tests, nil
}

// ValidateSuite checks tag safety and exact manifest ownership.
func ValidateSuite(suite Suite, manifest Manifest) error {
	known := map[string]bool{"e2e": true, "full": true}
	for _, identity := range manifest.Identities {
		known[identity.Name] = true
	}
	for _, file := range suite.Files {
		for _, tag := range expressionTags(file.Expr) {
			if !known[tag] {
				return fmt.Errorf("%s: unknown build tag %q", file.Name, tag)
			}
		}
		if file.Expr.Eval(func(tag string) bool { return tag != "e2e" && known[tag] }) {
			return fmt.Errorf("%s: build expression does not require e2e", file.Name)
		}
		if len(file.Tests) > 0 {
			if file.Expr.Eval(func(tag string) bool { return tag == "e2e" }) {
				return fmt.Errorf("%s: runnable tests use suite-only support constraint", file.Name)
			}
			if !file.Expr.Eval(func(tag string) bool { return tag == "e2e" || tag == "full" }) {
				return fmt.Errorf("%s: runnable tests are missing full fallback", file.Name)
			}
		}
	}

	seenTests := make(map[string]string)
	for _, file := range suite.Files {
		for _, test := range file.Tests {
			if previous, duplicate := seenTests[test]; duplicate {
				return fmt.Errorf("test %s declared in both %s and %s", test, previous, file.Name)
			}
			seenTests[test] = file.Name
		}
	}

	for _, identity := range manifest.Identities {
		actual := suite.Inventory(identity.Name)
		if len(actual.Tests) == 0 {
			return fmt.Errorf("identity %q selects no tests", identity.Name)
		}
		if err := compareIdentity(identity, actual); err != nil {
			return err
		}
	}
	if err := compareIdentity(manifest.FullOnly, suite.FullOnly(manifest)); err != nil {
		return err
	}
	return nil
}

// ExpandedManifest fills the checked-in identity names with their parsed files and tests.
func ExpandedManifest(suite Suite, manifest Manifest) Manifest {
	expanded := Manifest{Identities: make([]Identity, 0, len(manifest.Identities))}
	for _, identity := range manifest.Identities {
		expanded.Identities = append(expanded.Identities, suite.Inventory(identity.Name))
	}
	expanded.FullOnly = suite.FullOnly(manifest)
	return expanded
}

// WriteManifest writes a deterministic, reviewable identity inventory.
func WriteManifest(writer io.Writer, manifest Manifest) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

func compareIdentity(expected, actual Identity) error {
	if !slices.Equal(expected.Files, actual.Files) {
		return fmt.Errorf("identity %q files mismatch: expected %v, got %v", expected.Name, expected.Files, actual.Files)
	}
	if !slices.Equal(expected.Tests, actual.Tests) {
		return fmt.Errorf("identity %q tests mismatch: expected %v, got %v", expected.Name, expected.Tests, actual.Tests)
	}
	return nil
}

// Inventory returns the exact files and Test functions selected by one identity.
func (suite Suite) Inventory(identity string) Identity {
	tags := map[string]bool{"e2e": true, identity: true}
	return suite.inventory(identity, func(tag string) bool { return tags[tag] })
}

// FullInventory returns all tests compiled by e2e,full.
func (suite Suite) FullInventory() Identity {
	tags := map[string]bool{"e2e": true, "full": true}
	return suite.inventory("full", func(tag string) bool { return tags[tag] })
}

// FullOnly returns the full tests not selected by any focused identity.
func (suite Suite) FullOnly(manifest Manifest) Identity {
	focused := make(map[string]bool)
	for _, identity := range manifest.Identities {
		for _, test := range suite.Inventory(identity.Name).Tests {
			focused[test] = true
		}
	}
	full := suite.FullInventory()
	result := Identity{Name: "full_only"}
	for _, file := range suite.Files {
		selected := false
		for _, test := range file.Tests {
			if slices.Contains(full.Tests, test) && !focused[test] {
				result.Tests = append(result.Tests, test)
				selected = true
			}
		}
		if selected {
			result.Files = append(result.Files, file.Name)
		}
	}
	slices.Sort(result.Tests)
	return result
}

func (suite Suite) inventory(name string, match func(string) bool) Identity {
	result := Identity{Name: name}
	for _, file := range suite.Files {
		if !file.Expr.Eval(match) {
			continue
		}
		if len(file.Tests) > 0 {
			result.Files = append(result.Files, file.Name)
			result.Tests = append(result.Tests, file.Tests...)
		}
	}
	slices.Sort(result.Tests)
	return result
}

func expressionTags(expr constraint.Expr) []string {
	seen := make(map[string]bool)
	var visit func(constraint.Expr)
	visit = func(current constraint.Expr) {
		switch value := current.(type) {
		case *constraint.TagExpr:
			seen[value.Tag] = true
		case *constraint.NotExpr:
			visit(value.X)
		case *constraint.AndExpr:
			visit(value.X)
			visit(value.Y)
		case *constraint.OrExpr:
			visit(value.X)
			visit(value.Y)
		}
	}
	visit(expr)
	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}
	slices.Sort(tags)
	return tags
}

// RunCompileMatrix asks the Go compiler to build every focused identity and full suite.
func RunCompileMatrix(ctx context.Context, siteDir string, manifest Manifest, output io.Writer) error {
	temporary, err := os.MkdirTemp("", "goshtoso-e2e-matrix-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	identities := append(identityNames(manifest), "full")
	for _, identity := range identities {
		if _, err := fmt.Fprintf(output, "compile e2e,%s\n", identity); err != nil {
			return fmt.Errorf("report compile identity: %w", err)
		}
		command := exec.CommandContext(ctx, "go", "test", "-c", "-tags=e2e,"+identity,
			"-o", filepath.Join(temporary, identity+".test"), "./tests/e2e")
		command.Dir = siteDir
		command.Stdout = output
		command.Stderr = output
		if err := command.Run(); err != nil {
			return fmt.Errorf("compile e2e,%s: %w", identity, err)
		}
	}
	return nil
}

// RunListMatrix compares Go's real test inventory with the parsed constraints.
func RunListMatrix(ctx context.Context, siteDir string, suite Suite, manifest Manifest, output io.Writer) error {
	identities := append(identityNames(manifest), "full")
	for _, identity := range identities {
		if _, err := fmt.Fprintf(output, "list e2e,%s\n", identity); err != nil {
			return fmt.Errorf("report list identity: %w", err)
		}
		command := exec.CommandContext(ctx, "go", "test", "-tags=e2e,"+identity, "-list", "^Test", "./tests/e2e")
		command.Dir = siteDir
		command.Env = append(os.Environ(), "GOSHTOSO_E2E_LIST_ONLY=1")
		data, err := command.CombinedOutput()
		if err != nil {
			_, _ = output.Write(data)
			return fmt.Errorf("list e2e,%s: %w", identity, err)
		}
		actual := parseGoTestList(string(data))
		expected := suite.Inventory(identity).Tests
		if identity == "full" {
			expected = suite.FullInventory().Tests
		}
		if !slices.Equal(expected, actual) {
			return fmt.Errorf("e2e,%s list mismatch: expected %v, got %v", identity, expected, actual)
		}
	}
	return nil
}

func parseGoTestList(output string) []string {
	var tests []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Test") && !strings.ContainsAny(line, " \t") {
			tests = append(tests, line)
		}
	}
	slices.Sort(tests)
	return tests
}

func identityNames(manifest Manifest) []string {
	names := make([]string, 0, len(manifest.Identities))
	for _, identity := range manifest.Identities {
		names = append(names, identity.Name)
	}
	return names
}
