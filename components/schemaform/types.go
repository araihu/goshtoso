package schemaform

import (
	"fmt"
	"sort"
	"strings"
)

// Kind is the rendered input kind.
type Kind string

const (
	KindString  Kind = "string"
	KindNumber  Kind = "number"
	KindInteger Kind = "integer"
	KindBoolean Kind = "boolean"
	KindEnum    Kind = "enum"
	KindObject  Kind = "object"  // group of nested fields
	KindArray   Kind = "array"   // simple array of scalars (renders as tagslist)
	KindUnknown Kind = "unknown" // fallback — render read-only JSON
)

// AllowMode is the per-path constraint expressed by an allow_list entry.
// AllowList leaves can be bool (true → managed) or string ("managed"/"disabled").
type AllowMode string

const (
	// AllowModeManaged: platform-controlled — render but disable the input.
	AllowModeManaged AllowMode = "managed"
	// AllowModeDisabled: hide the field entirely — the platform won't accept
	// overrides at this path.
	AllowModeDisabled AllowMode = "disabled"
)

// Field describes one renderable form control derived from a JSON Schema node + defaults.
type Field struct {
	// Path is a dotted JSONPath from the root of the values object (e.g. "auth.password").
	Path string
	// Name is the input name attribute; usually Path.
	Name string
	// Label is the human-readable field label. For unwrapped 1-child objects
	// it includes the parent label (e.g. "Crds › Enabled") so the rendered
	// context is preserved without a wrapping section.
	Label string
	// HelperText is an optional helper text.
	HelperText string
	// Kind determines rendering.
	Kind Kind
	// Required marks the field with an asterisk.
	Required bool
	// Managed means the allow_list marks this path as platform-controlled —
	// rendered read-only with a lock.
	Managed bool
	// Default is the JSON-serialized default value (quoted scalars preserved).
	Default string
	// ArrayDefault is the element list when Kind == KindArray. Populated in
	// parallel with Default so the renderer can feed it straight to the
	// tagslist component without re-splitting the CSV representation.
	ArrayDefault []string
	// Value is the current form value (JSON-serialized).
	Value string
	// Enum lists valid options when Kind == KindEnum.
	Enum []string
	// Children are nested fields when Kind == KindObject.
	Children []Field
	// Errors hold per-field validation errors.
	Errors []string
}

// FlattenAllowList walks a nested allow_list map and produces a flat map of
// dotted JSONPath → AllowMode. Accepted leaf shapes:
//   - bool true  → AllowModeManaged (backward-compatible encoding)
//   - string "managed" → AllowModeManaged
//   - string "disabled" → AllowModeDisabled
//
// Anything else (bool false, other strings, numbers) is ignored — the path
// behaves as unrestricted.
func FlattenAllowList(m *map[string]any) map[string]AllowMode {
	out := map[string]AllowMode{}
	if m == nil {
		return out
	}
	flattenAllowListInto(*m, "", out)
	return out
}

func flattenAllowListInto(m map[string]any, prefix string, out map[string]AllowMode) {
	for k, v := range m {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		switch x := v.(type) {
		case map[string]any:
			flattenAllowListInto(x, path, out)
		case bool:
			if x {
				out[path] = AllowModeManaged
			}
		case string:
			switch strings.ToLower(x) {
			case "managed":
				out[path] = AllowModeManaged
			case "disabled":
				out[path] = AllowModeDisabled
			}
		}
	}
}

// Walk builds a list of Fields from a JSON Schema subset.
//   - schema: the JSON Schema root (typically an object with "properties").
//   - defaults: default values map (from AddonVersion.DefaultValues).
//   - values: current form values (user-submitted so far).
//   - allowList: flattened allow_list (see FlattenAllowList).
//
// Paths marked AllowModeDisabled are skipped entirely.
// Objects with exactly one (post-filter) child are unwrapped: the single
// descendant bubbles up with its label prefixed "Parent › Child" so the UI
// avoids a section wrapper with a single input inside.
func Walk(schema map[string]any, defaults, values map[string]any, allowList map[string]AllowMode) []Field {
	if schema == nil {
		return nil
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok || len(props) == 0 {
		return nil
	}
	return buildSchemaFields(props, stringSet(schema["required"]), "", defaults, values, allowList)
}

func buildSchemaFields(props map[string]any, required map[string]bool, prefix string, defaults, values map[string]any, allowList map[string]AllowMode) []Field {
	keys := sortedKeys(props)
	out := make([]Field, 0, len(keys))
	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		if allowList[path] == AllowModeDisabled {
			continue
		}
		node, _ := props[k].(map[string]any)
		if node == nil {
			continue
		}
		f := buildSchemaField(path, k, node, defaults, values, allowList, required[k])
		if f.Kind == KindObject && len(f.Children) == 0 {
			continue
		}
		out = append(out, normalizeField(f))
	}
	return out
}

