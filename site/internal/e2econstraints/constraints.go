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

const currentSourceTag = "goshtoso_current_source"

// Identity records the source files and runnable tests selected by one build tag.
type Identity struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
	Tests []string `json:"tests"`
}

// SpecializedSuite records a deliberately opt-in E2E contract that cannot run
// under the standard full suite. Its exact tag set, files, and tests are a
// checked-in ownership contract; specialized suites cannot hide behind generic
// tag fallback.
type SpecializedSuite struct {
	Name          string   `json:"name"`
	Tags          []string `json:"tags"`
	Files         []string `json:"files"`
	Tests         []string `json:"tests"`
	SelectedTests []string `json:"selected_tests"`
}

// Manifest is the checked-in contract for focused and full-only E2E tests.
type Manifest struct {
	Identities        []Identity         `json:"identities"`
	SpecializedSuites []SpecializedSuite `json:"specialized_suites,omitempty"`
	FullOnly          Identity           `json:"full_only"`
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
	seen := make(map[string]struct{}, len(manifest.Identities)+len(manifest.SpecializedSuites))
	for i, identity := range manifest.Identities {
		if !identityPattern.MatchString(identity.Name) {
			return Manifest{}, fmt.Errorf("identity %d has invalid name %q", i, identity.Name)
		}
		if _, duplicate := seen[identity.Name]; duplicate {
			return Manifest{}, fmt.Errorf("duplicate identity %q", identity.Name)
		}
		seen[identity.Name] = struct{}{}
	}
	for i, specialized := range manifest.SpecializedSuites {
		if !identityPattern.MatchString(specialized.Name) {
			return Manifest{}, fmt.Errorf("specialized suite %d has invalid name %q", i, specialized.Name)
		}
		if _, duplicate := seen[specialized.Name]; duplicate {
			return Manifest{}, fmt.Errorf("duplicate specialized suite %q", specialized.Name)
		}
		seen[specialized.Name] = struct{}{}
		if len(specialized.Tags) == 0 {
			return Manifest{}, fmt.Errorf("specialized suite %q has no tags", specialized.Name)
		}
		seenTags := make(map[string]struct{}, len(specialized.Tags))
		for _, tag := range specialized.Tags {
			if !identityPattern.MatchString(tag) || tag == "e2e" || tag == "full" || tag == currentSourceTag {
				return Manifest{}, fmt.Errorf("specialized suite %q has invalid tag %q", specialized.Name, tag)
			}
			if _, duplicate := seenTags[tag]; duplicate {
				return Manifest{}, fmt.Errorf("specialized suite %q has duplicate tag %q", specialized.Name, tag)
			}
			seenTags[tag] = struct{}{}
		}
		if err := validateSpecializedInventoryDeclaration(specialized); err != nil {
			return Manifest{}, err
		}
	}
	if len(manifest.Identities) == 0 {
		return Manifest{}, fmt.Errorf("identity manifest is empty")
	}
	return manifest, nil
}

func validateSpecializedInventoryDeclaration(specialized SpecializedSuite) error {
	for name, values := range map[string][]string{
		"files":          specialized.Files,
		"tests":          specialized.Tests,
		"selected_tests": specialized.SelectedTests,
	} {
		if len(values) == 0 {
			return fmt.Errorf("specialized suite %q has no %s", specialized.Name, name)
		}
		previous := ""
		for _, value := range values {
			if value == "" || value <= previous {
				return fmt.Errorf("specialized suite %q %s must be unique and sorted", specialized.Name, name)
			}
			previous = value
		}
	}
	return nil
}

