package demo

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/araihu/goshtoso/components"
)

// APIPropDoc documents one struct field, functional option, constructor, or
// deliberate non-render helper in a structured API section.
type APIPropDoc struct {
	Name        string
	Signature   string
	Default     string
	Allowed     []string
	Description string
	Required    bool
}

// APISection describes one named public surface in a component API reference.
type APISection struct {
	ID          string
	Title       string
	Description string
	Constructor string
	Kind        components.Kind
	StructType  reflect.Type
	Props       []APIPropDoc
}

// StructAPI describes a public struct. Field signatures are derived from
// reflection so documentation cannot silently drift from the Go type.
func StructAPI[T any](
	kind components.Kind,
	title string,
	constructor string,
	description string,
	props []APIPropDoc,
) APISection {
	typ := reflect.TypeFor[T]()
	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("demo: StructAPI[%s] requires a struct type", typ))
	}
	validateAPIProps(props, false)

	return newAPISection(kind, title, constructor, description, typ, props)
}

// OptionsAPI describes a constructor configured by functional options.
func OptionsAPI(
	kind components.Kind,
	title string,
	constructor string,
	description string,
	props []APIPropDoc,
) APISection {
	validateAPIProps(props, true)
	return newAPISection(kind, title, constructor, description, nil, props)
}

// FunctionsAPI describes constructors and deliberate non-render helpers that
// have explicit function signatures rather than reflected struct fields.
func FunctionsAPI(
	kind components.Kind,
	title string,
	constructor string,
	description string,
	props []APIPropDoc,
) APISection {
	validateAPIProps(props, true)
	return newAPISection(kind, title, constructor, description, nil, props)
}

func newAPISection(
	kind components.Kind,
	title string,
	constructor string,
	description string,
	structType reflect.Type,
	props []APIPropDoc,
) APISection {
	id := slugify(title)
	if id == "" {
		panic("demo: API section title must produce a non-empty ID")
	}

	return APISection{
		ID:          id,
		Title:       title,
		Description: description,
		Constructor: constructor,
		Kind:        kind,
		StructType:  structType,
		Props:       cloneAPIProps(props),
	}
}

func validateAPIProps(props []APIPropDoc, explicitSignatures bool) {
	seen := make(map[string]struct{}, len(props))
	for _, prop := range props {
		if prop.Name == "" {
			panic("demo: API prop name must not be empty")
		}
		if _, duplicate := seen[prop.Name]; duplicate {
			panic(fmt.Sprintf("demo: duplicate API prop %q", prop.Name))
		}
		seen[prop.Name] = struct{}{}

		if explicitSignatures && prop.Signature == "" {
			panic(fmt.Sprintf("demo: API prop %q requires an explicit signature", prop.Name))
		}
		if !explicitSignatures && prop.Signature != "" {
			panic(fmt.Sprintf("demo: reflected struct field %q must not declare a signature", prop.Name))
		}
	}
}

func cloneAPIProps(props []APIPropDoc) []APIPropDoc {
	cloned := slices.Clone(props)
	for i := range cloned {
		cloned[i].Allowed = slices.Clone(cloned[i].Allowed)
	}
	return cloned
}

func normalizedAPISections(sections []APISection) []APISection {
	normalized := slices.Clone(sections)
	seen := make(map[string]int, len(normalized))
	for i := range normalized {
		normalized[i].Props = cloneAPIProps(normalized[i].Props)

		base := normalized[i].ID
		if base == "" {
			base = slugify(normalized[i].Title)
		}
		if base == "" {
			base = "api-section"
		}

		seen[base]++
		normalized[i].ID = base
		if seen[base] > 1 {
			normalized[i].ID = fmt.Sprintf("%s-%d", base, seen[base])
		}
	}
	return normalized
}

func apiPropSignature(section APISection, prop APIPropDoc) string {
	if prop.Signature != "" {
		return prop.Signature
	}
	if section.StructType == nil || section.StructType.Kind() != reflect.Struct {
		return ""
	}
	field, ok := section.StructType.FieldByName(prop.Name)
	if !ok {
		return ""
	}
	return field.Type.String()
}

func apiPropAllowed(prop APIPropDoc) string {
	return strings.Join(prop.Allowed, ", ")
}
