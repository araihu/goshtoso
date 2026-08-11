package schemaform

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// render is a small helper that renders a templ component to a string.
func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultSchemafield(t *testing.T) {
	var buf bytes.Buffer
	if err := Fields(FieldsConfig{}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render default schemaform: %v", err)
	}
	// Empty config renders the wrapper but no fields.
	if !strings.Contains(buf.String(), "flex flex-col gap-5") {
		t.Fatalf("default render missing wrapper: %s", buf.String())
	}
}

func TestFieldsUseReadableSemanticRolesForRequiredAndErrors(t *testing.T) {
	html := render(t, Fields(FieldsConfig{Fields: []Field{{
		Path:     "name",
		Label:    "Name",
		Kind:     KindString,
		Required: true,
		Errors:   []string{"Name is required"},
	}}}))

	want := `text-danger-text dark:text-danger-text-dark`
	if got := strings.Count(html, want); got != 2 {
		t.Fatalf("required marker and error must use semantic danger text roles: got %d in %s", got, html)
	}
}

// schema used across the Walk-based render tests. Exercises every kind plus
// object unwrap, multi-child section, disabled skip, managed, and required.
func sampleSchema() map[string]any {
	return map[string]any{
		"required": []any{"name"},
		"properties": map[string]any{
			"name":   map[string]any{"type": "string", "title": "Full Name", "description": "your name"},
			"age":    map[string]any{"type": "integer"},
			"score":  map[string]any{"type": "number"},
			"active": map[string]any{"type": "boolean"},
			"color":  map[string]any{"enum": []any{"red", "green"}},
			"tags":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			// array of objects collapses to KindUnknown (textarea).
			"matrix": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
			// object with two children stays a section/fieldset.
			"config": map[string]any{
				"type":     "object",
				"required": []any{"host"},
				"properties": map[string]any{
					"host": map[string]any{"type": "string"},
					"port": map[string]any{"type": "integer"},
				},
			},
			// object with a single child gets unwrapped with "Parent › Child".
			"solo": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"enabled": map[string]any{"type": "boolean"},
				},
			},
			// object with no children is dropped entirely.
			"empty": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			"secret":       map[string]any{"type": "string"},
			"managedField": map[string]any{"type": "string"},
		},
	}
}

func TestWalkRendersAllKinds(t *testing.T) {
	defaults := map[string]any{
		"name":   "Alice",
		"age":    float64(30),
		"tags":   []any{"a", "b"},
		"active": true,
		"config": map[string]any{"host": "localhost", "port": float64(8080)},
	}
	values := map[string]any{"name": "Bob"}
	allowList := map[string]AllowMode{
		"secret":       AllowModeDisabled,
		"managedField": AllowModeManaged,
	}

	fields := Walk(sampleSchema(), defaults, values, allowList)
	if len(fields) == 0 {
		t.Fatalf("Walk returned no fields")
	}

	html := render(t, Fields(FieldsConfig{Fields: fields, NamePrefix: "values"}))

	checks := map[string]string{
		"title override":   "Full Name",
		"helper text":      "your name",
		"required marker":  `aria-label="required"`,
		"current value":    `value="Bob"`, // values wins over default
		"default fallback": `placeholder="Alice"`,
		"number input":     `type="number"`,
		"integer step":     `step="1"`,
		"number step":      `step="any"`,
		"checkbox":         `type="checkbox"`,
		"enum select":      "<select",
		"enum option":      `value="red"`,
		"array tagslist":   "+ Add item",
		"unknown textarea": "<textarea",
		"object section":   "<fieldset",
		"object legend":    "<legend",
		"managed badge":    "managed",
		"managed disabled": "disabled",
		"unwrap separator": "›",
		"input name":       `name="values.name"`,
	}
	for label, want := range checks {
		if !strings.Contains(html, want) {
			t.Errorf("%s: missing %q in rendered html", label, want)
		}
	}

	// Disabled path must not render.
	if strings.Contains(html, "values.secret") {
		t.Errorf("disabled path should be pruned, found values.secret")
	}
	// Empty object must not render.
	if strings.Contains(html, "Empty") {
		t.Errorf("empty object should be dropped")
	}
}