func buildSchemaField(path, name string, node map[string]any, defaults, values map[string]any, allowList map[string]AllowMode, required bool) Field {
	mode := allowList[path]
	f := Field{
		Path:     path,
		Name:     path,
		Label:    humanize(name),
		Required: required,
		Managed:  mode == AllowModeManaged,
	}
	if t, ok := getString(node, "title"); ok && t != "" {
		f.Label = t
	}
	if d, ok := getString(node, "description"); ok {
		f.HelperText = d
	}

	if enumRaw, ok := node["enum"].([]any); ok && len(enumRaw) > 0 {
		f.Kind = KindEnum
		f.Enum = toStringSlice(enumRaw)
	} else {
		tp, _ := getString(node, "type")
		f.Kind = kindFromJSONType(tp)
	}

	f.Default = jsonScalar(lookupDotted(defaults, path))
	f.Value = jsonScalar(lookupDotted(values, path))
	if f.Kind == KindArray {
		f.ArrayDefault = arrayElements(lookupDotted(defaults, path))
	}

	if f.Kind == KindObject {
		if childProps, ok := node["properties"].(map[string]any); ok && len(childProps) > 0 {
			childRequired := stringSet(node["required"])
			f.Children = buildSchemaFields(childProps, childRequired, path, defaults, values, allowList)
		}
	}

	if f.Kind == KindArray {
		if items, ok := node["items"].(map[string]any); ok {
			if tp, _ := getString(items, "type"); tp == "" || tp == "object" || tp == "array" {
				f.Kind = KindUnknown
			}
		}
	}

	return f
}

// FallbackFromDefaults builds fields when no schema is published. Recurses
// into nested maps so complex defaults still produce an editable form.
// Scalar kinds are inferred from the Go type of the default value; nested
// maps become KindObject with populated Children. Disabled paths are skipped;
// single-child objects are unwrapped (see normalizeField).
func FallbackFromDefaults(defaults, values map[string]any, allowList map[string]AllowMode) []Field {
	if len(defaults) == 0 {
		return nil
	}
	return buildDefaultsFields(defaults, "", values, allowList)
}

func buildDefaultsFields(m map[string]any, prefix string, values map[string]any, allowList map[string]AllowMode) []Field {
	keys := sortedKeys(m)
	out := make([]Field, 0, len(keys))
	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		mode := allowList[path]
		if mode == AllowModeDisabled {
			continue
		}
		f := buildDefaultField(path, k, m[k], values, allowList, mode)
		if f.Kind == KindObject && len(f.Children) == 0 {
			continue
		}
		out = append(out, normalizeField(f))
	}
	return out
}

func buildDefaultField(path, name string, val any, values map[string]any, allowList map[string]AllowMode, mode AllowMode) Field {
	f := Field{
		Path:    path,
		Name:    path,
		Label:   humanize(name),
		Kind:    kindFromGo(val),
		Managed: mode == AllowModeManaged,
		Default: jsonScalar(val),
		Value:   jsonScalar(lookupDotted(values, path)),
	}
	switch f.Kind {
	case KindObject:
		if m, ok := val.(map[string]any); ok {
			f.Children = buildDefaultsFields(m, path, values, allowList)
		}
	case KindArray:
		if arr, ok := val.([]any); ok {
			if hasComplexElements(arr) {
				f.Kind = KindUnknown
			} else {
				f.ArrayDefault = arrayElements(val)
			}
		}
	}
	return f
}

// arrayElements extracts a []string representation of an array-typed value
// for seeding tagslist inputs. Accepts []any (YAML-decoded path) and returns
// nil for anything else so the renderer can fall back to an empty list.
func arrayElements(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, el := range arr {
		out = append(out, jsonScalar(el))
	}
	return out
}

