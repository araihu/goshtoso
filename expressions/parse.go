package expressions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// ParseJSON validates and parses a partial expression set from JSON bytes.
// Property names match Set and its groups exactly. Unknown and duplicate keys,
// nulls, non-string expressions, and trailing documents are rejected.
// Function fields accept named placeholders documented in JSONSchema.
// Empty strings inherit defaults, including for function fields.
func ParseJSON(data []byte) (Set, error) {
	if !utf8.Valid(data) {
		return Set{}, fmt.Errorf("expressions: JSON must be UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	object, err := jsonObject(decoder)
	if err != nil {
		return Set{}, fmt.Errorf("expressions: %w", err)
	}
	if _, err = decoder.Token(); err != io.EOF {
		return Set{}, fmt.Errorf("expressions: expected one JSON object")
	}
	groups := make(map[string]map[string]string)
	for name, raw := range object {
		if name == "$schema" {
			if err := jsonString(raw); err != nil {
				return Set{}, fmt.Errorf("expressions: $schema: %w", err)
			}
			continue
		}
		fields, err := jsonObject(json.NewDecoder(bytes.NewReader(raw)))
		if err != nil {
			return Set{}, fmt.Errorf("expressions: %s: %w", name, err)
		}
		groups[name] = make(map[string]string)
		for key, value := range fields {
			if err := jsonString(value); err != nil {
				return Set{}, fmt.Errorf("expressions: %s.%s: %w", name, key, err)
			}
			var text string
			if err := json.Unmarshal(value, &text); err != nil {
				return Set{}, err
			}
			groups[name][key] = text
		}
	}
	return parseGroups(groups)
}

func jsonString(raw []byte) error {
	if len(raw) == 0 || raw[0] != '"' {
		return fmt.Errorf("expected a string")
	}
	var value string
	return json.Unmarshal(raw, &value)
}

func jsonObject(decoder *json.Decoder) (map[string]json.RawMessage, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if token != json.Delim('{') {
		return nil, fmt.Errorf("expected an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("expected a property name")
		}
		if _, ok := fields[name]; ok {
			return nil, fmt.Errorf("duplicate property %q", name)
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, err
		}
		fields[name] = raw
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return fields, nil
}

// ParseYAML validates and parses a partial expression set from one YAML document.
// It follows ParseJSON's property and placeholder rules. All property names and
// expression values must be strings. YAML aliases and merge keys are not supported.
func ParseYAML(data []byte) (Set, error) {
	if !utf8.Valid(data) {
		return Set{}, fmt.Errorf("expressions: YAML must be UTF-8")
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return Set{}, fmt.Errorf("expressions: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return Set{}, fmt.Errorf("expressions: expected one YAML document")
	}
	if len(document.Content) != 1 {
		return Set{}, fmt.Errorf("expressions: expected an object")
	}
	object, err := yamlObject(document.Content[0])
	if err != nil {
		return Set{}, err
	}
	groups := make(map[string]map[string]string)
	for name, node := range object {
		if name == "$schema" {
			if !yamlString(node) {
				return Set{}, fmt.Errorf("expressions: $schema: expected a string")
			}
			continue
		}
		fields, err := yamlObject(node)
		if err != nil {
			return Set{}, fmt.Errorf("expressions: %s: %w", name, err)
		}
		groups[name] = make(map[string]string)
		for key, value := range fields {
			if !yamlString(value) {
				return Set{}, fmt.Errorf("expressions: %s.%s: expected a string", name, key)
			}
			groups[name][key] = value.Value
		}
	}
	return parseGroups(groups)
}

func yamlString(node *yaml.Node) bool { return node.Kind == yaml.ScalarNode && node.Tag == "!!str" }

func yamlObject(node *yaml.Node) (map[string]*yaml.Node, error) {
	if node.Kind != yaml.MappingNode || node.Tag != "!!map" {
		return nil, fmt.Errorf("expected an object")
	}
	fields := make(map[string]*yaml.Node)
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		if !yamlString(key) {
			return nil, fmt.Errorf("line %d: property name must be a string", key.Line)
		}
		if _, ok := fields[key.Value]; ok {
			return nil, fmt.Errorf("line %d: duplicate property %q", key.Line, key.Value)
		}
		fields[key.Value] = node.Content[i+1]
	}
	return fields, nil
}

// LoadFile reads a JSON or YAML expression file from the operating system.
// The filename must end in .json, .yaml, or .yml (case-insensitive).
func LoadFile(filename string) (Set, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Set{}, fmt.Errorf("expressions: %w", err)
	}
	return parseFile(filename, data)
}

// LoadFS reads an expression file from fsys, including embed.FS and os.DirFS.
// The name follows fs.FS path rules and must end in .json, .yaml, or .yml.
func LoadFS(fsys fs.FS, name string) (Set, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return Set{}, fmt.Errorf("expressions: %w", err)
	}
	return parseFile(name, data)
}

func parseFile(name string, data []byte) (Set, error) {
	var set Set
	var err error
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json":
		set, err = ParseJSON(data)
	case ".yaml", ".yml":
		set, err = ParseYAML(data)
	default:
		err = fmt.Errorf("expected .json, .yaml, or .yml extension")
	}
	if err != nil {
		return Set{}, fmt.Errorf("expressions: %s: %w", name, err)
	}
	return set, nil
}

type fileSchema struct {
	Properties map[string]fileProperty `json:"properties"`
}
type fileProperty struct {
	Properties map[string]fileProperty `json:"properties"`
	Parameter  string                  `json:"x-parameter"`
}

func parseGroups(groups map[string]map[string]string) (Set, error) {
	var schema fileSchema
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return Set{}, fmt.Errorf("expressions: invalid built-in schema: %w", err)
	}
	var set Set
	root := reflect.ValueOf(&set).Elem()
	for group, fields := range groups {
		target := root.FieldByName(group)
		if !target.IsValid() {
			return Set{}, fmt.Errorf("expressions: unknown group %q", group)
		}
		for name, text := range fields {
			field := target.FieldByName(name)
			if !field.IsValid() {
				return Set{}, fmt.Errorf("expressions: unknown property %s.%s", group, name)
			}
			if text == "" {
				continue
			}
			if field.Kind() == reflect.String {
				field.SetString(text)
				continue
			}
			parameter := schema.Properties[group].Properties[name].Parameter
			parts, err := parseMessage(text, parameter)
			if err != nil {
				return Set{}, fmt.Errorf("expressions: %s.%s: %w", group, name, err)
			}
			field.Set(messageFunc(field.Type(), parts))
		}
	}
	return set, nil
}
