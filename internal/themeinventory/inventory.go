// Package themeinventory parses the built-in theme registry embedded in
// all-themes.css. It is tooling-only: runtime consumers use generated CSS and
// the public themes catalog.
package themeinventory

import (
	"fmt"
	"regexp"
	"strings"
)

// Ownership classifies whether a theme represents an Arai Hû organization
// identity or is a generic design-system theme.
type Ownership string

const (
	OwnershipGeneric      Ownership = "generic"
	OwnershipOrganization Ownership = "organization"
)

// Theme is one annotated canonical selector from all-themes.css.
type Theme struct {
	Key       string
	Label     string
	Ownership Ownership
	Block     string
}

var (
	metadataLine = regexp.MustCompile(`(?m)^[\t ]*/\*[\t ]*goshtoso-theme:[\t ]*([^|\r\n]+)[\t ]*\|[\t ]*([^|\r\n]+)[\t ]*\|[\t ]*([^*\r\n]+?)[\t ]*\*/[\t ]*$`)
	selectorLine = regexp.MustCompile(`(?m)^[\t ]*\[data-theme=([a-z0-9][a-z0-9-]*)\][\t ]*\{`)
	selectorHead = regexp.MustCompile(`^\[data-theme=([a-z0-9][a-z0-9-]*)\][\t ]*\{`)
	validKey     = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

// Parse returns the canonical built-in themes in stylesheet order. Every
// canonical [data-theme=key] block must have exactly one immediately preceding
// goshtoso-theme metadata comment, and metadata keys must be unique.
func Parse(source []byte) ([]Theme, error) {
	text := string(source)
	metadataMatches := metadataLine.FindAllStringSubmatchIndex(text, -1)
	selectorMatches := selectorLine.FindAllStringSubmatch(text, -1)

	selectorCounts := make(map[string]int, len(selectorMatches))
	for _, match := range selectorMatches {
		selectorCounts[match[1]]++
	}

	themes := make([]Theme, 0, len(metadataMatches))
	metadataKeys := make(map[string]struct{}, len(metadataMatches))
	for _, match := range metadataMatches {
		key := strings.TrimSpace(text[match[2]:match[3]])
		label := strings.TrimSpace(text[match[4]:match[5]])
		ownership := Ownership(strings.TrimSpace(text[match[6]:match[7]]))

		if !validKey.MatchString(key) {
			return nil, fmt.Errorf("theme metadata has invalid key %q", key)
		}
		if label == "" {
			return nil, fmt.Errorf("theme %q has empty label", key)
		}
		if ownership != OwnershipGeneric && ownership != OwnershipOrganization {
			return nil, fmt.Errorf("theme %q has invalid ownership %q", key, ownership)
		}
		if _, duplicate := metadataKeys[key]; duplicate {
			return nil, fmt.Errorf("duplicate theme metadata key %q", key)
		}
		metadataKeys[key] = struct{}{}

		rest := text[match[1]:]
		trimmed := strings.TrimLeft(rest, " \t\r\n")
		selector := selectorHead.FindStringSubmatchIndex(trimmed)
		if selector == nil {
			return nil, fmt.Errorf("theme metadata %q is not followed by a canonical selector", key)
		}
		selectorKey := trimmed[selector[2]:selector[3]]
		if selectorKey != key {
			return nil, fmt.Errorf("metadata key %q does not match selector %q", key, selectorKey)
		}

		selectorStart := match[1] + len(rest) - len(trimmed)
		openingBrace := selectorStart + strings.Index(text[selectorStart:], "{")
		closingBrace, err := findClosingBrace(text, openingBrace)
		if err != nil {
			return nil, fmt.Errorf("theme %q: %w", key, err)
		}
		themes = append(themes, Theme{
			Key:       key,
			Label:     label,
			Ownership: ownership,
			Block:     normalizeBlock(text, selectorStart, closingBrace+1),
		})
	}

	for key, count := range selectorCounts {
		if count > 1 {
			return nil, fmt.Errorf("duplicate canonical selector %q", key)
		}
		if _, ok := metadataKeys[key]; !ok {
			return nil, fmt.Errorf("selector %q has no theme metadata", key)
		}
	}
	for key := range metadataKeys {
		if selectorCounts[key] == 0 {
			return nil, fmt.Errorf("theme metadata %q has no canonical selector", key)
		}
	}
	if len(themes) == 0 {
		return nil, fmt.Errorf("theme inventory is empty")
	}
	return themes, nil
}

func findClosingBrace(source string, opening int) (int, error) {
	depth := 0
	for index := opening; index < len(source); index++ {
		switch source[index] {
		case '/', '*':
			if source[index] == '/' && index+1 < len(source) && source[index+1] == '*' {
				end := strings.Index(source[index+2:], "*/")
				if end < 0 {
					return 0, fmt.Errorf("unterminated CSS comment")
				}
				index += end + 3
			}
		case '\'', '"':
			quote := source[index]
			for index++; index < len(source); index++ {
				if source[index] == '\\' {
					index++
					continue
				}
				if source[index] == quote {
					break
				}
			}
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return index, nil
			}
		}
	}
	return 0, fmt.Errorf("unterminated selector block")
}

func normalizeBlock(source string, start, end int) string {
	lineStart := strings.LastIndex(source[:start], "\n") + 1
	indent := source[lineStart:start]
	lines := strings.Split(source[start:end], "\n")
	for index := 1; index < len(lines); index++ {
		lines[index] = strings.TrimPrefix(lines[index], indent)
	}
	return strings.Join(lines, "\n")
}
