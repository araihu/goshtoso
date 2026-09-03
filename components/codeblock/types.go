package codeblock

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"unicode"
)

// Density controls spacing inside code block regions.
type Density string

const (
	// DensityDefault uses standard documentation spacing.
	DensityDefault Density = ""
	// DensityCompact reduces header and code padding for short snippets.
	DensityCompact Density = "compact"
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
	// Density controls spacing in the header and highlighted code body.
	Density Density
	// ID overrides the auto-generated element ID
	ID string
	// DisableCopyButton omits the copy control and all copy-runtime hooks.
	// The zero value preserves the default copy control.
	DisableCopyButton bool
}

func (cfg Config) headerClasses() string {
	spacing := "px-4 py-2"
	if cfg.Density == DensityCompact {
		spacing = "px-3 py-1.5"
	}
	return "flex items-center justify-between " + spacing + " bg-surface-alt dark:bg-surface-dark-alt border-b border-outline dark:border-outline-dark"
}

func (cfg Config) bodyClasses() string {
	classes := "codeblock overflow-x-auto"
	if cfg.Density == DensityCompact {
		classes = "codeblock codeblock-compact overflow-x-auto"
	}
	return classes
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
