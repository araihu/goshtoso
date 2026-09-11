package expressions_test

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/araihu/goshtoso/expressions"
)

//go:embed testdata/pt.*
var expressionFiles embed.FS

func TestLoadExpressions(t *testing.T) {
	for _, name := range []string{"testdata/pt.json", "testdata/pt.yaml"} {
		t.Run(name, func(t *testing.T) {
			set, err := expressions.LoadFS(expressionFiles, name)
			if err != nil {
				t.Fatal(err)
			}
			if set.Pagination.PageAriaLabel(3) != "Página 3" || set.CodeBlock.CopyAriaLabel("{unsafe}") != "Copiar {unsafe}" {
				t.Fatal("message substitution failed")
			}
			if set.Pagination.PreviousLabel != "" {
				t.Fatal("parser must return partial overrides")
			}
			resolved := expressions.From(expressions.With(context.Background(), set))
			if resolved.Pagination.PreviousLabel != "Previous" || resolved.CodeBlock.CopyLabel != "Copiar" {
				t.Fatal("fallback lost")
			}
			data, err := fs.ReadFile(expressionFiles, name)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), filepath.Base(name))
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			disk, err := expressions.LoadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if disk.Pagination.PageAriaLabel(0) != "Página 0" {
				t.Fatal("disk load failed")
			}
		})
	}
	_, err := expressions.LoadFS(expressionFiles, "missing.yaml")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("lost filesystem error: %v", err)
	}
	_, err = expressions.LoadFS(fstest.MapFS{"file.txt": {Data: []byte(`{}`)}}, "file.txt")
	if err == nil {
		t.Fatal("accepted unknown extension")
	}
}

func TestExpressionDocumentErrors(t *testing.T) {
	jsonCases := []string{`null`, `[]`, ``, `{} {}`, `{"Unknown":{}}`, `{"pagination":{}}`, `{"Pagination":null}`, `{"Pagination":[]}`, `{"Pagination":{"Nope":"x"}}`, `{"Pagination":{"NextLabel":null}}`, `{"Pagination":{"NextLabel":1}}`, `{"Pagination":{"NextLabel":true}}`, `{"Pagination":{},"Pagination":{}}`, `{"Pagination":{"NextLabel":"x","NextLabel":"y"}}`, `{"$schema":null}`, `{"Pagination":{"PageAriaLabel":"Page {wrong}"}}`, `{"Pagination":{"PageAriaLabel":"Page {page"}}`}
	for _, input := range jsonCases {
		if _, err := expressions.ParseJSON([]byte(input)); err == nil {
			t.Errorf("accepted JSON %s", input)
		}
	}
	yamlCases := []string{"", "null", "[]", "{}\n---\n{}", "Pagination: null", "Pagination: [a]", "Pagination:\n  NextLabel: 2", "Pagination:\n  NextLabel: true", "Pagination:\n  NextLabel: null", "Pagination:\n  NextLabel: x\n  NextLabel: y", "Pagination: {}\nPagination: {}", "Unknown: {}", "Pagination:\n  Wrong: x", "Pagination: &p {}\nCodeBlock: *p", "Pagination:\n  <<: {}", "Pagination:\n  PageAriaLabel: 'Page {wrong}'"}
	for _, input := range yamlCases {
		if _, err := expressions.ParseYAML([]byte(input)); err == nil {
			t.Errorf("accepted YAML %s", input)
		}
	}
}

func TestMessageEscapesAndInheritance(t *testing.T) {
	set, err := expressions.ParseJSON([]byte(`{"CodeBlock":{"CopyAriaLabel":"{{{label}}} {label}"},"Pagination":{"PageAriaLabel":""}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := set.CodeBlock.CopyAriaLabel("{label}"); got != "{{label}} {label}" {
		t.Fatal(got)
	}
	if set.Pagination.PageAriaLabel != nil {
		t.Fatal("empty callback should inherit")
	}
}

func TestSchemaCoversEveryExpression(t *testing.T) {
	var schema struct {
		Properties map[string]struct {
			Properties map[string]struct {
				Type      string
				Parameter string `json:"x-parameter"`
				Pattern   string
			}
		}
	}
	if err := json.Unmarshal(expressions.JSONSchema(), &schema); err != nil {
		t.Fatal(err)
	}
	typ := reflect.TypeFor[expressions.Set]()
	if len(schema.Properties) != typ.NumField()+1 {
		t.Fatal("schema groups differ from Set")
	}
	document := map[string]map[string]string{}
	count := 0
	for group := range typ.Fields() {
		fields := schema.Properties[group.Name].Properties
		if len(fields) != group.Type.NumField() {
			t.Fatalf("schema fields differ for %s", group.Name)
		}
		document[group.Name] = map[string]string{}
		for field := range group.Type.Fields() {
			property := fields[field.Name]
			if property.Type != "string" {
				t.Fatalf("missing property %s.%s", group.Name, field.Name)
			}
			text := "translated"
			if field.Type.Kind() == reflect.Func {
				if property.Parameter == "" {
					t.Fatal("missing callback argument")
				}
				text = "value {" + property.Parameter + "}"
				match, err := regexp.MatchString(property.Pattern, text)
				if err != nil || !match {
					t.Fatalf("schema rejects valid message: %v", err)
				}
			}
			document[group.Name][field.Name] = text
			count++
		}
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	set, err := expressions.ParseJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	value := reflect.ValueOf(set)
	for i := 0; i < typ.NumField(); i++ {
		for j := 0; j < typ.Field(i).Type.NumField(); j++ {
			field := value.Field(i).Field(j)
			if field.Kind() == reflect.String {
				if field.String() != "translated" {
					t.Fatal("field lost")
				}
				continue
			}
			arg := reflect.ValueOf("test")
			if field.Type().In(0).Kind() == reflect.Int {
				arg = reflect.ValueOf(4)
			}
			if got := field.Call([]reflect.Value{arg})[0].String(); !strings.HasPrefix(got, "value ") {
				t.Fatal(got)
			}
		}
	}
	t.Logf("validated %d expression properties", count)
	a := expressions.JSONSchema()
	a[0] = 'x'
	if expressions.JSONSchema()[0] != '{' {
		t.Fatal("schema bytes are mutable")
	}
}

func FuzzParseExpressions(f *testing.F) {
	for _, seed := range []string{`{}`, `{"Pagination":{"PageAriaLabel":"第 {page} 页"}}`, "CodeBlock:\n  CopyLabel: Copiar", "x: &a [*a]"} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = expressions.ParseJSON(data)
		_, _ = expressions.ParseYAML(data)
	})
}
