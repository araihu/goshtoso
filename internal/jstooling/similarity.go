package jstooling

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const similarityShingleSize = 5

var (
	functionDeclarationPattern = regexp.MustCompile(`\bfunction\s*([A-Za-z_$][A-Za-z0-9_$]*)?\s*\([^)]*\)\s*\{`)
	arrowFunctionPattern       = regexp.MustCompile(`\b([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*(?:\([^)]*\)|[A-Za-z_$][A-Za-z0-9_$]*)\s*=>\s*\{`)
	javascriptTokenPattern     = regexp.MustCompile("(?s)\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'|`(?:\\\\.|[^`\\\\])*`|[A-Za-z_$][A-Za-z0-9_$]*|[0-9]+(?:\\.[0-9]+)?|===|!==|=>|==|!=|<=|>=|&&|\\|\\||\\+\\+|--|[{}()\\[\\].,:;+\\-*/%<>=!?&|]")
)

type FunctionRef struct {
	Path string
	Line int
	Name string
}

type SimilarityPair struct {
	Left  FunctionRef
	Right FunctionRef
	Score float64
}

type extractedFunction struct {
	ref    FunctionRef
	tokens []string
	start  int
	end    int
}

// SimilarFunctions reports pairs at or above threshold using a deterministic
// five-token shingle Dice coefficient over normalized function bodies.
func SimilarFunctions(sources map[string][]byte, threshold float64) []SimilarityPair {
	paths := make([]string, 0, len(sources))
	for path := range sources {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var functions []extractedFunction
	for _, path := range paths {
		functions = append(functions, extractFunctions(path, sources[path])...)
	}
	var pairs []SimilarityPair
	for left := 0; left < len(functions); left++ {
		for right := left + 1; right < len(functions); right++ {
			if functions[left].ref.Path == functions[right].ref.Path && rangesOverlap(functions[left], functions[right]) {
				continue
			}
			score := dice(functions[left].tokens, functions[right].tokens)
			if score >= threshold {
				pairs = append(pairs, SimilarityPair{Left: functions[left].ref, Right: functions[right].ref, Score: score})
			}
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Score != pairs[j].Score {
			return pairs[i].Score > pairs[j].Score
		}
		leftI := pairs[i].Left.Path + ":" + strconv.Itoa(pairs[i].Left.Line) + ":" + pairs[i].Left.Name
		leftJ := pairs[j].Left.Path + ":" + strconv.Itoa(pairs[j].Left.Line) + ":" + pairs[j].Left.Name
		if leftI != leftJ {
			return leftI < leftJ
		}
		rightI := pairs[i].Right.Path + ":" + strconv.Itoa(pairs[i].Right.Line) + ":" + pairs[i].Right.Name
		rightJ := pairs[j].Right.Path + ":" + strconv.Itoa(pairs[j].Right.Line) + ":" + pairs[j].Right.Name
		return rightI < rightJ
	})
	return pairs
}

func extractFunctions(path string, source []byte) []extractedFunction {
	masked := maskJavaScript(source)
	var functions []extractedFunction
	for _, candidate := range []struct {
		pattern *regexp.Regexp
		arrow   bool
	}{{functionDeclarationPattern, false}, {arrowFunctionPattern, true}} {
		for _, match := range candidate.pattern.FindAllSubmatchIndex(masked, -1) {
			open := match[1] - 1
			close := matchingBrace(masked, open)
			if close < 0 {
				continue
			}
			name := "anonymous"
			if match[2] >= 0 && match[3] >= 0 {
				name = string(source[match[2]:match[3]])
			}
			tokens := normalizeJavaScript(source[open+1 : close])
			if len(tokens) < similarityShingleSize*2 {
				continue
			}
			functions = append(functions, extractedFunction{
				ref:    FunctionRef{Path: path, Line: lineAt(source, match[0]), Name: name},
				tokens: tokens,
				start:  match[0],
				end:    close,
			})
		}
	}
	return functions
}

func rangesOverlap(left, right extractedFunction) bool {
	return left.start <= right.end && right.start <= left.end
}

func maskJavaScript(source []byte) []byte {
	masked := append([]byte(nil), source...)
	quote := byte(0)
	escaped := false
	lineComment := false
	blockComment := false
	for index := 0; index < len(masked); index++ {
		current := source[index]
		next := byte(0)
		if index+1 < len(source) {
			next = source[index+1]
		}
		if lineComment {
			if current == '\n' {
				lineComment = false
			} else {
				masked[index] = ' '
			}
			continue
		}
		if blockComment {
			masked[index] = ' '
			if current == '*' && next == '/' {
				masked[index+1] = ' '
				blockComment = false
				index++
			}
			continue
		}
		if quote != 0 {
			if current != '\n' {
				masked[index] = ' '
			}
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == quote {
				quote = 0
			}
			continue
		}
		if current == '/' && next == '/' {
			masked[index] = ' '
			masked[index+1] = ' '
			lineComment = true
			index++
			continue
		}
		if current == '/' && next == '*' {
			masked[index] = ' '
			masked[index+1] = ' '
			blockComment = true
			index++
			continue
		}
		if current == '\'' || current == '"' || current == '`' {
			masked[index] = ' '
			quote = current
		}
	}
	return masked
}

func normalizeJavaScript(body []byte) []string {
	keywords := map[string]struct{}{
		"break": {}, "case": {}, "catch": {}, "const": {}, "continue": {},
		"default": {}, "delete": {}, "do": {}, "else": {}, "false": {},
		"finally": {}, "for": {}, "function": {}, "if": {}, "in": {},
		"instanceof": {}, "let": {}, "new": {}, "null": {}, "return": {},
		"switch": {}, "this": {}, "throw": {}, "true": {}, "try": {},
		"typeof": {}, "undefined": {}, "var": {}, "void": {}, "while": {},
	}
	raw := javascriptTokenPattern.FindAllString(string(body), -1)
	tokens := make([]string, 0, len(raw))
	for _, current := range raw {
		switch {
		case current[0] == '\'' || current[0] == '"' || current[0] == '`':
			tokens = append(tokens, "str")
		case current[0] >= '0' && current[0] <= '9':
			tokens = append(tokens, "num")
		case isIdentifier(current):
			if _, ok := keywords[current]; ok {
				tokens = append(tokens, current)
			} else {
				tokens = append(tokens, "id")
			}
		default:
			tokens = append(tokens, current)
		}
	}
	return tokens
}

func dice(left, right []string) float64 {
	leftShingles := shingles(left)
	rightShingles := shingles(right)
	if len(leftShingles) == 0 || len(rightShingles) == 0 {
		return 0
	}
	intersection := 0
	remaining := make(map[string]int, len(rightShingles))
	for _, shingle := range rightShingles {
		remaining[shingle]++
	}
	for _, shingle := range leftShingles {
		if remaining[shingle] > 0 {
			intersection++
			remaining[shingle]--
		}
	}
	return 2 * float64(intersection) / float64(len(leftShingles)+len(rightShingles))
}

func shingles(tokens []string) []string {
	if len(tokens) < similarityShingleSize {
		return nil
	}
	result := make([]string, 0, len(tokens)-similarityShingleSize+1)
	for index := 0; index <= len(tokens)-similarityShingleSize; index++ {
		result = append(result, strings.Join(tokens[index:index+similarityShingleSize], "\x00"))
	}
	return result
}
