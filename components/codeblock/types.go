package codeblock

import "fmt"

var idCounter int

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

// GetID returns a unique ID for the code element
func (cfg Config) GetID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	idCounter++
	return fmt.Sprintf("codeblock-%d", idCounter)
}

// GetLabel returns the header label, defaulting to the language name
func (cfg Config) GetLabel() string {
	if cfg.Label != "" {
		return cfg.Label
	}
	return cfg.Language
}

func (cfg Config) maxHeightStyle() string {
	if cfg.MaxHeight != "" {
		return "max-height: " + cfg.MaxHeight + "; overflow-y: auto;"
	}
	return ""
}
