// Package schematree renders read-only, format-neutral schema documentation.
// Callers resolve references and convert their schema model before rendering.
package schematree

import "github.com/a-h/templ"

// Constraint is one named annotation, such as a default, pattern, or example.
type Constraint struct {
	Name  string
	Value string
}

// Node describes a field, array item, reference, or schema alternative.
// Children are rendered in caller order. Reference resolution and cycle limits
// belong to the caller; the component never fetches or parses schemas.
type Node struct {
	Name string
	// Path is an optional full property path, exposed as the name's tooltip.
	Path string
	Type string
	// TypeContent replaces Type, for example with a linked named type.
	TypeContent templ.Component
	Description string
	// DescriptionContent replaces Description when non-nil. It accepts block
	// content, including styled Markdown. The caller owns HTML sanitization.
	DescriptionContent templ.Component
	Required           bool
	// ShowOptional also labels fields that are not required. Default is false.
	ShowOptional bool
	Nullable     bool
	Deprecated   bool
	Constraints  []Constraint
	// MetadataContent follows Constraints and can contain examples or actions.
	MetadataContent templ.Component
	Children        []Node
	// ChildrenContent follows Children, supporting caller-owned lazy fragments.
	ChildrenContent templ.Component
	// Collapsed initially closes a branch. Native disclosure works without JS.
	Collapsed bool
	RootClass string
	RootAttrs templ.Attributes
}

// Config defines a schema tree. Zero nodes render EmptyLabel when supplied.
// The surrounding document owns its heading level through Header.
type Config struct {
	ID          string
	AriaLabel   string
	Header      templ.Component
	Description string
	// DescriptionContent replaces Description when non-nil.
	DescriptionContent templ.Component
	Nodes              []Node
	EmptyLabel         string
	RootClass          string
	RootAttrs          templ.Attributes
}