// normalizeField collapses object→single-child chains. Recursive: each child
// is normalized first, then if this field is an object with exactly one child
// the child bubbles up carrying "Parent › " prepended to its label.
// ≥2 children remain as sections.
func normalizeField(f Field) Field {
	if f.Kind != KindObject {
		return f
	}
	for i := range f.Children {
		f.Children[i] = normalizeField(f.Children[i])
	}
	if len(f.Children) == 1 {
		c := f.Children[0]
		c.Label = f.Label + " › " + c.Label
		return c
	}
	return f
}

// PruneDisabled returns a deep copy of `values` with every path marked
// AllowModeDisabled stripped out. Empty parent maps that remain after pruning
// are removed as well so the serialized YAML doesn't carry orphan keys.
// Used to seed the "raw YAML" tab of the drawer install flow with the same
// subset the schema form actually exposes.
func PruneDisabled(values map[string]any, allowList map[string]AllowMode) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	return pruneDisabledAt(values, "", allowList)
}

func pruneDisabledAt(m map[string]any, prefix string, allowList map[string]AllowMode) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		if allowList[path] == AllowModeDisabled {
			continue
		}
		if child, ok := v.(map[string]any); ok {
			pruned := pruneDisabledAt(child, path, allowList)
			if len(pruned) == 0 {
				continue
			}
			out[k] = pruned
			continue
		}
		out[k] = v
	}
	return out
}

// HasOnlySimpleScalars reports whether every field in the list is a simple scalar —
// no objects, arrays, or unknowns. Retained for backward compatibility; new
// callers should render Fields unconditionally and let the component handle
// groups/sections.
func HasOnlySimpleScalars(fields []Field) bool {
	for _, f := range fields {
		switch f.Kind {
		case KindObject, KindUnknown:
			return false
		case KindArray:
			if len(f.Children) > 0 {
				return false
			}
		}
	}
	return true
}

// --- helpers ---

func kindFromJSONType(tp string) Kind {
	switch tp {
	case "string":
		return KindString
	case "number":
		return KindNumber
	case "integer":
		return KindInteger
	case "boolean":
		return KindBoolean
	case "object":
		return KindObject
	case "array":
		return KindArray
	default:
		return KindUnknown
	}
}

func kindFromGo(v any) Kind {
	switch v.(type) {
	case string:
		return KindString
	case bool:
		return KindBoolean
	case int, int32, int64, uint, uint32, uint64:
		return KindInteger
	case float32, float64:
		return KindNumber
	case map[string]any:
		return KindObject
	case []any:
		return KindArray
	default:
		return KindUnknown
	}
}

func hasComplexElements(arr []any) bool {
	for _, v := range arr {
		switch v.(type) {
		case map[string]any, []any:
			return true
		}
	}
	return false
}

func getString(m map[string]any, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func stringSet(v any) map[string]bool {
	out := map[string]bool{}
	arr, ok := v.([]any)
	if !ok {
		return out
	}
	for _, a := range arr {
		if s, ok := a.(string); ok {
			out[s] = true
		}
	}
	return out
}

func toStringSlice(raw []any) []string {
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, fmt.Sprint(v))
	}
	return out
}

func jsonScalar(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%v", x)
	case []any:
		parts := make([]string, 0, len(x))
		for _, el := range x {
			parts = append(parts, jsonScalar(el))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprintf("%v", x)
	}
}

func lookupDotted(m map[string]any, path string) any {
	if m == nil || path == "" {
		return nil
	}
	segs := strings.Split(path, ".")
	var cur any = m
	for _, s := range segs {
		mp, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mp[s]
	}
	return cur
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func humanize(s string) string {
	if i := strings.LastIndex(s, "."); i >= 0 {
		s = s[i+1:]
	}
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	var b strings.Builder
	for i, r := range s {
		if i > 0 && isUpper(r) && !isUpper(prevRune(s, i)) {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	out := b.String()
	if out == "" {
		return s
	}
	return strings.ToUpper(out[:1]) + out[1:]
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }

func prevRune(s string, i int) rune {
	if i == 0 {
		return 0
	}
	for j := i - 1; j >= 0; j-- {
		r := rune(s[j])
		if r != 0 {
			return r
		}
	}
	return 0
}
