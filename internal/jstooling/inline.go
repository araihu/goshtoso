package jstooling

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"go/scanner"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const maxInlineExpressionBytes = 80

var (
	scriptPattern          = regexp.MustCompile(`(?is)<script(?:\s+([^>]*))?>(.*?)</script\s*>`)
	doubleAttributePattern = regexp.MustCompile(`(?is)(x-(?:data|init|effect|show|model|bind(?::[[:alnum:]_.:-]+)?|on:[[:alnum:]_.:-]+)|@[[:alnum:]_.:-]+|on(?:click|change|input|load|submit|keydown|keyup))\s*=\s*"([^"]*)"`)
	singleAttributePattern = regexp.MustCompile(`(?is)(x-(?:data|init|effect|show|model|bind(?::[[:alnum:]_.:-]+)?|on:[[:alnum:]_.:-]+)|@[[:alnum:]_.:-]+|on(?:click|change|input|load|submit|keydown|keyup))\s*=\s*'([^']*)'`)
	functionPattern        = regexp.MustCompile(`\bfunc\s+(?:\([^)]*\)\s*)?([[:alnum:]_]+)\s*\(`)
)

// Finding identifies JavaScript that should move from Go/templ markup into an
// authored first-party source file.
type Finding struct {
	Path    string
	Line    int
	Kind    string
	Summary string
}

// Key is stable across unrelated line shifts and suitable for an inventory
// baseline. Duplicate keys remain duplicate baseline lines.
func (finding Finding) Key() string {
	digest := sha256.Sum256([]byte(finding.Kind + "\x00" + finding.Path + "\x00" + finding.Summary))
	return fmt.Sprintf("%s|%s|%x", finding.Path, finding.Kind, digest[:8])
}

// DetectInlineJavaScript applies the extraction policy to one source file.
// Executable inline script blocks are always findings. Alpine/event expressions
// are allowed only when single-line and at most 80 bytes. Multiline Go string
// builders with JavaScript browser markers are findings.
func DetectInlineJavaScript(path string, source []byte) []Finding {
	findings := detectScriptBodies(path, source)
	findings = append(findings, detectAttributes(path, source, doubleAttributePattern)...)
	findings = append(findings, detectAttributes(path, source, singleAttributePattern)...)
	findings = append(findings, detectBuilders(path, source)...)
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		if findings[i].Kind != findings[j].Kind {
			return findings[i].Kind < findings[j].Kind
		}
		return findings[i].Summary < findings[j].Summary
	})
	return findings
}

func detectScriptBodies(path string, source []byte) []Finding {
	var findings []Finding
	searchable := maskSourceComments(source)
	for _, match := range scriptPattern.FindAllSubmatchIndex(searchable, -1) {
		attributes := ""
		if match[2] >= 0 {
			attributes = strings.ToLower(string(source[match[2]:match[3]]))
		}
		body := source[match[4]:match[5]]
		if strings.Contains(attributes, "src=") || strings.Contains(attributes, "application/ld+json") || strings.Contains(attributes, "application/json") || len(strings.TrimSpace(string(body))) == 0 {
			continue
		}
		kind := "script-body"
		contextStart := max(0, match[0]-512)
		prefix := source[contextStart:match[0]]
		if rawIndex := bytesLastIndex(prefix, []byte("templ.Raw")); rawIndex >= 0 {
			following := prefix[rawIndex:]
			if !bytesContains(following, []byte("\n}")) {
				kind = "templ-raw-script"
			}
		}
		findings = append(findings, Finding{
			Path: path, Line: lineAt(source, match[0]), Kind: kind,
			Summary: summarize(string(body)),
		})
	}
	return findings
}

func detectAttributes(path string, source []byte, pattern *regexp.Regexp) []Finding {
	var findings []Finding
	searchable := maskSourceComments(source)
	for _, match := range pattern.FindAllSubmatchIndex(searchable, -1) {
		value := string(source[match[4]:match[5]])
		trimmed := strings.TrimSpace(value)
		if !strings.Contains(value, "\n") && len(trimmed) <= maxInlineExpressionBytes {
			continue
		}
		findings = append(findings, Finding{
			Path: path, Line: lineAt(source, match[0]), Kind: "event-expression",
			Summary: summarize(trimmed),
		})
	}
	return findings
}

func detectBuilders(path string, source []byte) []Finding {
	var findings []Finding
	for _, match := range functionPattern.FindAllSubmatchIndex(source, -1) {
		open := bytesIndexByte(source, match[1], '{')
		if open < 0 {
			continue
		}
		close := matchingBrace(source, open)
		if close < 0 {
			continue
		}
		body := source[open+1 : close]
		literals, markers, strongMarkers := javascriptStringEvidence(body)
		if literals < 2 || markers < 2 || strongMarkers < 2 || bytesCount(body, '\n') < 2 {
			continue
		}
		name := string(source[match[2]:match[3]])
		findings = append(findings, Finding{
			Path: path, Line: lineAt(source, match[0]), Kind: "js-builder",
			Summary: "func " + name + " builds multiline JavaScript",
		})
	}
	return findings
}

