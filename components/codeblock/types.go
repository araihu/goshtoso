package codeblock

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"unicode"
)

// Config holds configuration for the code block component
type Config struct {
	// Language selects the Chroma lexer (e.g. "go", "bash", "html", "css").
	// "templ" aliases to the Go lexer. Unknown values fall back to plain text.
	Language string
	// Code is the source code to display
	Code string
	// Label is the header text (defaults to Language if empty)
	Label string
	// MaxHeight is an optional CSS max-height for scrollable long code (e.g. "400px")
	MaxHeight string
	// ID overrides the auto-generated element ID
	ID string
}

// GetID returns a stable ID for the code element.
func (cfg Config) getID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	base := slugPart(cfg.getLabel())
	if base == "" {
		base = slugPart(cfg.Language)
	}
	if base == "" {
		base = "snippet"
	}

	sum := sha1.Sum([]byte(cfg.Language + "\x00" + cfg.Label + "\x00" + cfg.Code))
	return "codeblock-" + base + "-" + hex.EncodeToString(sum[:])[:10]
}

// GetLabel returns the header label, defaulting to the language name.
func (cfg Config) getLabel() string {
	if cfg.Label != "" {
		return cfg.Label
	}
	return cfg.Language
}

func (cfg Config) copyLabel() string {
	label := cfg.getLabel()
	if label == "" {
		label = cfg.getID()
	}
	return "Copy " + label + " code"
}

func (cfg Config) maxHeightStyle() string {
	if cfg.MaxHeight != "" {
		return "max-height: " + cfg.MaxHeight + "; overflow-y: auto;"
	}
	return ""
}

func slugPart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case !lastDash && b.Len() > 0:
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func jsStringSingle(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\u2028':
			b.WriteString(`\u2028`)
		case '\u2029':
			b.WriteString(`\u2029`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