// SpecializedSelectedTests returns the one canonical required-test selection
// for an explicitly named specialized suite. Consumers such as required CI
// must derive this list from identities.json instead of copying it into a
// second runner-maintained array.
func SpecializedSelectedTests(manifest Manifest, name string) ([]string, error) {
	name = strings.TrimSpace(name)
	for _, specialized := range manifest.SpecializedSuites {
		if specialized.Name == name {
			return slices.Clone(specialized.SelectedTests), nil
		}
	}
	return nil, fmt.Errorf("specialized suite %q is not declared", name)
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
	known := map[string]bool{"e2e": true, "full": true, currentSourceTag: true}
	for _, identity := range manifest.Identities {
		known[identity.Name] = true
	}
	for _, specialized := range manifest.SpecializedSuites {
		for _, tag := range specialized.Tags {
			known[tag] = true
		}
	}
	for _, file := range suite.Files {
		if err := validateSourceFile(file, known, manifest.SpecializedSuites); err != nil {
			return err
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
	for _, specialized := range manifest.SpecializedSuites {
		owner := suite.SpecializedInventory(specialized)
		if len(owner.Tests) == 0 {
			return fmt.Errorf("specialized suite %q selects no tests", specialized.Name)
		}
		selected := suite.specializedSelectedInventory(specialized)
		if err := compareSpecializedSuite(specialized, owner, selected); err != nil {
			return err
		}
	}
	if err := compareIdentity(manifest.FullOnly, suite.FullOnly(manifest)); err != nil {
		return err
	}
	return nil
}

func validateSourceFile(file SourceFile, known map[string]bool, specializedSuites []SpecializedSuite) error {
	tags := expressionTags(file.Expr)
	for _, tag := range tags {
		if !known[tag] {
			return fmt.Errorf("%s: unknown build tag %q", file.Name, tag)
		}
	}
	if file.Expr.Eval(func(tag string) bool { return tag != "e2e" && known[tag] }) {
		return fmt.Errorf("%s: build expression does not require e2e", file.Name)
	}
	if specialized, ok := specializedSuiteForTags(tags, specializedSuites); ok {
		if err := validateSpecializedExpression(file, specialized); err != nil {
			return err
		}
		return nil
	}
	if len(file.Tests) == 0 {
		return nil
	}
	if file.Expr.Eval(func(tag string) bool { return tag == "e2e" }) {
		return fmt.Errorf("%s: runnable tests use suite-only support constraint", file.Name)
	}
	if slices.Contains(tags, currentSourceTag) {
		standardFull := func(tag string) bool { return tag == "e2e" || tag == "full" }
		if file.Expr.Eval(standardFull) {
			return fmt.Errorf("%s: current-source-only tests must be excluded from standard full", file.Name)
		}
		dedicated := func(tag string) bool {
			return tag == "e2e" || tag == "full" || tag == currentSourceTag
		}
		if !file.Expr.Eval(dedicated) {
			return fmt.Errorf("%s: current-source-only tests are missing dedicated tag inclusion", file.Name)
		}
		return nil
	}
	if !file.Expr.Eval(func(tag string) bool { return tag == "e2e" || tag == "full" }) {
		return fmt.Errorf("%s: runnable tests are missing full fallback", file.Name)
	}
	return nil
}

func specializedSuiteForTags(tags []string, specializedSuites []SpecializedSuite) (SpecializedSuite, bool) {
	for _, specialized := range specializedSuites {
		focused := append([]string{"e2e"}, specialized.Tags...)
		slices.Sort(focused)
		fullFallback := append(append([]string{}, focused...), "full")
		slices.Sort(fullFallback)
		if slices.Equal(tags, focused) || slices.Equal(tags, fullFallback) {
			return specialized, true
		}
	}
	return SpecializedSuite{}, false
}

func validateSpecializedExpression(file SourceFile, specialized SpecializedSuite) error {
	focused := make(map[string]bool, len(specialized.Tags)+1)
	focused["e2e"] = true
	for _, tag := range specialized.Tags {
		focused[tag] = true
	}
	if !file.Expr.Eval(func(tag string) bool { return focused[tag] }) {
		return fmt.Errorf("%s: specialized suite %q is not selected by its exact tags", file.Name, specialized.Name)
	}
	for _, required := range specialized.Tags {
		if file.Expr.Eval(func(tag string) bool { return tag != required && focused[tag] }) {
			return fmt.Errorf("%s: specialized suite %q does not require tag %q", file.Name, specialized.Name, required)
		}
	}
	return nil
}

// CurrentSourceInventory returns tests compiled by the dedicated root-plus-site
// E2E contract. It includes standard full tests because current-source-only
// files may use helpers declared behind the full suite tag.
func (suite Suite) CurrentSourceInventory() Identity {
	tags := map[string]bool{"e2e": true, "full": true, currentSourceTag: true}
	return suite.inventory(currentSourceTag, func(tag string) bool { return tags[tag] })
}

// ExpandedManifest fills the checked-in identity names with their parsed files and tests.
func ExpandedManifest(suite Suite, manifest Manifest) Manifest {
	expanded := Manifest{
		Identities:        make([]Identity, 0, len(manifest.Identities)),
		SpecializedSuites: make([]SpecializedSuite, 0, len(manifest.SpecializedSuites)),
	}
	for _, identity := range manifest.Identities {
		expanded.Identities = append(expanded.Identities, suite.Inventory(identity.Name))
	}
	for _, specialized := range manifest.SpecializedSuites {
		inventory := suite.SpecializedInventory(specialized)
		expanded.SpecializedSuites = append(expanded.SpecializedSuites, SpecializedSuite{
			Name:          specialized.Name,
			Tags:          slices.Clone(specialized.Tags),
			Files:         inventory.Files,
			Tests:         inventory.Tests,
			SelectedTests: suite.specializedSelectedInventory(specialized).Tests,
		})
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

func compareSpecializedSuite(expected SpecializedSuite, owner, selected Identity) error {
	if !slices.Equal(expected.Files, owner.Files) {
		return fmt.Errorf("specialized suite %q files mismatch: expected %v, got %v", expected.Name, expected.Files, owner.Files)
	}
	if !slices.Equal(expected.Tests, owner.Tests) {
		return fmt.Errorf("specialized suite %q tests mismatch: expected %v, got %v", expected.Name, expected.Tests, owner.Tests)
	}
	if !slices.Equal(expected.SelectedTests, selected.Tests) {
		return fmt.Errorf("specialized suite %q selected tests mismatch: expected %v, got %v", expected.Name, expected.SelectedTests, selected.Tests)
	}
	return nil
}

// Inventory returns the exact files and Test functions selected by one identity.
func (suite Suite) Inventory(identity string) Identity {
	tags := map[string]bool{"e2e": true, identity: true}
	return suite.inventory(identity, func(tag string) bool { return tags[tag] })
}

// SpecializedInventory returns tests compiled by exactly one explicit
// specialized tag set. The tags are supplied by the checked-in manifest rather
// than derived from an arbitrary caller command.
func (suite Suite) SpecializedInventory(specialized SpecializedSuite) Identity {
	result := Identity{Name: specialized.Name}
	for _, file := range suite.Files {
		owner, ok := specializedSuiteForTags(expressionTags(file.Expr), []SpecializedSuite{specialized})
		if !ok || owner.Name != specialized.Name {
			continue
		}
		result.Files = append(result.Files, file.Name)
		if len(file.Tests) > 0 {
			result.Tests = append(result.Tests, file.Tests...)
		}
	}
	slices.Sort(result.Tests)
	return result
}

func (suite Suite) specializedSelectedInventory(specialized SpecializedSuite) Identity {
	tags := make(map[string]bool, len(specialized.Tags)+1)
	tags["e2e"] = true
	for _, tag := range specialized.Tags {
		tags[tag] = true
	}
	return suite.inventory(specialized.Name, func(tag string) bool { return tags[tag] })
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
	for _, specialized := range manifest.SpecializedSuites {
		for _, test := range suite.specializedSelectedInventory(specialized).Tests {
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

type suiteMatrixSelection struct {
	Name string
	Tags []string
}

func suiteMatrixSelections(manifest Manifest) []suiteMatrixSelection {
	selections := make([]suiteMatrixSelection, 0, len(manifest.Identities)+len(manifest.SpecializedSuites)+1)
	for _, identity := range manifest.Identities {
		selections = append(selections, suiteMatrixSelection{Name: identity.Name, Tags: []string{"e2e", identity.Name}})
	}
	for _, specialized := range manifest.SpecializedSuites {
		selections = append(selections, suiteMatrixSelection{
			Name: specialized.Name,
			Tags: append([]string{"e2e"}, specialized.Tags...),
		})
	}
	return append(selections, suiteMatrixSelection{Name: "full", Tags: []string{"e2e", "full"}})
}

func matrixInventory(suite Suite, manifest Manifest, selection suiteMatrixSelection) Identity {
	if selection.Name == "full" {
		return suite.FullInventory()
	}
	for _, specialized := range manifest.SpecializedSuites {
		if specialized.Name == selection.Name {
			return Identity{Name: specialized.Name, Tests: slices.Clone(specialized.SelectedTests)}
		}
	}
	return suite.Inventory(selection.Name)
}

// RunCompileMatrix asks the Go compiler to build every focused, specialized,
// and full E2E suite recorded by the manifest.
func RunCompileMatrix(ctx context.Context, siteDir string, manifest Manifest, output io.Writer) error {
	temporary, err := os.MkdirTemp("", "goshtoso-e2e-matrix-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	for _, selection := range suiteMatrixSelections(manifest) {
		tagSet := strings.Join(selection.Tags, ",")
		if _, err := fmt.Fprintf(output, "compile %s\n", tagSet); err != nil {
			return fmt.Errorf("report compile identity: %w", err)
		}
		command := exec.CommandContext(ctx, "go", "test", "-c", "-tags="+tagSet,
			"-o", filepath.Join(temporary, selection.Name+".test"), "./tests/e2e")
		command.Dir = siteDir
		command.Stdout = output
		command.Stderr = output
		if err := command.Run(); err != nil {
			return fmt.Errorf("compile %s: %w", tagSet, err)
		}
	}
	return nil
}

// RunListMatrix compares Go's real test inventory with the parsed constraints.
func RunListMatrix(ctx context.Context, siteDir string, suite Suite, manifest Manifest, output io.Writer) error {
	for _, selection := range suiteMatrixSelections(manifest) {
		tagSet := strings.Join(selection.Tags, ",")
		if _, err := fmt.Fprintf(output, "list %s\n", tagSet); err != nil {
			return fmt.Errorf("report list identity: %w", err)
		}
		command := exec.CommandContext(ctx, "go", "test", "-tags="+tagSet, "-list", "^Test", "./tests/e2e")
		command.Dir = siteDir
		command.Env = append(os.Environ(), "GOSHTOSO_E2E_LIST_ONLY=1")
		data, err := command.CombinedOutput()
		if err != nil {
			_, _ = output.Write(data)
			return fmt.Errorf("list %s: %w", tagSet, err)
		}
		actual := parseGoTestList(string(data))
		expected := matrixInventory(suite, manifest, selection).Tests
		if !slices.Equal(expected, actual) {
			return fmt.Errorf("%s list mismatch: expected %v, got %v", tagSet, expected, actual)
		}
	}
	return nil
}

func parseGoTestList(output string) []string {
	var tests []string
	for line := range strings.SplitSeq(output, "\n") {
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
