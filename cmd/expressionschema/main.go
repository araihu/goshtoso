// Command expressionschema generates the expression file schema from the Go API.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/expressions"
)

func main() {
	source := flag.String("source", "expressions/set.go", "Go expression definitions")
	output := flag.String("output", "expressions/schema.json", "schema output")
	flag.Parse()
	if err := run(*source, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(source, output string) error {
	file, err := parser.ParseFile(token.NewFileSet(), source, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	groups := map[string]*ast.StructType{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typ, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := typ.Type.(*ast.StructType)
			if ok {
				groups[typ.Name.Name] = st
			}
		}
	}
	props := map[string]any{"$schema": map[string]any{"type": "string", "description": "Optional editor schema reference. Goshtoso does not fetch this URL."}}
	defaults := reflect.ValueOf(expressions.English())
	for _, group := range groups["Set"].Fields.List {
		name := group.Names[0].Name
		fields, err := properties(groups[name], defaults.FieldByName(name))
		if err != nil {
			return err
		}
		props[name] = map[string]any{"type": "object", "description": name + " expressions. Omitted or empty fields inherit defaults.", "additionalProperties": false, "properties": fields}
	}
	schema := map[string]any{"$schema": "https://json-schema.org/draft/2020-12/schema", "$id": "https://goshtoso.araihu.com/schemas/expressions.schema.json", "title": "Goshtoso expressions", "description": "Partial component text overrides. Property names match the Go API and are case-sensitive.", "type": "object", "additionalProperties": false, "properties": props}
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(output, append(data, '\n'), 0644)
}

func properties(group *ast.StructType, defaults reflect.Value) (map[string]any, error) {
	props := map[string]any{}
	for _, field := range group.Fields.List {
		name := field.Names[0].Name
		description := strings.TrimSpace(field.Doc.Text())
		property := map[string]any{"type": "string", "description": description}
		value := defaults.FieldByName(name)
		if fn, ok := field.Type.(*ast.FuncType); ok {
			arg := fn.Params.List[0].Names[0].Name
			property["x-parameter"] = arg
			property["pattern"] = `^(?:[^{}]|\{\{|\}\}|\{` + arg + `\})*$`
			property["description"] = description + " In files, use {" + arg + "} for the argument and {{ or }} for literal braces. An empty string inherits the default."
			sample := reflect.ValueOf("{" + arg + "}")
			if value.Type().In(0).Kind() == reflect.Int {
				sample = reflect.ValueOf(123456789)
			}
			example := value.Call([]reflect.Value{sample})[0].String()
			if value.Type().In(0).Kind() == reflect.Int {
				example = strings.ReplaceAll(example, strconv.Itoa(123456789), "{"+arg+"}")
			}
			property["examples"] = []string{example}
		} else {
			if value.Kind() != reflect.String {
				return nil, fmt.Errorf("unsupported field %s", name)
			}
			property["default"] = value.String()
		}
		props[name] = property
	}
	return props, nil
}