func TestWalkNilAndEmpty(t *testing.T) {
	if got := Walk(nil, nil, nil, nil); got != nil {
		t.Errorf("Walk(nil) = %v, want nil", got)
	}
	if got := Walk(map[string]any{}, nil, nil, nil); got != nil {
		t.Errorf("Walk(no properties) = %v, want nil", got)
	}
	// properties present but not a map
	if got := Walk(map[string]any{"properties": "x"}, nil, nil, nil); got != nil {
		t.Errorf("Walk(bad properties) = %v, want nil", got)
	}
}

func TestWalkSkipsNilNode(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"bad":  "not-a-map",
			"good": map[string]any{"type": "string"},
		},
	}
	fields := Walk(schema, nil, nil, nil)
	if len(fields) != 1 || fields[0].Path != "good" {
		t.Fatalf("expected only 'good' field, got %+v", fields)
	}
}

func TestFlattenAllowList(t *testing.T) {
	if got := FlattenAllowList(nil); len(got) != 0 {
		t.Errorf("FlattenAllowList(nil) = %v, want empty", got)
	}
	m := map[string]any{
		"a": true,
		"b": false, // ignored
		"c": "managed",
		"d": "DISABLED", // case-insensitive
		"e": "other",    // ignored
		"f": float64(1), // ignored
		"nested": map[string]any{
			"x": true,
			"y": "disabled",
		},
	}
	got := FlattenAllowList(&m)
	want := map[string]AllowMode{
		"a":        AllowModeManaged,
		"c":        AllowModeManaged,
		"d":        AllowModeDisabled,
		"nested.x": AllowModeManaged,
		"nested.y": AllowModeDisabled,
	}
	if len(got) != len(want) {
		t.Fatalf("FlattenAllowList size = %d (%v), want %d", len(got), got, len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("FlattenAllowList[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestFallbackFromDefaults(t *testing.T) {
	if got := FallbackFromDefaults(nil, nil, nil); got != nil {
		t.Errorf("FallbackFromDefaults(empty) = %v, want nil", got)
	}

	defaults := map[string]any{
		"name":    "Alice",
		"count":   int(3),
		"ratio":   float64(1.5),
		"enabled": true,
		"tags":    []any{"x", "y"},
		"matrix":  []any{map[string]any{"k": "v"}}, // complex -> unknown
		"nested": map[string]any{
			"host": "localhost",
			"port": int64(8080),
		},
		"solo": map[string]any{
			"only": "child",
		},
		"hidden": "secret",
	}
	values := map[string]any{"name": "Bob"}
	allowList := map[string]AllowMode{"hidden": AllowModeDisabled}

	fields := FallbackFromDefaults(defaults, values, allowList)
	if len(fields) == 0 {
		t.Fatalf("FallbackFromDefaults returned nothing")
	}

	html := render(t, Fields(FieldsConfig{Fields: fields}))
	wants := []string{
		`name="values.name"`, // default NamePrefix applied
		`value="Bob"`,
		"<fieldset", // nested object section
		"<textarea", // complex array -> unknown
		"+ Add item",
		"Solo › Only", // single-child unwrap
	}
	for _, w := range wants {
		if !strings.Contains(html, w) {
			t.Errorf("fallback render missing %q", w)
		}
	}
	if strings.Contains(html, "values.hidden") {
		t.Errorf("disabled fallback path should be pruned")
	}
}

func TestPruneDisabled(t *testing.T) {
	if got := PruneDisabled(nil, nil); len(got) != 0 {
		t.Errorf("PruneDisabled(nil) = %v, want empty map", got)
	}
	values := map[string]any{
		"keep":    "yes",
		"drop":    "no",
		"section": map[string]any{"inner": "v", "secret": "x"},
		"gone":    map[string]any{"onlybad": "x"},
	}
	allowList := map[string]AllowMode{
		"drop":           AllowModeDisabled,
		"section.secret": AllowModeDisabled,
		"gone.onlybad":   AllowModeDisabled,
	}
	got := PruneDisabled(values, allowList)
	if _, ok := got["drop"]; ok {
		t.Errorf("disabled scalar should be pruned")
	}
	if _, ok := got["gone"]; ok {
		t.Errorf("parent that becomes empty should be pruned")
	}
	sec, ok := got["section"].(map[string]any)
	if !ok {
		t.Fatalf("section should survive, got %T", got["section"])
	}
	if _, ok := sec["secret"]; ok {
		t.Errorf("nested disabled key should be pruned")
	}
	if sec["inner"] != "v" {
		t.Errorf("nested kept key missing")
	}
}

func TestHasOnlySimpleScalars(t *testing.T) {
	scalars := []Field{{Kind: KindString}, {Kind: KindBoolean}, {Kind: KindArray}}
	if !HasOnlySimpleScalars(scalars) {
		t.Errorf("simple scalars + flat array should be true")
	}
	if HasOnlySimpleScalars([]Field{{Kind: KindObject}}) {
		t.Errorf("object should be false")
	}
	if HasOnlySimpleScalars([]Field{{Kind: KindUnknown}}) {
		t.Errorf("unknown should be false")
	}
	if HasOnlySimpleScalars([]Field{{Kind: KindArray, Children: []Field{{Kind: KindString}}}}) {
		t.Errorf("array with children should be false")
	}
}

func TestKindFromJSONType(t *testing.T) {
	cases := map[string]Kind{
		"string":  KindString,
		"number":  KindNumber,
		"integer": KindInteger,
		"boolean": KindBoolean,
		"object":  KindObject,
		"array":   KindArray,
		"weird":   KindUnknown,
		"":        KindUnknown,
	}
	for in, want := range cases {
		if got := kindFromJSONType(in); got != want {
			t.Errorf("kindFromJSONType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestKindFromGo(t *testing.T) {
	cases := []struct {
		v    any
		want Kind
	}{
		{"s", KindString},
		{true, KindBoolean},
		{int(1), KindInteger},
		{int32(1), KindInteger},
		{int64(1), KindInteger},
		{uint(1), KindInteger},
		{uint32(1), KindInteger},
		{uint64(1), KindInteger},
		{float32(1), KindNumber},
		{float64(1), KindNumber},
		{map[string]any{}, KindObject},
		{[]any{}, KindArray},
		{nil, KindUnknown},
		{struct{}{}, KindUnknown},
	}
	for _, c := range cases {
		if got := kindFromGo(c.v); got != c.want {
			t.Errorf("kindFromGo(%T) = %q, want %q", c.v, got, c.want)
		}
	}
}

func TestHasComplexElements(t *testing.T) {
	if hasComplexElements([]any{"a", float64(1)}) {
		t.Errorf("flat array should be false")
	}
	if !hasComplexElements([]any{"a", map[string]any{}}) {
		t.Errorf("array with map should be true")
	}
	if !hasComplexElements([]any{[]any{}}) {
		t.Errorf("array with nested array should be true")
	}
}

func TestArrayElements(t *testing.T) {
	if got := arrayElements("not-array"); got != nil {
		t.Errorf("arrayElements(non-array) = %v, want nil", got)
	}
	got := arrayElements([]any{"a", float64(2), true})
	want := []string{"a", "2", "true"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("arrayElements = %v, want %v", got, want)
	}
}

func TestGetStringAndStringSet(t *testing.T) {
	m := map[string]any{"s": "v", "n": 1}
	if v, ok := getString(m, "s"); !ok || v != "v" {
		t.Errorf("getString existing = %q,%v", v, ok)
	}
	if _, ok := getString(m, "missing"); ok {
		t.Errorf("getString missing should be false")
	}
	if _, ok := getString(m, "n"); ok {
		t.Errorf("getString non-string should be false")
	}

	set := stringSet([]any{"a", "b", 3})
	if !set["a"] || !set["b"] || len(set) != 2 {
		t.Errorf("stringSet = %v", set)
	}
	if got := stringSet("nope"); len(got) != 0 {
		t.Errorf("stringSet(non-array) = %v, want empty", got)
	}
}

func TestToStringSlice(t *testing.T) {
	got := toStringSlice([]any{"a", 1, true})
	if strings.Join(got, ",") != "a,1,true" {
		t.Errorf("toStringSlice = %v", got)
	}
}

func TestJSONScalar(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{nil, ""},
		{"hello", "hello"},
		{true, "true"},
		{false, "false"},
		{float64(3.5), "3.5"},
		{[]any{"a", float64(2)}, "a, 2"},
		{int(7), "7"}, // default branch via fmt
	}
	for _, c := range cases {
		if got := jsonScalar(c.v); got != c.want {
			t.Errorf("jsonScalar(%v) = %q, want %q", c.v, got, c.want)
		}
	}
}

func TestLookupDotted(t *testing.T) {
	m := map[string]any{
		"a": map[string]any{"b": map[string]any{"c": "deep"}},
		"x": "shallow",
	}
	if got := lookupDotted(m, "a.b.c"); got != "deep" {
		t.Errorf("lookupDotted deep = %v", got)
	}
	if got := lookupDotted(m, "x"); got != "shallow" {
		t.Errorf("lookupDotted shallow = %v", got)
	}
	if got := lookupDotted(m, "a.missing.c"); got != nil {
		t.Errorf("lookupDotted missing = %v, want nil", got)
	}
	if got := lookupDotted(nil, "a"); got != nil {
		t.Errorf("lookupDotted nil map = %v", got)
	}
	if got := lookupDotted(m, ""); got != nil {
		t.Errorf("lookupDotted empty path = %v", got)
	}
	// descending into a non-map mid-path returns nil
	if got := lookupDotted(m, "x.deeper"); got != nil {
		t.Errorf("lookupDotted into scalar = %v, want nil", got)
	}
}

func TestSortedKeys(t *testing.T) {
	got := sortedKeys(map[string]any{"b": 1, "a": 1, "c": 1})
	if strings.Join(got, "") != "abc" {
		t.Errorf("sortedKeys = %v", got)
	}
}

func TestHumanize(t *testing.T) {
	cases := map[string]string{
		"name":       "Name",
		"my_field":   "My field",
		"my-field":   "My field",
		"camelCase":  "Camel Case",
		"a.b.cValue": "C Value",
		"":           "",
		"HTTPServer": "HTTPServer", // consecutive uppers not split
		"hostName":   "Host Name",
	}
	for in, want := range cases {
		if got := humanize(in); got != want {
			t.Errorf("humanize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsUpperAndPrevRune(t *testing.T) {
	if !isUpper('A') || isUpper('a') || isUpper('1') {
		t.Errorf("isUpper wrong")
	}
	if got := prevRune("abc", 0); got != 0 {
		t.Errorf("prevRune at 0 = %q, want 0", got)
	}
	if got := prevRune("abc", 2); got != 'b' {
		t.Errorf("prevRune = %q, want 'b'", got)
	}
}

func TestPrefillHelpers(t *testing.T) {
	if got := prefillValue(Field{Value: "v", Default: "d"}); got != "v" {
		t.Errorf("prefillValue value wins = %q", got)
	}
	if got := prefillValue(Field{Default: "d"}); got != "d" {
		t.Errorf("prefillValue default fallback = %q", got)
	}
	if !prefillChecked(Field{Value: "true"}) {
		t.Errorf("prefillChecked value true")
	}
	if prefillChecked(Field{Value: "false", Default: "true"}) {
		t.Errorf("prefillChecked value false wins over default")
	}
	if !prefillChecked(Field{Default: "true"}) {
		t.Errorf("prefillChecked default true")
	}
}

func TestInputClassesAndOrDefault(t *testing.T) {
	managed := inputClasses(true)
	if !strings.Contains(managed, "border-control-outline") || !strings.Contains(managed, "dark:border-control-outline-dark") {
		t.Errorf("managed classes missing control boundary roles: %q", managed)
	}
	if !strings.Contains(managed, "cursor-not-allowed") {
		t.Errorf("managed classes missing cursor-not-allowed: %q", managed)
	}
	if strings.Contains(managed, "opacity-70") {
		t.Errorf("managed opacity must not fade the control boundary: %q", managed)
	}
	if !strings.Contains(managed, "bg-surface-alt") || !strings.Contains(managed, "text-on-surface-muted") {
		t.Errorf("managed control must retain governed disabled surface/text semantics: %q", managed)
	}
	normal := inputClasses(false)
	if !strings.Contains(normal, "border-control-outline") || !strings.Contains(normal, "dark:border-control-outline-dark") {
		t.Errorf("normal classes missing control boundary roles: %q", normal)
	}
	if strings.Contains(normal, "cursor-not-allowed") {
		t.Errorf("normal classes should not be disabled: %q", normal)
	}
	if got := orDefault("", "fallback"); got != "fallback" {
		t.Errorf("orDefault empty = %q", got)
	}
	if got := orDefault("set", "fallback"); got != "set" {
		t.Errorf("orDefault set = %q", got)
	}
}

func TestBooleanInputRendersActualControlBoundary(t *testing.T) {
	html := render(t, Fields(FieldsConfig{Fields: []Field{
		{Path: "enabled", Label: "Enabled", Kind: KindBoolean, Default: "true"},
		{Path: "locked", Label: "Locked", Kind: KindBoolean, Managed: true},
	}}))
	if !strings.Contains(html, `class="size-4 appearance-none rounded border border-control-outline`) {
		t.Fatalf("boolean input needs border width plus semantic control role: %s", html)
	}
	if !strings.Contains(html, "appearance-none") {
		t.Fatalf("boolean input must opt out of native appearance so its semantic border actually renders: %s", html)
	}
	if !strings.Contains(html, "checked:bg-primary") || !strings.Contains(html, "dark:checked:bg-primary-dark") {
		t.Fatalf("appearance-none boolean must retain a governed checked affordance: %s", html)
	}
	for _, hook := range []string{
		`label for="values.enabled"`,
		`type="checkbox" id="values.enabled" name="values.enabled" value="true" checked`,
		`type="checkbox" id="values.locked" name="values.locked" value="true" disabled`,
		"focus-visible:outline-2",
	} {
		if !strings.Contains(html, hook) {
			t.Errorf("boolean native semantics/state hook missing %q: %s", hook, html)
		}
	}
}

func TestGetNamePrefix(t *testing.T) {
	if got := (FieldsConfig{}).getNamePrefix(); got != "values" {
		t.Errorf("getNamePrefix default = %q", got)
	}
	if got := (FieldsConfig{NamePrefix: "spec"}).getNamePrefix(); got != "spec" {
		t.Errorf("getNamePrefix set = %q", got)
	}
}

// TestFieldErrorsRender ensures per-field validation errors render and that an
// enum without a Required flag emits the empty placeholder option.
func TestFieldErrorsAndOptionalEnum(t *testing.T) {
	fields := []Field{
		{Path: "email", Name: "email", Label: "Email", Kind: KindString, Errors: []string{"is invalid"}},
		{Path: "color", Name: "color", Label: "Color", Kind: KindEnum, Enum: []string{"red"}, Value: "red"},
	}
	html := render(t, Fields(FieldsConfig{Fields: fields}))
	if !strings.Contains(html, "is invalid") {
		t.Errorf("field error not rendered")
	}
	if !strings.Contains(html, `<option value="">—</option>`) {
		t.Errorf("optional enum missing empty option")
	}
	if !strings.Contains(html, `value="red" selected`) {
		t.Errorf("selected enum option not marked")
	}
}
