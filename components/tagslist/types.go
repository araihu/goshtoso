package tagslist

import "strings"

// Config holds configuration for the TagsList component.
// TagsList renders a list of removable tag chips plus an input + Add button
// for appending new tags. Values submit as name[0], name[1], ... via hidden inputs.
type Config struct {
	// ID is the element ID for the outer container
	ID string
	// Name is the form field name prefix (submits as name[0], name[1], ...)
	Name string
	// Values is the initial list of tag values
	Values []string
	// Placeholder is shown in the add-tag input
	Placeholder string
	// AddActionLabel is the add button text (default: "Add")
	AddActionLabel string
	// Disabled renders chips only (no remove buttons, no input row)
	Disabled bool
	// RootClass allows additional CSS classes on the outer container.
	RootClass string
}

// GetAddLabel returns the add button label with default
func (c Config) getAddLabel() string {
	if c.AddActionLabel != "" {
		return c.AddActionLabel
	}
	return "Add"
}

// GetPlaceholder returns the input placeholder with default
func (c Config) getPlaceholder() string {
	if c.Placeholder != "" {
		return c.Placeholder
	}
	return "Add a tag..."
}

// AlpineData returns the x-data expression for Alpine.js initialization.
// Uses unquoted JS keys + single-quoted strings to survive templ HTML escaping.
func (c Config) alpineData() string {
	var sb strings.Builder
	sb.WriteString("{ items: [")
	for i, v := range c.Values {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("'")
		sb.WriteString(jsEscapeSingle(v))
		sb.WriteString("'")
	}
	sb.WriteString("], newTag: '', name: '")
	sb.WriteString(jsEscapeSingle(c.Name))
	sb.WriteString("', addTag() { const v = this.newTag.trim(); if (v) { this.items.push(v); this.newTag = ''; } } }")
	return sb.String()
}

// ContainerClasses returns CSS classes for the outer container
func (c Config) containerClasses() string {
	base := "flex flex-col gap-2"
	if c.RootClass != "" {
		return base + " " + c.RootClass
	}
	return base
}

// jsEscapeSingle escapes a string for safe use inside a single-quoted JS string
// embedded in an HTML attribute. Escapes backslash, single quote, and newlines.
func jsEscapeSingle(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\\':
			sb.WriteString(`\\`)
		case '\'':
			sb.WriteString(`\'`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		case '\u2028':
			sb.WriteString(`\u2028`)
		case '\u2029':
			sb.WriteString(`\u2029`)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