func javascriptStringEvidence(body []byte) (int, int, int) {
	var scan scanner.Scanner
	file := token.NewFileSet().AddFile("builder.go", -1, len(body))
	scan.Init(file, body, nil, scanner.ScanComments)
	literals := 0
	markers := 0
	strongMarkers := 0
	for {
		_, tok, literal := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok != token.STRING {
			continue
		}
		decoded, err := strconv.Unquote(literal)
		if err != nil {
			continue
		}
		literals++
		for _, marker := range []string{"document.", "window.", "Alpine.", "htmx.", "addEventListener", "querySelector", "=>", "function (", "function("} {
			if strings.Contains(decoded, marker) {
				markers++
			}
		}
		for _, marker := range []string{"document.", "window.", "addEventListener", "querySelector", "=>", ":function", "this.", "() {", "){"} {
			if strings.Contains(decoded, marker) {
				strongMarkers++
			}
		}
	}
	return literals, markers, strongMarkers
}

// ScanInlineJavaScript scans tracked source-shaped files in stable order.
func ScanInlineJavaScript(root string) ([]Finding, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		include, skipDirectory := includeInlineScanPath(root, path, entry)
		if skipDirectory {
			return filepath.SkipDir
		}
		if !include {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	var findings []Finding
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		findings = append(findings, DetectInlineJavaScript(filepath.ToSlash(relative), content)...)
	}
	return findings, nil
}

func includeInlineScanPath(root, path string, entry os.DirEntry) (bool, bool) {
	if entry.IsDir() {
		name := entry.Name()
		skip := path != root && (strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor")
		return false, skip
	}
	pathSlash := filepath.ToSlash(path)
	excluded := strings.HasSuffix(path, "_templ.go") || strings.HasSuffix(path, "_test.go") ||
		strings.Contains(pathSlash, "/fixtures/") || strings.HasPrefix(pathSlash, "internal/jstooling/") ||
		strings.Contains(pathSlash, "/internal/jstooling/") || strings.HasPrefix(pathSlash, "site/tests/") ||
		strings.Contains(pathSlash, "/site/tests/")
	if excluded {
		return false, false
	}
	extension := filepath.Ext(path)
	return extension == ".go" || extension == ".templ" || extension == ".html", false
}

func maskSourceComments(source []byte) []byte {
	masked := append([]byte(nil), source...)
	state := sourceMaskState{}
	for index := range source {
		if state.consumeActive(source, masked, index) {
			continue
		}
		state.start(source, masked, index)
	}
	return masked
}

type sourceMaskState struct {
	quote        byte
	escaped      bool
	lineComment  bool
	blockComment bool
	htmlComment  bool
}

func (state *sourceMaskState) consumeActive(source, masked []byte, index int) bool {
	current := source[index]
	if state.lineComment {
		if current == '\n' {
			state.lineComment = false
		} else {
			masked[index] = ' '
		}
		return true
	}
	if state.blockComment {
		masked[index] = ' '
		state.blockComment = index == 0 || source[index-1] != '*' || current != '/'
		return true
	}
	if state.htmlComment {
		masked[index] = ' '
		state.htmlComment = index < 2 || string(source[index-2:index+1]) != "-->"
		return true
	}
	if state.quote == 0 {
		return false
	}
	if state.escaped {
		state.escaped = false
	} else if current == '\\' && state.quote != '`' {
		state.escaped = true
	} else if current == state.quote {
		state.quote = 0
	}
	return true
}

func (state *sourceMaskState) start(source, masked []byte, index int) {
	current := source[index]
	if current == '\'' || current == '"' || current == '`' {
		state.quote = current
		return
	}
	if index+1 < len(source) && current == '/' && source[index+1] == '/' {
		masked[index] = ' '
		state.lineComment = true
		return
	}
	if index+1 < len(source) && current == '/' && source[index+1] == '*' {
		masked[index] = ' '
		state.blockComment = true
		return
	}
	if index+3 < len(source) && string(source[index:index+4]) == "<!--" {
		masked[index] = ' '
		state.htmlComment = true
	}
}

// ReadBaseline reads stable finding keys. Duplicate lines represent duplicate
// findings with the same content fingerprint.
func ReadBaseline(path string) (map[string]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	baseline := map[string]int{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		baseline[line]++
	}
	return baseline, scanner.Err()
}

// NewFindings subtracts the content baseline as a multiset.
func NewFindings(findings []Finding, baseline map[string]int) []Finding {
	remaining := make(map[string]int, len(baseline))
	maps.Copy(remaining, baseline)
	var result []Finding
	for _, finding := range findings {
		key := finding.Key()
		if remaining[key] > 0 {
			remaining[key]--
			continue
		}
		result = append(result, finding)
	}
	return result
}

func summarize(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 120 {
		return value[:117] + "..."
	}
	return value
}

func lineAt(source []byte, offset int) int {
	return 1 + bytesCount(source[:min(offset, len(source))], '\n')
}

func matchingBrace(source []byte, open int) int {
	masked := maskJavaScript(source)
	depth := 0
	for index := open; index < len(masked); index++ {
		current := masked[index]
		switch current {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func bytesCount(value []byte, target byte) int {
	count := 0
	for _, current := range value {
		if current == target {
			count++
		}
	}
	return count
}

func bytesIndexByte(value []byte, start int, target byte) int {
	for index := start; index < len(value); index++ {
		if value[index] == target {
			return index
		}
		if value[index] == '\n' && index-start > 500 {
			return -1
		}
	}
	return -1
}

func bytesLastIndex(value, target []byte) int {
	return strings.LastIndex(string(value), string(target))
}

func bytesContains(value, target []byte) bool {
	return strings.Contains(string(value), string(target))
}

func isIdentifier(value string) bool {
	for index, current := range value {
		if index == 0 && !unicode.IsLetter(current) && current != '_' && current != '$' {
			return false
		}
		if index > 0 && !unicode.IsLetter(current) && !unicode.IsDigit(current) && current != '_' && current != '$' {
			return false
		}
	}
	return value != ""
}
